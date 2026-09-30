package restapi

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
)

type statusResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
	Time     string `json:"time,omitempty"`
}

func NewRouter(database *sql.DB, log *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.Recover(log))

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
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
	})

	return router
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
