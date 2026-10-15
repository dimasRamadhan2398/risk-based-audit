package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"auth-service/pkg/config"
	"auth-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	logger.Log = zap.NewNop()
}

func TestEmailService_SendViaResend_Success(t *testing.T) {
	var requestCount int32
	var receivedPayload resendEmailRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer re_test_key_12345", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resendEmailResponse{
			ID: "msg_test_abc123",
		})
	}))
	defer server.Close()

	svc := &EmailService{
		resendConfig: &config.ResendConfig{
			APIKey: "re_test_key_12345",
			From:   "AuditSphere <no-reply@mail.auditsphere.app>",
		},
		appName: "AuditSphere",
		appEnv:  "production",
		httpClient: &http.Client{
			Transport: &testRoundTripper{serverURL: server.URL},
			Timeout:   5 * time.Second,
		},
	}

	ok, err := svc.SendOTPEmail(context.Background(), "user@example.com", "123456", "192.168.1.1", "Chrome on MacOS")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
	assert.Equal(t, "AuditSphere <no-reply@mail.auditsphere.app>", receivedPayload.From)
	assert.Equal(t, []string{"user@example.com"}, receivedPayload.To)
	assert.Contains(t, receivedPayload.HTML, "123456")
	assert.Contains(t, receivedPayload.Text, "123456")
}

func TestEmailService_SendViaResend_RetryOnTransientError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			// First two attempts fail with 429 Rate Limit
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message": "Rate limit exceeded"}`))
			return
		}
		// 3rd attempt succeeds
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resendEmailResponse{
			ID: "msg_success_after_retry",
		})
	}))
	defer server.Close()

	svc := &EmailService{
		resendConfig: &config.ResendConfig{
			APIKey: "re_test_key_12345",
			From:   "AuditSphere <no-reply@mail.auditsphere.app>",
		},
		appName: "AuditSphere",
		appEnv:  "production",
		httpClient: &http.Client{
			Transport: &testRoundTripper{serverURL: server.URL},
			Timeout:   5 * time.Second,
		},
	}

	ok, err := svc.SendOTPEmail(context.Background(), "user@example.com", "888999", "10.0.0.1", "Firefox")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestEmailService_MockMode_InDevelopment(t *testing.T) {
	svc := &EmailService{
		resendConfig: &config.ResendConfig{
			APIKey: "",
		},
		appName: "AuditSphere",
		appEnv:  "development",
	}

	ok, err := svc.SendOTPEmail(context.Background(), "dev@example.com", "654321", "127.0.0.1", "Postman")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestEmailService_Error_InProductionWithoutCredentials(t *testing.T) {
	svc := &EmailService{
		resendConfig: &config.ResendConfig{
			APIKey: "",
		},
		appName: "AuditSphere",
		appEnv:  "production",
	}

	ok, err := svc.SendOTPEmail(context.Background(), "prod@example.com", "654321", "127.0.0.1", "Web")
	assert.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "RESEND_API_KEY")
}

// testRoundTripper redirects requests destined for api.resend.com to the test server
type testRoundTripper struct {
	serverURL string
}

func (t *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.URL.Scheme = "http"
	newReq.URL.Host = stringsTrimScheme(t.serverURL)
	return http.DefaultTransport.RoundTrip(newReq)
}

func stringsTrimScheme(url string) string {
	if len(url) > 7 && url[:7] == "http://" {
		return url[7:]
	}
	if len(url) > 8 && url[:8] == "https://" {
		return url[8:]
	}
	return url
}
