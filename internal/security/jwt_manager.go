package security

import (
	"errors"
	"go-auth-service/internal/config"
	"go-auth-service/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   int64     `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:          cfg.SecretKey,
		accessTokenExpiry:  cfg.AccessTokenExpiry,
		refreshTokenExpiry: cfg.RefreshTokenExpiry,
		issuer:             cfg.Issuer,
	}
}

// GenerateAccessToken generates a new access token for a user
func (jm *JWTManager) GenerateAccessToken(userID int64, email, username string) (string, error) {
	return jm.generateToken(userID, email, username, AccessToken, jm.accessTokenExpiry)
}

// GenerateRefreshToken generates a new refresh token for a user
func (jm *JWTManager) GenerateRefreshToken(userID int64, email, username string) (string, error) {
	return jm.generateToken(userID, email, username, RefreshToken, jm.refreshTokenExpiry)
}

// generateToken is a helper function to generate JWT tokens
func (jm *JWTManager) generateToken(userID int64, email, username string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    jm.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jm.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(constants.ErrInvalidToken)
		}
		return []byte(jm.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New(constants.ErrInvalidToken)
	}

	return claims, nil
}

// ExtractUserID extracts the user ID from a token without full validation
func (jm *JWTManager) ExtractUserID(tokenString string) (int64, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
