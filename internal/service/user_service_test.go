package service_test

import (
	"context"
	"database/sql"
	"errors"
	"go-auth-service/internal/domain/user"
	"go-auth-service/internal/service"
	"testing"
)

// MockPasswordHasher is a mock implementation of PasswordHasherInterface
type MockPasswordHasher struct {
	HashPasswordFunc   func(password string) (string, error)
	VerifyPasswordFunc func(hashedPassword, password string) error
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	if m.HashPasswordFunc != nil {
		return m.HashPasswordFunc(password)
	}
	return "hashed_" + password, nil
}

func (m *MockPasswordHasher) VerifyPassword(hashedPassword, password string) error {
	if m.VerifyPasswordFunc != nil {
		return m.VerifyPasswordFunc(hashedPassword, password)
	}
	if hashedPassword == "hashed_"+password {
		return nil
	}
	return errors.New("password mismatch")
}

// MockUserRepository is a mock implementation of RepositoryInterface
type MockUserRepository struct {
	ExistsByEmailFunc    func(ctx context.Context, email string) (bool, error)
	ExistsByUsernameFunc func(ctx context.Context, username string) (bool, error)
	CreateUserTxFunc     func(ctx context.Context, tx *sql.Tx, user *user.User) error
	BeginTxFunc          func(ctx context.Context) (*sql.Tx, error)
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsByEmailFunc != nil {
		return m.ExistsByEmailFunc(ctx, email)
	}
	return false, nil
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.ExistsByUsernameFunc != nil {
		return m.ExistsByUsernameFunc(ctx, username)
	}
	return false, nil
}

func (m *MockUserRepository) CreateUserTx(ctx context.Context, tx *sql.Tx, u *user.User) error {
	if m.CreateUserTxFunc != nil {
		return m.CreateUserTxFunc(ctx, tx, u)
	}
	u.ID = 1 // Simulate auto-increment ID
	return nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	return nil
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, nil
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id int64) (*user.User, error) {
	return nil, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int64) error {
	return nil
}

// TestUserService_ValidateEmail demonstrates unit testing with mocks
func TestUserService_ValidateEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email - missing @",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "invalid email - empty",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use reflection or internal testing helper to access validateEmail
			// For demonstration, we test RegisterUser which calls validateEmail
			ctx := context.Background()
			_, err := userService.RegisterUser(ctx, tt.email, "validuser", "validpass123")

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestUserService_RegisterUser_EmailAlreadyExists demonstrates mocking repository behavior
func TestUserService_RegisterUser_EmailAlreadyExists(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Simulate email already exists
		},
	}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	ctx := context.Background()
	_, err := userService.RegisterUser(ctx, "existing@example.com", "newuser", "password123")

	if err == nil {
		t.Error("Expected error for existing email, got nil")
		return
	}

	// The actual error message from the service
	expectedMsg := "email already in use"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s' error, got: %v", expectedMsg, err)
	}
}

// TestUserService_RegisterUser_Success demonstrates successful registration
func TestUserService_RegisterUser_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email doesn't exist
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil // Username doesn't exist
		},
		CreateUserTxFunc: func(ctx context.Context, tx *sql.Tx, u *user.User) error {
			u.ID = 123 // Simulate database auto-increment
			return nil
		},
	}
	mockHasher := &MockPasswordHasher{
		HashPasswordFunc: func(password string) (string, error) {
			return "hashed_password", nil
		},
	}

	userService := service.NewUserService(mockRepo, mockHasher)

	ctx := context.Background()
	newUser, err := userService.RegisterUser(ctx, "new@example.com", "newuser", "Password123!")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if newUser == nil {
		t.Fatal("Expected user to be created, got nil")
	}

	if newUser.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got: %s", newUser.Email)
	}

	if newUser.PasswordHash != "hashed_password" {
		t.Errorf("Expected password to be hashed, got: %s", newUser.PasswordHash)
	}

	if newUser.IsActive {
		t.Error("Expected user to be inactive until email verification")
	}
}
