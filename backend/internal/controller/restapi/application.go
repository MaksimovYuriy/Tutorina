package restapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type ApplicationService interface {
	ListAll(context.Context) ([]entity.Application, error)
	ListMine(context.Context, int64) ([]entity.Application, error)
	Create(context.Context, entity.Application) (entity.Application, error)
	UpdateAdmin(context.Context, int64, string) (entity.Application, error)
	UpdateMine(context.Context, int64, int64, string) (entity.Application, error)
}

type applicationController struct{ service ApplicationService }

type applicationRequest struct {
	LessonID int64  `json:"lessonId"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Comment  string `json:"comment"`
}

type applicationStatusRequest struct {
	Status string `json:"status"`
}

type applicationAttributes struct {
	LessonID           string    `json:"lessonId"`
	OfferTitle         string    `json:"offerTitle"`
	TeacherProfileID   string    `json:"teacherProfileId"`
	TeacherDisplayName string    `json:"teacherDisplayName"`
	LessonStartsAt     time.Time `json:"lessonStartsAt"`
	LessonCapacity     int       `json:"lessonCapacity"`
	FullName           string    `json:"fullName"`
	Phone              string    `json:"phone"`
	Email              string    `json:"email"`
	Comment            string    `json:"comment"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
}

type applicationResource struct {
	Type       string                `json:"type"`
	ID         string                `json:"id"`
	Attributes applicationAttributes `json:"attributes"`
}

func (c applicationController) create(w http.ResponseWriter, r *http.Request) {
	var body applicationRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	created, err := c.service.Create(r.Context(), entity.Application{
		LessonID: body.LessonID, FullName: body.FullName, Phone: body.Phone,
		Email: body.Email, Comment: body.Comment,
	})
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapApplication(created)})
}

func (c applicationController) listAll(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.ListAll(r.Context())
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeApplicationList(w, items)
}

func (c applicationController) listMine(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	items, err := c.service.ListMine(r.Context(), user.ID)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeApplicationList(w, items)
}

func (c applicationController) updateAdmin(w http.ResponseWriter, r *http.Request) {
	c.update(w, r, false)
}

func (c applicationController) updateMine(w http.ResponseWriter, r *http.Request) {
	c.update(w, r, true)
}

func (c applicationController) update(w http.ResponseWriter, r *http.Request, mine bool) {
	id, ok := positiveID(w, r, "id", "application")
	if !ok {
		return
	}
	var body applicationStatusRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	var (
		updated entity.Application
		err     error
	)
	if mine {
		user, ok := middleware.CurrentUser(r.Context())
		if !ok {
			apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		updated, err = c.service.UpdateMine(r.Context(), user.ID, id, body.Status)
	} else {
		updated, err = c.service.UpdateAdmin(r.Context(), id, body.Status)
	}
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapApplication(updated)})
}

func writeApplicationList(w http.ResponseWriter, items []entity.Application) {
	data := make([]applicationResource, 0, len(items))
	for _, item := range items {
		data = append(data, mapApplication(item))
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": data})
}

func mapApplication(item entity.Application) applicationResource {
	return applicationResource{
		Type: "applications", ID: strconv.FormatInt(item.ID, 10),
		Attributes: applicationAttributes{
			LessonID: strconv.FormatInt(item.LessonID, 10), OfferTitle: item.OfferTitle,
			TeacherProfileID: strconv.FormatInt(item.TeacherProfileID, 10), TeacherDisplayName: item.TeacherDisplayName,
			LessonStartsAt: item.LessonStartsAt, LessonCapacity: item.LessonCapacity,
			FullName: item.FullName, Phone: item.Phone, Email: item.Email, Comment: item.Comment,
			Status: item.Status, CreatedAt: item.CreatedAt,
		},
	}
}

func writeApplicationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_input", "Invalid application", err.Error())
	case errors.Is(err, repo.ErrNotFound):
		apiresponse.WriteError(w, http.StatusNotFound, "not_found", "Lesson or application not found", "")
	case errors.Is(err, usecase.ErrConflict), errors.Is(err, repo.ErrConflict):
		apiresponse.WriteError(w, http.StatusConflict, "lesson_full", "No places left", "The lesson capacity has been reached")
	default:
		slog.Error("Application request failed", slog.Any("error", err))
		apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
	}
}
