package mfa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth-service/models"
	"auth-service/pkg/validations"
	mfaService "auth-service/services/mfa"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeMfaService struct {
	mfaService.MfaServiceInterface
	gotReq models.VerifyMfaRequest
	gotUA  string
	result *models.EnableMfaResult
}

func (f *fakeMfaService) EnableMFA(_ context.Context, _ uuid.UUID, req models.VerifyMfaRequest, _, ua string) (*models.EnableMfaResult, error) {
	f.gotReq, f.gotUA = req, ua
	return f.result, nil
}

func TestVerifyMFA_ReturnsAlreadyEnabledAndPassesDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, already := range []bool{false, true} {
		svc := &fakeMfaService{result: &models.EnableMfaResult{AlreadyEnabled: already}}
		ctrl := NewMfaController(validations.New(), svc)
		r := gin.New()
		r.POST("/api/v1/mfa/verify", func(c *gin.Context) { c.Set("user_id", uuid.NewString()) }, ctrl.VerifyMFA)

		body := `{"code":"123456","device_fingerprint":"fp1","device_name":"Chrome on macOS","device_type":"desktop"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/mfa/verify", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("already=%v: status %d body %s", already, w.Code, w.Body.String())
		}
		var resp struct {
			Success bool `json:"success"`
			Data    struct {
				Message        string `json:"message"`
				AlreadyEnabled bool   `json:"already_enabled"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Success || resp.Data.AlreadyEnabled != already {
			t.Fatalf("already=%v: response %s", already, w.Body.String())
		}
		if svc.gotReq.DeviceFingerprint != "fp1" || svc.gotReq.DeviceName != "Chrome on macOS" || svc.gotUA != "Mozilla/5.0" {
			t.Fatalf("device not passed through: %+v ua=%q", svc.gotReq, svc.gotUA)
		}
	}
}
