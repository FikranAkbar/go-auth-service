package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP context
type AppError struct {
	Err        error
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Predefined error types
var (
	// Validation errors
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrEmailRequired    = errors.New("email is required")
	ErrUsernameRequired = errors.New("username is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 128 characters")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.New("username must not exceed 50 characters")
	ErrEmailTooLong     = errors.New("email must not exceed 100 characters")

	// Business logic errors
	ErrEmailAlreadyExists    = errors.New("email already in use")
	ErrUsernameAlreadyExists = errors.New("username already in use")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")

	// Authentication errors
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrMissingToken      = errors.New("missing authentication token")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInsufficientPerms = errors.New("insufficient permissions")

	// System errors
	ErrInternal       = errors.New("internal server error")
	ErrDatabaseError  = errors.New("database error")
	ErrHashingFailed  = errors.New("password hashing failed")
	ErrTokenGenFailed = errors.New("token generation failed")
)

// NewAppError creates a new application error with HTTP status code
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
	}
}

// BadRequest creates a 400 error
func BadRequest(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusBadRequest)
}

// Conflict creates a 409 error
func Conflict(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusConflict)
}

// NotFound creates a 404 error
func NotFound(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusNotFound)
}

// Unauthorized creates a 401 error
func Unauthorized(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusUnauthorized)
}

// Internal creates a 500 error
func Internal(err error, message string) *AppError {
	return NewAppError(err, message, http.StatusInternalServerError)
}

// GetHTTPStatus returns the appropriate HTTP status code for an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	// Map known errors to status codes
	switch {
	case errors.Is(err, ErrInvalidEmail),
		errors.Is(err, ErrEmailRequired),
		errors.Is(err, ErrUsernameRequired),
		errors.Is(err, ErrPasswordRequired),
		errors.Is(err, ErrPasswordTooShort),
		errors.Is(err, ErrPasswordTooLong),
		errors.Is(err, ErrUsernameTooShort),
		errors.Is(err, ErrUsernameTooLong),
		errors.Is(err, ErrEmailTooLong):
		return http.StatusBadRequest

	case errors.Is(err, ErrEmailAlreadyExists),
		errors.Is(err, ErrUsernameAlreadyExists):
		return http.StatusConflict

	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound

	case errors.Is(err, ErrInvalidCredentials),
		errors.Is(err, ErrInvalidToken),
		errors.Is(err, ErrExpiredToken),
		errors.Is(err, ErrMissingToken),
		errors.Is(err, ErrUnauthorized),
		errors.Is(err, ErrInsufficientPerms):
		return http.StatusUnauthorized

	default:
		return http.StatusInternalServerError
	}
}

// GetMessage returns a user-friendly error message
func GetMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) && appErr.Message != "" {
		return appErr.Message
	}
	return err.Error()
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
