package security

import (
	"errors"
	"go-auth-service/internal/config"
	domainSecurity "go-auth-service/internal/domain/security"
	"go-auth-service/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
	return jm.generateToken(userID, email, username, domainSecurity.AccessToken, jm.accessTokenExpiry)
}

// GenerateRefreshToken generates a new refresh token for a user
func (jm *JWTManager) GenerateRefreshToken(userID int64, email, username string) (string, error) {
	return jm.generateToken(userID, email, username, domainSecurity.RefreshToken, jm.refreshTokenExpiry)
}

// GenerateVerificationToken generates a token for email verification (24 hour expiry)
func (jm *JWTManager) GenerateVerificationToken(userID int64, email string) (string, error) {
	return jm.generateToken(userID, email, "", domainSecurity.VerificationToken, 24*time.Hour)
}

// generateToken is a helper function to generate JWT tokens
func (jm *JWTManager) generateToken(userID int64, email, username string, tokenType domainSecurity.TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := domainSecurity.Claims{
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
func (jm *JWTManager) ValidateToken(tokenString string) (*domainSecurity.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domainSecurity.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(constants.ErrInvalidToken)
		}
		return []byte(jm.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*domainSecurity.Claims)
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

// Compile-time check to ensure JWTManager implements the interface
var _ domainSecurity.JWTManagerInterface = (*JWTManager)(nil)
