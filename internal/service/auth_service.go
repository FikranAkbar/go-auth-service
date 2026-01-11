package service

import (
	"context"
	"go-auth-service/internal/domain/auth"
	"go-auth-service/internal/domain/email"
	"go-auth-service/internal/domain/user"
	"go-auth-service/internal/security"
	appErrors "go-auth-service/pkg/errors"
	"go-auth-service/pkg/logger"
)

type AuthService struct {
	userService  user.ServiceInterface
	emailService email.ServiceInterface
	jwtManager   *security.JWTManager
}

func NewAuthService(userService user.ServiceInterface, emailService email.ServiceInterface, jwtManager *security.JWTManager) *AuthService {
	return &AuthService{
		userService:  userService,
		emailService: emailService,
		jwtManager:   jwtManager,
	}
}

// RegisterUser handles the complete user registration flow with transaction support
func (s *AuthService) RegisterUser(ctx context.Context, email, username, password string) (int64, string, error) {
	// Begin database transaction
	tx, err := s.userService.BeginTx(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback() // Auto-rollback if not committed

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
		return err
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(foundUser.Email, verificationToken); err != nil {
		logger.Errorf("Failed to send verification email to user %d (%s): %v", foundUser.ID, foundUser.Email, err)
		return err
	}

	return nil
}

// Compile-time check to ensure AuthService implements auth.ServiceInterface
var _ auth.ServiceInterface = (*AuthService)(nil)
