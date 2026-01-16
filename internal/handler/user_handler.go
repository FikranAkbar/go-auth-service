package handler

import (
	"context"
	"go-auth-service/internal/domain/user"
	"go-auth-service/pkg/constants"
	appErrors "go-auth-service/pkg/errors"
	"go-auth-service/pkg/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	UserService user.ServiceInterface
}

func NewUserHandler(userService user.ServiceInterface) *UserHandler {
	return &UserHandler{UserService: userService}
}

// GetCurrentUser returns the current authenticated user's profile
// Protected by auth middleware (gets user_id from context)
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(constants.ContextKeyUserID).(int64)
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Get user from service
	foundUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	// Return user response (sensitive data like password_hash excluded)
	response.Success(w, foundUser.ToResponse(), "User retrieved successfully")
}

// GetUserByID retrieves a user by their ID (admin/public endpoint)
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from URL parameter
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid user ID format")
		return
	}

	// Get user from service
	foundUser, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	// Return user response (sensitive data like password_hash excluded)
	response.Success(w, foundUser.ToResponse(), "User retrieved successfully")
}
