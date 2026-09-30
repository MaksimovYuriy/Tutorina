package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const maximumJSONBodySize = 1 << 20

type AuthService interface {
	Login(context.Context, string, string) (entity.Session, error)
	Authenticate(context.Context, string) (entity.User, error)
	Logout(context.Context, string) error
}

type authController struct {
	service      AuthService
	cookieSecure bool
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userDocument struct {
	Data userResource `json:"data"`
}

type userResource struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Attributes userAttributes `json:"attributes"`
}

type userAttributes struct {
	Email string        `json:"email"`
	Roles []entity.Role `json:"roles"`
}

func newAuthController(service AuthService, cookieSecure bool) *authController {
	return &authController{service: service, cookieSecure: cookieSecure}
}

func (controller *authController) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(w, r, &request); err != nil {
		apiresponse.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request", err.Error())
		return
	}
	session, err := controller.service.Login(r.Context(), request.Email, request.Password)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		apiresponse.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password", "")
		return
	}
	if err != nil {
		apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
		return
	}
	controller.setSessionCookie(w, session.Token, session.ExpiresAt)
	w.WriteHeader(http.StatusNoContent)
}

func (controller *authController) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(middleware.SessionCookieName); err == nil {
		if err := controller.service.Logout(r.Context(), cookie.Value); err != nil {
			apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
			return
		}
	}
	controller.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (controller *authController) me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.CurrentUser(r.Context())
	if !ok {
		apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}
	apiresponse.Write(w, http.StatusOK, userDocument{Data: userResource{
		Type: "users",
		ID:   strconv.FormatInt(user.ID, 10),
		Attributes: userAttributes{
			Email: user.Email,
			Roles: user.Roles,
		},
	}})
}

func (controller *authController) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   controller.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (controller *authController) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   controller.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
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
