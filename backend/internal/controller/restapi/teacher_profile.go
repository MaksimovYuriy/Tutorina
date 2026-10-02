package restapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/localphotos"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type TeacherProfileService interface {
	ListPublic(context.Context) ([]entity.TeacherProfile, error)
	ListAll(context.Context) ([]entity.TeacherProfile, error)
	GetMine(context.Context, int64) (entity.TeacherProfile, error)
	UpdateMine(context.Context, int64, entity.TeacherProfile) (entity.TeacherProfile, error)
	ReplaceMinePhoto(context.Context, int64, io.Reader) (entity.TeacherProfile, error)
	RemoveMinePhoto(context.Context, int64) (entity.TeacherProfile, error)
	ResetPassword(context.Context, int64, string) error
	Create(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error)
	Update(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error)
	ReplacePhoto(context.Context, int64, io.Reader) (entity.TeacherProfile, error)
	RemovePhoto(context.Context, int64) (entity.TeacherProfile, error)
	CreateAccount(context.Context, int64, string, string) (entity.TeacherProfile, error)
	Archive(context.Context, int64) error
}

type teacherProfileController struct{ service TeacherProfileService }
type teacherProfileRequest struct {
	UserID      *int64 `json:"userId"`
	DisplayName string `json:"displayName"`
	Education   string `json:"education"`
	Experience  string `json:"experience"`
	Approach    string `json:"approach"`
	PhotoURL    string `json:"photoUrl"`
	IsPublished bool   `json:"isPublished"`
}
type teacherAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type teacherProfileResource struct {
	Type       string                `json:"type"`
	ID         string                `json:"id"`
	Attributes teacherProfileRequest `json:"attributes"`
}
type publicTeacherProfileAttributes struct {
	DisplayName string `json:"displayName"`
	Education   string `json:"education"`
	Experience  string `json:"experience"`
	Approach    string `json:"approach"`
	PhotoURL    string `json:"photoUrl"`
}

type teacherPasswordResetRequest struct {
	Password string `json:"password"`
}

func (c teacherProfileController) getMine(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	profile, err := c.service.GetMine(r.Context(), user.ID)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) updateMine(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	var body teacherProfileRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	profile, err := c.service.UpdateMine(r.Context(), user.ID, body.entity(0))
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) replaceMinePhoto(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	file, cleanup, err := teacherPhotoFile(w, r)
	if err != nil {
		return
	}
	defer cleanup()
	profile, err := c.service.ReplaceMinePhoto(r.Context(), user.ID, file)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) removeMinePhoto(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	profile, err := c.service.RemoveMinePhoto(r.Context(), user.ID)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) resetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid teacher id", "")
		return
	}
	var body teacherPasswordResetRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	if err := c.service.ResetPassword(r.Context(), id, body.Password); err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func teacherPhotoFile(w http.ResponseWriter, r *http.Request) (io.ReadCloser, func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, localphotos.MaxFileSize+(256<<10))
	if err := r.ParseMultipartForm(256 << 10); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_photo", "Invalid teacher photo", "Use a JPEG, PNG or WebP file up to 5 MB")
		return nil, func() {}, err
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		r.MultipartForm.RemoveAll()
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_photo", "Invalid teacher photo", "Multipart field photo is required")
		return nil, func() {}, err
	}
	cleanup := func() {
		file.Close()
		r.MultipartForm.RemoveAll()
	}
	return file, cleanup, nil
}
func (c teacherProfileController) listPublic(w http.ResponseWriter, r *http.Request) {
	c.list(w, r, true)
}
func (c teacherProfileController) listAll(w http.ResponseWriter, r *http.Request) {
	c.list(w, r, false)
}
func (c teacherProfileController) list(w http.ResponseWriter, r *http.Request, public bool) {
	var profiles []entity.TeacherProfile
	var err error
	if public {
		profiles, err = c.service.ListPublic(r.Context())
	} else {
		profiles, err = c.service.ListAll(r.Context())
	}
	if err != nil {
		apiresponse.WriteError(w, 500, "internal_error", "Internal server error", "")
		return
	}
	if public {
		data := make([]map[string]any, 0, len(profiles))
		for _, profile := range profiles {
			data = append(data, map[string]any{"type": "teacherProfiles", "id": strconv.FormatInt(profile.ID, 10), "attributes": publicTeacherProfileAttributes{DisplayName: profile.DisplayName, Education: profile.Education, Experience: profile.Experience, Approach: profile.Approach, PhotoURL: profile.PhotoURL}})
		}
		apiresponse.Write(w, http.StatusOK, map[string]any{"data": data})
		return
	}
	data := make([]teacherProfileResource, 0, len(profiles))
	for _, profile := range profiles {
		data = append(data, mapTeacherProfile(profile))
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": data})
}
func (c teacherProfileController) create(w http.ResponseWriter, r *http.Request) {
	var body teacherProfileRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, 400, "invalid_request", "Invalid request", err.Error())
		return
	}
	created, err := c.service.Create(r.Context(), body.entity(0))
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapTeacherProfile(created)})
}
func (c teacherProfileController) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, 400, "invalid_id", "Invalid teacher id", "")
		return
	}
	var body teacherProfileRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, 400, "invalid_request", "Invalid request", err.Error())
		return
	}
	updated, err := c.service.Update(r.Context(), body.entity(id))
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(updated)})
}
func (c teacherProfileController) replacePhoto(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid teacher id", "")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, localphotos.MaxFileSize+(256<<10))
	if err := r.ParseMultipartForm(256 << 10); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_photo", "Invalid teacher photo", "Use a JPEG, PNG or WebP file up to 5 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, _, err := r.FormFile("photo")
	if err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_photo", "Invalid teacher photo", "Multipart field photo is required")
		return
	}
	defer file.Close()

	profile, err := c.service.ReplacePhoto(r.Context(), id, file)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) removePhoto(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid teacher id", "")
		return
	}
	profile, err := c.service.RemovePhoto(r.Context(), id)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusOK, map[string]any{"data": mapTeacherProfile(profile)})
}

