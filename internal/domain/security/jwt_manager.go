package security

import (
	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken       TokenType = "access"
	RefreshToken      TokenType = "refresh"
	VerificationToken TokenType = "verification"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   int64     `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTManagerInterface defines the contract for JWT token operations
// This allows mocking in tests
type JWTManagerInterface interface {
	GenerateAccessToken(userID int64, email, username string) (string, error)
	GenerateRefreshToken(userID int64, email, username string) (string, error)
	GenerateVerificationToken(userID int64, email string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	ExtractUserID(tokenString string) (int64, error)
}
