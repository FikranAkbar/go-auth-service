package middleware

import (
	"context"
	domainSecurity "go-auth-service/internal/domain/security"
	"go-auth-service/pkg/constants"
	"go-auth-service/pkg/response"
	"net/http"
	"strings"
)

// AuthMiddleware provides JWT authentication for protected routes
type AuthMiddleware struct {
	jwtManager domainSecurity.JWTManagerInterface
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager domainSecurity.JWTManagerInterface) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth is a middleware that validates JWT access token and sets user context
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "Missing authorization header")
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(w, "Invalid authorization header format. Expected: Bearer <token>")
			return
		}

		accessToken := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateToken(accessToken)
		if err != nil {
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Verify token type is access token
		if claims.Type != domainSecurity.AccessToken {
			response.Unauthorized(w, "Invalid token type. Expected access token")
			return
		}

		// Set user context for downstream handlers
		ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, constants.ContextKeyUserEmail, claims.Email)
		ctx = context.WithValue(ctx, constants.ContextKeyUsername, claims.Username)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that extracts user info if token is present, but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// Invalid format, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		accessToken := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateToken(accessToken)
		if err != nil {
			// Invalid token, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Verify token type is access token
		if claims.Type != domainSecurity.AccessToken {
			// Invalid token type, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Set user context for downstream handlers
		ctx := context.WithValue(r.Context(), constants.ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, constants.ContextKeyUserEmail, claims.Email)
		ctx = context.WithValue(ctx, constants.ContextKeyUsername, claims.Username)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
