package restapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type SlotService interface {
	List(context.Context, bool) ([]entity.Slot, error)
	Save(context.Context, entity.Slot) (entity.Slot, error)
	Delete(context.Context, int64) error
}
type slotController struct{ service SlotService }

func (c slotController) listPublic(w http.ResponseWriter, r *http.Request) {
	slots, err := c.service.List(r.Context(), true)
	if err != nil {
		slotError(w, err)
		return
	}
	// Public output contains availability only, not internal occupancy or publication state.
	type publicSlot struct {
		ID         int64     `json:"id"`
		Title      string    `json:"title"`
		StartsAt   time.Time `json:"startsAt"`
		EndsAt     time.Time `json:"endsAt"`
		Format     string    `json:"format"`
		Kind       string    `json:"kind"`
		FreePlaces int       `json:"freePlaces"`
	}
	result := make([]publicSlot, 0, len(slots))
	for _, s := range slots {
		result = append(result, publicSlot{s.ID, s.Title, s.StartsAt, s.EndsAt, s.Format, s.Kind, s.Capacity - s.Occupied})
	}
	writeJSON(w, 200, map[string]any{"data": result})
}
func (c slotController) listAdmin(w http.ResponseWriter, r *http.Request) {
	slots, err := c.service.List(r.Context(), false)
	if err != nil {
		slotError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"data": slots})
}
func (c slotController) create(w http.ResponseWriter, r *http.Request) { c.save(w, r, false) }
func (c slotController) update(w http.ResponseWriter, r *http.Request) { c.save(w, r, true) }
func (c slotController) save(w http.ResponseWriter, r *http.Request, update bool) {
	var input entity.Slot
	if err := decodeJSON(w, r, &input); err != nil || input.ID != 0 {
		slotError(w, usecase.ErrInvalidInput)
		return
	}
	if update {
		id, err := slotID(r)
		if err != nil {
			slotError(w, err)
			return
		}
		input.ID = id
	}
	result, err := c.service.Save(r.Context(), input)
	if err != nil {
		slotError(w, err)
		return
	}
	status := http.StatusCreated
	if update {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"data": result})
}
func (c slotController) delete(w http.ResponseWriter, r *http.Request) {
	id, err := slotID(r)
	if err == nil {
		err = c.service.Delete(r.Context(), id)
	}
	if err != nil {
		slotError(w, err)
		return
	}
	w.WriteHeader(204)
}
func slotID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, usecase.ErrInvalidInput
	}
	return id, nil
}
func slotError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, 400, "invalid_input", "Проверьте время, вместимость и занятость слота.", "")
	case errors.Is(err, usecase.ErrNotFound):
		apiresponse.WriteError(w, 404, "not_found", "Слот не найден.", "")
	default:
		apiresponse.WriteError(w, 500, "internal_error", "Не удалось выполнить запрос.", "")
	}
}
