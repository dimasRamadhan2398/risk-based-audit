package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/kafka"
	"auth-service/pkg/logger"
	"auth-service/pkg/middleware"
	"auth-service/pkg/utils"
	"auth-service/pkg/validations"
	authService "auth-service/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	e2eSecret   = "e2e-secret-at-least-64-bytes-long-e2e-secret-at-least-64-bytes-long!"
	e2ePassword = "Correct-Horse-9!"
)

type e2eUsers struct{ user *models.User }

func (u e2eUsers) FindByID(id uuid.UUID) (*models.User, error) {
	if u.user.ID != id {
		return nil, apperrors.ErrNotFound
	}
	return u.user, nil
}

type e2eAttempts struct{ n map[string]int64 }

func (a *e2eAttempts) TrackLoginAttempt(_ context.Context, id string, _ time.Duration) (int64, error) {
	a.n[id]++
	return a.n[id], nil
}
func (a *e2eAttempts) GetLoginAttempts(_ context.Context, id string) (int64, error) {
	return a.n[id], nil
}
func (a *e2eAttempts) ClearLoginAttempts(_ context.Context, id string) error {
	delete(a.n, id)
	return nil
}

type e2ePublisher struct{ events []*kafka.Event }

func (p *e2ePublisher) Publish(_ context.Context, e *kafka.Event) error {
	p.events = append(p.events, e)
	return nil
}

// masterRequireStepUp is a FIXTURE copy of master-service
// pkg/middleware/step_up.go RequireStepUp (the two services are separate Go
// modules). Keep it in sync; the real cross-module run is documented in the
// step-up report. It returns "" when the token is accepted, else a reason.
func masterRequireStepUp(raw, accessUserID, purpose, secret string, used map[string]bool) string {
	const label = "auditsphere-step-up-v1"
	if raw == "" {
		return "missing_token"
	}
	key := utils.DeriveStepUpKey(secret) // checked against the label below
	if label != utils.StepUpKeyLabel {
		return "label_drift"
	}
	type stepUpClaims struct {
		Purpose string `json:"purpose"`
		jwt.RegisteredClaims
	}
	claims := &stepUpClaims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (interface{}, error) { return key, nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(30*time.Second))
	if err != nil || !tok.Valid {
		return "invalid_token"
	}
	if claims.Purpose != purpose {
		return "purpose_mismatch"
	}
	if claims.ID == "" || claims.Subject == "" || accessUserID == "" || claims.Subject != accessUserID {
		return "subject_mismatch"
	}
	if claims.IssuedAt == nil || claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) > 5*time.Minute+30*time.Second {
		return "lifetime_too_long"
	}
	if used[claims.ID] {
		return "replayed"
	}
	used[claims.ID] = true
	return ""
}

func TestVerifyPassword_EndToEndTokenAcceptedByMasterVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if logger.Log == nil {
		logger.Log = zap.NewNop()
	}
	hash, err := utils.HashPassword(e2ePassword)
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{ID: uuid.New(), Username: "admin1", PasswordHash: hash, IsActive: true}
	pub := &e2ePublisher{}
	svc := authService.NewStepUpService(e2eUsers{user: user}, &e2eAttempts{n: map[string]int64{}}, pub, e2eSecret)

	r := gin.New()
	auth := r.Group("/api/v1/auth")
	protected := auth.Group("")
	protected.Use(middleware.NewAuthMiddleware(e2eSecret).Authenticate())
	protected.POST("/verify-password", NewStepUpController(validations.New(), svc).VerifyPassword)

	access, err := utils.GenerateToken(user.ID.String(), user.Username, []string{"ADMIN"}, e2eSecret, 1)
	if err != nil {
		t.Fatal(err)
	}
	call := func(bearer, body string) (int, map[string]any) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	if code, _ := call("", `{"password":"`+e2ePassword+`","purpose":"org_levels.change"}`); code != http.StatusUnauthorized {
		t.Fatalf("no access token: %d", code)
	}
	if code, _ := call(access, `{"purpose":"org_levels.change"}`); code != http.StatusBadRequest {
		t.Fatalf("missing password: %d", code)
	}
	if code, out := call(access, `{"password":"nope","purpose":"org_levels.change"}`); code != http.StatusForbidden {
		t.Fatalf("wrong password: %d %v", code, out)
	}
	if code, _ := call(access, `{"password":"`+e2ePassword+`","purpose":"other"}`); code != http.StatusBadRequest {
		t.Fatalf("bad purpose: %d", code)
	}

	code, out := call(access, `{"password":"`+e2ePassword+`","purpose":"org_levels.change"}`)
	if code != http.StatusOK {
		t.Fatalf("verify-password: %d %v", code, out)
	}
	data, _ := out["data"].(map[string]any)
	stepUp, _ := data["step_up_token"].(string)
	if stepUp == "" || data["expires_in"] != float64(300) {
		t.Fatalf("data = %v", data)
	}

	used := map[string]bool{}
	// The issued token passes master's checks for the same user...
	if reason := masterRequireStepUp(stepUp, user.ID.String(), utils.StepUpPurposeOrgLevels, e2eSecret, used); reason != "" {
		t.Fatalf("auth-service token rejected by master verification: %s", reason)
	}
	// ...once.
	if reason := masterRequireStepUp(stepUp, user.ID.String(), utils.StepUpPurposeOrgLevels, e2eSecret, used); reason != "replayed" {
		t.Fatalf("second use: %q", reason)
	}
	// The access token is not a step-up token.
	if reason := masterRequireStepUp(access, user.ID.String(), utils.StepUpPurposeOrgLevels, e2eSecret, used); reason != "invalid_token" {
		t.Fatalf("access token as step-up: %q", reason)
	}
	// Another tenant's secret gives another key.
	if reason := masterRequireStepUp(stepUp, user.ID.String(), utils.StepUpPurposeOrgLevels, e2eSecret+"-other-tenant", map[string]bool{}); reason != "invalid_token" {
		t.Fatalf("other tenant secret: %q", reason)
	}
	// The step-up token is not an access token for auth-service either.
	if code, _ := call(stepUp, `{"password":"`+e2ePassword+`","purpose":"org_levels.change"}`); code != http.StatusUnauthorized {
		t.Fatalf("step-up token used as bearer: %d", code)
	}

	// One security.step_up event per verify attempt that reached the service
	// (wrong password, bad purpose, success); none carries a secret.
	if len(pub.events) != 3 {
		t.Fatalf("events = %d, want 3", len(pub.events))
	}
	for _, ev := range pub.events {
		b, _ := json.Marshal(ev)
		if ev.Type != kafka.EventStepUp || bytes.Contains(b, []byte(e2ePassword)) || bytes.Contains(b, []byte(stepUp)) || bytes.Contains(b, []byte("nope")) {
			t.Fatalf("event %s", b)
		}
	}
}
