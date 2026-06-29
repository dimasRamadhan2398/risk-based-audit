package email

import (
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type EmailSender struct {
	config *SMTPConfig
}

func NewEmailSender(cfg *SMTPConfig) *EmailSender {
	return &EmailSender{
		config: cfg,
	}
}

func (s *EmailSender) Send(to []string, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to[0], subject, body))

	return smtp.SendMail(addr, auth, s.config.From, to, msg)
}
