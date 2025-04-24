package entity

// User represents a user entity
type User struct {
	ID           int32
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	Role         string
}
