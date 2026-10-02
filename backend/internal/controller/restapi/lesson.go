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

type LessonService interface {
	ListPublic(context.Context, time.Time, time.Time) ([]entity.Lesson, error)
	ListAll(context.Context, time.Time, time.Time) ([]entity.Lesson, error)
	ListMine(context.Context, int64, time.Time, time.Time) ([]entity.Lesson, error)
	CreateAdmin(context.Context, entity.Lesson) (entity.Lesson, error)
	CreateMine(context.Context, int64, entity.Lesson) (entity.Lesson, error)
	UpdateAdmin(context.Context, entity.Lesson) (entity.Lesson, error)
	UpdateMine(context.Context, int64, entity.Lesson) (entity.Lesson, error)
	ArchiveAdmin(context.Context, int64) error
	ArchiveMine(context.Context, int64, int64) error
}

type lessonController struct{ service LessonService }

type lessonRequest struct {
	TeacherOfferID int64     `json:"teacherOfferId"`
	Description    string    `json:"description"`
	StartsAt       time.Time `json:"startsAt"`
	EndsAt         time.Time `json:"endsAt"`
	DeliveryFormat string    `json:"deliveryFormat"`
	LessonType     string    `json:"lessonType"`
	Capacity       int       `json:"capacity"`
	Status         string    `json:"status"`
	EnrollmentOpen bool      `json:"enrollmentOpen"`
	GroupGoal      string    `json:"groupGoal"`
	GroupLevel     string    `json:"groupLevel"`
}

type lessonAttributes struct {
	TeacherOfferID     string    `json:"teacherOfferId"`
	TeacherProfileID   string    `json:"teacherProfileId"`
	TeacherDisplayName string    `json:"teacherDisplayName"`
	OfferTitle         string    `json:"offerTitle"`
	PriceRubles        *int      `json:"priceRubles"`
	Description        string    `json:"description"`
	StartsAt           time.Time `json:"startsAt"`
	EndsAt             time.Time `json:"endsAt"`
	DeliveryFormat     string    `json:"deliveryFormat"`
	LessonType         string    `json:"lessonType"`
	Capacity           int       `json:"capacity"`
	Status             string    `json:"status"`
	EnrollmentOpen     bool      `json:"enrollmentOpen"`
	GroupGoal          string    `json:"groupGoal"`
	GroupLevel         string    `json:"groupLevel"`
}

type lessonResource struct {
	Type       string           `json:"type"`
	ID         string           `json:"id"`
	Attributes lessonAttributes `json:"attributes"`
}

func (c lessonController) listPublic(w http.ResponseWriter, r *http.Request) {
	from, to, ok := scheduleRange(w, r, true)
	if !ok {
		return
	}
	items, err := c.service.ListPublic(r.Context(), from, to)
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonList(w, items)
}

func (c lessonController) listAll(w http.ResponseWriter, r *http.Request) {
	from, to, ok := scheduleRange(w, r, false)
	if !ok {
		return
	}
	items, err := c.service.ListAll(r.Context(), from, to)
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonList(w, items)
}

func (c lessonController) listMine(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	from, to, ok := scheduleRange(w, r, false)
	if !ok {
		return
	}
	items, err := c.service.ListMine(r.Context(), user.ID, from, to)
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonList(w, items)
}

func (c lessonController) createAdmin(w http.ResponseWriter, r *http.Request) {
	c.create(w, r, false)
}

func (c lessonController) createMine(w http.ResponseWriter, r *http.Request) {
	c.create(w, r, true)
}

func (c lessonController) create(w http.ResponseWriter, r *http.Request, mine bool) {
	var body lessonRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	var (
		created entity.Lesson
		err     error
	)
	if mine {
		user, ok := middleware.CurrentUser(r.Context())
		if !ok {
			apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		created, err = c.service.CreateMine(r.Context(), user.ID, body.entity(0))
	} else {
		created, err = c.service.CreateAdmin(r.Context(), body.entity(0))
	}
	if err != nil {
		writeLessonError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapLesson(created)})
}

func (c lessonController) updateAdmin(w http.ResponseWriter, r *http.Request) {
	c.update(w, r, false)
}

func (c lessonController) updateMine(w http.ResponseWriter, r *http.Request) {
	c.update(w, r, true)
}

