package accesskey

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }
func (r *Repo) Hash(ctx context.Context) ([]byte, error) {
	var hash []byte
	err := r.database.QueryRowContext(ctx, `SELECT key_hash FROM access_keys WHERE id=1`).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		err = repo.ErrNotFound
	}
	return hash, err
}

// Replace revokes all sessions through ON DELETE CASCADE in the same transaction.
func (r *Repo) Replace(ctx context.Context, hash []byte) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serialize concurrent rotations, including the initial creation.
	if _, err = tx.ExecContext(ctx, `LOCK TABLE access_keys IN EXCLUSIVE MODE`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM access_keys`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO access_keys(id,key_hash) VALUES(1,$1)`, hash); err != nil {
		return err
	}
	return tx.Commit()
}
