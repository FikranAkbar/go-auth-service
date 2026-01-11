package user

import (
	"context"
	"database/sql"
)

// ServiceInterface defines the contract for user business logic.
// This interface allows the handler layer to work with different service implementations,
// making the code testable by enabling mock services in tests.
type ServiceInterface interface {
	// BeginTx starts a new database transaction
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// RegisterUser creates a new user account
	RegisterUser(ctx context.Context, email, username, password string) (*User, error)

	// RegisterUserTx creates a new user account within a transaction
	RegisterUserTx(ctx context.Context, tx *sql.Tx, email, username, password string) (*User, error)

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
}
