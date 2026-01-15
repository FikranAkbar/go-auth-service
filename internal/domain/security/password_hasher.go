package security

// PasswordHasherInterface defines the contract for password hashing operations
// This allows mocking in tests
type PasswordHasherInterface interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}
