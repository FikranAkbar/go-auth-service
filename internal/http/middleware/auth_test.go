package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-auth-service/internal/domain/security"
	"go-auth-service/pkg/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCKS
// ============================================================================

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateAccessToken(userID int64, email, username string) (string, error) {
	args := m.Called(userID, email, username)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) GenerateRefreshToken(userID int64, email, username string) (string, error) {
	args := m.Called(userID, email, username)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) GenerateVerificationToken(userID int64, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*security.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*security.Claims), args.Error(1)
}

func (m *MockJWTManager) ExtractUserID(tokenString string) (int64, error) {
	args := m.Called(tokenString)
	return args.Get(0).(int64), args.Error(1)
}

// Mock handler to test if request reaches downstream
type MockHandler struct {
	Called      bool
	UserID      int64
	Email       string
	Username    string
	HasUserID   bool
	HasEmail    bool
	HasUsername bool
}

func (h *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true

	// Check if user context is set
	if userID, ok := r.Context().Value(constants.ContextKeyUserID).(int64); ok {
		h.HasUserID = true
		h.UserID = userID
	}

	if email, ok := r.Context().Value(constants.ContextKeyUserEmail).(string); ok {
		h.HasEmail = true
		h.Email = email
	}

	if username, ok := r.Context().Value(constants.ContextKeyUsername).(string); ok {
		h.HasUsername = true
		h.Username = username
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ============================================================================
// REQUIRE AUTH MIDDLEWARE TESTS
// Contract: Validates JWT access token and rejects invalid requests
// Success: Token valid → sets user context → calls next handler
// Failures: Missing header, invalid format, invalid token, wrong token type
// ============================================================================

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.AccessToken,
	}

	mockJWT.On("ValidateToken", "valid_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was called
	assert.True(t, mockHandler.Called)

	// Verify user context was set
	assert.True(t, mockHandler.HasUserID)
	assert.Equal(t, int64(123), mockHandler.UserID)
	assert.True(t, mockHandler.HasEmail)
	assert.Equal(t, "test@example.com", mockHandler.Email)
	assert.True(t, mockHandler.HasUsername)
	assert.Equal(t, "testuser", mockHandler.Username)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_MissingAuthorizationHeader(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header set
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// No JWT validation should have been called
	mockJWT.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidHeaderFormat_NoBearer(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockJWT.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidHeaderFormat_OnlyBearer(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockJWT.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_RequireAuth_InvalidHeaderFormat_CaseInsensitiveBearer(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.AccessToken,
	}

	mockJWT.On("ValidateToken", "valid_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "BEARER valid_token") // Uppercase BEARER
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Should work because we use strings.ToLower
	assert.True(t, mockHandler.Called)
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	mockJWT.On("ValidateToken", "invalid_token").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_WrongTokenType_RefreshToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.RefreshToken, // Wrong type!
	}

	mockJWT.On("ValidateToken", "refresh_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer refresh_token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_WrongTokenType_VerificationToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.VerificationToken, // Wrong type!
	}

	mockJWT.On("ValidateToken", "verification_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer verification_token")
	w := httptest.NewRecorder()

	handler := middleware.RequireAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was NOT called
	assert.False(t, mockHandler.Called)

	// Verify unauthorized response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	mockJWT.AssertExpectations(t)
}

// ============================================================================
// OPTIONAL AUTH MIDDLEWARE TESTS
// Contract: Extracts user info if token present, but doesn't require it
// Success: Valid token → sets context, Invalid/No token → continues without context
// Failures: None (always continues to next handler)
// ============================================================================

func TestAuthMiddleware_OptionalAuth_Success_WithValidToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.AccessToken,
	}

	mockJWT.On("ValidateToken", "valid_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was called
	assert.True(t, mockHandler.Called)

	// Verify user context was set
	assert.True(t, mockHandler.HasUserID)
	assert.Equal(t, int64(123), mockHandler.UserID)
	assert.True(t, mockHandler.HasEmail)
	assert.Equal(t, "test@example.com", mockHandler.Email)
	assert.True(t, mockHandler.HasUsername)
	assert.Equal(t, "testuser", mockHandler.Username)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_Success_NoToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	// No Authorization header
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was called
	assert.True(t, mockHandler.Called)

	// Verify user context was NOT set
	assert.False(t, mockHandler.HasUserID)
	assert.False(t, mockHandler.HasEmail)
	assert.False(t, mockHandler.HasUsername)

	// Verify response is still OK
	assert.Equal(t, http.StatusOK, w.Code)

	// No JWT validation should have been called
	mockJWT.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_OptionalAuth_InvalidHeaderFormat_ContinuesWithoutAuth(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was still called
	assert.True(t, mockHandler.Called)

	// Verify user context was NOT set
	assert.False(t, mockHandler.HasUserID)

	// Verify response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertNotCalled(t, "ValidateToken")
}

func TestAuthMiddleware_OptionalAuth_InvalidToken_ContinuesWithoutAuth(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	mockJWT.On("ValidateToken", "invalid_token").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was still called
	assert.True(t, mockHandler.Called)

	// Verify user context was NOT set
	assert.False(t, mockHandler.HasUserID)

	// Verify response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WrongTokenType_ContinuesWithoutAuth(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	claims := &security.Claims{
		UserID:   123,
		Email:    "test@example.com",
		Username: "testuser",
		Type:     security.RefreshToken, // Wrong type
	}

	mockJWT.On("ValidateToken", "refresh_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer refresh_token")
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was still called
	assert.True(t, mockHandler.Called)

	// Verify user context was NOT set
	assert.False(t, mockHandler.HasUserID)

	// Verify response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_EmptyToken_ContinuesWithoutAuth(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)
	mockHandler := &MockHandler{}

	// Mock validation for empty string token
	mockJWT.On("ValidateToken", "").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	handler := middleware.OptionalAuth(mockHandler)
	handler.ServeHTTP(w, req)

	// Verify handler was still called
	assert.True(t, mockHandler.Called)

	// Verify user context was NOT set
	assert.False(t, mockHandler.HasUserID)

	// Verify response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	mockJWT.AssertExpectations(t)
}

// ============================================================================
// CONSTRUCTOR TEST
// ============================================================================

func TestNewAuthMiddleware(t *testing.T) {
	mockJWT := new(MockJWTManager)
	middleware := NewAuthMiddleware(mockJWT)

	assert.NotNil(t, middleware)
	assert.Equal(t, mockJWT, middleware.jwtManager)
}
