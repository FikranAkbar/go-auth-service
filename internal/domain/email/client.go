package email

// ClientInterface defines the contract for sending emails
type ClientInterface interface {
	SendEmail(to, subject, body string) error
}
