package user

import (
	"context"
	"go-auth-service/internal/domain/repository"
)

// ServiceInterface defines the contract for user business logic.
type ServiceInterface interface {
	// BeginTx starts a new database transaction
	BeginTx(ctx context.Context) (repository.TransactionInterface, error)

	// RegisterUser creates a new user account
	RegisterUser(ctx context.Context, email, username, password string) (*User, error)

	// RegisterUserTx creates a new user account within a transaction
	RegisterUserTx(ctx context.Context, tx repository.TransactionInterface, email, username, password string) (*User, error)

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, userID int64) (*User, error)

	// GetUserByEmail retrieves a user by their email
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, userID int64) error

	// VerifyEmail activates a user account after email verification
	VerifyEmail(ctx context.Context, userID int64) error

	// VerifyPassword compares a hashed password with a plain text password
	VerifyPassword(hashedPassword, password string) error
}
