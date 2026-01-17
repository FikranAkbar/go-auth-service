package security

import (
	"go-auth-service/internal/config"
	domainSecurity "go-auth-service/internal/domain/security"
	"go-auth-service/pkg/constants"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTManager(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	manager := NewJWTManager(cfg)

	if manager == nil {
		t.Fatal("NewJWTManager returned nil")
	}

	if manager.secretKey != cfg.SecretKey {
		t.Errorf("Expected secret key %s, got %s", cfg.SecretKey, manager.secretKey)
	}

	if manager.accessTokenExpiry != cfg.AccessTokenExpiry {
		t.Errorf("Expected access token expiry %v, got %v", cfg.AccessTokenExpiry, manager.accessTokenExpiry)
	}

	if manager.refreshTokenExpiry != cfg.RefreshTokenExpiry {
		t.Errorf("Expected refresh token expiry %v, got %v", cfg.RefreshTokenExpiry, manager.refreshTokenExpiry)
	}

	if manager.issuer != cfg.Issuer {
		t.Errorf("Expected issuer %s, got %s", cfg.Issuer, manager.issuer)
	}
}

func TestGenerateAccessToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name     string
		userID   int64
		email    string
		username string
	}{
		{
			name:     "Valid user data",
			userID:   123,
			email:    "test@example.com",
			username: "testuser",
		},
		{
			name:     "Empty username",
			userID:   456,
			email:    "user@example.com",
			username: "",
		},
		{
			name:     "Long email",
			userID:   789,
			email:    "very.long.email.address@subdomain.example.com",
			username: "longuser",
		},
		{
			name:     "Zero user ID",
			userID:   0,
			email:    "zero@example.com",
			username: "zerouser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateAccessToken(tt.userID, tt.email, tt.username)

			if err != nil {
				t.Fatalf("GenerateAccessToken failed: %v", err)
			}

			if token == "" {
				t.Error("Generated token should not be empty")
			}

			// Validate the token
			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate generated token: %v", err)
			}

			// Verify claims
			if claims.UserID != tt.userID {
				t.Errorf("Expected userID %d, got %d", tt.userID, claims.UserID)
			}
			if claims.Email != tt.email {
				t.Errorf("Expected email %s, got %s", tt.email, claims.Email)
			}
			if claims.Username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, claims.Username)
			}
			if claims.Type != domainSecurity.AccessToken {
				t.Errorf("Expected token type %s, got %s", domainSecurity.AccessToken, claims.Type)
			}
			if claims.Issuer != cfg.Issuer {
				t.Errorf("Expected issuer %s, got %s", cfg.Issuer, claims.Issuer)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	userID := int64(123)
	email := "test@example.com"
	username := "testuser"

	token, err := manager.GenerateRefreshToken(userID, email, username)

	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Validate the token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	// Verify claims
	if claims.UserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}
	if claims.Type != domainSecurity.RefreshToken {
		t.Errorf("Expected token type %s, got %s", domainSecurity.RefreshToken, claims.Type)
	}
}

func TestGenerateVerificationToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	userID := int64(123)
	email := "test@example.com"

	token, err := manager.GenerateVerificationToken(userID, email)

	if err != nil {
		t.Fatalf("GenerateVerificationToken failed: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Validate the token
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate generated token: %v", err)
	}

	// Verify claims
	if claims.UserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
	if claims.Username != "" {
		t.Errorf("Expected empty username, got %s", claims.Username)
	}
	if claims.Type != domainSecurity.VerificationToken {
		t.Errorf("Expected token type %s, got %s", domainSecurity.VerificationToken, claims.Type)
	}

	// Verify expiry is approximately 24 hours
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	diff := actualExpiry.Sub(expectedExpiry).Abs()
	if diff > 5*time.Second {
		t.Errorf("Expected expiry around %v, got %v (diff: %v)", expectedExpiry, actualExpiry, diff)
	}
}

