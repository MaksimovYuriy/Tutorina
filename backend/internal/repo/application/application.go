package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(database *sql.DB) *Repo { return &Repo{database: database} }

func (r *Repo) ListAll(ctx context.Context) ([]entity.Application, error) {
	return r.list(ctx, nil)
}

func (r *Repo) ListMine(ctx context.Context, userID int64) ([]entity.Application, error) {
	return r.list(ctx, &userID)
}

func (r *Repo) list(ctx context.Context, userID *int64) ([]entity.Application, error) {
	query := applicationSelect + ` WHERE applications.archived_at IS NULL`
	args := make([]any, 0, 1)
	if userID != nil {
		query += ` AND teacher_profiles.user_id=$1`
		args = append(args, *userID)
	}
	query += ` ORDER BY applications.created_at DESC, applications.id DESC`
	rows, err := r.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()
	items := make([]entity.Application, 0)
	for rows.Next() {
		var item entity.Application
		if err := scan(rows, &item); err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applications: %w", err)
	}
	return items, nil
}

func (r *Repo) Create(ctx context.Context, item entity.Application) (entity.Application, error) {
	const query = `INSERT INTO applications (lesson_id, full_name, phone, email, comment, status)
		SELECT lessons.id, $2, $3, $4, $5, 'new'
		FROM lessons
		JOIN teacher_offers ON teacher_offers.id=lessons.teacher_offer_id AND teacher_offers.archived_at IS NULL AND teacher_offers.is_published=TRUE
		JOIN offers ON offers.id=teacher_offers.offer_id AND offers.archived_at IS NULL AND offers.is_published=TRUE
		JOIN teacher_profiles ON teacher_profiles.id=lessons.teacher_profile_id AND teacher_profiles.archived_at IS NULL AND teacher_profiles.is_published=TRUE
		WHERE lessons.id=$1 AND lessons.archived_at IS NULL AND lessons.lesson_type='group'
			AND lessons.status='planned' AND lessons.enrollment_open=TRUE AND lessons.starts_at>CURRENT_TIMESTAMP
		RETURNING id`
	var id int64
	err := r.database.QueryRowContext(ctx, query, item.LessonID, item.FullName, item.Phone, item.Email, item.Comment).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Application{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Application{}, fmt.Errorf("create application: %w", err)
	}
	return r.get(ctx, r.database, id, nil)
}

func (r *Repo) UpdateStatus(ctx context.Context, applicationID int64, status string, userID *int64) (entity.Application, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return entity.Application{}, fmt.Errorf("begin application status transaction: %w", err)
	}
	defer tx.Rollback()

	const lockQuery = `SELECT applications.lesson_id, lessons.capacity
		FROM applications
		JOIN lessons ON lessons.id=applications.lesson_id AND lessons.archived_at IS NULL
		JOIN teacher_profiles ON teacher_profiles.id=lessons.teacher_profile_id
		WHERE applications.id=$1 AND applications.archived_at IS NULL
			AND ($2::bigint IS NULL OR teacher_profiles.user_id=$2)
		FOR UPDATE OF applications, lessons`
	var lessonID int64
	var capacity int
	if err := tx.QueryRowContext(ctx, lockQuery, applicationID, userID).Scan(&lessonID, &capacity); errors.Is(err, sql.ErrNoRows) {
		return entity.Application{}, repo.ErrNotFound
	} else if err != nil {
		return entity.Application{}, fmt.Errorf("lock application: %w", err)
	}

	if status == entity.ApplicationStatusAccepted || status == entity.ApplicationStatusCompleted {
		var occupied int
		const countQuery = `SELECT COUNT(*) FROM applications
			WHERE lesson_id=$1 AND id<>$2 AND archived_at IS NULL AND status IN ('accepted','completed')`
		if err := tx.QueryRowContext(ctx, countQuery, lessonID, applicationID).Scan(&occupied); err != nil {
			return entity.Application{}, fmt.Errorf("count occupied lesson places: %w", err)
		}
		if occupied >= capacity {
			return entity.Application{}, repo.ErrConflict
		}
	}

	if _, err := tx.ExecContext(ctx, `UPDATE applications SET status=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, applicationID, status); err != nil {
		return entity.Application{}, fmt.Errorf("update application status: %w", err)
	}
	updated, err := r.get(ctx, tx, applicationID, userID)
	if err != nil {
		return entity.Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return entity.Application{}, fmt.Errorf("commit application status: %w", err)
	}
	return updated, nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (r *Repo) get(ctx context.Context, database queryer, applicationID int64, userID *int64) (entity.Application, error) {
	query := applicationSelect + ` WHERE applications.id=$1 AND applications.archived_at IS NULL
		AND ($2::bigint IS NULL OR teacher_profiles.user_id=$2)`
	var item entity.Application
	err := scan(database.QueryRowContext(ctx, query, applicationID, userID), &item)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Application{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Application{}, fmt.Errorf("get application: %w", err)
	}
	return item, nil
}

const applicationSelect = `SELECT applications.id, applications.lesson_id, lessons.offer_title,
	lessons.teacher_profile_id, teacher_profiles.display_name, lessons.starts_at, lessons.capacity,
	applications.full_name, applications.phone, applications.email, applications.comment,
	applications.status, applications.created_at, applications.updated_at
	FROM applications
	JOIN lessons ON lessons.id=applications.lesson_id
	JOIN teacher_profiles ON teacher_profiles.id=lessons.teacher_profile_id`

type scanner interface{ Scan(...any) error }

func scan(row scanner, item *entity.Application) error {
	return row.Scan(
		&item.ID, &item.LessonID, &item.OfferTitle, &item.TeacherProfileID,
		&item.TeacherDisplayName, &item.LessonStartsAt, &item.LessonCapacity,
		&item.FullName, &item.Phone, &item.Email, &item.Comment, &item.Status,
		&item.CreatedAt, &item.UpdatedAt,
	)
}
