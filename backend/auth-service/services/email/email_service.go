package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"auth-service/pkg/config"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"

	"go.uber.org/zap"
)

const (
	resendEmailsEndpoint = "https://api.resend.com/emails"
	maxResendRetries     = 3
	initialRetryBackoff  = 500 * time.Millisecond
)

type EmailServiceInterface interface {
	SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error)
	SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error)
	SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error)

	SendNewDeviceTrustedEmail(email, username, deviceName, ipAddress string) error
	SendDeviceRemovedEmail(email, deviceName, ipAddress, location string) error
}

type EmailService struct {
	resendConfig *config.ResendConfig
	smtpConfig   *config.SMTPConfig
	appName      string
	appEnv       string
	httpClient   *http.Client
}

func NewEmailService(
	resendCfg *config.ResendConfig,
	smtpCfg *config.SMTPConfig,
	appName string,
	appEnv string,
) EmailServiceInterface {
	return &EmailService{
		resendConfig: resendCfg,
		smtpConfig:   smtpCfg,
		appName:      appName,
		appEnv:       appEnv,
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

// resendEmailRequest is the JSON payload sent to POST https://api.resend.com/emails
type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// resendEmailResponse is the response returned by Resend upon successful dispatch
type resendEmailResponse struct {
	ID string `json:"id"`
}

// sendEmail orchestrates sending an email:
// 1. Attempts direct Resend REST API delivery with retries if RESEND_API_KEY is configured.
// 2. Falls back to SMTP if SMTP is configured.
// 3. In non-production environments with no credentials, safely logs mock delivery.
// 4. In production with no credentials, returns a clear configuration error.
func (e *EmailService) sendEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	if len(to) == 0 {
		return apperrors.Wrap("EMAIL_INVALID_RECIPIENT", "No recipients specified", 400, nil)
	}

	apiKey := ""
	if e.resendConfig != nil {
		apiKey = e.resendConfig.GetAPIKey()
	}

	// 1. Purely wired Resend Email API path
	if apiKey != "" && !strings.HasPrefix(apiKey, "mock") && !strings.HasPrefix(apiKey, "re_mock") {
		from := e.resendConfig.GetFrom(e.appName)
		err := e.sendViaResendWithRetry(ctx, apiKey, from, to, subject, htmlBody, textBody)
		if err == nil {
			return nil
		}

		logger.Warn("Failed to send email via Resend API after retries, checking SMTP fallback...",
			logger.LogField("error", err.Error()),
			logger.LogField("to", to[0]),
		)
	}

	// 2. SMTP fallback path
	if e.isSMTPConfigured() {
		logger.Info("Attempting email delivery via SMTP relay...", logger.LogField("to", to[0]))
		err := e.sendViaSMTP(to, subject, htmlBody, textBody)
		if err == nil {
			logger.Info("Email delivered successfully via SMTP fallback", logger.LogField("to", to[0]))
			return nil
		}
		logger.Error("SMTP delivery also failed", logger.LogField("error", err.Error()))
	}

	// 3. Mock fallback for development / staging
	if e.appEnv != "production" {
		logger.Warn(fmt.Sprintf("\n"+
			"================================================================================\n"+
			"  [EMAIL MOCK MODE - NO LIVE RESEND KEY CONFIGURED]\n"+
			"  To:       %s\n"+
			"  Subject:  %s\n"+
			"  Time:     %s\n"+
			"--------------------------------------------------------------------------------\n"+
			"  Message Body:\n"+
			"  %s\n"+
			"================================================================================",
			strings.Join(to, ", "), subject, time.Now().Format(time.RFC3339),
			strings.ReplaceAll(textBody, "\n", "\n  "),
		))
		return nil
	}

	// 4. Critical error in production
	logger.Error("[EMAIL CRITICAL] Outbound email failed! No working email provider configured in production environment.",
		logger.LogField("to", to[0]),
		logger.LogField("subject", subject),
	)
	return apperrors.Wrap(
		"EMAIL_DELIVERY_FAILED",
		"Unable to deliver email. Please ensure RESEND_API_KEY is configured in production environment.",
		500,
		nil,
	)
}

// sendViaResendWithRetry calls POST https://api.resend.com/emails with exponential backoff
func (e *EmailService) sendViaResendWithRetry(
	ctx context.Context,
	apiKey string,
	from string,
	to []string,
	subject string,
	htmlBody string,
	textBody string,
) error {
	payload := resendEmailRequest{
		From:    from,
		To:      to,
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return apperrors.Wrap("RESEND_SERIALIZE_ERROR", "Failed to marshal Resend payload", 500, err)
	}

	var lastErr error
	backoff := initialRetryBackoff

	for attempt := 1; attempt <= maxResendRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEmailsEndpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return apperrors.Wrap("RESEND_REQUEST_BUILD_ERROR", "Failed to build Resend request", 500, err)
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "AuditSphere-AuthService/1.0")

		resp, err := e.httpClient.Do(req)
		if err != nil {
			lastErr = err
			logger.Warn("Network error calling Resend API, retrying...",
				logger.LogField("attempt", attempt),
				logger.LogField("max_attempts", maxResendRetries),
				logger.LogField("error", err.Error()),
			)
			if attempt < maxResendRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
				}
				continue
			}
			break
		}

		rawBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 200 OK or 201 Created indicates successful email acceptance by Resend
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var successResp resendEmailResponse
			if err := json.Unmarshal(rawBody, &successResp); err == nil {
				logger.Info("Email dispatched successfully via Resend API",
					zap.String("resend_id", successResp.ID),
					zap.String("to", to[0]),
					zap.String("subject", subject),
				)
			} else {
				logger.Info("Email dispatched successfully via Resend API",
					zap.String("to", to[0]),
					zap.String("subject", subject),
				)
			}
			return nil
		}

		// Retryable HTTP status codes: 429 (rate limited), 500, 502, 503, 504
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("resend API HTTP %d: %s", resp.StatusCode, string(rawBody))
			logger.Warn("Transient error calling Resend API, retrying...",
				logger.LogField("attempt", attempt),
				logger.LogField("status", resp.StatusCode),
				logger.LogField("backoff_ms", backoff.Milliseconds()),
			)
			if attempt < maxResendRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
				}
				continue
			}
			break
		}

		// Permanent error (e.g. 401 Unauthorized, 403 Forbidden, 422 Unprocessable / Unverified Domain)
		logger.Error("Resend API rejected email payload permanently",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(rawBody)),
			zap.String("from", from),
		)
		return apperrors.Wrap(
			"RESEND_API_ERROR",
			fmt.Sprintf("Resend API rejected email (HTTP %d): %s", resp.StatusCode, string(rawBody)),
			resp.StatusCode,
			nil,
		)
	}

	return apperrors.Wrap("RESEND_RETRIES_EXHAUSTED", "Failed to deliver email via Resend after retries", 502, lastErr)
}

