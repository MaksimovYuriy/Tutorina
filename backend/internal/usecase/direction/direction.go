package direction

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
	List(context.Context) ([]entity.Direction, error)
	Save(context.Context, entity.Direction) (entity.Direction, error)
	Delete(context.Context, int64) error
}
type Service struct{ repository Repository }

func New(r Repository) *Service { return &Service{r} }
func (s *Service) List(ctx context.Context) ([]entity.Direction, error) {
	return s.repository.List(ctx)
}
func (s *Service) Save(ctx context.Context, v entity.Direction) (entity.Direction, error) {
	v.Name = strings.TrimSpace(v.Name)
	if v.ID < 0 || v.Name == "" || utf8.RuneCountInString(v.Name) > 120 {
		return entity.Direction{}, usecase.ErrInvalidInput
	}
	result, err := s.repository.Save(ctx, v)
	return result, mapError(err)
}
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return usecase.ErrInvalidInput
	}
	return mapError(s.repository.Delete(ctx, id))
}
func mapError(err error) error {
	if errors.Is(err, repo.ErrNotFound) {
		return usecase.ErrNotFound
	}
	if errors.Is(err, repo.ErrConflict) {
		return usecase.ErrConflict
	}
	return err
}
