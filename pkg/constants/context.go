package constants

// Context Keys
type ContextKey string

const (
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyUserEmail ContextKey = "user_email"
	ContextKeyUserRole  ContextKey = "user_role"
	ContextKeySessionID ContextKey = "session_id"
	ContextKeyTraceID   ContextKey = "trace_id"
)

// String returns the string representation of the context key
func (c ContextKey) String() string {
	return string(c)
}
