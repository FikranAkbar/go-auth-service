package user

// ServiceInterface defines the contract for user business logic.
// This interface allows the handler layer to work with different service implementations,
// making the code testable by enabling mock services in tests.
type ServiceInterface interface {
	// TODO: Add service methods here as you implement them.
	// Example methods you'll likely need:
	//
	// Register(ctx context.Context, email, password, name string) (*User, error)
	// GetUserByID(ctx context.Context, userID string) (*User, error)
	// GetUserByEmail(ctx context.Context, email string) (*User, error)
	// UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) (*User, error)
	// DeleteUser(ctx context.Context, userID string) error
}
