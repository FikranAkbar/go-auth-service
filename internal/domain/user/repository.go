package user

// RepositoryInterface defines the contract for user data operations.
// This interface allows the service layer to work with different repository implementations,
// making the code testable by enabling mock repositories in tests.
type RepositoryInterface interface {
	// TODO: Add repository methods here as you implement them.
	// Example methods you'll likely need:
	//
	// CreateUser(ctx context.Context, user *User) error
	// FindUserByEmail(ctx context.Context, email string) (*User, error)
	// FindUserByID(ctx context.Context, id string) (*User, error)
	// UpdateUser(ctx context.Context, user *User) error
	// DeleteUser(ctx context.Context, id string) error
	// UserExistsByEmail(ctx context.Context, email string) (bool, error)
}
