package restapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type OfferService interface {
	ListPublic(context.Context) ([]entity.Offer, error)
	ListAll(context.Context) ([]entity.Offer, error)
	Create(context.Context, entity.Offer) (entity.Offer, error)
	Update(context.Context, entity.Offer) (entity.Offer, error)
	Archive(context.Context, int64) error
	AssignTeacher(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error)
	UpdateAssignment(context.Context, entity.TeacherOffer) (entity.TeacherOffer, error)
	RemoveAssignment(context.Context, int64) error
}

type offerController struct{ service OfferService }

type offerRequest struct {
	Title                  string `json:"title"`
	Description            string `json:"description"`
	Goal                   string `json:"goal"`
	DefaultDurationMinutes int    `json:"defaultDurationMinutes"`
	Format                 string `json:"format"`
	PriceRubles            *int   `json:"priceRubles"`
	IsPublished            bool   `json:"isPublished"`
}

type teacherOfferRequest struct {
	TeacherProfileID int64 `json:"teacherProfileId"`
	DurationMinutes  *int  `json:"durationMinutes"`
	PriceRubles      *int  `json:"priceRubles"`
	IsPublished      bool  `json:"isPublished"`
}

type teacherOfferAttributes struct {
	OfferID            string `json:"offerId"`
	TeacherProfileID   string `json:"teacherProfileId"`
	TeacherDisplayName string `json:"teacherDisplayName"`
	DurationMinutes    *int   `json:"durationMinutes"`
	PriceRubles        *int   `json:"priceRubles"`
	IsPublished        bool   `json:"isPublished"`
}

type teacherOfferResource struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Attributes teacherOfferAttributes `json:"attributes"`
}

type offerAttributes struct {
	Title                  string                 `json:"title"`
	Description            string                 `json:"description"`
	Goal                   string                 `json:"goal"`
	DefaultDurationMinutes int                    `json:"defaultDurationMinutes"`
	Format                 string                 `json:"format"`
	PriceRubles            *int                   `json:"priceRubles"`
	IsPublished            bool                   `json:"isPublished"`
	Teachers               []teacherOfferResource `json:"teachers"`
}

type offerResource struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes offerAttributes `json:"attributes"`
}

func (c offerController) listPublic(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.ListPublic(r.Context())
	if err != nil {
		writeOfferError(w, err)
		return
	}
	writeOfferList(w, items)
}

func (c offerController) listAll(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.ListAll(r.Context())
	if err != nil {
		writeOfferError(w, err)
		return
	}
	writeOfferList(w, items)
}

func (c offerController) create(w http.ResponseWriter, r *http.Request) {
	var body offerRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	created, err := c.service.Create(r.Context(), body.entity(0))
	if err != nil {
		writeOfferError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapOffer(created)})
}

func (c offerController) update(w http.ResponseWriter, r *http.Request) {
	id, ok := positiveID(w, r, "id", "offer")
	if !ok {
		return
	}
	var body offerRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	updated, err := c.service.Update(r.Context(), body.entity(id))
	if err != nil {
		writeOfferError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapOffer(updated)})
}

func (c offerController) archive(w http.ResponseWriter, r *http.Request) {
	id, ok := positiveID(w, r, "id", "offer")
	if !ok {
		return
	}
	if err := c.service.Archive(r.Context(), id); err != nil {
		writeOfferError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c offerController) assignTeacher(w http.ResponseWriter, r *http.Request) {
	offerID, ok := positiveID(w, r, "id", "offer")
	if !ok {
		return
	}
	var body teacherOfferRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	created, err := c.service.AssignTeacher(r.Context(), body.entity(0, offerID))
	if err != nil {
		writeOfferError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapTeacherOffer(created)})
}

func (c offerController) updateAssignment(w http.ResponseWriter, r *http.Request) {
	id, ok := positiveID(w, r, "id", "teacher offer")
	if !ok {
		return
	}
	var body teacherOfferRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	updated, err := c.service.UpdateAssignment(r.Context(), body.entity(id, 0))
	if err != nil {
		writeOfferError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherOffer(updated)})
}

func (c offerController) removeAssignment(w http.ResponseWriter, r *http.Request) {
	id, ok := positiveID(w, r, "id", "teacher offer")
	if !ok {
		return
	}
	if err := c.service.RemoveAssignment(r.Context(), id); err != nil {
		writeOfferError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func positiveID(w http.ResponseWriter, r *http.Request, parameter, resource string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, parameter), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid "+resource+" id", "")
		return 0, false
	}
	return id, true
}

func (request offerRequest) entity(id int64) entity.Offer {
	return entity.Offer{
		ID: id, Title: request.Title, Description: request.Description, Goal: request.Goal,
		DefaultDurationMinutes: request.DefaultDurationMinutes, Format: request.Format,
		PriceRubles: request.PriceRubles, IsPublished: request.IsPublished,
	}
}

func (request teacherOfferRequest) entity(id, offerID int64) entity.TeacherOffer {
	return entity.TeacherOffer{
		ID: id, OfferID: offerID, TeacherProfileID: request.TeacherProfileID,
		DurationMinutes: request.DurationMinutes, PriceRubles: request.PriceRubles,
		IsPublished: request.IsPublished,
	}
}

func writeOfferList(w http.ResponseWriter, items []entity.Offer) {
	data := make([]offerResource, 0, len(items))
	for _, item := range items {
		data = append(data, mapOffer(item))
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": data})
}

func mapOffer(item entity.Offer) offerResource {
	teachers := make([]teacherOfferResource, 0, len(item.Teachers))
	for _, assignment := range item.Teachers {
		teachers = append(teachers, mapTeacherOffer(assignment))
	}
	return offerResource{
		Type: "offers", ID: strconv.FormatInt(item.ID, 10),
		Attributes: offerAttributes{
			Title: item.Title, Description: item.Description, Goal: item.Goal,
			DefaultDurationMinutes: item.DefaultDurationMinutes, Format: item.Format,
			PriceRubles: item.PriceRubles, IsPublished: item.IsPublished, Teachers: teachers,
		},
	}
}

func mapTeacherOffer(item entity.TeacherOffer) teacherOfferResource {
	return teacherOfferResource{
		Type: "teacherOffers", ID: strconv.FormatInt(item.ID, 10),
		Attributes: teacherOfferAttributes{
			OfferID: strconv.FormatInt(item.OfferID, 10), TeacherProfileID: strconv.FormatInt(item.TeacherProfileID, 10),
			TeacherDisplayName: item.TeacherDisplayName, DurationMinutes: item.DurationMinutes,
			PriceRubles: item.PriceRubles, IsPublished: item.IsPublished,
		},
	}
}

func writeOfferError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_input", "Invalid offer", err.Error())
	case errors.Is(err, repo.ErrNotFound):
		apiresponse.WriteError(w, http.StatusNotFound, "not_found", "Offer or teacher not found", "")
	case errors.Is(err, usecase.ErrConflict), errors.Is(err, repo.ErrConflict):
		apiresponse.WriteError(w, http.StatusConflict, "conflict", "Teacher is already assigned", "")
	default:
		slog.Error("Offer request failed", slog.Any("error", err))
		apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
	}
}
