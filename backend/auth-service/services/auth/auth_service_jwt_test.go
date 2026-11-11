package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"auth-service/models"
	"auth-service/pkg/config"
	"auth-service/pkg/logger"
	"auth-service/pkg/redis"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// jwtTestSecret is synthetic and only used inside these tests.
const jwtTestSecret = "test-jwt-secret-for-claims-only"

const jwtTestEmployeeUUID = "6f1c1b1e-0000-4000-8000-0000000000e1"

// stubUserRepo implements only Update, which completeLogin calls. Any other
// method would panic through the nil embedded interface, which is intended.
type stubUserRepo struct {
	repositories.UserRepositoryInterface
}

func (stubUserRepo) Update(*models.User) error { return nil }

// newClaimsTestService builds an AuthService whose master-service is a stub
// that answers GET /api/v1/employees/by-code/:code with the given status.
// It returns the service and a counter of master-service calls.
func newClaimsTestService(t *testing.T, masterStatus int, masterAuth *string) (*AuthService, *int32) {
	t.Helper()
	// completeLogin logs through the package-level logger, which is only set up
	// by Init in production.
	if logger.Log == nil {
		logger.Log = zap.NewNop()
	}
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if masterAuth != nil {
			*masterAuth = r.Header.Get("Authorization")
		}
		if masterStatus != http.StatusOK {
			w.WriteHeader(masterStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]string{
				"id":            jwtTestEmployeeUUID,
				"employee_code": "EMP001",
				"department_id": "6f1c1b1e-0000-4000-8000-0000000000d1",
			},
		})
	}))
	t.Cleanup(srv.Close)

	// Redis is unreachable on purpose: the token store only logs a warning on
	// failure, and the test does not depend on it.
	rdb := &redis.Client{Client: goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})}
	t.Cleanup(func() { _ = rdb.Close() })

	cfg := &config.Config{
		JWT:           config.JWTConfig{Secret: jwtTestSecret, ExpiryHour: 1},
		MasterService: config.MasterServiceConfig{URL: srv.URL, TimeoutSeconds: 2},
	}
	svc := NewAuthService(stubUserRepo{}, nil, nil, nil, rdb, cfg, nil, nil).(*AuthService)
	return svc, &calls
}

func claimsTestUser(employeeCode *string) *models.User {
	return &models.User{
		ID:         uuid.New(),
		Username:   "auditor1",
		Email:      "auditor1@example.test",
		EmployeeID: employeeCode,
		Roles:      []models.Role{{Name: "AUDITOR"}},
	}
}

func TestCompleteLogin_IssuesEmployeeClaimsWhenLinked(t *testing.T) {
	var masterAuth string
	svc, calls := newClaimsTestService(t, http.StatusOK, &masterAuth)
	code := "EMP001"

	resp, err := svc.completeLogin(context.Background(), claimsTestUser(&code), "", "", "", false)
	if err != nil {
		t.Fatalf("completeLogin: %v", err)
	}

	claims, err := utils.ParseToken(resp.Token, jwtTestSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.EmployeeID != jwtTestEmployeeUUID {
		t.Fatalf("employee_id claim = %q, want %q", claims.EmployeeID, jwtTestEmployeeUUID)
	}
	if claims.EmployeeCode != "EMP001" {
		t.Fatalf("employee_code claim = %q, want EMP001", claims.EmployeeCode)
	}
	// Existing claims are unchanged.
	if claims.UserID == "" || claims.Username != "auditor1" || len(claims.Roles) != 1 || claims.Roles[0] != "AUDITOR" {
		t.Fatalf("existing claims changed: user_id=%q username=%q roles=%v", claims.UserID, claims.Username, claims.Roles)
	}
	if resp.User.EmployeeID == nil || *resp.User.EmployeeID != jwtTestEmployeeUUID {
		t.Fatalf("response employee_id = %v, want %s", resp.User.EmployeeID, jwtTestEmployeeUUID)
	}

	// The lookup used a bootstrap token that carries no employee claims, and
	// the returned token is a different token.
	if atomic.LoadInt32(calls) != 1 {
		t.Fatalf("master-service calls = %d, want 1", *calls)
	}
	bootstrap, err := utils.ParseToken(strings.TrimPrefix(masterAuth, "Bearer "), jwtTestSecret)
	if err != nil {
		t.Fatalf("bootstrap token not valid for master-service: %v", err)
	}
	if bootstrap.EmployeeID != "" || bootstrap.EmployeeCode != "" {
		t.Fatalf("bootstrap token leaked employee claims: %+v", bootstrap)
	}
	if strings.TrimPrefix(masterAuth, "Bearer ") == resp.Token {
		t.Fatal("returned token must not be the bootstrap token")
	}
}

func TestCompleteLogin_OmitsEmployeeClaimsWhenUnlinked(t *testing.T) {
	svc, calls := newClaimsTestService(t, http.StatusOK, nil)

	resp, err := svc.completeLogin(context.Background(), claimsTestUser(nil), "", "", "", false)
	if err != nil {
		t.Fatalf("completeLogin: %v", err)
	}
	claims, err := utils.ParseToken(resp.Token, jwtTestSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.EmployeeID != "" || claims.EmployeeCode != "" {
		t.Fatalf("unlinked user got employee claims: id=%q code=%q", claims.EmployeeID, claims.EmployeeCode)
	}
	if claims.Username != "auditor1" {
		t.Fatalf("username claim = %q", claims.Username)
	}
	if atomic.LoadInt32(calls) != 0 {
		t.Fatalf("unlinked user must not call master-service, got %d calls", *calls)
	}
}

func TestCompleteLogin_KeepsEmployeeCodeWhenMasterUnavailable(t *testing.T) {
	// The code comes from the local users row, so it survives a master-service
	// miss. The UUID is not known and must be left out.
	svc, _ := newClaimsTestService(t, http.StatusInternalServerError, nil)
	code := "EMP001"

	resp, err := svc.completeLogin(context.Background(), claimsTestUser(&code), "", "", "", false)
	if err != nil {
		t.Fatalf("completeLogin must not fail when master-service is down: %v", err)
	}
	claims, err := utils.ParseToken(resp.Token, jwtTestSecret)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.EmployeeID != "" {
		t.Fatalf("employee_id claim = %q, want empty when master is down", claims.EmployeeID)
	}
	if claims.EmployeeCode != "EMP001" {
		t.Fatalf("employee_code claim = %q, want EMP001", claims.EmployeeCode)
	}
}

func TestGenerateToken_WithoutEmployeeOmitsClaims(t *testing.T) {
	tok, err := utils.GenerateToken("u1", "name", []string{"ADMIN"}, jwtTestSecret, 1)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := utils.ParseToken(tok, jwtTestSecret)
	if err != nil {
		t.Fatal(err)
	}
	if claims.EmployeeID != "" || claims.EmployeeCode != "" {
		t.Fatalf("GenerateToken must not add employee claims: %+v", claims)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatal("exp claim missing or in the past")
	}
}
