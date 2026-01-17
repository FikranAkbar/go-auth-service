package infra

import (
	"go-auth-service/internal/config"
	"testing"
)

// ============================================================================
// EMAIL CLIENT TESTS
// Contract: Creates email client for sending emails via SMTP
// Success: Valid config → returns *EmailClient
// Note: Tests validate client creation and configuration
// ============================================================================

func TestNewEmailClient_ValidConfig(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     "587",
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "noreply@example.com",
		FromName:     "Test Service",
	}

	client := NewEmailClient(cfg)

	if client == nil {
		t.Fatal("Expected email client to be non-nil")
	}

	// Verify configuration is set correctly
	if client.smtpHost != "smtp.gmail.com" {
		t.Errorf("Expected SMTP host smtp.gmail.com, got %s", client.smtpHost)
	}

	if client.smtpPort != 587 {
		t.Errorf("Expected SMTP port 587, got %d", client.smtpPort)
	}

	if client.smtpUsername != "test@example.com" {
		t.Errorf("Expected SMTP username test@example.com, got %s", client.smtpUsername)
	}

	if client.smtpPassword != "password" {
		t.Errorf("Expected SMTP password to be set, got %s", client.smtpPassword)
	}

	if client.fromEmail != "noreply@example.com" {
		t.Errorf("Expected from email noreply@example.com, got %s", client.fromEmail)
	}

	if client.fromName != "Test Service" {
		t.Errorf("Expected from name Test Service, got %s", client.fromName)
	}
}

func TestNewEmailClient_DifferentSMTPProvider(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "smtp.sendgrid.net",
		SMTPPort:     "465",
		SMTPUsername: "apikey",
		SMTPPassword: "SG.xxxxx",
		FromEmail:    "noreply@sendgrid.com",
		FromName:     "SendGrid Service",
	}

	client := NewEmailClient(cfg)

	if client == nil {
		t.Fatal("Expected email client to be non-nil")
	}

	if client.smtpHost != "smtp.sendgrid.net" {
		t.Errorf("Expected SMTP host smtp.sendgrid.net, got %s", client.smtpHost)
	}

	if client.smtpPort != 465 {
		t.Errorf("Expected SMTP port 465, got %d", client.smtpPort)
	}
}

func TestNewEmailClient_EmptyPassword(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     "1025", // MailHog default port
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "test@localhost",
		FromName:     "Test",
	}

	client := NewEmailClient(cfg)

	if client == nil {
		t.Fatal("Expected email client to be non-nil")
	}

	if client.smtpPassword != "" {
		t.Errorf("Expected empty password, got %s", client.smtpPassword)
	}
}

func TestNewEmailClient_InvalidPort_ParsesAsZero(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     "invalid", // Invalid port number
		SMTPUsername: "user",
		SMTPPassword: "pass",
		FromEmail:    "test@example.com",
		FromName:     "Test",
	}

	client := NewEmailClient(cfg)

	// strconv.Atoi returns 0 on error
	if client.smtpPort != 0 {
		t.Errorf("Expected port 0 for invalid port string, got %d", client.smtpPort)
	}
}

func TestNewEmailClient_CustomPort(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "smtp.custom.com",
		SMTPPort:     "2525",
		SMTPUsername: "custom@example.com",
		SMTPPassword: "custompass",
		FromEmail:    "noreply@custom.com",
		FromName:     "Custom Service",
	}

	client := NewEmailClient(cfg)

	if client.smtpPort != 2525 {
		t.Errorf("Expected custom port 2525, got %d", client.smtpPort)
	}
}

func TestNewEmailClient_SpecialCharactersInName(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     "587",
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "noreply@example.com",
		FromName:     "Test Service™ & Co.",
	}

	client := NewEmailClient(cfg)

	if client.fromName != "Test Service™ & Co." {
		t.Errorf("Expected from name with special chars, got %s", client.fromName)
	}
}

// ============================================================================
// SEND EMAIL TESTS
// Contract: Sends email via SMTP
// Success: Valid SMTP connection → sends email
// Note: These are integration tests requiring SMTP server
// ============================================================================

func TestEmailClient_SendEmail_Integration(t *testing.T) {
	// This test requires actual SMTP server connection
	t.Skip("Skipping integration test - requires SMTP server")

	cfg := &config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     "1025", // MailHog or similar
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "test@localhost",
		FromName:     "Test Service",
	}

	client := NewEmailClient(cfg)

	err := client.SendEmail(
		"recipient@example.com",
		"Test Subject",
		"<h1>Test Email</h1><p>This is a test email body.</p>",
	)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestEmailClient_SendEmail_HTMLBody(t *testing.T) {
	// This test documents sending HTML email
	t.Skip("Skipping integration test - requires SMTP server")

	cfg := &config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     "1025",
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "noreply@example.com",
		FromName:     "HTML Service",
	}

	client := NewEmailClient(cfg)

	htmlBody := `
		<!DOCTYPE html>
		<html>
		<head><title>Test Email</title></head>
		<body>
			<h1>Welcome!</h1>
			<p>This is a <strong>HTML</strong> email.</p>
		</body>
		</html>
	`

	err := client.SendEmail("user@example.com", "HTML Test", htmlBody)

	if err != nil {
		t.Errorf("Expected no error sending HTML email, got %v", err)
	}
}

func TestEmailClient_SendEmail_InvalidSMTP_ReturnsError(t *testing.T) {
	// This test verifies error handling for invalid SMTP
	t.Skip("Skipping integration test - would attempt real connection")

	cfg := &config.EmailConfig{
		SMTPHost:     "invalid.smtp.server",
		SMTPPort:     "587",
		SMTPUsername: "invalid",
		SMTPPassword: "invalid",
		FromEmail:    "test@example.com",
		FromName:     "Test",
	}

	client := NewEmailClient(cfg)

	err := client.SendEmail("recipient@example.com", "Test", "Test Body")

	if err == nil {
		t.Error("Expected error for invalid SMTP server, got nil")
	}
}

func TestEmailClient_SendEmail_MultipleRecipients(t *testing.T) {
	// This test documents that SendEmail currently supports single recipient
	// To support multiple recipients, the method would need to be modified
	t.Skip("Skipping - current implementation supports single recipient only")

	cfg := &config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     "1025",
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "test@localhost",
		FromName:     "Test",
	}

	client := NewEmailClient(cfg)

	// Current implementation only accepts single 'to' address
	err := client.SendEmail("user1@example.com", "Test", "Body")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Note: Email client tests are typically integration tests
// For production, consider:
// 1. Use MailHog or similar for local SMTP testing
// 2. Mock SMTP client for unit tests
// 3. Use email service providers' test modes (e.g., SendGrid sandbox)
// 4. Separate integration tests from unit tests
