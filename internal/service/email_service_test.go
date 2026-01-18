package service_test

import (
	"errors"
	"go-auth-service/internal/domain/email"
	"go-auth-service/internal/service"
	"testing"
)

// MockEmailClient implements email.ClientInterface
type MockEmailClient struct {
	SendEmailFunc func(to, subject, body string) error
}

func (m *MockEmailClient) SendEmail(to, subject, body string) error {
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, body)
	}
	return nil
}

// Compile-time check
var _ email.ClientInterface = (*MockEmailClient)(nil)

func TestNewEmailService(t *testing.T) {
	mockClient := &MockEmailClient{}
	appURL := "https://example.com"

	emailService := service.NewEmailService(mockClient, appURL)

	if emailService == nil {
		t.Fatal("NewEmailService returned nil")
	}
}

func TestEmailService_SendVerificationEmail_Success(t *testing.T) {
	var capturedEmail, capturedSubject, capturedBody string

	mockClient := &MockEmailClient{
		SendEmailFunc: func(to, subject, body string) error {
			capturedEmail = to
			capturedSubject = subject
			capturedBody = body
			return nil
		},
	}

	appURL := "https://example.com"
	emailService := service.NewEmailService(mockClient, appURL)

	email := "test@example.com"
	token := "test-verification-token-123"

	err := emailService.SendVerificationEmail(email, token)

	if err != nil {
		t.Fatalf("SendVerificationEmail failed: %v", err)
	}

	// Verify the email was sent to the correct address
	if capturedEmail != email {
		t.Errorf("Expected email to be sent to %s, got %s", email, capturedEmail)
	}

	// Verify the subject
	expectedSubject := "Verify your email address"
	if capturedSubject != expectedSubject {
		t.Errorf("Expected subject '%s', got '%s'", expectedSubject, capturedSubject)
	}

	// Verify the body contains the verification link
	expectedLink := "https://example.com/api/auth/verify-email?token=test-verification-token-123"
	if len(capturedBody) == 0 {
		t.Error("Email body should not be empty")
	}

	// Check if the body contains the verification link
	if !contains(capturedBody, expectedLink) {
		t.Errorf("Email body should contain verification link '%s'", expectedLink)
	}

	// Check if the body contains expected text
	if !containsIgnoreCase(capturedBody, "verify your email") {
		t.Errorf("Email body should contain 'verify your email'. Body: %s", capturedBody)
	}

	if !contains(capturedBody, "24 hours") {
		t.Errorf("Email body should mention token expiry (24 hours). Body: %s", capturedBody)
	}
}

func TestEmailService_SendVerificationEmail_ClientError(t *testing.T) {
	expectedError := errors.New("email client failed")

	mockClient := &MockEmailClient{
		SendEmailFunc: func(to, subject, body string) error {
			return expectedError
		},
	}

	appURL := "https://example.com"
	emailService := service.NewEmailService(mockClient, appURL)

	err := emailService.SendVerificationEmail("test@example.com", "token123")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

func TestEmailService_SendVerificationEmail_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		email string
		token string
	}{
		{
			name:  "Empty email",
			email: "",
			token: "token123",
		},
		{
			name:  "Empty token",
			email: "test@example.com",
			token: "",
		},
		{
			name:  "Long token",
			email: "test@example.com",
			token: "very-long-token-" + string(make([]byte, 100)),
		},
		{
			name:  "Special characters in email",
			email: "test+tag@sub-domain.example.com",
			token: "token123",
		},
		{
			name:  "Special characters in token",
			email: "test@example.com",
			token: "token-with-special-chars_!@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockEmailClient{
				SendEmailFunc: func(to, subject, body string) error {
					return nil
				},
			}

			appURL := "https://example.com"
			emailService := service.NewEmailService(mockClient, appURL)

			err := emailService.SendVerificationEmail(tt.email, tt.token)

			if err != nil {
				t.Errorf("SendVerificationEmail failed: %v", err)
			}
		})
	}
}

func TestEmailService_SendVerificationEmail_DifferentAppURLs(t *testing.T) {
	tests := []struct {
		name        string
		appURL      string
		token       string
		expectedURL string
	}{
		{
			name:        "HTTPS URL",
			appURL:      "https://example.com",
			token:       "token123",
			expectedURL: "https://example.com/api/auth/verify-email?token=token123",
		},
		{
			name:        "HTTP URL",
			appURL:      "http://localhost:8080",
			token:       "token456",
			expectedURL: "http://localhost:8080/api/auth/verify-email?token=token456",
		},
		{
			name:        "URL with path",
			appURL:      "https://example.com/app",
			token:       "token789",
			expectedURL: "https://example.com/app/api/auth/verify-email?token=token789",
		},
		{
			name:        "URL with trailing slash",
			appURL:      "https://example.com/",
			token:       "tokenABC",
			expectedURL: "https://example.com//api/auth/verify-email?token=tokenABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedBody string

			mockClient := &MockEmailClient{
				SendEmailFunc: func(to, subject, body string) error {
					capturedBody = body
					return nil
				},
			}

			emailService := service.NewEmailService(mockClient, tt.appURL)

			err := emailService.SendVerificationEmail("test@example.com", tt.token)

			if err != nil {
				t.Fatalf("SendVerificationEmail failed: %v", err)
			}

			if !contains(capturedBody, tt.expectedURL) {
				t.Errorf("Email body should contain URL '%s', body: %s", tt.expectedURL, capturedBody)
			}
		})
	}
}

func TestEmailService_SendVerificationEmail_HTMLFormat(t *testing.T) {
	var capturedBody string

	mockClient := &MockEmailClient{
		SendEmailFunc: func(to, subject, body string) error {
			capturedBody = body
			return nil
		},
	}

	appURL := "https://example.com"
	emailService := service.NewEmailService(mockClient, appURL)

	err := emailService.SendVerificationEmail("test@example.com", "token123")

	if err != nil {
		t.Fatalf("SendVerificationEmail failed: %v", err)
	}

	// Verify HTML structure
	htmlElements := []string{
		"<html>",
		"</html>",
		"<body>",
		"</body>",
		"<h2>",
		"</h2>",
		"<p>",
		"</p>",
		"<a href=",
		"</a>",
	}

	for _, element := range htmlElements {
		if !contains(capturedBody, element) {
			t.Errorf("Email body should contain HTML element '%s'", element)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function for case-insensitive contains
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}
