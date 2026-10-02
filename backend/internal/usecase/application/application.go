package application

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type Repository interface {
	ListAll(context.Context) ([]entity.Application, error)
	ListMine(context.Context, int64) ([]entity.Application, error)
	Create(context.Context, entity.Application) (entity.Application, error)
	UpdateStatus(context.Context, int64, string, *int64) (entity.Application, error)
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListAll(ctx context.Context) ([]entity.Application, error) {
	return s.repository.ListAll(ctx)
}

func (s *Service) ListMine(ctx context.Context, userID int64) ([]entity.Application, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	return s.repository.ListMine(ctx, userID)
}

func (s *Service) Create(ctx context.Context, item entity.Application) (entity.Application, error) {
	item.ID = 0
	item.Status = entity.ApplicationStatusNew
	if err := validateContact(&item); err != nil {
		return entity.Application{}, err
	}
	return s.repository.Create(ctx, item)
}

func (s *Service) UpdateAdmin(ctx context.Context, applicationID int64, status string) (entity.Application, error) {
	return s.update(ctx, applicationID, status, nil)
}

func (s *Service) UpdateMine(ctx context.Context, userID, applicationID int64, status string) (entity.Application, error) {
	if userID <= 0 {
		return entity.Application{}, fmt.Errorf("%w: invalid user id", usecase.ErrInvalidInput)
	}
	return s.update(ctx, applicationID, status, &userID)
}

func (s *Service) update(ctx context.Context, applicationID int64, status string, userID *int64) (entity.Application, error) {
	if applicationID <= 0 || !validStatus(status) {
		return entity.Application{}, fmt.Errorf("%w: invalid application status", usecase.ErrInvalidInput)
	}
	updated, err := s.repository.UpdateStatus(ctx, applicationID, status, userID)
	if errors.Is(err, repo.ErrConflict) {
		return entity.Application{}, usecase.ErrConflict
	}
	return updated, err
}

func validateContact(item *entity.Application) error {
	item.FullName = strings.TrimSpace(item.FullName)
	item.Phone = strings.TrimSpace(item.Phone)
	item.Email = strings.ToLower(strings.TrimSpace(item.Email))
	item.Comment = strings.TrimSpace(item.Comment)
	if item.LessonID <= 0 || item.FullName == "" || item.Phone == "" {
		return fmt.Errorf("%w: lesson, full name and phone are required", usecase.ErrInvalidInput)
	}
	if len([]rune(item.FullName)) > 200 || len([]rune(item.Phone)) > 50 || len([]rune(item.Email)) > 320 || len([]rune(item.Comment)) > 2000 {
		return fmt.Errorf("%w: application fields are too long", usecase.ErrInvalidInput)
	}
	if item.Email != "" {
		address, err := mail.ParseAddress(item.Email)
		if err != nil || address.Address != item.Email {
			return fmt.Errorf("%w: invalid email", usecase.ErrInvalidInput)
		}
	}
	return nil
}

func validStatus(status string) bool {
	switch status {
	case entity.ApplicationStatusNew, entity.ApplicationStatusAccepted,
		entity.ApplicationStatusRejected, entity.ApplicationStatusCompleted:
		return true
	default:
		return false
	}
}
