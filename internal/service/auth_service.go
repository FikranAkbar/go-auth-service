package service

import (
	"context"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/domain/email"
	domainRepository "go-auth-service/internal/domain/repository"
	domainSecurity "go-auth-service/internal/domain/security"
	"go-auth-service/internal/domain/user"
	appErrors "go-auth-service/pkg/errors"
	"go-auth-service/pkg/logger"
	"time"
)

type AuthService struct {
	userService  user.ServiceInterface
	emailService email.ServiceInterface
	jwtManager   domainSecurity.JWTManagerInterface
	tokenRepo    domainRepository.TokenRepositoryInterface
}

func NewAuthService(
	userService user.ServiceInterface,
	emailService email.ServiceInterface,
	jwtManager domainSecurity.JWTManagerInterface,
	tokenRepo domainRepository.TokenRepositoryInterface,
) *AuthService {
	return &AuthService{
		userService:  userService,
		emailService: emailService,
		jwtManager:   jwtManager,
		tokenRepo:    tokenRepo,
	}
}

// RegisterUser handles the complete user registration flow with transaction support
func (s *AuthService) RegisterUser(ctx context.Context, email, username, password string) (int64, string, error) {
	// Begin database transaction
	tx, err := s.userService.BeginTx(ctx)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }() // Auto-rollback if not committed

	// Register user within transaction
	newUser, err := s.userService.RegisterUserTx(ctx, tx, email, username, password)
	if err != nil {
		return 0, "", err
	}

	// Generate verification token (24 hour expiry)
	verificationToken, err := s.jwtManager.GenerateVerificationToken(newUser.ID, newUser.Email)
	if err != nil {
		logger.Errorf("Failed to generate verification token for user %d: %v", newUser.ID, err)
		return 0, "", err
	}

	// Commit transaction FIRST - only send email if user is successfully saved
	if err := tx.Commit(); err != nil {
		logger.Errorf("Failed to commit transaction for user %d: %v", newUser.ID, err)
		return 0, "", err
	}

	// Send verification email AFTER successful commit
	// If this fails, user is already saved - they can use resend verification
	if err := s.emailService.SendVerificationEmail(newUser.Email, verificationToken); err != nil {
		logger.Errorf("Failed to send verification email to user %d (%s): %v", newUser.ID, newUser.Email, err)
		// User is saved but email failed - return error so handler can inform user
		return 0, "", err
	}

	return newUser.ID, newUser.Email, nil
}

// Login authenticates a user and returns user data with tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*user.User, string, string, error) {
	// Get user by email
	foundUser, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Warnf("Login attempt with non-existent email: %s", email)
		return nil, "", "", appErrors.ErrInvalidCredentials
	}

	// Check if user is active (email verified)
	if !foundUser.IsActive {
		logger.Warnf("Login attempt with unverified email: %s (user ID: %d)", email, foundUser.ID)
		return nil, "", "", appErrors.ErrUserNotVerified
	}

	// Verify password
	if err := s.userService.VerifyPassword(foundUser.PasswordHash, password); err != nil {
		logger.Warnf("Login attempt with invalid password for email: %s (user ID: %d)", email, foundUser.ID)
		return nil, "", "", appErrors.ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(foundUser.ID, foundUser.Email, foundUser.Username)
	if err != nil {
		logger.Errorf("Failed to generate access token for user %d: %v", foundUser.ID, err)
		return nil, "", "", err
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(foundUser.ID, foundUser.Email, foundUser.Username)
	if err != nil {
		logger.Errorf("Failed to generate refresh token for user %d: %v", foundUser.ID, err)
		return nil, "", "", err
	}

	// Store refresh token in Redis
	// Calculate expiration time from config
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour) // Default 7 days, should match JWT config
	if err := s.tokenRepo.StoreRefreshToken(ctx, foundUser.ID, refreshToken, refreshTokenExpiry); err != nil {
		logger.Errorf("Failed to store refresh token in Redis for user %d: %v", foundUser.ID, err)
		return nil, "", "", err
	}

	return foundUser, accessToken, refreshToken, nil
}

