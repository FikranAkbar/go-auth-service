package auth

// ServiceInterface defines the contract for authentication business logic.
// This interface allows the handler layer to work with different auth service implementations.
type ServiceInterface interface {
	// TODO: Add auth service methods here as you implement them.
	// Example methods you'll likely need:
	//
	// Login(ctx context.Context, email, password string) (*TokenPair, error)
	// RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	// Logout(ctx context.Context, refreshToken string) error
	// ValidateAccessToken(ctx context.Context, accessToken string) (userID string, err error)
}

// TokenRepositoryInterface defines the contract for token storage operations.
type TokenRepositoryInterface interface {
	// TODO: Add token repository methods here as you implement them.
	// Example methods:
	//
	// SaveRefreshToken(ctx context.Context, data *RefreshTokenData) error
	// GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error)
	// RevokeRefreshToken(ctx context.Context, token string) error
	// DeleteExpiredTokens(ctx context.Context) error
}
