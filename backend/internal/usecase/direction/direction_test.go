package direction

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type repositoryStub struct {
	saved bool
	err   error
}

func (r *repositoryStub) List(context.Context) ([]entity.Direction, error) {
	return []entity.Direction{}, nil
}
func (r *repositoryStub) Save(_ context.Context, v entity.Direction) (entity.Direction, error) {
	r.saved = true
	return v, r.err
}
func (r *repositoryStub) Delete(context.Context, int64) error { return r.err }
func TestInvalidDirectionIsNotSaved(t *testing.T) {
	for _, name := range []string{" ", strings.Repeat("Я", 121)} {
		r := &repositoryStub{}
		_, err := New(r).Save(context.Background(), entity.Direction{Name: name})
		if !errors.Is(err, usecase.ErrInvalidInput) || r.saved {
			t.Fatalf("err=%v saved=%v", err, r.saved)
		}
	}
}
func TestDirectionNameTrimmed(t *testing.T) {
	v, err := New(&repositoryStub{}).Save(context.Background(), entity.Direction{Name: "  Математика  "})
	if err != nil || v.Name != "Математика" {
		t.Fatalf("name=%q err=%v", v.Name, err)
	}
}
func TestConflictsAndMissingDirections(t *testing.T) {
	for _, tc := range []struct{ storage, want error }{{repo.ErrConflict, usecase.ErrConflict}, {repo.ErrNotFound, usecase.ErrNotFound}} {
		s := New(&repositoryStub{err: tc.storage})
		if err := s.Delete(context.Background(), 1); !errors.Is(err, tc.want) {
			t.Fatal(err)
		}
		if _, err := s.Save(context.Background(), entity.Direction{ID: 1, Name: "Математика"}); !errors.Is(err, tc.want) {
			t.Fatal(err)
		}
	}
}
