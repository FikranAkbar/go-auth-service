package infra

import (
	"fmt"
	"go-auth-service/internal/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailClient struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

func NewEmailClient(cfg *config.EmailConfig) *EmailClient {
	port, _ := strconv.Atoi(cfg.SMTPPort)
	return &EmailClient{
		smtpHost:     cfg.SMTPHost,
		smtpPort:     port,
		smtpUsername: cfg.SMTPUsername,
		smtpPassword: cfg.SMTPPassword,
		fromEmail:    cfg.FromEmail,
		fromName:     cfg.FromName,
	}
}

func (e *EmailClient) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()

	// Set sender
	m.SetHeader("From", fmt.Sprintf("%s <%s>", e.fromName, e.fromEmail))

	// Set recipient
	m.SetHeader("To", to)

	// Set subject
	m.SetHeader("Subject", subject)

	// Set HTML body
	m.SetBody("text/html", body)

	// Configure SMTP dialer
	d := gomail.NewDialer(e.smtpHost, e.smtpPort, e.smtpUsername, e.smtpPassword)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
