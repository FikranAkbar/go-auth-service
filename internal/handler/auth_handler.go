package handler

import (
	"context"
	"encoding/json"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/security"
	"go-auth-service/pkg/constants"
	appErrors "go-auth-service/pkg/errors"
	"go-auth-service/pkg/response"
	"net/http"
)

type AuthHandler struct {
	authService auth.ServiceInterface
	jwtManager  *security.JWTManager
}

func NewAuthHandler(authService auth.ServiceInterface, jwtManager *security.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
	}
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, constants.ErrInvalidRequestBody)
		return
	}
	_, userEmail, err := h.authService.RegisterUser(ctx, req.Email, req.Username, req.Password)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}
	registerResp := map[string]interface{}{
		"message": "Registration successful. Please check your email to verify your account.",
		"email":   userEmail,
	}
	response.Created(w, registerResp, "User registered successfully. Please verify your email.")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, constants.ErrInvalidRequestBody)
		return
	}

	loggedInUser, accessToken, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	loginResp := auth.LoginResponse{
		User:         loggedInUser.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.Success(w, loginResp, "Login successful")
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	token := r.URL.Query().Get("token")
	if token == "" {
		response.BadRequest(w, "Verification token is required")
		return
	}
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		response.BadRequest(w, "Invalid or expired verification token")
		return
	}
	if claims.Type != security.VerificationToken {
		response.BadRequest(w, "Invalid token type")
		return
	}
	verifiedUser, err := h.authService.VerifyEmail(ctx, claims.UserID)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	// Generate tokens and store refresh token in Redis
	accessToken, refreshToken, err := h.authService.GenerateAndStoreTokens(ctx, verifiedUser)
	if err != nil {
		response.InternalServerError(w, "Failed to generate authentication tokens")
		return
	}

	verifyResp := auth.RegisterResponse{
		User:         verifiedUser.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	response.Success(w, verifyResp, "Email verified successfully. You are now logged in.")
}
func (h *AuthHandler) ResendVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req auth.ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, constants.ErrInvalidRequestBody)
		return
	}
	if req.Email == "" {
		response.BadRequest(w, "Email is required")
		return
	}
	if err := h.authService.ResendVerificationEmail(ctx, req.Email); err != nil {
		if err == appErrors.ErrUserAlreadyVerified {
			response.BadRequest(w, "This account is already verified. Please login.")
			return
		}
		response.Success(w, map[string]interface{}{
			"message": "If an unverified account exists for this email, a verification email has been sent.",
		}, "Verification email sent")
		return
	}
	response.Success(w, map[string]interface{}{
		"message": "Verification email sent. Please check your email to verify your account.",
	}, "Verification email sent successfully")
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Logout user
	if err := h.authService.Logout(ctx, userID); err != nil {
		response.InternalServerError(w, "Failed to logout")
		return
	}

	response.Success(w, map[string]interface{}{
		"message": "Logged out successfully",
	}, "Logout successful")
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req auth.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, constants.ErrInvalidRequestBody)
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "Refresh token is required")
		return
	}

	// Validate refresh token and generate new access token
	newAccessToken, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		statusCode := appErrors.GetHTTPStatus(err)
		message := appErrors.GetMessage(err)
		response.Error(w, statusCode, message)
		return
	}

	response.Success(w, map[string]interface{}{
		"access_token": newAccessToken,
	}, "Access token refreshed successfully")
}