func TestValidateToken_Success(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	userID := int64(123)
	email := "test@example.com"
	username := "testuser"

	token, err := manager.GenerateAccessToken(userID, email, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  -1 * time.Hour, // Already expired
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	userID := int64(123)
	email := "test@example.com"
	username := "testuser"

	token, err := manager.GenerateAccessToken(userID, email, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be expired
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("Expected 'token is expired' error, got: %v", err)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	cfg1 := &config.JWTConfig{
		SecretKey:          "first-secret-key-at-least-32-characters",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager1 := NewJWTManager(cfg1)

	cfg2 := &config.JWTConfig{
		SecretKey:          "second-secret-key-at-least-32-characters",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager2 := NewJWTManager(cfg2)

	// Generate token with first manager
	token, err := manager1.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with second manager (different secret)
	_, err = manager2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "signature is invalid") {
		t.Errorf("Expected 'signature is invalid' error, got: %v", err)
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name        string
		token       string
		description string
	}{
		{
			name:        "Empty token",
			token:       "",
			description: "Should fail with empty token",
		},
		{
			name:        "Invalid format",
			token:       "not-a-jwt-token",
			description: "Should fail with invalid format",
		},
		{
			name:        "Incomplete JWT",
			token:       "header.payload",
			description: "Should fail with incomplete JWT",
		},
		{
			name:        "Random string",
			token:       "random.random.random",
			description: "Should fail with random string",
		},
		{
			name:        "Only dots",
			token:       "...",
			description: "Should fail with only dots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.description)
			}
		})
	}
}

func TestValidateToken_WrongSigningMethod(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// Create a token with a different signing method (e.g., none)
	claims := domainSecurity.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     domainSecurity.AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	// Create token with 'none' algorithm (no signature)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	// This should fail because we expect HMAC
	_, err := manager.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for wrong signing method, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), constants.ErrInvalidToken) {
		t.Errorf("Expected '%s' error, got: %v", constants.ErrInvalidToken, err)
	}
}

func TestValidateToken_TamperedClaims(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// Generate a valid token
	token, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Tamper with the token by modifying a character in the payload
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("Token doesn't have 3 parts")
	}

	// Modify the payload
	tamperedPayload := parts[1]
	if len(tamperedPayload) > 0 {
		// Change the last character
		runes := []rune(tamperedPayload)
		if runes[len(runes)-1] == 'A' {
			runes[len(runes)-1] = 'B'
		} else {
			runes[len(runes)-1] = 'A'
		}
		tamperedPayload = string(runes)
	}

	tamperedToken := parts[0] + "." + tamperedPayload + "." + parts[2]

	// Try to validate the tampered token
	_, err = manager.ValidateToken(tamperedToken)
	if err == nil {
		t.Error("Expected error for tampered token, got nil")
	}
}

func TestExtractUserID_Success(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	expectedUserID := int64(123)
	token, err := manager.GenerateAccessToken(expectedUserID, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userID, err := manager.ExtractUserID(token)
	if err != nil {
		t.Fatalf("ExtractUserID failed: %v", err)
	}

	if userID != expectedUserID {
		t.Errorf("Expected userID %d, got %d", expectedUserID, userID)
	}
}

func TestExtractUserID_InvalidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "Invalid token",
			token: "invalid-token",
		},
		{
			name:  "Malformed token",
			token: "header.payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := manager.ExtractUserID(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token, got nil")
			}
			if userID != 0 {
				t.Errorf("Expected userID 0 for invalid token, got %d", userID)
			}
		})
	}
}

func TestExtractUserID_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  -1 * time.Hour, // Already expired
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	token, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	userID, err := manager.ExtractUserID(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
	if userID != 0 {
		t.Errorf("Expected userID 0 for expired token, got %d", userID)
	}
}

