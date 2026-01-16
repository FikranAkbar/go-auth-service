package handler

import (
	"context"
	"database/sql"
	"go-auth-service/internal/domain/user"
	"go-auth-service/pkg/constants"
	appErrors "go-auth-service/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// MOCKS
// ============================================================================

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) BeginTx(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockUserService) RegisterUser(ctx context.Context, email, username, password string) (*user.User, error) {
	args := m.Called(ctx, email, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) RegisterUserTx(ctx context.Context, tx *sql.Tx, email, username, password string) (*user.User, error) {
	args := m.Called(ctx, tx, email, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) VerifyEmail(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) VerifyPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// ============================================================================
// GET CURRENT USER ENDPOINT TESTS
// Contract: GET /api/users/me
// Input: user_id from context (set by auth middleware)
// Success: 200 with user data
// Failures: 401 (not authenticated), 404 (user not found), 500 (server error)
// ============================================================================

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	testUser := &user.User{
		ID:        1,
		Email:     "test@example.com",
		Username:  "testuser",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserService.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	ctx := context.WithValue(req.Context(), constants.ContextKeyUserID, int64(1))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_NotAuthenticated(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	// No user ID in context
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandler_GetCurrentUser_UserNotFound(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	mockUserService.On("GetUserByID", mock.Anything, int64(999)).
		Return(nil, appErrors.NotFound(appErrors.ErrUserNotFound, "User not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	ctx := context.WithValue(req.Context(), constants.ContextKeyUserID, int64(999))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetCurrentUser_InvalidUserIDType(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	// Wrong type in context (string instead of int64)
	ctx := context.WithValue(req.Context(), constants.ContextKeyUserID, "invalid")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetCurrentUser(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ============================================================================
// GET USER BY ID ENDPOINT TESTS
// Contract: GET /api/users/{id}
// Input: id from URL parameter
// Success: 200 with user data
// Failures: 400 (invalid id), 404 (user not found), 500 (server error)
// ============================================================================

func TestUserHandler_GetUserByID_Success(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	testUser := &user.User{
		ID:        123,
		Email:     "user@example.com",
		Username:  "username123",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserService.On("GetUserByID", mock.Anything, int64(123)).Return(testUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)

	// Setup chi context for URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetUserByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetUserByID_MissingID(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodGet, "/api/users/", nil)

	// Setup chi context without id param
	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetUserByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetUserByID_InvalidIDFormat(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)

	// Setup chi context with invalid id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetUserByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetUserByID_UserNotFound(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	mockUserService.On("GetUserByID", mock.Anything, int64(999)).
		Return(nil, appErrors.NotFound(appErrors.ErrUserNotFound, "User not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)

	// Setup chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetUserByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetUserByID_NegativeID(t *testing.T) {
	mockUserService := new(MockUserService)
	handler := NewUserHandler(mockUserService)

	// Even though -1 is invalid, it will be parsed as int64(-1)
	// The service layer should handle this validation
	mockUserService.On("GetUserByID", mock.Anything, int64(-1)).
		Return(nil, appErrors.BadRequest(appErrors.ErrUserNotFound, "Invalid user ID"))

	req := httptest.NewRequest(http.MethodGet, "/api/users/-1", nil)

	// Setup chi context with negative id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetUserByID(w, req)

	// Should return error from service
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUserService.AssertExpectations(t)
}
