package auth

import (
	"context"
	"go-auth-service/internal/domain/user"
)

// ServiceInterface defines the contract for authentication business logic
type ServiceInterface interface {
	// RegisterUser handles the complete user registration flow with transaction support
	// Returns error if registration or email sending fails (with automatic rollback)
	RegisterUser(ctx context.Context, email, username, password string) (userID int64, userEmail string, error error)

	// VerifyEmail verifies a user's email and activates their account
	// Returns the verified user
	VerifyEmail(ctx context.Context, userID int64) (*user.User, error)

	// ResendVerificationEmail resends verification email to unverified user
	ResendVerificationEmail(ctx context.Context, email string) error
}
