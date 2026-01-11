package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher handles password hashing and verification using bcrypt
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher with specified cost
// Recommended cost: 4 for development (~100ms), 10 for production (~3s)
func NewPasswordHasher(cost int) *PasswordHasher {
	// Validate cost range (bcrypt accepts 4-31)
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost // 4
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost // 31
	}

	return &PasswordHasher{
		cost: cost,
	}
}

// NewPasswordHasherWithDefaultCost creates hasher with bcrypt default cost (10)
func NewPasswordHasherWithDefaultCost() *PasswordHasher {
	return &PasswordHasher{
		cost: bcrypt.DefaultCost, // 10
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
