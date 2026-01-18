package repository

import (
	"context"
	"database/sql"
	"errors"
	domainRepository "go-auth-service/internal/domain/repository"
	"go-auth-service/internal/domain/user"
	appErrors "go-auth-service/pkg/errors"

	"github.com/redis/go-redis/v9"
)

type UserRepository struct {
	Db          *sql.DB
	RedisClient *redis.Client
}

func NewUserRepository(db *sql.DB, redisClient *redis.Client) *UserRepository {
	return &UserRepository{Db: db, RedisClient: redisClient}
}

// BeginTx starts a new database transaction
func (r *UserRepository) BeginTx(ctx context.Context) (domainRepository.TransactionInterface, error) {
	return r.Db.BeginTx(ctx, nil)
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.Db.QueryRowContext(
		ctx,
		query,
		u.Email,
		u.Username,
		u.PasswordHash,
		u.IsActive,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// CreateUserTx creates a new user within a transaction
func (r *UserRepository) CreateUserTx(ctx context.Context, tx domainRepository.TransactionInterface, u *user.User) error {
	// Type assert to *sql.Tx since we need it for QueryRowContext
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return errors.New("invalid transaction type")
	}

	query := `
		INSERT INTO users (email, username, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := sqlTx.QueryRowContext(
		ctx,
		query,
		u.Email,
		u.Username,
		u.PasswordHash,
		u.IsActive,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// FindUserByEmail finds a user by their email address
func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	u := &user.User{}
	err := r.Db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, appErrors.Wrap(err, "failed to query user by email")
	}

	return u, nil
}

// FindUserByID finds a user by their ID
func (r *UserRepository) FindUserByID(ctx context.Context, id int64) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	u := &user.User{}
	err := r.Db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.ErrUserNotFound
		}
		return nil, appErrors.Wrap(err, "failed to query user by id")
	}

	return u, nil
}

// ExistsByEmail checks if a user with the given email already exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.Db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ExistsByUsername checks if a user with the given username already exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.Db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, password_hash = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.Db.QueryRowContext(
		ctx,
		query,
		u.Email,
		u.Username,
		u.PasswordHash,
		u.IsActive,
		u.ID,
	).Scan(&u.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.ErrUserNotFound
		}
		return appErrors.Wrap(err, "failed to update user")
	}

	return nil
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.Db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.Wrap(err, "failed to delete user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return appErrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return appErrors.ErrUserNotFound
	}

	return nil
}

// Compile-time check to ensure UserRepository implements user.RepositoryInterface
var _ user.RepositoryInterface = (*UserRepository)(nil)
