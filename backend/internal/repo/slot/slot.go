package slot

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db} }

const columns = `id,title,starts_at,ends_at,format,kind,capacity,occupied,status,published`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity.Slot, error) {
	var v entity.Slot
	err := row.Scan(&v.ID, &v.Title, &v.StartsAt, &v.EndsAt, &v.Format, &v.Kind, &v.Capacity, &v.Occupied, &v.Status, &v.Published)
	if errors.Is(err, sql.ErrNoRows) {
		err = repo.ErrNotFound
	}
	return v, err
}
func (r *Repo) List(ctx context.Context, public bool) ([]entity.Slot, error) {
	query := `SELECT ` + columns + ` FROM slots`
	if public {
		query += ` WHERE published AND status='planned' AND starts_at>CURRENT_TIMESTAMP AND occupied<capacity`
	}
	query += ` ORDER BY starts_at,id`
	rows, err := r.database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Slot, 0)
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (r *Repo) Save(ctx context.Context, v entity.Slot) (entity.Slot, error) {
	args := []any{v.Title, v.StartsAt, v.EndsAt, v.Format, v.Kind, v.Capacity, v.Occupied, v.Status, v.Published}
	query := `INSERT INTO slots(title,starts_at,ends_at,format,kind,capacity,occupied,status,published) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	if v.ID != 0 {
		args = append(args, v.ID)
		query = `UPDATE slots SET title=$1,starts_at=$2,ends_at=$3,format=$4,kind=$5,capacity=$6,occupied=$7,status=$8,published=$9 WHERE id=$10`
	}
	return scan(r.database.QueryRowContext(ctx, query+` RETURNING `+columns, args...))
}
func (r *Repo) Delete(ctx context.Context, id int64) error {
	result, err := r.database.ExecContext(ctx, `DELETE FROM slots WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repo.ErrNotFound
	}
	return nil
}
