package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const SessionCookieName = "tutorina_session"

type AuthService interface {
	Authenticate(context.Context, string) (entity.User, error)
}

type userContextKey struct{}

func RequireAuth(service AuthService, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
				return
			}
			user, err := service.Authenticate(r.Context(), cookie.Value)
			if errors.Is(err, usecase.ErrUnauthorized) {
				apiresponse.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
				return
			}
			if err != nil {
				log.Error("Authentication failed", slog.Any("error", err))
				apiresponse.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal server error", "")
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentUser(ctx context.Context) (entity.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(entity.User)
	return user, ok
}
