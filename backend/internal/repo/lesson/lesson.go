package lesson

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(database *sql.DB) *Repo { return &Repo{database: database} }

func (r *Repo) ListPublic(ctx context.Context, from, to time.Time) ([]entity.Lesson, error) {
	const filter = ` AND lessons.lesson_type='group' AND lessons.status='planned' AND lessons.enrollment_open=TRUE
		AND teacher_offers.archived_at IS NULL AND offers.archived_at IS NULL AND teacher_profiles.archived_at IS NULL
		AND teacher_offers.is_published=TRUE AND offers.is_published=TRUE AND teacher_profiles.is_published=TRUE`
	return r.list(ctx, from, to, filter, nil)
}

func (r *Repo) ListAll(ctx context.Context, from, to time.Time) ([]entity.Lesson, error) {
	return r.list(ctx, from, to, "", nil)
}

func (r *Repo) ListMine(ctx context.Context, userID int64, from, to time.Time) ([]entity.Lesson, error) {
	return r.list(ctx, from, to, ` AND teacher_profiles.user_id=$3`, []any{userID})
}

func (r *Repo) list(ctx context.Context, from, to time.Time, filter string, extraArgs []any) ([]entity.Lesson, error) {
	query := `SELECT lessons.id, lessons.teacher_offer_id, lessons.teacher_profile_id, teacher_profiles.display_name,
		lessons.offer_title, lessons.price_rubles, lessons.starts_at, lessons.ends_at, lessons.delivery_format,
		lessons.lesson_type, lessons.capacity, lessons.status, lessons.enrollment_open, lessons.group_goal,
		lessons.group_level, lessons.created_at, lessons.updated_at
		FROM lessons
		JOIN teacher_offers ON teacher_offers.id=lessons.teacher_offer_id
		JOIN offers ON offers.id=teacher_offers.offer_id
		JOIN teacher_profiles ON teacher_profiles.id=lessons.teacher_profile_id
		WHERE lessons.archived_at IS NULL AND lessons.starts_at < $2 AND lessons.ends_at > $1` + filter +
		` ORDER BY lessons.starts_at, teacher_profiles.display_name, lessons.id`
	args := []any{from, to}
	args = append(args, extraArgs...)
	rows, err := r.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	defer rows.Close()
	items := make([]entity.Lesson, 0)
	for rows.Next() {
		var item entity.Lesson
		if err := scan(rows, &item); err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lessons: %w", err)
	}
	return items, nil
}

func (r *Repo) Create(ctx context.Context, item entity.Lesson, userID *int64) (entity.Lesson, error) {
	const query = `INSERT INTO lessons (
		teacher_offer_id, teacher_profile_id, offer_title, price_rubles, starts_at, ends_at,
		delivery_format, lesson_type, capacity, status, enrollment_open, group_goal, group_level
	)
	SELECT teacher_offers.id, teacher_profiles.id, offers.title,
		COALESCE(teacher_offers.price_rubles, offers.price_rubles), $2, $3, $4, $5, $6, $7, $8, $9, $10
	FROM teacher_offers
	JOIN offers ON offers.id=teacher_offers.offer_id AND offers.archived_at IS NULL
	JOIN teacher_profiles ON teacher_profiles.id=teacher_offers.teacher_profile_id AND teacher_profiles.archived_at IS NULL
	WHERE teacher_offers.id=$1 AND teacher_offers.archived_at IS NULL
		AND (offers.format='both' OR offers.format=$4)
		AND ($11::bigint IS NULL OR teacher_profiles.user_id=$11)
	RETURNING id, teacher_offer_id, teacher_profile_id, '', offer_title, price_rubles, starts_at, ends_at,
		delivery_format, lesson_type, capacity, status, enrollment_open, group_goal, group_level, created_at, updated_at`
	var created entity.Lesson
	err := scan(r.database.QueryRowContext(ctx, query,
		item.TeacherOfferID, item.StartsAt, item.EndsAt, item.DeliveryFormat, item.LessonType,
		item.Capacity, item.Status, item.EnrollmentOpen, item.GroupGoal, item.GroupLevel, userID,
	), &created)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Lesson{}, repo.ErrNotFound
	}
	if isExclusionViolation(err) {
		return entity.Lesson{}, repo.ErrConflict
	}
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("create lesson: %w", err)
	}
	if err := r.teacherName(ctx, &created); err != nil {
		return entity.Lesson{}, err
	}
	return created, nil
}

