package constants

import "time"

// JWT Token Types
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// JWT Claims Keys
const (
	JWTClaimUserID    = "user_id"
	JWTClaimEmail     = "email"
	JWTClaimTokenType = "token_type"
	JWTClaimIssuedAt  = "iat"
	JWTClaimExpiresAt = "exp"
	JWTClaimSubject   = "sub"
)

// JWT Token Expiration
const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// JWT Error Messages
const (
	ErrJWTInvalidToken     = "Invalid JWT token"
	ErrJWTExpiredToken     = "JWT token has expired"
	ErrJWTMalformedToken   = "Malformed JWT token"
	ErrJWTSignatureInvalid = "Invalid JWT signature"
)
