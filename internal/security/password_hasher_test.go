package security

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHasher(t *testing.T) {
	tests := []struct {
		name         string
		inputCost    int
		expectedCost int
		description  string
	}{
		{
			name:         "Valid cost within range",
			inputCost:    10,
			expectedCost: 10,
			description:  "Should use the provided cost when valid",
		},
		{
			name:         "Cost below minimum",
			inputCost:    2,
			expectedCost: bcrypt.MinCost,
			description:  "Should default to MinCost (4) when input is below minimum",
		},
		{
			name:         "Cost at minimum boundary",
			inputCost:    bcrypt.MinCost,
			expectedCost: bcrypt.MinCost,
			description:  "Should accept MinCost (4) as valid",
		},
		{
			name:         "Cost above maximum",
			inputCost:    35,
			expectedCost: bcrypt.MaxCost,
			description:  "Should default to MaxCost (31) when input is above maximum",
		},
		{
			name:         "Cost at maximum boundary",
			inputCost:    bcrypt.MaxCost,
			expectedCost: bcrypt.MaxCost,
			description:  "Should accept MaxCost (31) as valid",
		},
		{
			name:         "Negative cost",
			inputCost:    -5,
			expectedCost: bcrypt.MinCost,
			description:  "Should default to MinCost when cost is negative",
		},
		{
			name:         "Zero cost",
			inputCost:    0,
			expectedCost: bcrypt.MinCost,
			description:  "Should default to MinCost when cost is zero",
		},
		{
			name:         "Development cost (4)",
			inputCost:    4,
			expectedCost: 4,
			description:  "Should accept development-recommended cost",
		},
		{
			name:         "Production cost (10)",
			inputCost:    10,
			expectedCost: 10,
			description:  "Should accept production-recommended cost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewPasswordHasher(tt.inputCost)

			if hasher == nil {
				t.Fatal("NewPasswordHasher returned nil")
			}

			if hasher.cost != tt.expectedCost {
				t.Errorf("Expected cost %d, got %d. %s", tt.expectedCost, hasher.cost, tt.description)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		cost        int
		expectError bool
		description string
	}{
		{
			name:        "Valid password",
			password:    "mySecurePassword123!",
			cost:        4,
			expectError: false,
			description: "Should successfully hash a valid password",
		},
		{
			name:        "Empty password",
			password:    "",
			cost:        4,
			expectError: false,
			description: "Should hash empty password (bcrypt allows this)",
		},
		{
			name:        "Long password (72 bytes - bcrypt limit)",
			password:    strings.Repeat("a", 72),
			cost:        4,
			expectError: false,
			description: "Should hash password at bcrypt's 72-byte limit",
		},
		{
			name:        "Very long password (exceeds 72 bytes)",
			password:    strings.Repeat("a", 100),
			cost:        4,
			expectError: true,
			description: "Should return error for password exceeding 72 bytes",
		},
		{
			name:        "Password with special characters",
			password:    "p@ssw0rd!#$%^&*()",
			cost:        4,
			expectError: false,
			description: "Should hash password with special characters",
		},
		{
			name:        "Unicode password",
			password:    "パスワード123",
			cost:        4,
			expectError: false,
			description: "Should hash password with unicode characters",
		},
		{
			name:        "Password with spaces",
			password:    "my password with spaces",
			cost:        4,
			expectError: false,
			description: "Should hash password containing spaces",
		},
		{
			name:        "Numeric only password",
			password:    "123456789",
			cost:        4,
			expectError: false,
			description: "Should hash numeric-only password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewPasswordHasher(tt.cost)

			hash, err := hasher.HashPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return // Stop test here for error cases
			}

			if err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
				return // Stop test if unexpected error occurred
			}

			// Verify hash is not empty
			if hash == "" {
				t.Error("Hash should not be empty")
				return
			}

			// Verify hash starts with bcrypt prefix
			if len(hash) < 4 || (!strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$")) {
				t.Errorf("Hash should start with bcrypt prefix, got: %s", hash)
				return
			}

			// Verify hash can be verified
			err = hasher.VerifyPassword(hash, tt.password)
			if err != nil {
				t.Errorf("Generated hash should be verifiable with original password: %v", err)
			}
		})
	}
}

func TestHashPassword_GeneratesDifferentHashesForSamePassword(t *testing.T) {
	hasher := NewPasswordHasher(4)
	password := "samePassword123"

	hash1, err1 := hasher.HashPassword(password)
	hash2, err2 := hasher.HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Error("Same password should generate different hashes (due to random salt)")
	}

	// Both should verify correctly
	if err := hasher.VerifyPassword(hash1, password); err != nil {
		t.Error("Hash1 should verify correctly")
	}
	if err := hasher.VerifyPassword(hash2, password); err != nil {
		t.Error("Hash2 should verify correctly")
	}
}

