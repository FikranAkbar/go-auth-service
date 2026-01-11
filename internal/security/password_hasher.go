package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher handles password hashing and verification using bcrypt
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher with default cost
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: bcrypt.DefaultCost, // Cost factor of 10
	}
}

// HashPassword hashes a plain text password using bcrypt
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a hashed password with a plain text password
func (ph *PasswordHasher) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