func TestTokenExpiry(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  1 * time.Hour,
		RefreshTokenExpiry: 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// Generate access token
	accessToken, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := manager.GenerateRefreshToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Validate access token expiry
	accessClaims, err := manager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	expectedAccessExpiry := time.Now().Add(1 * time.Hour)
	actualAccessExpiry := accessClaims.ExpiresAt.Time
	diff := actualAccessExpiry.Sub(expectedAccessExpiry).Abs()
	if diff > 5*time.Second {
		t.Errorf("Access token expiry mismatch. Expected ~%v, got %v (diff: %v)",
			expectedAccessExpiry, actualAccessExpiry, diff)
	}

	// Validate refresh token expiry
	refreshClaims, err := manager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	expectedRefreshExpiry := time.Now().Add(24 * time.Hour)
	actualRefreshExpiry := refreshClaims.ExpiresAt.Time
	diff = actualRefreshExpiry.Sub(expectedRefreshExpiry).Abs()
	if diff > 5*time.Second {
		t.Errorf("Refresh token expiry mismatch. Expected ~%v, got %v (diff: %v)",
			expectedRefreshExpiry, actualRefreshExpiry, diff)
	}
}

func TestJWTManager_InterfaceImplementation(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// This test ensures the JWTManager implements the interface
	// by attempting to use it
	token, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	_, err = manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
}

func BenchmarkGenerateAccessToken(b *testing.B) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateAccessToken(123, "test@example.com", "testuser")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)
	token, _ := manager.GenerateAccessToken(123, "test@example.com", "testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ValidateToken(token)
	}
}

func BenchmarkExtractUserID(b *testing.B) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)
	token, _ := manager.GenerateAccessToken(123, "test@example.com", "testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ExtractUserID(token)
	}
}

// Additional edge case tests for comprehensive coverage

func TestGenerateAccessToken_EdgeCases(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name        string
		userID      int64
		email       string
		username    string
		description string
	}{
		{
			name:        "Negative user ID",
			userID:      -1,
			email:       "test@example.com",
			username:    "testuser",
			description: "Should handle negative user ID",
		},
		{
			name:        "Very large user ID",
			userID:      9223372036854775807, // max int64
			email:       "test@example.com",
			username:    "testuser",
			description: "Should handle maximum int64 value",
		},
		{
			name:        "Special characters in email",
			userID:      123,
			email:       "test+tag@sub-domain.example.com",
			username:    "testuser",
			description: "Should handle special characters in email",
		},
		{
			name:        "Special characters in username",
			userID:      123,
			email:       "test@example.com",
			username:    "test_user-123",
			description: "Should handle special characters in username",
		},
		{
			name:        "Unicode in username",
			userID:      123,
			email:       "test@example.com",
			username:    "用户名",
			description: "Should handle unicode characters",
		},
		{
			name:        "Empty email",
			userID:      123,
			email:       "",
			username:    "testuser",
			description: "Should handle empty email",
		},
		{
			name:        "All fields empty except userID",
			userID:      123,
			email:       "",
			username:    "",
			description: "Should handle minimal data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateAccessToken(tt.userID, tt.email, tt.username)

			if err != nil {
				t.Fatalf("GenerateAccessToken failed: %v. %s", err, tt.description)
			}

			if token == "" {
				t.Errorf("Token should not be empty. %s", tt.description)
			}

			// Validate the token
			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v. %s", err, tt.description)
			}

			if claims.UserID != tt.userID {
				t.Errorf("UserID mismatch. Expected %d, got %d", tt.userID, claims.UserID)
			}
			if claims.Email != tt.email {
				t.Errorf("Email mismatch. Expected %s, got %s", tt.email, claims.Email)
			}
			if claims.Username != tt.username {
				t.Errorf("Username mismatch. Expected %s, got %s", tt.username, claims.Username)
			}
		})
	}
}

