package offer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
)

type Repo struct{ database *sql.DB }

func New(database *sql.DB) *Repo { return &Repo{database: database} }

func (r *Repo) List(ctx context.Context, publishedOnly bool) ([]entity.Offer, error) {
	offerQuery := `SELECT id, title, description, goal, default_duration_minutes, format, price_rubles, is_published, created_at, updated_at
		FROM offers WHERE archived_at IS NULL`
	if publishedOnly {
		offerQuery += ` AND is_published=TRUE`
	}
	offerQuery += ` ORDER BY title, id`

	rows, err := r.database.QueryContext(ctx, offerQuery)
	if err != nil {
		return nil, fmt.Errorf("list offers: %w", err)
	}
	offers := make([]entity.Offer, 0)
	byID := make(map[int64]int)
	for rows.Next() {
		var item entity.Offer
		if err := scanOffer(rows, &item); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		item.Teachers = make([]entity.TeacherOffer, 0)
		byID[item.ID] = len(offers)
		offers = append(offers, item)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close offer rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate offers: %w", err)
	}
	if len(offers) == 0 {
		return offers, nil
	}

	assignmentQuery := `SELECT teacher_offers.id, teacher_offers.offer_id, teacher_offers.teacher_profile_id,
		teacher_profiles.display_name, teacher_offers.duration_minutes, teacher_offers.price_rubles,
		teacher_offers.is_published, teacher_offers.created_at, teacher_offers.updated_at
		FROM teacher_offers
		JOIN offers ON offers.id=teacher_offers.offer_id AND offers.archived_at IS NULL
		JOIN teacher_profiles ON teacher_profiles.id=teacher_offers.teacher_profile_id AND teacher_profiles.archived_at IS NULL
		WHERE teacher_offers.archived_at IS NULL`
	if publishedOnly {
		assignmentQuery += ` AND offers.is_published=TRUE AND teacher_offers.is_published=TRUE AND teacher_profiles.is_published=TRUE`
	}
	assignmentQuery += ` ORDER BY teacher_profiles.display_name, teacher_offers.id`

	assignmentRows, err := r.database.QueryContext(ctx, assignmentQuery)
	if err != nil {
		return nil, fmt.Errorf("list teacher offers: %w", err)
	}
	defer assignmentRows.Close()
	for assignmentRows.Next() {
		var assignment entity.TeacherOffer
		if err := scanTeacherOffer(assignmentRows, &assignment); err != nil {
			return nil, fmt.Errorf("scan teacher offer: %w", err)
		}
		if index, ok := byID[assignment.OfferID]; ok {
			offers[index].Teachers = append(offers[index].Teachers, assignment)
		}
	}
	if err := assignmentRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teacher offers: %w", err)
	}
	return offers, nil
}

func (r *Repo) Get(ctx context.Context, offerID int64) (entity.Offer, error) {
	const query = `SELECT id, title, description, goal, default_duration_minutes, format, price_rubles, is_published, created_at, updated_at
		FROM offers WHERE id=$1 AND archived_at IS NULL`
	var item entity.Offer
	err := scanOffer(r.database.QueryRowContext(ctx, query, offerID), &item)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Offer{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Offer{}, fmt.Errorf("get offer: %w", err)
	}
	item.Teachers, err = r.listAssignments(ctx, offerID)
	if err != nil {
		return entity.Offer{}, err
	}
	return item, nil
}

func (r *Repo) Create(ctx context.Context, item entity.Offer) (entity.Offer, error) {
	const query = `INSERT INTO offers (title, description, goal, default_duration_minutes, format, price_rubles, is_published)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, title, description, goal, default_duration_minutes, format, price_rubles, is_published, created_at, updated_at`
	var created entity.Offer
	err := scanOffer(r.database.QueryRowContext(ctx, query, item.Title, item.Description, item.Goal, item.DefaultDurationMinutes, item.Format, item.PriceRubles, item.IsPublished), &created)
	if err != nil {
		return entity.Offer{}, fmt.Errorf("create offer: %w", err)
	}
	created.Teachers = make([]entity.TeacherOffer, 0)
	return created, nil
}

func (r *Repo) Update(ctx context.Context, item entity.Offer) (entity.Offer, error) {
	const query = `UPDATE offers SET title=$2, description=$3, goal=$4, default_duration_minutes=$5,
		format=$6, price_rubles=$7, is_published=$8, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND archived_at IS NULL
		RETURNING id, title, description, goal, default_duration_minutes, format, price_rubles, is_published, created_at, updated_at`
	var updated entity.Offer
	err := scanOffer(r.database.QueryRowContext(ctx, query, item.ID, item.Title, item.Description, item.Goal, item.DefaultDurationMinutes, item.Format, item.PriceRubles, item.IsPublished), &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Offer{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Offer{}, fmt.Errorf("update offer: %w", err)
	}
	var listErr error
	updated.Teachers, listErr = r.listAssignments(ctx, updated.ID)
	return updated, listErr
}

func (r *Repo) Archive(ctx context.Context, offerID int64) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin archive offer: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE offers SET is_published=FALSE, archived_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND archived_at IS NULL`, offerID)
	if err != nil {
		return fmt.Errorf("archive offer: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count archived offers: %w", err)
	}
	if changed == 0 {
		return repo.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE teacher_offers SET is_published=FALSE, archived_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE offer_id=$1 AND archived_at IS NULL`, offerID); err != nil {
		return fmt.Errorf("archive offer assignments: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive offer: %w", err)
	}
	return nil
}

