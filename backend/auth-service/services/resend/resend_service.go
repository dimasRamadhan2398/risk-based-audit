package resend

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"

	"go.uber.org/zap"
)

const resendBaseURL = "https://api.resend.com"

// ----- Request / Response models -----

// ----- Service interface & implementation -----

// ResendWebhookPayload represents an inbound webhook event from Resend
type ResendWebhookPayload struct {
	Type      string            `json:"type"`
	CreatedAt string            `json:"created_at"`
	Data      ResendWebhookData `json:"data"`
}

type ResendWebhookData struct {
	CreatedAt string            `json:"created_at"`
	EmailID   string            `json:"email_id"`
	From      string            `json:"from"`
	To        []string          `json:"to"`
	Subject   string            `json:"subject"`
	Bounce    *ResendBounceInfo `json:"bounce,omitempty"`
}

type ResendBounceInfo struct {
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
	Message string `json:"message"`
}

// ResendServiceInterface defines the Resend domain and webhook operations
type ResendServiceInterface interface {
	VerifyWebhookSignature(payload []byte, headers http.Header) bool
	HandleWebhookEvent(ctx context.Context, payload []byte) error
}

// ResendService holds the master Resend API key, webhook secret, and HTTP client
type ResendService struct {
	masterAPIKey  string
	webhookSecret string
	httpClient    *http.Client
}

// NewResendService creates a new ResendService.
func NewResendService(masterAPIKey, webhookSecret string) ResendServiceInterface {
	return &ResendService{
		masterAPIKey:  masterAPIKey,
		webhookSecret: webhookSecret,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// VerifyWebhookSignature verifies the authenticity of incoming Resend/Svix webhook events
func (s *ResendService) VerifyWebhookSignature(payload []byte, headers http.Header) bool {
	// If no webhook secret is configured, allow in development but log a warning
	if s.webhookSecret == "" {
		logger.Warn("[RESEND WEBHOOK] Webhook secret is not configured; skipping signature verification")
		return true
	}

	// 1. Direct secret token header check (e.g. X-Resend-Webhook-Secret or Authorization header)
	if token := headers.Get("X-Resend-Webhook-Secret"); token != "" && token == s.webhookSecret {
		return true
	}
	if authHeader := headers.Get("Authorization"); authHeader != "" && (authHeader == "Bearer "+s.webhookSecret || authHeader == s.webhookSecret) {
		return true
	}

	// 2. Standard Svix webhook signature check (used by Resend webhooks)
	svixID := headers.Get("svix-id")
	svixTimestamp := headers.Get("svix-timestamp")
	svixSignature := headers.Get("svix-signature")

	if svixID == "" || svixTimestamp == "" || svixSignature == "" {
		logger.Warn("[RESEND WEBHOOK] Missing Svix signature headers in webhook request",
			logger.LogField("svix_id", svixID),
			logger.LogField("has_timestamp", svixTimestamp != ""),
			logger.LogField("has_signature", svixSignature != ""),
		)
		return false
	}

	// Resend Svix secrets usually prefix with "whsec_"
	rawSecret := strings.TrimPrefix(s.webhookSecret, "whsec_")
	secretBytes, err := base64.StdEncoding.DecodeString(rawSecret)
	if err != nil {
		secretBytes = []byte(s.webhookSecret)
	}

	toSign := fmt.Sprintf("%s.%s.%s", svixID, svixTimestamp, string(payload))
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(toSign))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// svix-signature may contain multiple space-separated signatures (e.g. "v1,signature1 v1,signature2")
	signatures := strings.Split(svixSignature, " ")
	for _, sig := range signatures {
		parts := strings.Split(sig, ",")
		if len(parts) == 2 && parts[0] == "v1" {
			if hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
				return true
			}
		}
	}

	logger.Warn("[RESEND WEBHOOK] Webhook signature mismatch",
		logger.LogField("svix_id", svixID),
	)
	return false
}

// HandleWebhookEvent processes parsed Resend webhook events (delivery, bounces, spam complaints)
func (s *ResendService) HandleWebhookEvent(ctx context.Context, payload []byte) error {
	var event ResendWebhookPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("[RESEND WEBHOOK] Failed to parse webhook payload JSON",
			logger.LogField("error", err.Error()),
			logger.LogField("raw", string(payload)),
		)
		return apperrors.Wrap("RESEND_WEBHOOK_PARSE_ERROR", "Invalid webhook JSON", 400, err)
	}

	switch event.Type {
	case "email.bounced":
		bounceType := "unknown"
		bounceMsg := ""
		if event.Data.Bounce != nil {
			bounceType = fmt.Sprintf("%s/%s", event.Data.Bounce.Type, event.Data.Bounce.SubType)
			bounceMsg = event.Data.Bounce.Message
		}
		logger.Error("🚨 [RESEND WEBHOOK] Email BOUNCED - Recipient unable to receive mail",
			zap.String("email_id", event.Data.EmailID),
			zap.Strings("to", event.Data.To),
			zap.String("from", event.Data.From),
			zap.String("subject", event.Data.Subject),
			zap.String("bounce_type", bounceType),
			zap.String("bounce_message", bounceMsg),
		)

	case "email.complained":
		logger.Warn("⚠️ [RESEND WEBHOOK] Email SPAM COMPLAINT received",
			zap.String("email_id", event.Data.EmailID),
			zap.Strings("to", event.Data.To),
			zap.String("subject", event.Data.Subject),
		)

	case "email.delivered":
		logger.Info("✅ [RESEND WEBHOOK] Email DELIVERED successfully to recipient",
			zap.String("email_id", event.Data.EmailID),
			zap.Strings("to", event.Data.To),
			zap.String("subject", event.Data.Subject),
		)

	case "email.delivery_delayed":
		logger.Warn("⏳ [RESEND WEBHOOK] Email delivery DELAYED by recipient mailserver",
			zap.String("email_id", event.Data.EmailID),
			zap.Strings("to", event.Data.To),
			zap.String("subject", event.Data.Subject),
		)

	default:
		logger.Info("[RESEND WEBHOOK] Event received: "+event.Type,
			zap.String("email_id", event.Data.EmailID),
			zap.Strings("to", event.Data.To),
		)
	}

	return nil
}