func TestGenerateRefreshToken_EdgeCases(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name     string
		userID   int64
		email    string
		username string
	}{
		{
			name:     "Minimal data",
			userID:   1,
			email:    "a@b.c",
			username: "u",
		},
		{
			name:     "Very long email",
			userID:   123,
			email:    strings.Repeat("a", 100) + "@example.com",
			username: "testuser",
		},
		{
			name:     "Very long username",
			userID:   123,
			email:    "test@example.com",
			username: strings.Repeat("user", 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateRefreshToken(tt.userID, tt.email, tt.username)

			if err != nil {
				t.Fatalf("GenerateRefreshToken failed: %v", err)
			}

			if token == "" {
				t.Error("Token should not be empty")
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			if claims.Type != domainSecurity.RefreshToken {
				t.Errorf("Expected token type %s, got %s", domainSecurity.RefreshToken, claims.Type)
			}
		})
	}
}

func TestGenerateVerificationToken_EdgeCases(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name   string
		userID int64
		email  string
	}{
		{
			name:   "Minimal user ID",
			userID: 1,
			email:  "test@example.com",
		},
		{
			name:   "Zero user ID",
			userID: 0,
			email:  "test@example.com",
		},
		{
			name:   "Empty email",
			userID: 123,
			email:  "",
		},
		{
			name:   "Unicode email",
			userID: 123,
			email:  "用户@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateVerificationToken(tt.userID, tt.email)

			if err != nil {
				t.Fatalf("GenerateVerificationToken failed: %v", err)
			}

			if token == "" {
				t.Error("Token should not be empty")
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("Expected userID %d, got %d", tt.userID, claims.UserID)
			}
			if claims.Email != tt.email {
				t.Errorf("Expected email %s, got %s", tt.email, claims.Email)
			}
			if claims.Type != domainSecurity.VerificationToken {
				t.Errorf("Expected token type %s, got %s", domainSecurity.VerificationToken, claims.Type)
			}
		})
	}
}

func TestValidateToken_MalformedJWT(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// Generate a valid token first
	validToken, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		description string
	}{
		{
			name:        "Token with invalid base64 in payload",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid!!!.signature",
			description: "Should fail with invalid base64 encoding",
		},
		{
			name:        "Token with only header",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			description: "Should fail with incomplete token",
		},
		{
			name:        "Token with extra parts",
			token:       validToken + ".extra",
			description: "Should fail with extra parts",
		},
		{
			name:        "Token with missing signature",
			token:       strings.Join(strings.Split(validToken, ".")[:2], "."),
			description: "Should fail without signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.description)
			}
		})
	}
}

func TestValidateToken_DifferentTokenTypes(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	// Test that all token types can be validated
	accessToken, _ := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	refreshToken, _ := manager.GenerateRefreshToken(123, "test@example.com", "testuser")
	verificationToken, _ := manager.GenerateVerificationToken(123, "test@example.com")

	tests := []struct {
		name               string
		token              string
		expectedType       domainSecurity.TokenType
		shouldHaveUsername bool
	}{
		{
			name:               "Access token validation",
			token:              accessToken,
			expectedType:       domainSecurity.AccessToken,
			shouldHaveUsername: true,
		},
		{
			name:               "Refresh token validation",
			token:              refreshToken,
			expectedType:       domainSecurity.RefreshToken,
			shouldHaveUsername: true,
		},
		{
			name:               "Verification token validation",
			token:              verificationToken,
			expectedType:       domainSecurity.VerificationToken,
			shouldHaveUsername: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tt.token)
			if err != nil {
				t.Fatalf("Validation failed: %v", err)
			}

			if claims.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, claims.Type)
			}

			if tt.shouldHaveUsername && claims.Username == "" {
				t.Error("Expected username to be present")
			}

			if !tt.shouldHaveUsername && claims.Username != "" {
				t.Errorf("Expected empty username, got %s", claims.Username)
			}
		})
	}
}