// VerifyEmail verifies a user's email and activates their account
func (s *AuthService) VerifyEmail(ctx context.Context, userID int64) (*user.User, error) {
	// Activate user account
	if err := s.userService.VerifyEmail(ctx, userID); err != nil {
		return nil, err
	}

	// Get updated user
	verifiedUser, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return verifiedUser, nil
}

// GenerateAndStoreTokens generates access and refresh tokens, stores refresh token in Redis
func (s *AuthService) GenerateAndStoreTokens(ctx context.Context, u *user.User) (accessToken, refreshToken string, err error) {
	// Generate access token
	accessToken, err = s.jwtManager.GenerateAccessToken(u.ID, u.Email, u.Username)
	if err != nil {
		logger.Errorf("Failed to generate access token for user %d: %v", u.ID, err)
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err = s.jwtManager.GenerateRefreshToken(u.ID, u.Email, u.Username)
	if err != nil {
		logger.Errorf("Failed to generate refresh token for user %d: %v", u.ID, err)
		return "", "", err
	}

	// Store refresh token in Redis
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour) // Default 7 days, should match JWT config
	if err = s.tokenRepo.StoreRefreshToken(ctx, u.ID, refreshToken, refreshTokenExpiry); err != nil {
		logger.Errorf("Failed to store refresh token in Redis for user %d: %v", u.ID, err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ResendVerificationEmail resends verification email to unverified user
func (s *AuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	// Get user by email
	foundUser, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Check if user is already verified
	if foundUser.IsActive {
		return appErrors.ErrUserAlreadyVerified
	}

	// Generate new verification token (24 hour expiry)
	verificationToken, err := s.jwtManager.GenerateVerificationToken(foundUser.ID, foundUser.Email)
	if err != nil {
		logger.Errorf("Failed to generate verification token for user %d: %v", foundUser.ID, err)
		return appErrors.Internal(appErrors.ErrTokenGenFailed, "Failed to generate verification token")
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(foundUser.Email, verificationToken); err != nil {
		logger.Errorf("Failed to send verification email to user %d (%s): %v", foundUser.ID, foundUser.Email, err)
		return appErrors.Internal(appErrors.ErrEmailSendFailed, "Failed to send verification email. Please try again later.")
	}

	return nil
}

// RefreshToken validates refresh token and generates new access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate refresh token JWT
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		logger.Warnf("Invalid refresh token: %v", err)
		return "", appErrors.ErrInvalidToken
	}

	// Check token type
	if claims.Type != domainSecurity.RefreshToken {
		logger.Warnf("Token is not a refresh token, got type: %s", claims.Type)
		return "", appErrors.ErrInvalidToken
	}

	// Validate refresh token against Redis
	isValid, err := s.tokenRepo.ValidateRefreshToken(ctx, claims.UserID, refreshToken)
	if err != nil || !isValid {
		logger.Warnf("Refresh token validation failed for user %d: %v", claims.UserID, err)
		return "", appErrors.ErrInvalidToken
	}

	// Generate new access token
	newAccessToken, err := s.jwtManager.GenerateAccessToken(claims.UserID, claims.Email, claims.Username)
	if err != nil {
		logger.Errorf("Failed to generate new access token for user %d: %v", claims.UserID, err)
		return "", err
	}

	logger.Infof("Refresh token validated and new access token generated for user %d", claims.UserID)
	return newAccessToken, nil
}

// Logout invalidates user's refresh token
func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	// Delete refresh token from Redis
	if err := s.tokenRepo.DeleteRefreshToken(ctx, userID); err != nil {
		logger.Errorf("Failed to logout user %d: %v", userID, err)
		return err
	}

	logger.Infof("User %d logged out successfully", userID)
	return nil
}

// Compile-time check to ensure AuthService implements auth.ServiceInterface
var _ auth.ServiceInterface = (*AuthService)(nil)
