package restapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type DirectionService interface {
	List(context.Context) ([]entity.Direction, error)
	Save(context.Context, entity.Direction) (entity.Direction, error)
	Delete(context.Context, int64) error
}
type directionController struct{ service DirectionService }

func (c directionController) list(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.List(r.Context())
	if err != nil {
		directionError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"data": result})
}
func (c directionController) create(w http.ResponseWriter, r *http.Request) { c.save(w, r, false) }
func (c directionController) update(w http.ResponseWriter, r *http.Request) { c.save(w, r, true) }
func (c directionController) save(w http.ResponseWriter, r *http.Request, update bool) {
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		directionError(w, usecase.ErrInvalidInput)
		return
	}
	v := entity.Direction{Name: input.Name}
	if update {
		id, err := resourceID(r)
		if err != nil {
			directionError(w, err)
			return
		}
		v.ID = id
	}
	result, err := c.service.Save(r.Context(), v)
	if err != nil {
		directionError(w, err)
		return
	}
	status := http.StatusCreated
	if update {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"data": result})
}
func (c directionController) delete(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r)
	if err == nil {
		err = c.service.Delete(r.Context(), id)
	}
	if errors.Is(err, usecase.ErrConflict) {
		apiresponse.WriteError(w, 409, "direction_in_use", "Направление используется в расписании. Измените или удалите связанные слоты перед удалением направления.", "")
		return
	}
	if err != nil {
		directionError(w, err)
		return
	}
	w.WriteHeader(204)
}
func directionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, 400, "invalid_input", "Введите название направления длиной до 120 символов.", "")
	case errors.Is(err, usecase.ErrNotFound):
		apiresponse.WriteError(w, 404, "not_found", "Направление не найдено.", "")
	case errors.Is(err, usecase.ErrConflict):
		apiresponse.WriteError(w, 409, "conflict", "Направление с таким названием уже существует.", "")
	default:
		apiresponse.WriteError(w, 500, "internal_error", "Не удалось выполнить запрос.", "")
	}
}