func (c lessonController) update(w http.ResponseWriter, r *http.Request, mine bool) {
	id, ok := positiveID(w, r, "id", "lesson")
	if !ok {
		return
	}
	var body lessonRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	var (
		updated entity.Lesson
		err     error
	)
	if mine {
		user, ok := middleware.CurrentUser(r.Context())
		if !ok {
			apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		updated, err = c.service.UpdateMine(r.Context(), user.ID, body.entity(id))
	} else {
		updated, err = c.service.UpdateAdmin(r.Context(), body.entity(id))
	}
	if err != nil {
		writeLessonError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapLesson(updated)})
}

func (c lessonController) archiveAdmin(w http.ResponseWriter, r *http.Request) {
	c.archive(w, r, false)
}

func (c lessonController) archiveMine(w http.ResponseWriter, r *http.Request) {
	c.archive(w, r, true)
}

func (c lessonController) archive(w http.ResponseWriter, r *http.Request, mine bool) {
	id, ok := positiveID(w, r, "id", "lesson")
	if !ok {
		return
	}
	var err error
	if mine {
		user, ok := middleware.CurrentUser(r.Context())
		if !ok {
			apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
		err = c.service.ArchiveMine(r.Context(), user.ID, id)
	} else {
		err = c.service.ArchiveAdmin(r.Context(), id)
	}
	if err != nil {
		writeLessonError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func scheduleRange(w http.ResponseWriter, r *http.Request, public bool) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	from := now.Add(-30 * 24 * time.Hour)
	to := now.Add(180 * 24 * time.Hour)
	if public {
		from = now
		to = now.Add(90 * 24 * time.Hour)
	}
	var err error
	if raw := r.URL.Query().Get("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			apiresponse.WriteError(w, http.StatusBadRequest, "invalid_range", "Invalid schedule range", "from must be RFC3339")
			return time.Time{}, time.Time{}, false
		}
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			apiresponse.WriteError(w, http.StatusBadRequest, "invalid_range", "Invalid schedule range", "to must be RFC3339")
			return time.Time{}, time.Time{}, false
		}
	}
	return from, to, true
}

func (request lessonRequest) entity(id int64) entity.Lesson {
	return entity.Lesson{
		ID: id, TeacherOfferID: request.TeacherOfferID, Description: request.Description, StartsAt: request.StartsAt, EndsAt: request.EndsAt,
		DeliveryFormat: request.DeliveryFormat, LessonType: request.LessonType, Capacity: request.Capacity,
		Status: request.Status, EnrollmentOpen: request.EnrollmentOpen, GroupGoal: request.GroupGoal,
		GroupLevel: request.GroupLevel,
	}
}

func writeLessonList(w http.ResponseWriter, items []entity.Lesson) {
	data := make([]lessonResource, 0, len(items))
	for _, item := range items {
		data = append(data, mapLesson(item))
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": data})
}

func mapLesson(item entity.Lesson) lessonResource {
	return lessonResource{
		Type: "lessons", ID: strconv.FormatInt(item.ID, 10),
		Attributes: lessonAttributes{
			TeacherOfferID:     strconv.FormatInt(item.TeacherOfferID, 10),
			TeacherProfileID:   strconv.FormatInt(item.TeacherProfileID, 10),
			TeacherDisplayName: item.TeacherDisplayName, OfferTitle: item.OfferTitle,
			PriceRubles: item.PriceRubles, Description: item.Description, StartsAt: item.StartsAt, EndsAt: item.EndsAt,
			DeliveryFormat: item.DeliveryFormat, LessonType: item.LessonType, Capacity: item.Capacity,
			Status: item.Status, EnrollmentOpen: item.EnrollmentOpen, GroupGoal: item.GroupGoal,
			GroupLevel: item.GroupLevel,
		},
	}
}

func writeLessonError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_input", "Invalid lesson", err.Error())
	case errors.Is(err, repo.ErrNotFound):
		apiresponse.WriteError(w, http.StatusNotFound, "not_found", "Lesson or teacher offer not found", "")
	case errors.Is(err, usecase.ErrConflict), errors.Is(err, repo.ErrConflict):
		apiresponse.WriteError(w, http.StatusConflict, "schedule_conflict", "Lesson overlaps another lesson", "Choose another time for this teacher")
	default:
		slog.Error("Lesson request failed", slog.Any("error", err))
		apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
	}
}
