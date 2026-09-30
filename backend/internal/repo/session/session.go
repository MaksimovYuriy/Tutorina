package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct {
	database *sql.DB
}

func New(database *sql.DB) *Repo {
	return &Repo{database: database}
}

func (r *Repo) Create(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	const query = `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := r.database.ExecContext(ctx, query, userID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *Repo) FindActiveUserID(ctx context.Context, tokenHash []byte, now time.Time) (int64, error) {
	const query = `
		SELECT sessions.user_id
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1
			AND sessions.expires_at > $2
			AND sessions.revoked_at IS NULL
			AND users.is_active = TRUE
	`
	var userID int64
	if err := r.database.QueryRowContext(ctx, query, tokenHash, now).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, repo.ErrNotFound
		}
		return 0, fmt.Errorf("find active session: %w", err)
	}
	return userID, nil
}

func (r *Repo) Revoke(ctx context.Context, tokenHash []byte, now time.Time) error {
	const query = `
		UPDATE sessions
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	if _, err := r.database.ExecContext(ctx, query, tokenHash, now); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}
