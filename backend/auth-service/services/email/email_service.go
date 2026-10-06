package email

import (
	"context"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"time"

	"auth-service/pkg/config"
	apperrors "auth-service/pkg/errors"
)

type EmailServiceInterface interface {
	SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error)
	SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error)
	SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error)

	SendNewDeviceTrustedEmail(email, username, deviceName, ipAddress string) error
	SendDeviceRemovedEmail(email, deviceName, ipAddress, location string) error
}

type EmailService struct {
	config  *config.SMTPConfig
	appName string
}

func NewEmailService(cfg *config.SMTPConfig, appName string) EmailServiceInterface {
	return &EmailService{
		config:  cfg,
		appName: appName,
	}
}

func (e *EmailService) sendEmail(to []string, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)

	// From may carry a display name ("PT BAI via AuditSphere <no-reply@...>").
	// The envelope sender (MAIL FROM) must be the bare address, while the From
	// header keeps the display name. Providers such as Resend reject messages
	// without a From header.
	from, err := mail.ParseAddress(e.config.From)
	if err != nil {
		return apperrors.Wrap("EMAIL_CONFIG_ERROR", "Invalid sender address", 500, err)
	}

	msg := []byte("From: " + from.String() + "\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	err = smtp.SendMail(addr, auth, from.Address, to, msg)
	if err != nil {
		return apperrors.Wrap("EMAIL_SEND_ERROR", "Error sending email", 500, err)
	}
	return nil
}

// SendDeviceRemovedEmail implements EmailServiceInterface.
func (e *EmailService) SendDeviceRemovedEmail(email string, deviceName string, ipAddress string, location string) error {
	subject := "Device Removed"
	body := "Device " + deviceName + " removed in " + location + " by " + ipAddress
	return e.sendEmail([]string{email}, subject, body)
}

// SendNewDeviceTrustedEmail implements EmailServiceInterface.
func (e *EmailService) SendNewDeviceTrustedEmail(email string, username string, deviceName string, ipAddress string) error {
	subject := "New Device Enrolled"
	body := "Device " + deviceName + " enrolled by " + username + " in " + ipAddress
	return e.sendEmail([]string{email}, subject, body)
}

// SendPromotionEmail implements EmailServiceInterface.
func (e *EmailService) SendPromotionEmail(ctx context.Context, email string, subject string, body string) (bool, error) {
	err := e.sendEmail([]string{email}, subject, body)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SendOTPEmail implements EmailServiceInterface.
func (e *EmailService) SendOTPEmail(ctx context.Context, toEmail string, otp string, ipAddress string, userAgent string) (bool, error) {
	subject := "OTP Verification New Login Device"
	body := fmt.Sprintf(`Hello,

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
%s Security Team
`, ipAddress, userAgent, time.Now().Format("02 Jan 2006 15:04:05"), otp, e.appName)

	err := e.sendEmail([]string{toEmail}, subject, body)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SendSpecificEmail implements EmailServiceInterface.
func (e *EmailService) SendSpecificEmail(ctx context.Context, toEmail string, subject string, body string) (bool, error) {
	err := e.sendEmail([]string{toEmail}, subject, body)
	if err != nil {
		return false, err
	}
	return true, nil
}
