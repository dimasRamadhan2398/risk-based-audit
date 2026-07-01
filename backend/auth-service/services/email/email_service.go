package email

import (
	"context"
	"fmt"
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

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to[0], subject, body))

	err := smtp.SendMail(addr, auth, e.config.From, to, msg)
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
