package slot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type repositoryStub struct {
	saved bool
	err   error
}

func (r *repositoryStub) List(context.Context, bool) ([]entity.Slot, error) {
	return []entity.Slot{}, nil
}
func (r *repositoryStub) Save(_ context.Context, v entity.Slot) (entity.Slot, error) {
	r.saved = true
	return v, r.err
}
func (r *repositoryStub) Delete(context.Context, int64) error { return r.err }
func validSlot() entity.Slot {
	start := time.Now().Add(time.Hour)
	return entity.Slot{DirectionID: 1, StartsAt: start, EndsAt: start.Add(time.Hour), Format: "online", Kind: "group", Capacity: 4, Occupied: 2, Status: "planned", Published: true}
}
func TestInvalidSlotNeverReachesRepository(t *testing.T) {
	tests := map[string]func(*entity.Slot){
		"long level":          func(v *entity.Slot) { v.Level = strings.Repeat("Я", 81) },
		"missing direction":   func(v *entity.Slot) { v.DirectionID = 0 },
		"end before start":    func(v *entity.Slot) { v.EndsAt = v.StartsAt.Add(-time.Minute) },
		"zero duration":       func(v *entity.Slot) { v.EndsAt = v.StartsAt },
		"overbooking":         func(v *entity.Slot) { v.Occupied = 5 },
		"negative occupancy":  func(v *entity.Slot) { v.Occupied = -1 },
		"zero capacity":       func(v *entity.Slot) { v.Capacity = 0 },
		"individual capacity": func(v *entity.Slot) { v.Kind = "individual" },
		"unknown status":      func(v *entity.Slot) { v.Status = "accepted" },
		"unknown format":      func(v *entity.Slot) { v.Format = "both" },
		"unknown kind":        func(v *entity.Slot) { v.Kind = "other" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			v := validSlot()
			mutate(&v)
			r := &repositoryStub{}
			_, err := New(r).Save(context.Background(), v)
			if !errors.Is(err, usecase.ErrInvalidInput) || r.saved {
				t.Fatalf("err=%v saved=%v", err, r.saved)
			}
		})
	}
}
func TestAdminCanFillAndReleaseSlot(t *testing.T) {
	for _, occupied := range []int{4, 0} {
		v := validSlot()
		v.Occupied = occupied
		got, err := New(&repositoryStub{}).Save(context.Background(), v)
		if err != nil || got.Occupied != occupied {
			t.Fatalf("slot=%+v err=%v", got, err)
		}
	}
}
func TestMissingSlotMapsToNotFound(t *testing.T) {
	s := New(&repositoryStub{err: repo.ErrNotFound})
	v := validSlot()
	v.ID = 7
	if _, err := s.Save(context.Background(), v); !errors.Is(err, usecase.ErrNotFound) {
		t.Fatal(err)
	}
	if err := s.Delete(context.Background(), 7); !errors.Is(err, usecase.ErrNotFound) {
		t.Fatal(err)
	}
}

func TestUnknownDirectionReturnsInvalidInput(t *testing.T) {
	_, err := New(&repositoryStub{err: repo.ErrInvalidInput}).Save(context.Background(), validSlot())
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatal(err)
	}
}
