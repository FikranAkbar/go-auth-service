package user

import "context"

// RepositoryInterface defines the contract for user data operations.
// This interface allows the service layer to work with different repository implementations,
// making the code testable by enabling mock repositories in tests.
type RepositoryInterface interface {
	// CreateUser creates a new user in the database
	CreateUser(ctx context.Context, user *User) error

	// FindUserByEmail finds a user by their email address
	FindUserByEmail(ctx context.Context, email string) (*User, error)

	// FindUserByID finds a user by their ID
	FindUserByID(ctx context.Context, id int64) (*User, error)

	// ExistsByEmail checks if a user with the given email already exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username already exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id int64) error
}