func (r *Repo) CreateAssignment(ctx context.Context, assignment entity.TeacherOffer) (entity.TeacherOffer, error) {
	const query = `INSERT INTO teacher_offers (offer_id, teacher_profile_id, duration_minutes, price_rubles, is_published)
		SELECT offers.id, teacher_profiles.id, $3, $4, $5
		FROM offers, teacher_profiles
		WHERE offers.id=$1 AND offers.archived_at IS NULL
			AND teacher_profiles.id=$2 AND teacher_profiles.archived_at IS NULL
		RETURNING id, offer_id, teacher_profile_id, duration_minutes, price_rubles, is_published, created_at, updated_at`
	var created entity.TeacherOffer
	err := r.database.QueryRowContext(ctx, query, assignment.OfferID, assignment.TeacherProfileID, assignment.DurationMinutes, assignment.PriceRubles, assignment.IsPublished).
		Scan(&created.ID, &created.OfferID, &created.TeacherProfileID, &created.DurationMinutes, &created.PriceRubles, &created.IsPublished, &created.CreatedAt, &created.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherOffer{}, repo.ErrNotFound
	}
	if isUniqueViolation(err) {
		return entity.TeacherOffer{}, repo.ErrConflict
	}
	if err != nil {
		return entity.TeacherOffer{}, fmt.Errorf("create teacher offer: %w", err)
	}
	if err := r.database.QueryRowContext(ctx, `SELECT display_name FROM teacher_profiles WHERE id=$1`, created.TeacherProfileID).Scan(&created.TeacherDisplayName); err != nil {
		return entity.TeacherOffer{}, fmt.Errorf("get assigned teacher: %w", err)
	}
	return created, nil
}

func (r *Repo) UpdateAssignment(ctx context.Context, assignment entity.TeacherOffer) (entity.TeacherOffer, error) {
	const query = `UPDATE teacher_offers SET duration_minutes=$2, price_rubles=$3, is_published=$4, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND archived_at IS NULL
		RETURNING id, offer_id, teacher_profile_id, duration_minutes, price_rubles, is_published, created_at, updated_at`
	var updated entity.TeacherOffer
	err := r.database.QueryRowContext(ctx, query, assignment.ID, assignment.DurationMinutes, assignment.PriceRubles, assignment.IsPublished).
		Scan(&updated.ID, &updated.OfferID, &updated.TeacherProfileID, &updated.DurationMinutes, &updated.PriceRubles, &updated.IsPublished, &updated.CreatedAt, &updated.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherOffer{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.TeacherOffer{}, fmt.Errorf("update teacher offer: %w", err)
	}
	if err := r.database.QueryRowContext(ctx, `SELECT display_name FROM teacher_profiles WHERE id=$1 AND archived_at IS NULL`, updated.TeacherProfileID).Scan(&updated.TeacherDisplayName); errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherOffer{}, repo.ErrNotFound
	} else if err != nil {
		return entity.TeacherOffer{}, fmt.Errorf("get assigned teacher: %w", err)
	}
	return updated, nil
}

func (r *Repo) ArchiveAssignment(ctx context.Context, assignmentID int64) error {
	result, err := r.database.ExecContext(ctx, `UPDATE teacher_offers SET is_published=FALSE, archived_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND archived_at IS NULL`, assignmentID)
	if err != nil {
		return fmt.Errorf("archive teacher offer: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count archived teacher offers: %w", err)
	}
	if changed == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *Repo) listAssignments(ctx context.Context, offerID int64) ([]entity.TeacherOffer, error) {
	const query = `SELECT teacher_offers.id, teacher_offers.offer_id, teacher_offers.teacher_profile_id,
		teacher_profiles.display_name, teacher_offers.duration_minutes, teacher_offers.price_rubles,
		teacher_offers.is_published, teacher_offers.created_at, teacher_offers.updated_at
		FROM teacher_offers
		JOIN teacher_profiles ON teacher_profiles.id=teacher_offers.teacher_profile_id AND teacher_profiles.archived_at IS NULL
		WHERE teacher_offers.offer_id=$1 AND teacher_offers.archived_at IS NULL
		ORDER BY teacher_profiles.display_name, teacher_offers.id`
	rows, err := r.database.QueryContext(ctx, query, offerID)
	if err != nil {
		return nil, fmt.Errorf("list offer assignments: %w", err)
	}
	defer rows.Close()
	items := make([]entity.TeacherOffer, 0)
	for rows.Next() {
		var item entity.TeacherOffer
		if err := scanTeacherOffer(rows, &item); err != nil {
			return nil, fmt.Errorf("scan offer assignment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanOffer(row scanner, item *entity.Offer) error {
	return row.Scan(&item.ID, &item.Title, &item.Description, &item.Goal, &item.DefaultDurationMinutes, &item.Format, &item.PriceRubles, &item.IsPublished, &item.CreatedAt, &item.UpdatedAt)
}

func scanTeacherOffer(row scanner, item *entity.TeacherOffer) error {
	return row.Scan(&item.ID, &item.OfferID, &item.TeacherProfileID, &item.TeacherDisplayName, &item.DurationMinutes, &item.PriceRubles, &item.IsPublished, &item.CreatedAt, &item.UpdatedAt)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
