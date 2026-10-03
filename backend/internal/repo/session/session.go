package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }
func (r *Repo) Create(ctx context.Context, keyHash, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.database.ExecContext(ctx, `INSERT INTO sessions(key_hash,token_hash,expires_at) VALUES($1,$2,$3)`, keyHash, tokenHash, expiresAt)
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "23503" {
		return repo.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}
func (r *Repo) FindActive(ctx context.Context, tokenHash []byte, now time.Time) error {
	var exists bool
	err := r.database.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sessions WHERE token_hash=$1 AND expires_at>$2 AND revoked_at IS NULL)`, tokenHash, now).Scan(&exists)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	if !exists {
		return repo.ErrNotFound
	}
	return nil
}
func (r *Repo) Revoke(ctx context.Context, tokenHash []byte, now time.Time) error {
	_, err := r.database.ExecContext(ctx, `UPDATE sessions SET revoked_at=$2 WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash, now)
	return err
}
