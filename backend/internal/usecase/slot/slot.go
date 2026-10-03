package slot

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type Repository interface {
	List(context.Context, bool) ([]entity.Slot, error)
	Save(context.Context, entity.Slot) (entity.Slot, error)
	Delete(context.Context, int64) error
}
type Service struct{ repository Repository }

func New(r Repository) *Service { return &Service{r} }
func (s *Service) List(ctx context.Context, public bool) ([]entity.Slot, error) {
	return s.repository.List(ctx, public)
}
func (s *Service) Save(ctx context.Context, v entity.Slot) (entity.Slot, error) {
	v.Level = strings.TrimSpace(v.Level)
	if v.ID < 0 || v.DirectionID <= 0 || utf8.RuneCountInString(v.Level) > 80 || v.StartsAt.IsZero() || !v.EndsAt.After(v.StartsAt) ||
		(v.Format != "online" && v.Format != "offline") || (v.Kind != "individual" && v.Kind != "group") ||
		v.Capacity < 1 || v.Capacity > 1000 || v.Occupied < 0 || v.Occupied > v.Capacity ||
		(v.Kind == "individual" && v.Capacity != 1) ||
		(v.Status != "planned" && v.Status != "completed" && v.Status != "cancelled") {
		return entity.Slot{}, usecase.ErrInvalidInput
	}
	result, err := s.repository.Save(ctx, v)
	if errors.Is(err, repo.ErrNotFound) {
		err = usecase.ErrNotFound
	}
	if errors.Is(err, repo.ErrInvalidInput) {
		err = usecase.ErrInvalidInput
	}
	return result, err
}
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return usecase.ErrInvalidInput
	}
	err := s.repository.Delete(ctx, id)
	if errors.Is(err, repo.ErrNotFound) {
		return usecase.ErrNotFound
	}
	return err
}