func (e *EmailService) isSMTPConfigured() bool {
	if e.smtpConfig == nil || e.smtpConfig.Host == "" {
		return false
	}
	// Mailtrap sandbox with blank credentials is dummy default
	if strings.Contains(e.smtpConfig.Host, "mailtrap") && e.smtpConfig.Username == "" && e.smtpConfig.Password == "" {
		return false
	}
	return true
}

func (e *EmailService) sendViaSMTP(to []string, subject, htmlBody, textBody string) error {
	addr := fmt.Sprintf("%s:%d", e.smtpConfig.Host, e.smtpConfig.Port)

	var auth smtp.Auth
	if e.smtpConfig.Username != "" || e.smtpConfig.Password != "" {
		auth = smtp.PlainAuth("", e.smtpConfig.Username, e.smtpConfig.Password, e.smtpConfig.Host)
	}

	fromAddr := e.smtpConfig.From
	if fromAddr == "" {
		fromAddr = fmt.Sprintf("no-reply@%s", e.appName)
	}

	parsedFrom, err := mail.ParseAddress(fromAddr)
	if err != nil {
		parsedFrom = &mail.Address{Address: fromAddr}
	}

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	bodyContent := textBody
	contentType := "text/plain; charset=UTF-8"

	// If HTML is provided, send multipart/alternative
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", parsedFrom.String()))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(textBody + "\r\n\r\n")
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(htmlBody + "\r\n\r\n")
		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", contentType))
		msg.WriteString(bodyContent + "\r\n")
	}

	err = smtp.SendMail(addr, auth, parsedFrom.Address, to, msg.Bytes())
	if err != nil {
		return apperrors.Wrap("SMTP_SEND_ERROR", "Error sending email via SMTP", 500, err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Public Email Methods
// -----------------------------------------------------------------------------

// SendOTPEmail sends an OTP verification code with modern enterprise styling
func (e *EmailService) SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error) {
	subject := "🔐 Verification Code for Login - AuditSphere"
	currentTime := time.Now().Format("02 Jan 2006 15:04:05 MST")

	textBody := fmt.Sprintf(`Hello,

We detected a login attempt to your AuditSphere account from a new device.

Login Details:
- IP Address : %s
- Device     : %s
- Time       : %s

Your One-Time Password (OTP):
%s

This code is valid for 5 minutes.
Never share this OTP with anyone, including AuditSphere staff.

If you did not initiate this login request, please secure your account immediately.

Best regards,
AuditSphere Security Team
`, ipAddress, userAgent, currentTime, otp)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verification Code</title>
</head>
<body style="margin: 0; padding: 0; background-color: #0b0f17; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #334155;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background-color: #0b0f17; padding: 40px 15px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" style="max-width: 560px; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 10px 25px rgba(0,0,0,0.3); border: 1px solid #1e293b;" cellspacing="0" cellpadding="0">
          <!-- Header Banner -->
          <tr>
            <td style="background: linear-gradient(135deg, #0f172a 0%%, #1e3a8a 100%%); padding: 32px 36px; text-align: left;">
              <div style="font-size: 20px; font-weight: 800; color: #ffffff; letter-spacing: 1.5px; text-transform: uppercase;">
                AUDIT<span style="color: #60a5fa;">SPHERE</span>
              </div>
              <div style="font-size: 13px; color: #94a3b8; margin-top: 4px; font-weight: 500;">
                Risk-Based Internal Audit System
              </div>
            </td>
          </tr>
          <!-- Body Content -->
          <tr>
            <td style="padding: 36px;">
              <h2 style="margin: 0 0 12px 0; color: #0f172a; font-size: 22px; font-weight: 700;">
                Login Verification Code
              </h2>
              <p style="margin: 0 0 24px 0; color: #475569; font-size: 15px; line-height: 1.6;">
                A sign-in attempt requires verification. Please use the One-Time Password (OTP) below to complete your authentication.
              </p>

              <!-- OTP Box -->
              <div style="background-color: #f8fafc; border: 2px dashed #93c5fd; border-radius: 10px; padding: 24px 16px; text-align: center; margin: 28px 0;">
                <div style="font-size: 12px; font-weight: 700; text-transform: uppercase; color: #2563eb; letter-spacing: 1.5px; margin-bottom: 8px;">
                  Your Verification Code
                </div>
                <div style="font-family: 'SF Pro Mono', Menlo, Consolas, Monaco, monospace; font-size: 38px; font-weight: 800; letter-spacing: 10px; color: #1e3a8a; padding: 4px 0;">
                  %s
                </div>
                <div style="font-size: 13px; color: #64748b; margin-top: 8px;">
                  ⏱ Valid for <strong>5 minutes</strong> only
                </div>
              </div>

              <!-- Metadata Details Table -->
              <table role="presentation" width="100%%" style="background-color: #f1f5f9; border-radius: 8px; padding: 14px 18px; margin-bottom: 24px; font-size: 13px; color: #334155;" cellspacing="0" cellpadding="6">
                <tr>
                  <td style="color: #64748b; width: 110px;"><strong>IP Address:</strong></td>
                  <td><code>%s</code></td>
                </tr>
                <tr>
                  <td style="color: #64748b;"><strong>Client / Device:</strong></td>
                  <td>%s</td>
                </tr>
                <tr>
                  <td style="color: #64748b;"><strong>Request Time:</strong></td>
                  <td>%s</td>
                </tr>
              </table>

              <!-- Security Advice -->
              <div style="border-left: 3px solid #f59e0b; background-color: #fffbeb; padding: 12px 16px; border-radius: 4px; margin-bottom: 20px;">
                <p style="margin: 0; font-size: 13px; color: #92400e; line-height: 1.5;">
                  <strong>Security Reminder:</strong> Never share this code with anyone. AuditSphere administrators will never ask for your verification code.
                </p>
              </div>

              <p style="margin: 0; color: #64748b; font-size: 13px; line-height: 1.5;">
                If you did not initiate this login request, please disregard this email and notify your IT Security administrator immediately.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="background-color: #f8fafc; padding: 20px 36px; text-align: center; border-top: 1px solid #e2e8f0;">
              <p style="margin: 0; font-size: 12px; color: #94a3b8; line-height: 1.5;">
                This is an automated system notification from AuditSphere. Please do not reply to this email.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, otp, ipAddress, userAgent, currentTime)

	err := e.sendEmail(ctx, []string{toEmail}, subject, htmlBody, textBody)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SendNewDeviceTrustedEmail sends a notification when a new trusted device is registered
func (e *EmailService) SendNewDeviceTrustedEmail(email string, username string, deviceName string, ipAddress string) error {
	subject := "🛡️ New Trusted Device Registered - AuditSphere"
	currentTime := time.Now().Format("02 Jan 2006 15:04:05 MST")

	textBody := fmt.Sprintf(`Hello %s,

A new device was successfully registered and trusted on your AuditSphere account:

- Device Name : %s
- IP Address  : %s
- Timestamp   : %s

If this was you, no further action is needed.
If you do not recognize this activity, please revoke the device immediately from your Profile Settings.

AuditSphere Security Team
`, username, deviceName, ipAddress, currentTime)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>New Trusted Device</title></head>
<body style="margin:0;padding:0;background-color:#0b0f17;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="padding:40px 15px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" style="max-width:560px;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #1e293b;">
        <tr>
          <td style="background:linear-gradient(135deg, #0f172a, #047857);padding:28px 36px;color:#fff;">
            <div style="font-size:20px;font-weight:800;letter-spacing:1.5px;">AUDITSPHERE</div>
            <div style="font-size:13px;color:#a7f3d0;">Device Security Notification</div>
          </td>
        </tr>
        <tr>
          <td style="padding:36px;color:#334155;">
            <h2 style="margin:0 0 12px;font-size:20px;color:#0f172a;">New Trusted Device Enrolled</h2>
            <p style="margin:0 0 20px;font-size:14px;color:#475569;line-height:1.6;">
              Hello <strong>%s</strong>, a new device has been added to your trusted devices list:
            </p>
            <table role="presentation" width="100%%" style="background:#f8fafc;border-radius:8px;padding:14px;font-size:13px;margin-bottom:24px;">
              <tr><td style="color:#64748b;width:120px;"><strong>Device:</strong></td><td>%s</td></tr>
              <tr><td style="color:#64748b;"><strong>IP Address:</strong></td><td><code>%s</code></td></tr>
              <tr><td style="color:#64748b;"><strong>Time:</strong></td><td>%s</td></tr>
            </table>
            <p style="font-size:13px;color:#64748b;margin:0;">
              If you did not approve this device, please log in and revoke it immediately under Settings > Trusted Devices.
            </p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, username, deviceName, ipAddress, currentTime)

	return e.sendEmail(context.Background(), []string{email}, subject, htmlBody, textBody)
}

