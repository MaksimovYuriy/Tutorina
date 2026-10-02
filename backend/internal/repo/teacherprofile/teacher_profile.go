package teacherprofile

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

func (r *Repo) Get(ctx context.Context, profileID int64) (entity.TeacherProfile, error) {
	const query = `SELECT id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at
		FROM teacher_profiles WHERE id=$1 AND archived_at IS NULL`
	profile, err := scan(r.database.QueryRowContext(ctx, query, profileID))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherProfile{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("get teacher profile: %w", err)
	}
	return profile, nil
}

func (r *Repo) GetByUserID(ctx context.Context, userID int64) (entity.TeacherProfile, error) {
	const query = `SELECT id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at
		FROM teacher_profiles WHERE user_id=$1 AND archived_at IS NULL`
	profile, err := scan(r.database.QueryRowContext(ctx, query, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherProfile{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("get teacher profile by user: %w", err)
	}
	return profile, nil
}
func (r *Repo) List(ctx context.Context, publishedOnly bool) ([]entity.TeacherProfile, error) {
	query := `SELECT id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at FROM teacher_profiles`
	if publishedOnly {
		query += ` WHERE is_published = TRUE AND archived_at IS NULL`
	} else {
		query += ` WHERE archived_at IS NULL`
	}
	query += ` ORDER BY display_name, id`
	rows, err := r.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list teacher profiles: %w", err)
	}
	defer rows.Close()
	profiles := make([]entity.TeacherProfile, 0)
	for rows.Next() {
		profile, err := scan(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

func (r *Repo) Archive(ctx context.Context, profileID int64) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin archive teacher transaction: %w", err)
	}
	defer tx.Rollback()
	var userID sql.NullInt64
	err = tx.QueryRowContext(ctx, `UPDATE teacher_profiles SET is_published=FALSE, archived_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND archived_at IS NULL RETURNING user_id`, profileID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("archive teacher profile: %w", err)
	}
	if userID.Valid {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET is_active=FALSE, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, userID.Int64); err != nil {
			return fmt.Errorf("deactivate teacher account: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND revoked_at IS NULL`, userID.Int64); err != nil {
			return fmt.Errorf("revoke teacher sessions: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive teacher: %w", err)
	}
	return nil
}

func (r *Repo) CreateAccount(ctx context.Context, profileID int64, email, passwordHash string) (entity.TeacherProfile, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("begin teacher account transaction: %w", err)
	}
	defer tx.Rollback()
	var userID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id`, email, passwordHash).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.TeacherProfile{}, repo.ErrConflict
		}
		return entity.TeacherProfile{}, fmt.Errorf("create teacher account: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) SELECT $1, id FROM roles WHERE code='teacher'`, userID); err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("assign teacher role: %w", err)
	}
	const bind = `UPDATE teacher_profiles SET user_id=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND user_id IS NULL RETURNING id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at`
	profile, err := scan(tx.QueryRowContext(ctx, bind, profileID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherProfile{}, repo.ErrConflict
	}
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("bind teacher account: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("commit teacher account: %w", err)
	}
	return profile, nil
}

func (r *Repo) ReplacePhoto(ctx context.Context, profileID int64, photoURL string) (entity.TeacherProfile, string, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return entity.TeacherProfile{}, "", fmt.Errorf("begin replace teacher photo transaction: %w", err)
	}
	defer tx.Rollback()

	var oldPhotoURL string
	err = tx.QueryRowContext(ctx, `SELECT photo_url FROM teacher_profiles WHERE id=$1 AND archived_at IS NULL FOR UPDATE`, profileID).Scan(&oldPhotoURL)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherProfile{}, "", repo.ErrNotFound
	}
	if err != nil {
		return entity.TeacherProfile{}, "", fmt.Errorf("get current teacher photo: %w", err)
	}

	const update = `UPDATE teacher_profiles SET photo_url=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND archived_at IS NULL
		RETURNING id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at`
	profile, err := scan(tx.QueryRowContext(ctx, update, profileID, photoURL))
	if err != nil {
		return entity.TeacherProfile{}, "", fmt.Errorf("replace teacher photo: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return entity.TeacherProfile{}, "", fmt.Errorf("commit replace teacher photo: %w", err)
	}
	return profile, oldPhotoURL, nil
}

func (r *Repo) ResetPassword(ctx context.Context, profileID int64, passwordHash string) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset teacher password transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM teacher_profiles WHERE id=$1 AND user_id IS NOT NULL AND archived_at IS NULL FOR UPDATE`, profileID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return repo.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("find teacher account for password reset: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND is_active=TRUE`, userID, passwordHash); err != nil {
		return fmt.Errorf("reset teacher password: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke teacher sessions after password reset: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset teacher password: %w", err)
	}
	return nil
}
func (r *Repo) Create(ctx context.Context, profile entity.TeacherProfile) (entity.TeacherProfile, error) {
	const query = `INSERT INTO teacher_profiles (user_id, display_name, education, experience, approach, photo_url, is_published)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at`
	created, err := scan(r.database.QueryRowContext(ctx, query, profile.UserID, profile.DisplayName, profile.Education, profile.Experience, profile.Approach, profile.PhotoURL, profile.IsPublished))
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("create teacher profile: %w", err)
	}
	return created, nil
}

func (r *Repo) Update(ctx context.Context, profile entity.TeacherProfile) (entity.TeacherProfile, error) {
	const query = `UPDATE teacher_profiles SET user_id=$2, display_name=$3, education=$4, experience=$5, approach=$6, photo_url=$7, is_published=$8, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND archived_at IS NULL RETURNING id, user_id, display_name, education, experience, approach, photo_url, is_published, created_at, updated_at`
	updated, err := scan(r.database.QueryRowContext(ctx, query, profile.ID, profile.UserID, profile.DisplayName, profile.Education, profile.Experience, profile.Approach, profile.PhotoURL, profile.IsPublished))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeacherProfile{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.TeacherProfile{}, fmt.Errorf("update teacher profile: %w", err)
	}
	return updated, nil
}

type scanner interface{ Scan(...any) error }

func scan(row scanner) (entity.TeacherProfile, error) {
	var profile entity.TeacherProfile
	err := row.Scan(&profile.ID, &profile.UserID, &profile.DisplayName, &profile.Education, &profile.Experience, &profile.Approach, &profile.PhotoURL, &profile.IsPublished, &profile.CreatedAt, &profile.UpdatedAt)
	return profile, err
}
