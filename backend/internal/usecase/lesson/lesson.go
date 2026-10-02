package lesson

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type Repository interface {
	ListPublic(context.Context, time.Time, time.Time) ([]entity.Lesson, error)
	ListAll(context.Context, time.Time, time.Time) ([]entity.Lesson, error)
	ListMine(context.Context, int64, time.Time, time.Time) ([]entity.Lesson, error)
	Create(context.Context, entity.Lesson, *int64) (entity.Lesson, error)
	Update(context.Context, entity.Lesson, *int64) (entity.Lesson, error)
	Archive(context.Context, int64, *int64) error
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListPublic(ctx context.Context, from, to time.Time) ([]entity.Lesson, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	return s.repository.ListPublic(ctx, from, to)
}

func (s *Service) ListAll(ctx context.Context, from, to time.Time) ([]entity.Lesson, error) {
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	return s.repository.ListAll(ctx, from, to)
}

func (s *Service) ListMine(ctx context.Context, userID int64, from, to time.Time) ([]entity.Lesson, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	return s.repository.ListMine(ctx, userID, from, to)
}

func (s *Service) CreateAdmin(ctx context.Context, item entity.Lesson) (entity.Lesson, error) {
	return s.create(ctx, item, nil)
}

func (s *Service) CreateMine(ctx context.Context, userID int64, item entity.Lesson) (entity.Lesson, error) {
	if userID <= 0 {
		return entity.Lesson{}, fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	return s.create(ctx, item, &userID)
}

func (s *Service) create(ctx context.Context, item entity.Lesson, userID *int64) (entity.Lesson, error) {
	item.ID = 0
	if item.Status == "" {
		item.Status = entity.LessonStatusPlanned
	}
	if err := validateLesson(&item); err != nil {
		return entity.Lesson{}, err
	}
	created, err := s.repository.Create(ctx, item, userID)
	return created, mapConflict(err)
}

func (s *Service) UpdateAdmin(ctx context.Context, item entity.Lesson) (entity.Lesson, error) {
	return s.update(ctx, item, nil)
}

func (s *Service) UpdateMine(ctx context.Context, userID int64, item entity.Lesson) (entity.Lesson, error) {
	if userID <= 0 {
		return entity.Lesson{}, fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	return s.update(ctx, item, &userID)
}

func (s *Service) update(ctx context.Context, item entity.Lesson, userID *int64) (entity.Lesson, error) {
	if item.ID <= 0 {
		return entity.Lesson{}, fmt.Errorf("%w: invalid lesson id", usecase.ErrInvalidInput)
	}
	if err := validateLesson(&item); err != nil {
		return entity.Lesson{}, err
	}
	updated, err := s.repository.Update(ctx, item, userID)
	return updated, mapConflict(err)
}

func (s *Service) ArchiveAdmin(ctx context.Context, lessonID int64) error {
	return s.archive(ctx, lessonID, nil)
}

func (s *Service) ArchiveMine(ctx context.Context, userID, lessonID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	return s.archive(ctx, lessonID, &userID)
}

func (s *Service) archive(ctx context.Context, lessonID int64, userID *int64) error {
	if lessonID <= 0 {
		return fmt.Errorf("%w: invalid lesson id", usecase.ErrInvalidInput)
	}
	return s.repository.Archive(ctx, lessonID, userID)
}

func validateRange(from, to time.Time) error {
	if from.IsZero() || to.IsZero() || !to.After(from) || to.Sub(from) > 366*24*time.Hour {
		return fmt.Errorf("%w: invalid schedule range", usecase.ErrInvalidInput)
	}
	return nil
}

func validateLesson(item *entity.Lesson) error {
	item.GroupGoal = strings.TrimSpace(item.GroupGoal)
	item.GroupLevel = strings.TrimSpace(item.GroupLevel)
	item.StartsAt = item.StartsAt.UTC()
	item.EndsAt = item.EndsAt.UTC()
	if item.TeacherOfferID <= 0 {
		return fmt.Errorf("%w: teacher offer is required", usecase.ErrInvalidInput)
	}
	if item.StartsAt.IsZero() || item.EndsAt.IsZero() || !item.EndsAt.After(item.StartsAt) {
		return fmt.Errorf("%w: end must be after start", usecase.ErrInvalidInput)
	}
	if item.EndsAt.Sub(item.StartsAt) > 24*time.Hour {
		return fmt.Errorf("%w: lesson cannot last more than 24 hours", usecase.ErrInvalidInput)
	}
	if item.DeliveryFormat != entity.LessonDeliveryOnline && item.DeliveryFormat != entity.LessonDeliveryOffline {
		return fmt.Errorf("%w: unsupported delivery format", usecase.ErrInvalidInput)
	}
	switch item.LessonType {
	case entity.LessonTypeIndividual:
		if item.Capacity != 1 {
			return fmt.Errorf("%w: individual lesson capacity must be 1", usecase.ErrInvalidInput)
		}
		item.GroupGoal = ""
		item.GroupLevel = ""
	case entity.LessonTypeGroup:
		if item.Capacity < 2 {
			return fmt.Errorf("%w: group lesson capacity must be at least 2", usecase.ErrInvalidInput)
		}
		if item.GroupGoal == "" || item.GroupLevel == "" {
			return fmt.Errorf("%w: group goal and level are required", usecase.ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: unsupported lesson type", usecase.ErrInvalidInput)
	}
	switch item.Status {
	case entity.LessonStatusPlanned, entity.LessonStatusCompleted, entity.LessonStatusCancelled:
	default:
		return fmt.Errorf("%w: unsupported lesson status", usecase.ErrInvalidInput)
	}
	if item.Status != entity.LessonStatusPlanned {
		item.EnrollmentOpen = false
	}
	if len([]rune(item.GroupGoal)) > 500 || len([]rune(item.GroupLevel)) > 160 {
		return fmt.Errorf("%w: group details are too long", usecase.ErrInvalidInput)
	}
	return nil
}

func mapConflict(err error) error {
	if errors.Is(err, repo.ErrConflict) {
		return usecase.ErrConflict
	}
	return err
}
