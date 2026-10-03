package slot

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

const columns = `s.id,s.direction_id,d.name,s.level,s.starts_at,s.ends_at,s.format,s.kind,s.capacity,s.occupied,s.status,s.published`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity.Slot, error) {
	var v entity.Slot
	err := row.Scan(&v.ID, &v.DirectionID, &v.Title, &v.Level, &v.StartsAt, &v.EndsAt, &v.Format, &v.Kind, &v.Capacity, &v.Occupied, &v.Status, &v.Published)
	if errors.Is(err, sql.ErrNoRows) {
		err = repo.ErrNotFound
	}
	return v, err
}
func (r *Repo) List(ctx context.Context, public bool) ([]entity.Slot, error) {
	query := `SELECT ` + columns + ` FROM slots s JOIN directions d ON d.id=s.direction_id`
	if public {
		query += ` WHERE published AND status='planned' AND starts_at>CURRENT_TIMESTAMP AND occupied<capacity`
	}
	query += ` ORDER BY starts_at,s.id`
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

	args := []any{v.DirectionID, v.Level, v.StartsAt, v.EndsAt, v.Format, v.Kind, v.Capacity, v.Occupied, v.Status, v.Published}
	query := `INSERT INTO slots(direction_id,level,starts_at,ends_at,format,kind,capacity,occupied,status,published) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	if v.ID != 0 {
		args = append(args, v.ID)
		query = `UPDATE slots SET direction_id=$1,level=$2,starts_at=$3,ends_at=$4,format=$5,kind=$6,capacity=$7,occupied=$8,status=$9,published=$10 WHERE id=$11`
	}
	// Return the current direction name, without duplicating it in slots.
	result, err := scan(r.database.QueryRowContext(ctx, `WITH saved AS (`+query+` RETURNING *) SELECT `+columns+` FROM saved s JOIN directions d ON d.id=s.direction_id`, args...))
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23503" {
		return entity.Slot{}, repo.ErrInvalidInput
	}
	return result, err
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
