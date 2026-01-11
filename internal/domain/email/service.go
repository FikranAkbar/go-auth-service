package email

// ServiceInterface defines the contract for email operations
type ServiceInterface interface {
	SendVerificationEmail(email, verificationToken string) error
}
