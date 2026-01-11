package auth

import (
	"go-auth-service/internal/domain/user"
	"time"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogoutRequest represents the request body for user logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ResendVerificationRequest represents the request body for resending verification email
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// RegisterResponse represents the response for successful registration
type RegisterResponse struct {
	User         *user.UserResponse `json:"user"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	User         *user.UserResponse `json:"user"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // Access token expiration in seconds
}

// RefreshTokenData represents stored refresh token information in Redis
type RefreshTokenData struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
}
