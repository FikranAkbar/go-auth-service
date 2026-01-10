package constants

// Validation Constraints
const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MinNameLength     = 2
	MaxNameLength     = 100
	MaxEmailLength    = 255
)

// Validation Messages
const (
	ValidMsgEmailRequired    = "Email is required"
	ValidMsgEmailInvalid     = "Email format is invalid"
	ValidMsgPasswordRequired = "Password is required"
	ValidMsgPasswordTooShort = "Password must be at least 8 characters long"
	ValidMsgNameRequired     = "Name is required"
	ValidMsgNameTooShort     = "Name must be at least 2 characters long"
)

// Regular Expression Patterns
const (
	RegexEmail    = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	RegexPassword = `^.{8,}$`
	RegexUUID     = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
)
