package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func SessionToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(value, "Bearer ")
}

type AuthService interface {
	Authenticate(context.Context, string) error
}

func RequireAuth(service AuthService, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := SessionToken(r)
			if token == "" {
				apiresponse.WriteError(w, 401, "unauthorized", "Authentication required", "")
				return
			}
			err := service.Authenticate(r.Context(), token)
			if errors.Is(err, usecase.ErrUnauthorized) {
				apiresponse.WriteError(w, 401, "unauthorized", "Authentication required", "")
				return
			}
			if err != nil {
				log.Error("Authentication failed", slog.Any("error", err))
				apiresponse.WriteError(w, 500, "internal_error", "Internal server error", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