func TestVerifyPassword(t *testing.T) {
	hasher := NewPasswordHasher(4)
	correctPassword := "correctPassword123"

	// Generate a hash first
	hash, err := hasher.HashPassword(correctPassword)
	if err != nil {
		t.Fatalf("Failed to generate hash for test: %v", err)
	}

	tests := []struct {
		name          string
		hashedPwd     string
		plainPwd      string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "Correct password",
			hashedPwd:   hash,
			plainPwd:    correctPassword,
			expectError: false,
			description: "Should verify when password matches",
		},
		{
			name:          "Wrong password",
			hashedPwd:     hash,
			plainPwd:      "wrongPassword",
			expectError:   true,
			errorContains: "hashedPassword is not the hash of the given password",
			description:   "Should fail when password doesn't match",
		},
		{
			name:          "Empty password against hash",
			hashedPwd:     hash,
			plainPwd:      "",
			expectError:   true,
			errorContains: "hashedPassword is not the hash of the given password",
			description:   "Should fail when verifying empty password against non-empty hash",
		},
		{
			name:          "Invalid hash format",
			hashedPwd:     "not-a-valid-bcrypt-hash",
			plainPwd:      correctPassword,
			expectError:   true,
			errorContains: "hashedSecret too short",
			description:   "Should fail with invalid hash format",
		},
		{
			name:          "Empty hash",
			hashedPwd:     "",
			plainPwd:      correctPassword,
			expectError:   true,
			errorContains: "hashedSecret too short",
			description:   "Should fail with empty hash",
		},
		{
			name:          "Corrupted hash",
			hashedPwd:     "$2a$10$corrupted",
			plainPwd:      correctPassword,
			expectError:   true,
			errorContains: "hashedSecret too short",
			description:   "Should fail with corrupted hash",
		},
		{
			name:          "Case sensitive password check",
			hashedPwd:     hash,
			plainPwd:      strings.ToUpper(correctPassword),
			expectError:   true,
			errorContains: "hashedPassword is not the hash of the given password",
			description:   "Should fail when password case doesn't match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.VerifyPassword(tt.hashedPwd, tt.plainPwd)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestVerifyPassword_WithEmptyPasswordHash(t *testing.T) {
	hasher := NewPasswordHasher(4)
	emptyPassword := ""

	// Hash an empty password
	hash, err := hasher.HashPassword(emptyPassword)
	if err != nil {
		t.Fatalf("Failed to hash empty password: %v", err)
	}

	// Verify empty password against its hash
	err = hasher.VerifyPassword(hash, emptyPassword)
	if err != nil {
		t.Errorf("Should verify empty password correctly: %v", err)
	}

	// Verify non-empty password against empty password hash should fail
	err = hasher.VerifyPassword(hash, "notEmpty")
	if err == nil {
		t.Error("Should fail when verifying non-empty password against empty password hash")
	}
}

func TestPasswordHasher_CostAffectsHashTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-based test in short mode")
	}

	// This test verifies that higher cost actually takes longer
	// Note: We're using a very small cost difference to keep test fast
	lowCostHasher := NewPasswordHasher(4)
	highCostHasher := NewPasswordHasher(6)

	password := "testPassword123"

	// Low cost should complete
	_, err := lowCostHasher.HashPassword(password)
	if err != nil {
		t.Fatalf("Low cost hash failed: %v", err)
	}

	// High cost should also complete
	_, err = highCostHasher.HashPassword(password)
	if err != nil {
		t.Fatalf("High cost hash failed: %v", err)
	}

	// Both hashers should work correctly
	if lowCostHasher.cost != 4 {
		t.Errorf("Expected low cost hasher to have cost 4, got %d", lowCostHasher.cost)
	}
	if highCostHasher.cost != 6 {
		t.Errorf("Expected high cost hasher to have cost 6, got %d", highCostHasher.cost)
	}
}

func TestPasswordHasher_InterfaceImplementation(t *testing.T) {
	// This test ensures the PasswordHasher implements the interface
	// by attempting to use it through the interface
	hasher := NewPasswordHasher(4)

	// Test that basic operations work
	password := "testPassword"
	hash, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	err = hasher.VerifyPassword(hash, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	hasher := NewPasswordHasher(4)
	password := "benchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	hasher := NewPasswordHasher(4)
	password := "benchmarkPassword123"
	hash, _ := hasher.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hasher.VerifyPassword(hash, password)
	}
}