func (c teacherProfileController) createAccount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, 400, "invalid_id", "Invalid teacher id", "")
		return
	}
	var body teacherAccountRequest
	if err := decodeJSON(w, r, &body); err != nil {
		apiresponse.WriteError(w, 400, "invalid_request", "Invalid request", err.Error())
		return
	}
	profile, err := c.service.CreateAccount(r.Context(), id, body.Email, body.Password)
	if err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	apiresponse.Write(w, http.StatusCreated, map[string]any{"data": mapTeacherProfile(profile)})
}
func (c teacherProfileController) archive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		apiresponse.WriteError(w, 400, "invalid_id", "Invalid teacher id", "")
		return
	}
	if err := c.service.Archive(r.Context(), id); err != nil {
		writeTeacherProfileError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (request teacherProfileRequest) entity(id int64) entity.TeacherProfile {
	return entity.TeacherProfile{ID: id, UserID: request.UserID, DisplayName: request.DisplayName, Education: request.Education, Experience: request.Experience, Approach: request.Approach, PhotoURL: request.PhotoURL, IsPublished: request.IsPublished}
}
func mapTeacherProfile(profile entity.TeacherProfile) teacherProfileResource {
	return teacherProfileResource{Type: "teacherProfiles", ID: strconv.FormatInt(profile.ID, 10), Attributes: teacherProfileRequest{UserID: profile.UserID, DisplayName: profile.DisplayName, Education: profile.Education, Experience: profile.Experience, Approach: profile.Approach, PhotoURL: profile.PhotoURL, IsPublished: profile.IsPublished}}
}
func writeTeacherProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		apiresponse.WriteError(w, 400, "invalid_input", "Invalid teacher profile", err.Error())
	case errors.Is(err, repo.ErrNotFound):
		apiresponse.WriteError(w, 404, "not_found", "Teacher profile not found", "")
	case errors.Is(err, usecase.ErrConflict):
		apiresponse.WriteError(w, 409, "conflict", "Account cannot be linked", "Email is already used or the profile already has an account")
	default:
		slog.Error("Teacher profile request failed", slog.Any("error", err))
		apiresponse.WriteError(w, 500, "internal_error", "Internal server error", "")
	}
}
