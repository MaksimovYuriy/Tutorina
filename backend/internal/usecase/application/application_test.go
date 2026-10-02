package application

import (
	"context"
	"errors"
	"testing"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestCreateRequiresPhone(t *testing.T) {
	service := New(&repositoryStub{})
	_, err := service.Create(context.Background(), entity.Application{LessonID: 1, FullName: "Иван Иванов"})
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
}

func TestUpdateMapsCapacityConflict(t *testing.T) {
	service := New(&repositoryStub{updateErr: repo.ErrConflict})
	_, err := service.UpdateAdmin(context.Background(), 1, entity.ApplicationStatusAccepted)
	if !errors.Is(err, usecase.ErrConflict) {
		t.Fatalf("error = %v", err)
	}
}

type repositoryStub struct{ updateErr error }

func (*repositoryStub) ListAll(context.Context) ([]entity.Application, error) { return nil, nil }
func (*repositoryStub) ListMine(context.Context, int64) ([]entity.Application, error) {
	return nil, nil
}
func (*repositoryStub) Create(_ context.Context, item entity.Application) (entity.Application, error) {
	return item, nil
}
func (r *repositoryStub) UpdateStatus(context.Context, int64, string, *int64) (entity.Application, error) {
	return entity.Application{}, r.updateErr
}
