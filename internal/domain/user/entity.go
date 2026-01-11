package user
import "time"
// User represents a user entity in the system
type User struct {
	ID        string
	Email     string
	Password  string // Hashed password
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
// TODO: Add user-related value objects or methods here as needed
// Example:
// func (u *User) IsActive() bool
// func (u *User) Validate() error
