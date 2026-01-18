package service_test

import (
	"context"
	"errors"
	domainRepository "go-auth-service/internal/domain/repository"
	"go-auth-service/internal/domain/user"
	"go-auth-service/internal/service"
	"strings"
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
	CreateUserTxFunc     func(ctx context.Context, tx domainRepository.TransactionInterface, user *user.User) error
	BeginTxFunc          func(ctx context.Context) (domainRepository.TransactionInterface, error)
	CreateUserFunc       func(ctx context.Context, u *user.User) error
	FindUserByEmailFunc  func(ctx context.Context, email string) (*user.User, error)
	FindUserByIDFunc     func(ctx context.Context, id int64) (*user.User, error)
	UpdateUserFunc       func(ctx context.Context, u *user.User) error
	DeleteUserFunc       func(ctx context.Context, id int64) error
}

func (m *MockUserRepository) BeginTx(ctx context.Context) (domainRepository.TransactionInterface, error) {
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

func (m *MockUserRepository) CreateUserTx(ctx context.Context, tx domainRepository.TransactionInterface, u *user.User) error {
	if m.CreateUserTxFunc != nil {
		return m.CreateUserTxFunc(ctx, tx, u)
	}
	u.ID = 1 // Simulate auto-increment ID
	return nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, u)
	}
	u.ID = 1
	return nil
}

func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.FindUserByEmailFunc != nil {
		return m.FindUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id int64) (*user.User, error) {
	if m.FindUserByIDFunc != nil {
		return m.FindUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, u)
	}
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int64) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

// mockTransaction implements repository.TransactionInterface for testing
type mockTransaction struct {
	commitCalled   bool
	rollbackCalled bool
}

func (m *mockTransaction) Commit() error {
	m.commitCalled = true
	return nil
}

