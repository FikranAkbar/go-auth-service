package service

import (
	"context"
	"database/sql"
	"go-auth-service/internal/domain/user"
	"go-auth-service/internal/security"
	appErrors "go-auth-service/pkg/errors"
	"regexp"
	"strings"
)

type UserService struct {
	userRepository user.RepositoryInterface
	passwordHasher *security.PasswordHasher
}

func NewUserService(userRepo user.RepositoryInterface, passwordHasher *security.PasswordHasher) *UserService {
	return &UserService{
		userRepository: userRepo,
		passwordHasher: passwordHasher,
	}
}

// BeginTx starts a new database transaction
func (s *UserService) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.userRepository.BeginTx(ctx)
}

// RegisterUser creates a new user account with validation
func (s *UserService) RegisterUser(ctx context.Context, email, username, password string) (*user.User, error) {
	// Validate email
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Validate username
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Check if email already exists
	emailExists, err := s.userRepository.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to check email existence")
	}
	if emailExists {
		return nil, appErrors.ErrEmailAlreadyExists
	}

	// Check if username already exists
	usernameExists, err := s.userRepository.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to check username existence")
	}
	if usernameExists {
		return nil, appErrors.ErrUsernameAlreadyExists
	}

	// Hash the password
	hashedPassword, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, appErrors.Wrap(appErrors.ErrHashingFailed, err.Error())
	}

	// Create user entity (inactive until email verification)
	newUser := &user.User{
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Username:     strings.TrimSpace(username),
		PasswordHash: hashedPassword,
		IsActive:     false, // User must verify email first
	}

	// Save to database
	if err := s.userRepository.CreateUser(ctx, newUser); err != nil {
		return nil, appErrors.Wrap(err, "failed to create user")
	}

	return newUser, nil
}

// RegisterUserTx creates a new user account within a transaction
// This allows the caller to rollback if subsequent operations fail
func (s *UserService) RegisterUserTx(ctx context.Context, tx *sql.Tx, email, username, password string) (*user.User, error) {
	// Validate email
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Validate username
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Check if email already exists
	emailExists, err := s.userRepository.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to check email existence")
	}
	if emailExists {
		return nil, appErrors.ErrEmailAlreadyExists
	}

	// Check if username already exists
	usernameExists, err := s.userRepository.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, appErrors.Wrap(err, "failed to check username existence")
	}
	if usernameExists {
		return nil, appErrors.ErrUsernameAlreadyExists
	}

	// Hash the password
	hashedPassword, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, appErrors.Wrap(appErrors.ErrHashingFailed, err.Error())
	}

	// Create user entity (inactive until email verification)
	newUser := &user.User{
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Username:     strings.TrimSpace(username),
		PasswordHash: hashedPassword,
		IsActive:     false, // User must verify email first
	}

	// Save to database within transaction
	if err := s.userRepository.CreateUserTx(ctx, tx, newUser); err != nil {
		return nil, appErrors.Wrap(err, "failed to create user")
	}

	return newUser, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	return s.userRepository.FindUserByID(ctx, userID)
}

// GetUserByEmail retrieves a user by their email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.userRepository.FindUserByEmail(ctx, email)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, u *user.User) error {
	return s.userRepository.UpdateUser(ctx, u)
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userRepository.DeleteUser(ctx, userID)
}

// VerifyEmail activates a user account after email verification
func (s *UserService) VerifyEmail(ctx context.Context, userID int64) error {
	// Get user
	u, err := s.userRepository.FindUserByID(ctx, userID)
	if err != nil {
		return appErrors.Wrap(err, "failed to find user")
	}

	// Check if already active
	if u.IsActive {
		return nil // Already verified, no action needed
	}

	// Activate user
	u.IsActive = true
	if err := s.userRepository.UpdateUser(ctx, u); err != nil {
		return appErrors.Wrap(err, "failed to activate user")
	}

	return nil
}

// validateEmail validates email format and length
func (s *UserService) validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return appErrors.ErrEmailRequired
	}

	if len(email) > 100 {
		return appErrors.ErrEmailTooLong
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return appErrors.ErrInvalidEmail
	}

	return nil
}

// validateUsername validates username length and format
func (s *UserService) validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return appErrors.ErrUsernameRequired
	}

	if len(username) > 50 {
		return appErrors.ErrUsernameTooLong
	}

	if len(username) < 3 {
		return appErrors.ErrUsernameTooShort
	}

	return nil
}

// validatePassword validates password strength
func (s *UserService) validatePassword(password string) error {
	if password == "" {
		return appErrors.ErrPasswordRequired
	}

	if len(password) < 8 {
		return appErrors.ErrPasswordTooShort
	}

	if len(password) > 128 {
		return appErrors.ErrPasswordTooLong
	}

	return nil
}

// Compile-time check to ensure UserService implements user.ServiceInterface
var _ user.ServiceInterface = (*UserService)(nil)
