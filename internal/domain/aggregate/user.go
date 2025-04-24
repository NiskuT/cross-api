package aggregate

import (
	"strings"

	"github.com/NiskuT/cross-api/internal/domain/entity"
)

// User is the aggregate root for user domain
type User struct {
	user *entity.User
}

// NewUser creates a new user aggregate
func NewUser() *User {
	return &User{user: &entity.User{}}
}

// GetID returns the user ID
func (u *User) GetID() int32 {
	return u.user.ID
}

// GetEmail returns the user email
func (u *User) GetEmail() string {
	return u.user.Email
}

// GetFirstName returns the user first name
func (u *User) GetFirstName() string {
	return u.user.FirstName
}

// GetLastName returns the user last name
func (u *User) GetLastName() string {
	return u.user.LastName
}

// GetPasswordHash returns the user password hash
func (u *User) GetPasswordHash() string {
	return u.user.PasswordHash
}

// GetRole returns the user role
func (u *User) GetRole() string {
	return u.user.Role
}

// SetID sets the user ID
func (u *User) SetID(id int32) {
	u.user.ID = id
}

// SetEmail sets the user email
func (u *User) SetEmail(email string) {
	u.user.Email = email
}

// SetFirstName sets the user first name
func (u *User) SetFirstName(firstName string) {
	u.user.FirstName = firstName
}

// SetLastName sets the user last name
func (u *User) SetLastName(lastName string) {
	u.user.LastName = lastName
}

// SetPasswordHash sets the user password hash
func (u *User) SetPasswordHash(passwordHash string) {
	u.user.PasswordHash = passwordHash
}

// SetRole sets the user role
func (u *User) SetRole(role string) {
	u.user.Role = role
}

func (u *User) AddRole(newRole string) {
	// Check if the user already has this role
	roles := strings.Split(u.GetRole(), ",")
	for _, role := range roles {
		if strings.TrimSpace(role) == newRole {
			// User already has this role
			return
		}
	}

	// Add the new role
	var updatedRoles string
	if u.GetRole() == "" {
		updatedRoles = newRole
	} else {
		updatedRoles = u.GetRole() + "," + newRole
	}

	// Update the user
	u.SetRole(updatedRoles)
}
