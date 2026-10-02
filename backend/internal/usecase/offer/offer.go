package offer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type Repository interface {
	List(context.Context, bool) ([]entity.Offer, error)
	Get(context.Context, int64) (entity.Offer, error)
	Create(context.Context, entity.Offer) (entity.Offer, error)
	Update(context.Context, entity.Offer) (entity.Offer, error)
	Archive(context.Context, int64) error
	CreateAssignment(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error)
	UpdateAssignment(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error)
	ArchiveAssignment(context.Context, int64) error
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListPublic(ctx context.Context) ([]entity.Offer, error) {
	return s.repository.List(ctx, true)
}

func (s *Service) ListAll(ctx context.Context) ([]entity.Offer, error) {
	return s.repository.List(ctx, false)
}

func (s *Service) Create(ctx context.Context, item entity.Offer) (entity.Offer, error) {
	item.ID = 0
	if err := validateOffer(&item); err != nil {
		return entity.Offer{}, err
	}
	return s.repository.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, item entity.Offer) (entity.Offer, error) {
	if item.ID <= 0 {
		return entity.Offer{}, fmt.Errorf("%w: invalid offer id", usecase.ErrInvalidInput)
	}
	if err := validateOffer(&item); err != nil {
		return entity.Offer{}, err
	}
	return s.repository.Update(ctx, item)
}

func (s *Service) Archive(ctx context.Context, offerID int64) error {
	if offerID <= 0 {
		return fmt.Errorf("%w: invalid offer id", usecase.ErrInvalidInput)
	}
	return s.repository.Archive(ctx, offerID)
}

func (s *Service) AssignTeacher(ctx context.Context, assignment entity.TeacherOffer) (entity.TeacherOffer, error) {
	assignment.ID = 0
	if assignment.OfferID <= 0 || assignment.TeacherProfileID <= 0 {
		return entity.TeacherOffer{}, fmt.Errorf("%w: offer and teacher are required", usecase.ErrInvalidInput)
	}
	if err := validateOverrides(assignment); err != nil {
		return entity.TeacherOffer{}, err
	}
	created, err := s.repository.CreateAssignment(ctx, assignment)
	if errors.Is(err, repo.ErrConflict) {
		return entity.TeacherOffer{}, usecase.ErrConflict
	}
	return created, err
}

func (s *Service) UpdateAssignment(ctx context.Context, assignment entity.TeacherOffer) (entity.TeacherOffer, error) {
	if assignment.ID <= 0 {
		return entity.TeacherOffer{}, fmt.Errorf("%w: invalid assignment id", usecase.ErrInvalidInput)
	}
	if err := validateOverrides(assignment); err != nil {
		return entity.TeacherOffer{}, err
	}
	return s.repository.UpdateAssignment(ctx, assignment)
}

func (s *Service) RemoveAssignment(ctx context.Context, assignmentID int64) error {
	if assignmentID <= 0 {
		return fmt.Errorf("%w: invalid assignment id", usecase.ErrInvalidInput)
	}
	return s.repository.ArchiveAssignment(ctx, assignmentID)
}

func validateOffer(item *entity.Offer) error {
	item.Title = strings.TrimSpace(item.Title)
	item.Description = strings.TrimSpace(item.Description)
	item.Goal = strings.TrimSpace(item.Goal)
	if item.Title == "" {
		return fmt.Errorf("%w: title is required", usecase.ErrInvalidInput)
	}
	if len([]rune(item.Title)) > 160 {
		return fmt.Errorf("%w: title is too long", usecase.ErrInvalidInput)
	}
	if item.DefaultDurationMinutes < 15 || item.DefaultDurationMinutes > 480 {
		return fmt.Errorf("%w: duration must be between 15 and 480 minutes", usecase.ErrInvalidInput)
	}
	switch item.Format {
	case entity.OfferFormatOnline, entity.OfferFormatOffline, entity.OfferFormatBoth:
	default:
		return fmt.Errorf("%w: unsupported format", usecase.ErrInvalidInput)
	}
	if item.PriceRubles != nil && *item.PriceRubles < 0 {
		return fmt.Errorf("%w: price cannot be negative", usecase.ErrInvalidInput)
	}
	return nil
}

func validateOverrides(assignment entity.TeacherOffer) error {
	if assignment.DurationMinutes != nil && (*assignment.DurationMinutes < 15 || *assignment.DurationMinutes > 480) {
		return fmt.Errorf("%w: duration must be between 15 and 480 minutes", usecase.ErrInvalidInput)
	}
	if assignment.PriceRubles != nil && *assignment.PriceRubles < 0 {
		return fmt.Errorf("%w: price cannot be negative", usecase.ErrInvalidInput)
	}
	return nil
}
