package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/domain/security"
	"go-auth-service/internal/domain/user"
	"go-auth-service/pkg/constants"
	appErrors "go-auth-service/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCKS
// ============================================================================

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx context.Context, email, username, password string) (int64, string, error) {
	args := m.Called(ctx, email, username, password)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*user.User, string, string, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, "", "", args.Error(3)
	}
	return args.Get(0).(*user.User), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) VerifyEmail(ctx context.Context, userID int64) (*user.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockAuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthService) Logout(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GenerateAndStoreTokens(ctx context.Context, u *user.User) (string, string, error) {
	args := m.Called(ctx, u)
	return args.String(0), args.String(1), args.Error(2)
}

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

// ============================================================================
// REGISTER ENDPOINT TESTS
// Contract: POST /api/auth/register
// Input: {"email": "user@example.com", "username": "user123", "password": "securePass123"}
// Success: 201 with message and email
// Failures: 400 (invalid input), 409 (duplicate), 500 (server error)
// ============================================================================

func TestAuthHandler_Register_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RegisterUser", mock.Anything, "test@example.com", "testuser", "password123").
		Return(int64(1), "test@example.com", nil)

	reqBody := auth.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify response body contains the error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, constants.ErrInvalidRequestBody, response["error"])
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RegisterUser", mock.Anything, "existing@example.com", "testuser", "password123").
		Return(int64(0), "", appErrors.Conflict(appErrors.ErrEmailAlreadyExists, "Email already in use"))

	reqBody := auth.RegisterRequest{
		Email:    "existing@example.com",
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "Email already in use", response["error"])

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Register_UsernameAlreadyExists(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RegisterUser", mock.Anything, "test@example.com", "existinguser", "password123").
		Return(int64(0), "", appErrors.Conflict(appErrors.ErrUsernameAlreadyExists, "Username already in use"))

	reqBody := auth.RegisterRequest{
		Email:    "test@example.com",
		Username: "existinguser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// LOGIN ENDPOINT TESTS
// Contract: POST /api/auth/login
// Input: {"email": "user@example.com", "password": "securePass123"}
// Success: 200 with user data, access_token, refresh_token
// Failures: 400 (invalid input), 401 (wrong credentials/unverified), 500 (server error)
// ============================================================================

func TestAuthHandler_Login_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	testUser := &user.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockAuthService.On("Login", mock.Anything, "test@example.com", "password123").
		Return(testUser, "access_token_here", "refresh_token_here", nil)

	reqBody := auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("Login", mock.Anything, "test@example.com", "wrongpassword").
		Return(nil, "", "", appErrors.Unauthorized(appErrors.ErrInvalidCredentials, "Invalid credentials"))

	reqBody := auth.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "Invalid credentials", response["error"])

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_UserNotVerified(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("Login", mock.Anything, "test@example.com", "password123").
		Return(nil, "", "", appErrors.Unauthorized(appErrors.ErrUserNotVerified, "Please verify your email first"))

	reqBody := auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// VERIFY EMAIL ENDPOINT TESTS
// Contract: GET /api/auth/verify-email?token=xyz
// Input: token in query parameter
// Success: 200 with user data, access_token, refresh_token
// Failures: 400 (missing/invalid token), 404 (user not found), 500 (server error)
// ============================================================================

func TestAuthHandler_VerifyEmail_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	testUser := &user.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	claims := &security.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Type:   security.VerificationToken,
	}

	mockJWTManager.On("ValidateToken", "valid_token").Return(claims, nil)
	mockAuthService.On("VerifyEmail", mock.Anything, int64(1)).Return(testUser, nil)
	mockAuthService.On("GenerateAndStoreTokens", mock.Anything, testUser).
		Return("access_token", "refresh_token", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email?token=valid_token", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_MissingToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_VerifyEmail_InvalidToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockJWTManager.On("ValidateToken", "invalid_token").
		Return(nil, errors.New("invalid token"))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email?token=invalid_token", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockJWTManager.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_WrongTokenType(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	claims := &security.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Type:   security.AccessToken, // Wrong type!
	}

	mockJWTManager.On("ValidateToken", "access_token").Return(claims, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email?token=access_token", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockJWTManager.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_VerifyEmailFails(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	claims := &security.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Type:   security.VerificationToken,
	}

	mockJWTManager.On("ValidateToken", "valid_token").Return(claims, nil)
	mockAuthService.On("VerifyEmail", mock.Anything, int64(1)).
		Return(nil, appErrors.NotFound(appErrors.ErrUserNotFound, "User not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email?token=valid_token", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "User not found", response["error"])

	mockAuthService.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestAuthHandler_VerifyEmail_GenerateTokensFails(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	testUser := &user.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	claims := &security.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Type:   security.VerificationToken,
	}

	mockJWTManager.On("ValidateToken", "valid_token").Return(claims, nil)
	mockAuthService.On("VerifyEmail", mock.Anything, int64(1)).Return(testUser, nil)
	mockAuthService.On("GenerateAndStoreTokens", mock.Anything, testUser).
		Return("", "", errors.New("failed to generate tokens"))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify-email?token=valid_token", nil)
	w := httptest.NewRecorder()

	handler.VerifyEmail(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "Failed to generate authentication tokens", response["error"])

	mockAuthService.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

// ============================================================================
// RESEND VERIFICATION EMAIL ENDPOINT TESTS
// Contract: POST /api/auth/resend-verification
// Input: {"email": "user@example.com"}
// Success: 200 with message
// Failures: 400 (invalid input/already verified), 500 (email/token error)
// Note: For security, returns 200 even if email doesn't exist
// ============================================================================

func TestAuthHandler_ResendVerificationEmail_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("ResendVerificationEmail", mock.Anything, "test@example.com").Return(nil)

	reqBody := auth.ResendVerificationRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ResendVerificationEmail_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ResendVerificationEmail_EmptyEmail(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	reqBody := auth.ResendVerificationRequest{
		Email: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ResendVerificationEmail_AlreadyVerified(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("ResendVerificationEmail", mock.Anything, "test@example.com").
		Return(appErrors.ErrUserAlreadyVerified)

	reqBody := auth.ResendVerificationRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ResendVerificationEmail_TokenGenerationFailed(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("ResendVerificationEmail", mock.Anything, "test@example.com").
		Return(appErrors.Internal(appErrors.ErrTokenGenFailed, "Failed to generate token"))

	reqBody := auth.ResendVerificationRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "Failed to generate verification token. Please try again later.", response["error"])

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ResendVerificationEmail_EmailSendFailed(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("ResendVerificationEmail", mock.Anything, "test@example.com").
		Return(appErrors.Internal(appErrors.ErrEmailSendFailed, "Failed to send email"))

	reqBody := auth.ResendVerificationRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Verify response body contains specific error message
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "Failed to send verification email. Please try again later.", response["error"])

	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ResendVerificationEmail_UserNotFound_ReturnsSuccessForSecurity(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	// Service returns unknown error (like user not found)
	mockAuthService.On("ResendVerificationEmail", mock.Anything, "nonexistent@example.com").
		Return(errors.New("some unknown error"))

	reqBody := auth.ResendVerificationRequest{
		Email: "nonexistent@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	// Should return 200 to not reveal if user exists
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_ResendVerificationEmail_UnknownAppError_ReturnsSuccessForSecurity(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	// Return an AppError but with a different underlying error (not TokenGenFailed or EmailSendFailed)
	unknownErr := errors.New("some other database error")
	mockAuthService.On("ResendVerificationEmail", mock.Anything, "test@example.com").
		Return(appErrors.Internal(unknownErr, "Some other internal error"))

	reqBody := auth.ResendVerificationRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/resend-verification", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ResendVerificationEmail(w, req)

	// Should return 200 with generic message for security (default case in switch)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// LOGOUT ENDPOINT TESTS
// Contract: POST /api/auth/logout
// Input: user_id from context (set by auth middleware)
// Success: 200 with success message
// Failures: 401 (not authenticated), 500 (server error)
// ============================================================================

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("Logout", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	ctx := context.WithValue(req.Context(), constants.ContextKeyUserID, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Logout_NotAuthenticated(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	// No user ID in context
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_ServiceError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("Logout", mock.Anything, int64(1)).
		Return(errors.New("redis error"))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	ctx := context.WithValue(req.Context(), constants.ContextKeyUserID, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// REFRESH TOKEN ENDPOINT TESTS
// Contract: POST /api/auth/refresh-token
// Input: {"refresh_token": "token_here"}
// Success: 200 with new access_token
// Failures: 400 (invalid input), 401 (invalid/expired token), 500 (server error)
// ============================================================================

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RefreshToken", mock.Anything, "valid_refresh_token").
		Return("new_access_token", nil)

	reqBody := auth.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_EmptyToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	reqBody := auth.RefreshTokenRequest{
		RefreshToken: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RefreshToken", mock.Anything, "invalid_token").
		Return("", appErrors.Unauthorized(appErrors.ErrInvalidToken, "Invalid token"))

	reqBody := auth.RefreshTokenRequest{
		RefreshToken: "invalid_token",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_ExpiredToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTManager := new(MockJWTManager)
	handler := NewAuthHandler(mockAuthService, mockJWTManager)

	mockAuthService.On("RefreshToken", mock.Anything, "expired_token").
		Return("", appErrors.Unauthorized(appErrors.ErrExpiredToken, "Token expired"))

	reqBody := auth.RefreshTokenRequest{
		RefreshToken: "expired_token",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh-token", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}
