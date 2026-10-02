package lesson

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestCreateIndividualRequiresCapacityOne(t *testing.T) {
	service := New(&repositoryStub{})
	_, err := service.CreateAdmin(context.Background(), validLesson(entity.LessonTypeIndividual, 2))
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateGroupRequiresPublicDetails(t *testing.T) {
	service := New(&repositoryStub{})
	item := validLesson(entity.LessonTypeGroup, 4)
	item.GroupGoal = ""
	_, err := service.CreateAdmin(context.Background(), item)
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateMapsOverlapConflict(t *testing.T) {
	service := New(&repositoryStub{createErr: repo.ErrConflict})
	_, err := service.CreateAdmin(context.Background(), validLesson(entity.LessonTypeGroup, 4))
	if !errors.Is(err, usecase.ErrConflict) {
		t.Fatalf("error = %v", err)
	}
}

func validLesson(lessonType string, capacity int) entity.Lesson {
	start := time.Now().Add(time.Hour)
	return entity.Lesson{
		TeacherOfferID: 1, StartsAt: start, EndsAt: start.Add(time.Hour),
		DeliveryFormat: entity.LessonDeliveryOnline, LessonType: lessonType,
		Capacity: capacity, Status: entity.LessonStatusPlanned, EnrollmentOpen: true,
		GroupGoal: "Разговорная практика", GroupLevel: "A2",
	}
}

type repositoryStub struct{ createErr error }

func (*repositoryStub) ListPublic(context.Context, time.Time, time.Time) ([]entity.Lesson, error) {
	return nil, nil
}
func (*repositoryStub) ListAll(context.Context, time.Time, time.Time) ([]entity.Lesson, error) {
	return nil, nil
}
func (*repositoryStub) ListMine(context.Context, int64, time.Time, time.Time) ([]entity.Lesson, error) {
	return nil, nil
}
func (r *repositoryStub) Create(context.Context, entity.Lesson, *int64) (entity.Lesson, error) {
	return entity.Lesson{}, r.createErr
}
func (*repositoryStub) Update(context.Context, entity.Lesson, *int64) (entity.Lesson, error) {
	return entity.Lesson{}, nil
}
func (*repositoryStub) Archive(context.Context, int64, *int64) error { return nil }
