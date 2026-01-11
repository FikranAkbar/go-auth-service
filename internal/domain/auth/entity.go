package auth

import "time"

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // Access token expiration in seconds
}

// RefreshTokenData represents stored refresh token information
type RefreshTokenData struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// TODO: Add other auth-related entities as needed
// Example:
// type LoginRequest struct { ... }
// type RegisterRequest struct { ... }
