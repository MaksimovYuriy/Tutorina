package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct {
	database *sql.DB
}

func New(database *sql.DB) *Repo {
	return &Repo{database: database}
}

func (r *Repo) FindCredentialsByEmail(ctx context.Context, email string) (entity.Credentials, error) {
	const query = `
		SELECT id, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var credentials entity.Credentials
	err := r.database.QueryRowContext(ctx, query, email).Scan(
		&credentials.ID,
		&credentials.Email,
		&credentials.PasswordHash,
		&credentials.IsActive,
		&credentials.CreatedAt,
		&credentials.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Credentials{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Credentials{}, fmt.Errorf("find user credentials: %w", err)
	}

	roles, err := r.roles(ctx, r.database, credentials.ID)
	if err != nil {
		return entity.Credentials{}, err
	}
	credentials.Roles = roles
	return credentials, nil
}

func (r *Repo) FindByID(ctx context.Context, id int64) (entity.User, error) {
	const query = `
		SELECT id, email, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err := r.database.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.User{}, fmt.Errorf("find user: %w", err)
	}

	roles, err := r.roles(ctx, r.database, user.ID)
	if err != nil {
		return entity.User{}, err
	}
	user.Roles = roles
	return user, nil
}

func (r *Repo) FindCredentialsByID(ctx context.Context, id int64) (entity.Credentials, error) {
	const query = `
		SELECT id, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var credentials entity.Credentials
	err := r.database.QueryRowContext(ctx, query, id).Scan(
		&credentials.ID,
		&credentials.Email,
		&credentials.PasswordHash,
		&credentials.IsActive,
		&credentials.CreatedAt,
		&credentials.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Credentials{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Credentials{}, fmt.Errorf("find user credentials by id: %w", err)
	}
	roles, err := r.roles(ctx, r.database, credentials.ID)
	if err != nil {
		return entity.Credentials{}, err
	}
	credentials.Roles = roles
	return credentials, nil
}

func (r *Repo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	result, err := r.database.ExecContext(ctx, `
		UPDATE users SET password_hash=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND is_active=TRUE
	`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user password rows: %w", err)
	}
	if affected != 1 {
		return repo.ErrNotFound
	}
	return nil
}
func (r *Repo) Create(ctx context.Context, email, passwordHash string, roles []entity.Role) (entity.User, error) {
	transaction, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return entity.User{}, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer transaction.Rollback()

	const createUser = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, is_active, created_at, updated_at
	`
	var user entity.User
	err = transaction.QueryRowContext(ctx, createUser, email, passwordHash).Scan(
		&user.ID, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return entity.User{}, repo.ErrConflict
		}
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}

	const assignRole = `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE code = $2
	`
	for _, role := range roles {
		result, err := transaction.ExecContext(ctx, assignRole, user.ID, role)
		if err != nil {
			return entity.User{}, fmt.Errorf("assign user role: %w", err)
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			return entity.User{}, fmt.Errorf("assign user role %q: role does not exist", role)
		}
	}

	if err := transaction.Commit(); err != nil {
		return entity.User{}, fmt.Errorf("commit create user transaction: %w", err)
	}
	user.Roles = append([]entity.Role(nil), roles...)
	return user, nil
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (r *Repo) roles(ctx context.Context, queryer queryer, userID int64) ([]entity.Role, error) {
	const query = `
		SELECT roles.code
		FROM roles
		JOIN user_roles ON user_roles.role_id = roles.id
		WHERE user_roles.user_id = $1
		ORDER BY roles.code
	`
	rows, err := queryer.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]entity.Role, 0, 2)
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user roles: %w", err)
	}
	return roles, nil
}
