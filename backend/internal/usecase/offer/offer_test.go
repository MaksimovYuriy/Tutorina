package offer

import (
	"context"
	"errors"
	"testing"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestCreateNormalizesOffer(t *testing.T) {
	repository := &repositoryStub{}
	service := New(repository)
	price := 1500
	_, err := service.Create(context.Background(), entity.Offer{
		Title: "  Разговорный английский  ", DefaultDurationMinutes: 60,
		Format: entity.OfferFormatOnline, PriceRubles: &price,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.created.Title != "Разговорный английский" {
		t.Fatalf("normalized title = %q", repository.created.Title)
	}
}

func TestCreateRejectsInvalidDuration(t *testing.T) {
	service := New(&repositoryStub{})
	_, err := service.Create(context.Background(), entity.Offer{Title: "Английский", DefaultDurationMinutes: 10, Format: entity.OfferFormatOnline})
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
}

func TestAssignTeacherMapsConflict(t *testing.T) {
	service := New(&repositoryStub{assignmentErr: repo.ErrConflict})
	_, err := service.AssignTeacher(context.Background(), entity.TeacherOffer{OfferID: 1, TeacherProfileID: 2})
	if !errors.Is(err, usecase.ErrConflict) {
		t.Fatalf("error = %v", err)
	}
}

type repositoryStub struct {
	created       entity.Offer
	assignmentErr error
}

func (r *repositoryStub) List(context.Context, bool) ([]entity.Offer, error) { return nil, nil }
func (r *repositoryStub) Get(context.Context, int64) (entity.Offer, error) {
	return entity.Offer{}, nil
}
func (r *repositoryStub) Create(_ context.Context, item entity.Offer) (entity.Offer, error) {
	r.created = item
	return item, nil
}
func (r *repositoryStub) Update(context.Context, entity.Offer) (entity.Offer, error) {
	return entity.Offer{}, nil
}
func (r *repositoryStub) Archive(context.Context, int64) error { return nil }
func (r *repositoryStub) CreateAssignment(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error) {
	return entity.TeacherOffer{}, r.assignmentErr
}
func (r *repositoryStub) UpdateAssignment(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error) {
	return entity.TeacherOffer{}, nil
}
func (r *repositoryStub) ArchiveAssignment(context.Context, int64) error { return nil }
