package restapi

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
)

func NewRouter(database *sql.DB, authService AuthService, slots SlotService, authConfig config.AuthConfig, log *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.Recover(log))
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	auth := newAuthController(authService, authConfig.CookieSecure)
	controller := slotController{slots}
	router.Route("/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			if err := database.PingContext(r.Context()); err != nil {
				writeJSON(w, 503, map[string]string{"status": "unavailable"})
				return
			}
			writeJSON(w, 200, map[string]string{"status": "ok", "database": "connected"})
		})
		r.Post("/auth/sessions", auth.login)
		r.Delete("/auth/session", auth.logout)
		r.With(middleware.RequireAuth(authService, log)).Get("/auth/me", auth.me)
		r.With(middleware.RequireAuth(authService, log)).Put("/auth/password", auth.changePassword)
		r.Get("/slots", controller.listPublic)
		r.Route("/admin/slots", func(r chi.Router) {
			r.Use(middleware.RequireAuth(authService, log))
			r.Use(middleware.RequireRole(entity.RoleAdmin))
			r.Get("/", controller.listAdmin)
			r.Post("/", controller.create)
			r.Put("/{id}", controller.update)
			r.Delete("/{id}", controller.delete)
		})
	})
	return router
}
func writeJSON(w http.ResponseWriter, status int, value any) { apiresponse.Write(w, status, value) }