// SendDeviceRemovedEmail sends a notification when a device is removed from trusted list
func (e *EmailService) SendDeviceRemovedEmail(email string, deviceName string, ipAddress string, location string) error {
	subject := "⚠️ Device Removed From Account - AuditSphere"
	currentTime := time.Now().Format("02 Jan 2006 15:04:05 MST")

	textBody := fmt.Sprintf(`Hello,

A device was removed from your trusted devices list:

- Device   : %s
- Location : %s
- IP       : %s
- Time     : %s

AuditSphere Security Team
`, deviceName, location, ipAddress, currentTime)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Device Removed</title></head>
<body style="margin:0;padding:0;background-color:#0b0f17;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="padding:40px 15px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" style="max-width:560px;background:#fff;border-radius:12px;overflow:hidden;border:1px solid #1e293b;">
        <tr>
          <td style="background:linear-gradient(135deg, #0f172a, #b91c1c);padding:28px 36px;color:#fff;">
            <div style="font-size:20px;font-weight:800;letter-spacing:1.5px;">AUDITSPHERE</div>
            <div style="font-size:13px;color:#fca5a5;">Device Security Alert</div>
          </td>
        </tr>
        <tr>
          <td style="padding:36px;color:#334155;">
            <h2 style="margin:0 0 12px;font-size:20px;color:#0f172a;">Trusted Device Removed</h2>
            <p style="margin:0 0 20px;font-size:14px;color:#475569;line-height:1.6;">
              A previously authorized device has been removed from your AuditSphere account:
            </p>
            <table role="presentation" width="100%%" style="background:#f8fafc;border-radius:8px;padding:14px;font-size:13px;margin-bottom:24px;">
              <tr><td style="color:#64748b;width:120px;"><strong>Device:</strong></td><td>%s</td></tr>
              <tr><td style="color:#64748b;"><strong>Location:</strong></td><td>%s</td></tr>
              <tr><td style="color:#64748b;"><strong>IP Address:</strong></td><td><code>%s</code></td></tr>
              <tr><td style="color:#64748b;"><strong>Time:</strong></td><td>%s</td></tr>
            </table>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, deviceName, location, ipAddress, currentTime)

	return e.sendEmail(context.Background(), []string{email}, subject, htmlBody, textBody)
}

// SendPromotionEmail implements EmailServiceInterface
func (e *EmailService) SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error) {
	err := e.sendEmail(ctx, []string{email}, subject, "", body)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SendSpecificEmail implements EmailServiceInterface
func (e *EmailService) SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error) {
	err := e.sendEmail(ctx, []string{toEmail}, subject, "", body)
	if err != nil {
		return false, err
	}
	return true, nil
}
