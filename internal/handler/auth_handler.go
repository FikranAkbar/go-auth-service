package handler

import (
	"context"
	"encoding/json"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/security"
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
		response.BadRequest(w, "Invalid request body")
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
	response.InternalServerError(w, "Login endpoint not yet implemented")
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
	accessToken, err := h.jwtManager.GenerateAccessToken(verifiedUser.ID, verifiedUser.Email, verifiedUser.Username)
	if err != nil {
		response.InternalServerError(w, "Failed to generate access token")
		return
	}
	refreshToken, err := h.jwtManager.GenerateRefreshToken(verifiedUser.ID, verifiedUser.Email, verifiedUser.Username)
	if err != nil {
		response.InternalServerError(w, "Failed to generate refresh token")
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
		response.BadRequest(w, "Invalid request body")
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
	response.InternalServerError(w, "Logout endpoint not yet implemented")
}
