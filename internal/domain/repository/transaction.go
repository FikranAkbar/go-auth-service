package repository

// TransactionInterface defines the contract for database transactions
// This allows us to mock transactions in tests across all repositories
type TransactionInterface interface {
	Commit() error
	Rollback() error
}
