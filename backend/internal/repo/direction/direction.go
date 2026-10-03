package direction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }
func (r *Repo) List(ctx context.Context) ([]entity.Direction, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT id,name FROM directions ORDER BY LOWER(name),id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Direction, 0)
	for rows.Next() {
		var v entity.Direction
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (r *Repo) Save(ctx context.Context, v entity.Direction) (entity.Direction, error) {
	query := `INSERT INTO directions(name) VALUES($1) RETURNING id,name`
	args := []any{v.Name}
	if v.ID != 0 {
		query = `UPDATE directions SET name=$1 WHERE id=$2 RETURNING id,name`
		args = append(args, v.ID)
	}
	err := r.database.QueryRowContext(ctx, query, args...).Scan(&v.ID, &v.Name)
	return v, mapError(err)
}
func (r *Repo) Delete(ctx context.Context, id int64) error {
	result, err := r.database.ExecContext(ctx, `DELETE FROM directions WHERE id=$1`, id)
	if err != nil {
		return mapError(err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return repo.ErrNotFound
	}
	return nil
}
func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && (pg.Code == "23505" || pg.Code == "23503" || pg.Code == "23001") {
		return repo.ErrConflict
	}
	return err
}
