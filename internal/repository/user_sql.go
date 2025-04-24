package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/NiskuT/cross-api/internal/domain/aggregate"
	repo "github.com/NiskuT/cross-api/internal/domain/repository"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateUser is returned when a user with the same ID or email already exists
	ErrDuplicateUser = errors.New("user with this ID or email already exists")
)

// SQLUserRepository is an implementation of the UserRepository interface that uses SQL
type SQLUserRepository struct {
	db *sql.DB
}

// NewSQLUserRepository creates a new SQLUserRepository
func NewSQLUserRepository(db *sql.DB) repo.UserRepository {
	return &SQLUserRepository{
		db: db,
	}
}

// User is an internal representation of a user for DB operations
type User struct {
	ID           int32
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
	Role         string
}

// GetUser retrieves a user by ID
func (r *SQLUserRepository) GetUser(ctx context.Context, id int32) (*aggregate.User, error) {
	query := `
		SELECT id, email, first_name, last_name, password_hash, role
		FROM users
		WHERE id = ?
	`

	var user User
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	userAggregate := aggregate.NewUser()
	userAggregate.SetID(user.ID)
	userAggregate.SetEmail(user.Email)
	userAggregate.SetFirstName(user.FirstName)
	userAggregate.SetLastName(user.LastName)
	userAggregate.SetPasswordHash(user.PasswordHash)
	userAggregate.SetRole(user.Role)

	return userAggregate, nil
}

// CreateUser creates a new user
func (r *SQLUserRepository) CreateUser(ctx context.Context, user *aggregate.User) error {
	query := `
		INSERT INTO users (email, first_name, last_name, password_hash, role)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.GetEmail(),
		user.GetFirstName(),
		user.GetLastName(),
		user.GetPasswordHash(),
		user.GetRole(),
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return ErrDuplicateUser
		}
		return err
	}

	// Get the auto-incremented ID
	if id, err := result.LastInsertId(); err == nil {
		user.SetID(int32(id))
	}

	return nil
}

// UpdateUser updates an existing user
func (r *SQLUserRepository) UpdateUser(ctx context.Context, user *aggregate.User) error {
	query := `
		UPDATE users
		SET email = ?, first_name = ?, last_name = ?, password_hash = ?, role = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.GetEmail(),
		user.GetFirstName(),
		user.GetLastName(),
		user.GetPasswordHash(),
		user.GetRole(),
		user.GetID(),
	)

	if err != nil {
		// Check for duplicate key error (e.g., unique email constraint)
		if isDuplicateKeyError(err) {
			return ErrDuplicateUser
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes a user by ID
func (r *SQLUserRepository) DeleteUser(ctx context.Context, id int32) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (r *SQLUserRepository) GetUserByEmail(ctx context.Context, email string) (*aggregate.User, error) {
	query := `
		SELECT id, email, first_name, last_name, password_hash, role
		FROM users
		WHERE email = ?
	`

	var user User
	row := r.db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	userAggregate := aggregate.NewUser()
	userAggregate.SetID(user.ID)
	userAggregate.SetEmail(user.Email)
	userAggregate.SetFirstName(user.FirstName)
	userAggregate.SetLastName(user.LastName)
	userAggregate.SetPasswordHash(user.PasswordHash)
	userAggregate.SetRole(user.Role)

	return userAggregate, nil
}