func TestNewJWTManager_ConfigValues(t *testing.T) {
	tests := []struct {
		name   string
		config *config.JWTConfig
	}{
		{
			name: "Standard configuration",
			config: &config.JWTConfig{
				SecretKey:          "test-secret-key-at-least-32-characters-long",
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 7 * 24 * time.Hour,
				Issuer:             "test-issuer",
			},
		},
		{
			name: "Short expiry times",
			config: &config.JWTConfig{
				SecretKey:          "test-secret-key-at-least-32-characters-long",
				AccessTokenExpiry:  1 * time.Minute,
				RefreshTokenExpiry: 5 * time.Minute,
				Issuer:             "test-issuer",
			},
		},
		{
			name: "Long expiry times",
			config: &config.JWTConfig{
				SecretKey:          "test-secret-key-at-least-32-characters-long",
				AccessTokenExpiry:  24 * time.Hour,
				RefreshTokenExpiry: 365 * 24 * time.Hour,
				Issuer:             "test-issuer",
			},
		},
		{
			name: "Different issuer",
			config: &config.JWTConfig{
				SecretKey:          "test-secret-key-at-least-32-characters-long",
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 7 * 24 * time.Hour,
				Issuer:             "different-issuer-value",
			},
		},
		{
			name: "Very long secret key",
			config: &config.JWTConfig{
				SecretKey:          strings.Repeat("secret", 20),
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 7 * 24 * time.Hour,
				Issuer:             "test-issuer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewJWTManager(tt.config)

			if manager == nil {
				t.Fatal("NewJWTManager returned nil")
			}

			// Test that the manager works with this configuration
			token, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			claims, err := manager.ValidateToken(token)
			if err != nil {
				t.Fatalf("Failed to validate token: %v", err)
			}

			if claims.Issuer != tt.config.Issuer {
				t.Errorf("Expected issuer %s, got %s", tt.config.Issuer, claims.Issuer)
			}
		})
	}
}

func TestExtractUserID_DifferentTokenTypes(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	tests := []struct {
		name           string
		generateToken  func() (string, error)
		expectedUserID int64
	}{
		{
			name: "Extract from access token",
			generateToken: func() (string, error) {
				return manager.GenerateAccessToken(100, "test@example.com", "testuser")
			},
			expectedUserID: 100,
		},
		{
			name: "Extract from refresh token",
			generateToken: func() (string, error) {
				return manager.GenerateRefreshToken(200, "test@example.com", "testuser")
			},
			expectedUserID: 200,
		},
		{
			name: "Extract from verification token",
			generateToken: func() (string, error) {
				return manager.GenerateVerificationToken(300, "test@example.com")
			},
			expectedUserID: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.generateToken()
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}

			userID, err := manager.ExtractUserID(token)
			if err != nil {
				t.Fatalf("Failed to extract user ID: %v", err)
			}

			if userID != tt.expectedUserID {
				t.Errorf("Expected userID %d, got %d", tt.expectedUserID, userID)
			}
		})
	}
}

func TestTokenClaims_Timestamps(t *testing.T) {
	cfg := &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	manager := NewJWTManager(cfg)

	beforeGeneration := time.Now().Truncate(time.Second) // JWT uses second precision
	token, err := manager.GenerateAccessToken(123, "test@example.com", "testuser")
	afterGeneration := time.Now().Add(1 * time.Second).Truncate(time.Second)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Verify IssuedAt is within reasonable range (allowing for second precision)
	if claims.IssuedAt.Time.Before(beforeGeneration) || claims.IssuedAt.Time.After(afterGeneration) {
		t.Errorf("IssuedAt timestamp is outside expected range. Before: %v, IssuedAt: %v, After: %v",
			beforeGeneration, claims.IssuedAt.Time, afterGeneration)
	}

	// Verify ExpiresAt is in the future
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token should not be expired immediately after creation")
	}

	// Verify the expiry is set correctly
	expectedExpiry := claims.IssuedAt.Time.Add(15 * time.Minute)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry).Abs()
	if diff > 1*time.Second {
		t.Errorf("Expiry time mismatch. Expected ~%v, got %v (diff: %v)",
			expectedExpiry, claims.ExpiresAt.Time, diff)
	}
}
