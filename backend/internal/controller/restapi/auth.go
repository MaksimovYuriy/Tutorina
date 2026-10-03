package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const maximumJSONBodySize = 1 << 20

type AuthService interface {
	Login(context.Context, string) (entity.Session, error)
	Authenticate(context.Context, string) error
	Logout(context.Context, string) error
}

type authController struct {
	service AuthService
}

type loginRequest struct {
	Key string `json:"key"`
}

func newAuthController(service AuthService) *authController {
	return &authController{service: service}
}

func (controller *authController) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", "")
		return
	}
	session, err := controller.service.Login(r.Context(), request.Key)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		apiresponse.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный ключ доступа", "")
		return
	}
	if err != nil {
		apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{"token": session.Token, "expiresAt": session.ExpiresAt})
}

func (controller *authController) logout(w http.ResponseWriter, r *http.Request) {
	if token := middleware.SessionToken(r); token != "" {
		if err := controller.service.Logout(r.Context(), token); err != nil {
			apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maximumJSONBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain one JSON object")
		}
		return err
	}
	return nil
}
