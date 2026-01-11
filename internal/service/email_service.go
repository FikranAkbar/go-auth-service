package service

import (
	"fmt"
	"go-auth-service/internal/domain/email"
	"go-auth-service/internal/infra"
)

type EmailService struct {
	emailClient *infra.EmailClient
	appURL      string
}

func NewEmailService(emailClient *infra.EmailClient, appURL string) *EmailService {
	return &EmailService{
		emailClient: emailClient,
		appURL:      appURL,
	}
}

func (s *EmailService) SendVerificationEmail(email, verificationToken string) error {
	verificationLink := fmt.Sprintf("%s/api/auth/verify-email?token=%s", s.appURL, verificationToken)

	subject := "Verify your email address"
	body := fmt.Sprintf(`<html><body><h2>Welcome! Please verify your email</h2><p>Thank you for registering. Please click the link below to verify your email address:</p><p><a href="%s">Verify Email</a></p><p>Or copy and paste this link into your browser:</p><p>%s</p><p>This link will expire in 24 hours.</p><p>If you did not register for an account, you can safely ignore this email.</p></body></html>`, verificationLink, verificationLink)

	return s.emailClient.SendEmail(email, subject, body)
}

// Compile-time check to ensure EmailService implements email.ServiceInterface
var _ email.ServiceInterface = (*EmailService)(nil)
