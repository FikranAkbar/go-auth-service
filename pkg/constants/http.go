package constants

// HTTP Headers
const (
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderXRequestID    = "X-Request-ID"
	HeaderXRealIP       = "X-Real-IP"
	HeaderUserAgent     = "User-Agent"
)

// Content Types
const (
	ContentTypeJSON           = "application/json"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipartForm  = "multipart/form-data"
	ContentTypeTextPlain      = "text/plain"
)

// HTTP Methods
const (
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodPATCH  = "PATCH"
	MethodDELETE = "DELETE"
)

// Common HTTP Status Messages
const (
	MsgBadRequest          = "Bad request"
	MsgUnauthorized        = "Unauthorized"
	MsgForbidden           = "Forbidden"
	MsgNotFound            = "Not found"
	MsgInternalServerError = "Internal server error"
	MsgSuccess             = "Success"
	MsgCreated             = "Created successfully"
	MsgUpdated             = "Updated successfully"
	MsgDeleted             = "Deleted successfully"
)
