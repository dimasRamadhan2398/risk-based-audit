package resend

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"auth-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func init() {
	logger.Log = zap.NewNop()
}

func TestResendService_VerifyWebhookSignature_Svix(t *testing.T) {
	secret := "whsec_testsecret1234567890abcdef"
	svc := NewResendService("re_master_key", secret)

	payload := []byte(`{"type":"email.delivered","created_at":"2026-10-08T10:00:00Z","data":{"email_id":"msg_123","from":"no-reply@example.com","to":["user@example.com"],"subject":"Test"}}`)
	svixID := "msg_svix_test_001"
	svixTimestamp := fmt.Sprintf("%d", time.Now().Unix())

	rawSecret := "testsecret1234567890abcdef"
	secretBytes, err := base64.StdEncoding.DecodeString(rawSecret)
	if err != nil {
		secretBytes = []byte(secret)
	}

	toSign := fmt.Sprintf("%s.%s.%s", svixID, svixTimestamp, string(payload))
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(toSign))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	headers := http.Header{}
	headers.Set("svix-id", svixID)
	headers.Set("svix-timestamp", svixTimestamp)
	headers.Set("svix-signature", fmt.Sprintf("v1,%s", sig))

	// Valid signature
	assert.True(t, svc.VerifyWebhookSignature(payload, headers))

	// Tampered payload
	tamperedPayload := []byte(`{"type":"email.bounced"}`)
	assert.False(t, svc.VerifyWebhookSignature(tamperedPayload, headers))

	// Invalid signature header
	badHeaders := http.Header{}
	badHeaders.Set("svix-id", svixID)
	badHeaders.Set("svix-timestamp", svixTimestamp)
	badHeaders.Set("svix-signature", "v1,badsignature")
	assert.False(t, svc.VerifyWebhookSignature(payload, badHeaders))
}

func TestResendService_VerifyWebhookSignature_SecretToken(t *testing.T) {
	secret := "my_custom_secret_token"
	svc := NewResendService("re_master_key", secret)

	headers := http.Header{}
	headers.Set("X-Resend-Webhook-Secret", "my_custom_secret_token")

	assert.True(t, svc.VerifyWebhookSignature([]byte("{}"), headers))

	headers.Set("X-Resend-Webhook-Secret", "wrong_token")
	assert.False(t, svc.VerifyWebhookSignature([]byte("{}"), headers))
}

func TestResendService_HandleWebhookEvent_Bounce(t *testing.T) {
	svc := NewResendService("re_master_key", "secret")

	bouncePayload := []byte(`{
		"type": "email.bounced",
		"created_at": "2026-10-08T10:00:00Z",
		"data": {
			"email_id": "msg_bounced_001",
			"from": "no-reply@mail.auditsphere.app",
			"to": ["nonexistent@example.com"],
			"subject": "OTP Verification",
			"bounce": {
				"type": "hard_bounce",
				"sub_type": "recipient_not_found",
				"message": "Recipient address does not exist"
			}
		}
	}`)

	err := svc.HandleWebhookEvent(context.Background(), bouncePayload)
	assert.NoError(t, err)
}
