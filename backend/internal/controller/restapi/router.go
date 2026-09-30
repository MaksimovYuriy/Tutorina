package restapi

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
)

type statusResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
	Time     string `json:"time,omitempty"`
}

func NewRouter(database *sql.DB, authService AuthService, authConfig config.AuthConfig, log *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.Recover(log))

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
	auth := newAuthController(authService, authConfig.CookieSecure)
	router.Route("/v1", func(router chi.Router) {
		router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			if err := database.PingContext(r.Context()); err != nil {
				log.Error("Database health check failed", slog.Any("error", err))
				writeJSON(w, http.StatusServiceUnavailable, statusResponse{Status: "unavailable"})
				return
			}
			writeJSON(w, http.StatusOK, statusResponse{
				Status:   "ok",
				Database: "connected",
				Time:     time.Now().UTC().Format(time.RFC3339),
			})
		})

		router.Post("/auth/sessions", auth.login)
		router.Delete("/auth/session", auth.logout)
		router.With(middleware.RequireAuth(authService, log)).Get("/auth/me", auth.me)
	})

	return router
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	apiresponse.Write(w, status, value)
}