func (m *mockTransaction) Rollback() error {
	m.rollbackCalled = true
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
		CreateUserFunc: func(ctx context.Context, u *user.User) error {
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

// Test GetUserByID
func TestUserService_GetUserByID(t *testing.T) {
	expectedUser := &user.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id int64) (*user.User, error) {
			if id == 123 {
				return expectedUser, nil
			}
			return nil, errors.New("user not found")
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	// Test success
	u, err := userService.GetUserByID(context.Background(), 123)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if u.ID != 123 {
		t.Errorf("Expected user ID 123, got %d", u.ID)
	}

	// Test not found
	_, err = userService.GetUserByID(context.Background(), 999)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

// Test GetUserByEmail
func TestUserService_GetUserByEmail(t *testing.T) {
	expectedUser := &user.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
	}

	mockRepo := &MockUserRepository{
		FindUserByEmailFunc: func(ctx context.Context, email string) (*user.User, error) {
			if email == "test@example.com" {
				return expectedUser, nil
			}
			return nil, errors.New("user not found")
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	// Test success
	u, err := userService.GetUserByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if u.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", u.Email)
	}

	// Test not found
	_, err = userService.GetUserByEmail(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

// Test UpdateUser
func TestUserService_UpdateUser(t *testing.T) {
	var updatedUser *user.User

	mockRepo := &MockUserRepository{
		UpdateUserFunc: func(ctx context.Context, u *user.User) error {
			updatedUser = u
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	testUser := &user.User{
		ID:       123,
		Email:    "updated@example.com",
		Username: "updateduser",
		IsActive: true,
	}

	err := userService.UpdateUser(context.Background(), testUser)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Expected updated email, got %s", updatedUser.Email)
	}
}

// Test DeleteUser
func TestUserService_DeleteUser(t *testing.T) {
	var deletedID int64

	mockRepo := &MockUserRepository{
		DeleteUserFunc: func(ctx context.Context, id int64) error {
			deletedID = id
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	err := userService.DeleteUser(context.Background(), 123)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	if deletedID != 123 {
		t.Errorf("Expected to delete user 123, got %d", deletedID)
	}
}

// Test VerifyEmail
func TestUserService_VerifyEmail_Success(t *testing.T) {
	testUser := &user.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: false,
	}

	var updatedUser *user.User

	mockRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id int64) (*user.User, error) {
			return testUser, nil
		},
		UpdateUserFunc: func(ctx context.Context, u *user.User) error {
			updatedUser = u
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	err := userService.VerifyEmail(context.Background(), 123)
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	if updatedUser == nil {
		t.Fatal("User should have been updated")
	}

	if !updatedUser.IsActive {
		t.Error("User should be active after email verification")
	}
}

// Test VerifyEmail - Already Active
func TestUserService_VerifyEmail_AlreadyActive(t *testing.T) {
	testUser := &user.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
	}

	updateCalled := false

	mockRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id int64) (*user.User, error) {
			return testUser, nil
		},
		UpdateUserFunc: func(ctx context.Context, u *user.User) error {
			updateCalled = true
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	err := userService.VerifyEmail(context.Background(), 123)
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	if updateCalled {
		t.Error("UpdateUser should not be called for already active user")
	}
}

// Test VerifyEmail - User Not Found
func TestUserService_VerifyEmail_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id int64) (*user.User, error) {
			return nil, errors.New("user not found")
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	err := userService.VerifyEmail(context.Background(), 999)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

// Test VerifyPassword
func TestUserService_VerifyPassword(t *testing.T) {
	mockRepo := &MockUserRepository{}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	// Test correct password
	err := userService.VerifyPassword("hashed_password123", "password123")
	if err != nil {
		t.Errorf("VerifyPassword failed for correct password: %v", err)
	}

	// Test wrong password
	err = userService.VerifyPassword("hashed_password123", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}

// Test validateUsername edge cases
func TestUserService_ValidateUsername(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
	}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	tests := []struct {
		name      string
		username  string
		shouldErr bool
	}{
		{"valid username", "testuser", false},
		{"min length username", "abc", false},
		{"max length username", strings.Repeat("a", 50), false},
		{"too short username", "ab", true},
		{"too long username", strings.Repeat("a", 51), true},
		{"empty username", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := userService.RegisterUser(context.Background(), "test@example.com", tt.username, "password123")
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test validatePassword edge cases
func TestUserService_ValidatePassword(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
	}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	tests := []struct {
		name      string
		password  string
		shouldErr bool
	}{
		{"valid password", "password123", false},
		{"min length password", "12345678", false},
		{"max length password", strings.Repeat("a", 128), false},
		{"too short password", "1234567", true},
		{"too long password", strings.Repeat("a", 129), true},
		{"empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := userService.RegisterUser(context.Background(), "test@example.com", "testuser", tt.password)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test BeginTx
func TestUserService_BeginTx(t *testing.T) {
	expectedTx := &mockTransaction{}

	mockRepo := &MockUserRepository{
		BeginTxFunc: func(ctx context.Context) (domainRepository.TransactionInterface, error) {
			return expectedTx, nil
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	tx, err := userService.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	if tx == nil {
		t.Error("Expected transaction object, got nil")
	}
}

// Test RegisterUserTx
func TestUserService_RegisterUserTx(t *testing.T) {
	mockTx := &mockTransaction{}

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
		CreateUserTxFunc: func(ctx context.Context, tx domainRepository.TransactionInterface, u *user.User) error {
			u.ID = 123
			return nil
		},
	}

	mockHasher := &MockPasswordHasher{
		HashPasswordFunc: func(password string) (string, error) {
			return "hashed_" + password, nil
		},
	}

	userService := service.NewUserService(mockRepo, mockHasher)

	newUser, err := userService.RegisterUserTx(context.Background(), mockTx, "test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("RegisterUserTx failed: %v", err)
	}

	if newUser.ID != 123 {
		t.Errorf("Expected user ID 123, got %d", newUser.ID)
	}

	if newUser.IsActive {
		t.Error("User should be inactive initially")
	}
}

// Test RegisterUser with username already exists
func TestUserService_RegisterUser_UsernameExists(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return true, nil // Username exists
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUser(context.Background(), "new@example.com", "existinguser", "password123")
	if err == nil {
		t.Error("Expected error for existing username")
	}
}

// Test RegisterUser with hashing failure
func TestUserService_RegisterUser_HashingFails(t *testing.T) {
	expectedErr := errors.New("hashing failed")

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		HashPasswordFunc: func(password string) (string, error) {
			return "", expectedErr
		},
	}

	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUser(context.Background(), "test@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error from hashing failure")
	}
}

// Test RegisterUser with CreateUser failure
func TestUserService_RegisterUser_CreateUserFails(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
		CreateUserFunc: func(ctx context.Context, u *user.User) error {
			return expectedErr
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUser(context.Background(), "test@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error from CreateUser failure")
	}
}

// Test RegisterUserTx with email exists error
func TestUserService_RegisterUserTx_EmailExists(t *testing.T) {
	mockTx := &mockTransaction{}

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email exists
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUserTx(context.Background(), mockTx, "existing@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error for existing email")
	}
}

// Test RegisterUserTx with username exists error
func TestUserService_RegisterUserTx_UsernameExists(t *testing.T) {
	mockTx := &mockTransaction{}

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return true, nil // Username exists
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUserTx(context.Background(), mockTx, "test@example.com", "existinguser", "password123")
	if err == nil {
		t.Error("Expected error for existing username")
	}
}

// Test RegisterUserTx with hashing failure
func TestUserService_RegisterUserTx_HashingFails(t *testing.T) {
	mockTx := &mockTransaction{}
	expectedErr := errors.New("hashing failed")

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
	}

	mockHasher := &MockPasswordHasher{
		HashPasswordFunc: func(password string) (string, error) {
			return "", expectedErr
		},
	}

	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUserTx(context.Background(), mockTx, "test@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error from hashing failure")
	}
}

// Test RegisterUserTx with CreateUserTx failure
func TestUserService_RegisterUserTx_CreateUserTxFails(t *testing.T) {
	mockTx := &mockTransaction{}
	expectedErr := errors.New("database error")

	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
		CreateUserTxFunc: func(ctx context.Context, tx domainRepository.TransactionInterface, u *user.User) error {
			return expectedErr
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	_, err := userService.RegisterUserTx(context.Background(), mockTx, "test@example.com", "testuser", "password123")
	if err == nil {
		t.Error("Expected error from CreateUserTx failure")
	}
}

// Test VerifyEmail with UpdateUser failure
func TestUserService_VerifyEmail_UpdateUserFails(t *testing.T) {
	expectedErr := errors.New("update failed")

	testUser := &user.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: false,
	}

	mockRepo := &MockUserRepository{
		FindUserByIDFunc: func(ctx context.Context, id int64) (*user.User, error) {
			return testUser, nil
		},
		UpdateUserFunc: func(ctx context.Context, u *user.User) error {
			return expectedErr
		},
	}

	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	err := userService.VerifyEmail(context.Background(), 123)
	if err == nil {
		t.Error("Expected error from UpdateUser failure")
	}
}

// Test validateEmail with various error cases
func TestUserService_ValidateEmail_AdditionalCases(t *testing.T) {
	mockRepo := &MockUserRepository{
		ExistsByEmailFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		ExistsByUsernameFunc: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
	}
	mockHasher := &MockPasswordHasher{}
	userService := service.NewUserService(mockRepo, mockHasher)

	tests := []struct {
		name      string
		email     string
		shouldErr bool
	}{
		{"too long email", strings.Repeat("a", 91) + "@example.com", true},
		{"invalid format no domain", "test@", true},
		{"invalid format no @", "testexample.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := userService.RegisterUser(context.Background(), tt.email, "testuser", "password123")
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}
