package handler

import (
	"context"
	"encoding/json"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/domain/user"
	"go-auth-service/internal/security"
	appErrors "go-auth-service/pkg/errors"
	"go-auth-service/pkg/response"
	"net/http"
)

type AuthHandler struct {
	userService user.ServiceInterface
	jwtManager  *security.JWTManager
}

func NewAuthHandler(userService user.ServiceInterface, jwtManager *security.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse request body
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Register user (includes validation and password hashing)
	newUser, err := h.userService.RegisterUser(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	// Generate JWT tokens
	accessToken, err := h.jwtManager.GenerateAccessToken(newUser.ID, newUser.Email, newUser.Username)
	if err != nil {
		response.InternalServerError(w, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(newUser.ID, newUser.Email, newUser.Username)
	if err != nil {
		response.InternalServerError(w, "Failed to generate refresh token")
		return
	}

	// TODO: Store refresh token in Redis
	// For now, we're just generating it

	// Prepare response
	registerResp := auth.RegisterResponse{
		User:         newUser.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.Created(w, registerResp, "User registered successfully")
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement login
	response.InternalServerError(w, "Login endpoint not yet implemented")
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logout
	response.InternalServerError(w, "Logout endpoint not yet implemented")
}
