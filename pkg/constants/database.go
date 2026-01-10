package constants

// Database Tables
const (
	TableUsers    = "users"
	TableSessions = "sessions"
	TableTokens   = "tokens"
)

// Database Columns - Users Table
const (
	ColUserID           = "id"
	ColUserEmail        = "email"
	ColUserPasswordHash = "password_hash"
	ColUserName         = "name"
	ColUserCreatedAt    = "created_at"
	ColUserUpdatedAt    = "updated_at"
	ColUserDeletedAt    = "deleted_at"
)

// Database Query Errors
const (
	DBErrDuplicateKey = "duplicate key value violates unique constraint"
	DBErrNoRows       = "no rows in result set"
	DBErrConnRefused  = "connection refused"
)
