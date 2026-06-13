package email

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	apperrors "auth-service/pkg/errors"

	resend "github.com/resend/resend-go/v2"
)

type EmailServiceInterface interface {
	SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error)
	SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error)
	SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error)

	SendNewDeviceTrustedEmail(email, username, deviceName, ipAddress string) error
	SendDeviceRemovedEmail(email, username, deviceName, ipAddress string) error
}

type EmailService struct {
	client    *resend.Client
	fromEmail string
	appName   string
}

// SendDeviceRemovedEmail implements EmailServiceInterface.
func (e *EmailService) SendDeviceRemovedEmail(email string, deviceName string, ipAddress string, location string) error {
	params := &resend.SendEmailRequest{
		From:    e.fromEmail,
		To:      []string{email},
		Subject: "Device Removed",
		Text:    "Device " + deviceName + " removed in " + location + " by " + ipAddress,
	}

	_, err := e.client.Emails.Send(params)
	if err != nil {
		return apperrors.Wrap("EMAIL_SEND_ERROR", "Error sending email", 500, err)
	}

	return nil
}

// SendNewDeviceTrustedEmail implements EmailServiceInterface.
func (e *EmailService) SendNewDeviceTrustedEmail(email string, username string, deviceName string, ipAddress string) error {
	params := &resend.SendEmailRequest{
		From:    e.fromEmail,
		To:      []string{email},
		Subject: "New Device Enrolled",
		Text:    "Device " + deviceName + " enrolled by " + username + " in " + ipAddress,
	}

	_, err := e.client.Emails.Send(params)
	if err != nil {
		return apperrors.Wrap("EMAIL_SEND_ERROR", err.Error(), 500, err)
	}

	return nil
}

// SendPromotionEmail implements EmailServiceInterface.
func (e *EmailService) SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error) {
	panic("unimplemented")
}

// SendOTPEmail implements EmailServiceInterface.
func (e *EmailService) SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error) {
	params := &resend.SendEmailRequest{
		From:    e.fromEmail,
		To:      []string{toEmail},
		Subject: "OTP Verification New Login Device",
		Text:    fmt.Sprintf(`Hello,

We detected a login attempt to your account from a new device.

Here are the details:
- IP Address : %s
- Device     : %s
- Time       : %s

To continue, please use the One-Time Password (OTP) below:

🔐 OTP Code: %s

This code will expire in 5 minutes.

If this was you, you can safely proceed with the login.
If you did NOT initiate this request, please ignore this email or secure your account immediately.

For your security, never share this OTP with anyone.

Best regards,  
Your Security Team
`, ipAddress, userAgent, time.Now().Format("02 Jan 2006 15:04:05"), otp),
	}

	//idempotency key (add this if no rate limiter)
	// options := &resend.SendEmailOptions{

	// }

	_, err := e.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return false, apperrors.Wrap("EMAIL_SEND_ERROR", "Error sending email", 500, err)
	}

	return true, nil
}

// SendPromotionEmail implements EmailServiceInterface.
func (e *EmailService) SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error) {
	params := &resend.SendEmailRequest{
		From:    e.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Text:    body,
	}

	//idempotency key (add this if no rate limiter)
	// options := &resend.SendEmailOptions{

	// }

	_, err := e.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return false, apperrors.Wrap("EMAIL_SEND_ERROR", "Error sending email", 500, err)
	}

	return true, nil
}

func NewEmailService(apiKey string, fromEmail string, appName string) EmailServiceInterface {
	safeFromEmail := buildSafeFromEmail(fromEmail, appName)

	return &EmailService{
		client:    resend.NewClient(apiKey),
		fromEmail: safeFromEmail,
		appName:   appName,
	}
}

func buildSafeFromEmail(fromEmail, appName string) string {
	defaultAppName := strings.TrimSpace(appName)
	if defaultAppName == "" {
		defaultAppName = "Risk Based Internal Audit"
	}

	defaultFrom := fmt.Sprintf("%s <onboarding@resend.dev>", defaultAppName)
	candidate := strings.TrimSpace(fromEmail)
	if candidate == "" {
		return defaultFrom
	}

	parsed, err := mail.ParseAddress(candidate)
	if err != nil {
		return defaultFrom
	}

	parts := strings.Split(strings.ToLower(parsed.Address), "@")
	if len(parts) != 2 || parts[1] == "example.com" {
		return defaultFrom
	}

	return candidate
}