func (r *Repo) Update(ctx context.Context, item entity.Lesson, userID *int64) (entity.Lesson, error) {
	const query = `UPDATE lessons SET
		teacher_offer_id=teacher_offers.id,
		teacher_profile_id=teacher_profiles.id,
		offer_title=offers.title,
		price_rubles=COALESCE(teacher_offers.price_rubles, offers.price_rubles),
		starts_at=$3, ends_at=$4, delivery_format=$5, lesson_type=$6, capacity=$7,
		status=$8, enrollment_open=$9, group_goal=$10, group_level=$11, updated_at=CURRENT_TIMESTAMP
	FROM teacher_offers
	JOIN offers ON offers.id=teacher_offers.offer_id AND offers.archived_at IS NULL
	JOIN teacher_profiles ON teacher_profiles.id=teacher_offers.teacher_profile_id AND teacher_profiles.archived_at IS NULL
	WHERE lessons.id=$1 AND lessons.archived_at IS NULL
		AND teacher_offers.id=$2 AND teacher_offers.archived_at IS NULL
		AND (offers.format='both' OR offers.format=$5)
		AND ($12::bigint IS NULL OR (teacher_profiles.user_id=$12 AND lessons.teacher_profile_id=teacher_profiles.id))
	RETURNING lessons.id, lessons.teacher_offer_id, lessons.teacher_profile_id, '', lessons.offer_title,
		lessons.price_rubles, lessons.starts_at, lessons.ends_at, lessons.delivery_format, lessons.lesson_type,
		lessons.capacity, lessons.status, lessons.enrollment_open, lessons.group_goal, lessons.group_level,
		lessons.created_at, lessons.updated_at`
	var updated entity.Lesson
	err := scan(r.database.QueryRowContext(ctx, query,
		item.ID, item.TeacherOfferID, item.StartsAt, item.EndsAt, item.DeliveryFormat, item.LessonType,
		item.Capacity, item.Status, item.EnrollmentOpen, item.GroupGoal, item.GroupLevel, userID,
	), &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Lesson{}, repo.ErrNotFound
	}
	if isExclusionViolation(err) {
		return entity.Lesson{}, repo.ErrConflict
	}
	if err != nil {
		return entity.Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	if err := r.teacherName(ctx, &updated); err != nil {
		return entity.Lesson{}, err
	}
	return updated, nil
}

func (r *Repo) Archive(ctx context.Context, lessonID int64, userID *int64) error {
	const query = `UPDATE lessons SET archived_at=CURRENT_TIMESTAMP, enrollment_open=FALSE, updated_at=CURRENT_TIMESTAMP
		FROM teacher_profiles
		WHERE lessons.id=$1 AND lessons.archived_at IS NULL AND teacher_profiles.id=lessons.teacher_profile_id
			AND ($2::bigint IS NULL OR teacher_profiles.user_id=$2)`
	result, err := r.database.ExecContext(ctx, query, lessonID, userID)
	if err != nil {
		return fmt.Errorf("archive lesson: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count archived lessons: %w", err)
	}
	if changed == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *Repo) teacherName(ctx context.Context, item *entity.Lesson) error {
	if err := r.database.QueryRowContext(ctx, `SELECT display_name FROM teacher_profiles WHERE id=$1`, item.TeacherProfileID).Scan(&item.TeacherDisplayName); err != nil {
		return fmt.Errorf("get lesson teacher: %w", err)
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner, item *entity.Lesson) error {
	return row.Scan(
		&item.ID, &item.TeacherOfferID, &item.TeacherProfileID, &item.TeacherDisplayName,
		&item.OfferTitle, &item.PriceRubles, &item.StartsAt, &item.EndsAt, &item.DeliveryFormat,
		&item.LessonType, &item.Capacity, &item.Status, &item.EnrollmentOpen, &item.GroupGoal,
		&item.GroupLevel, &item.CreatedAt, &item.UpdatedAt,
	)
}

func isExclusionViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23P01"
}
