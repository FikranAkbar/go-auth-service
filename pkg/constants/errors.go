package constants

// Error Messages - General
const (
	ErrInvalidRequest       = "Invalid request"
	ErrInvalidRequestBody   = "Invalid request body"
	ErrFailedToEncodeJSON   = "Failed to encode JSON response"
	ErrFailedToDecodeJSON   = "Failed to decode JSON request"
	ErrMissingRequiredField = "Missing required field"
)

// Error Messages - Authentication
const (
	ErrInvalidCredentials     = "Invalid credentials"
	ErrInvalidToken           = "Invalid token"
	ErrExpiredToken           = "Token has expired"
	ErrMissingToken           = "Missing authentication token"
	ErrUnauthorizedAccess     = "Unauthorized access"
	ErrInsufficientPermission = "Insufficient permission"
)

// Error Messages - User
const (
	ErrUserNotFound      = "User not found"
	ErrUserAlreadyExists = "User already exists"
	ErrEmailAlreadyUsed  = "Email already in use"
	ErrInvalidEmail      = "Invalid email format"
	ErrInvalidPassword   = "Invalid password"
	ErrPasswordTooShort  = "Password must be at least 8 characters"
	ErrPasswordMismatch  = "Passwords do not match"
)

// Error Messages - Database
const (
	ErrDatabaseConnection  = "Database connection error"
	ErrDatabaseQuery       = "Database query error"
	ErrDatabaseTransaction = "Database transaction error"
	ErrRecordNotFound      = "Record not found"
	ErrRecordAlreadyExists = "Record already exists"
)

// Error Messages - Redis/Cache
const (
	ErrCacheConnection = "Cache connection error"
	ErrCacheNotFound   = "Cache entry not found"
	ErrCacheSet        = "Failed to set cache"
	ErrCacheGet        = "Failed to get cache"
	ErrCacheDelete     = "Failed to delete cache"
)

// Error Messages - Validation
const (
	ErrValidationFailed  = "Validation failed"
	ErrInvalidFormat     = "Invalid format"
	ErrInvalidValue      = "Invalid value"
	ErrValueTooLong      = "Value exceeds maximum length"
	ErrValueTooShort     = "Value below minimum length"
	ErrInvalidDateFormat = "Invalid date format"
	ErrInvalidUUID       = "Invalid UUID format"
)
