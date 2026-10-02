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
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
)

type statusResponse struct {
	Status   string `json:"status"`
	Database string `json:"database,omitempty"`
	Time     string `json:"time,omitempty"`
}

func NewRouter(database *sql.DB, authService AuthService, teacherProfiles TeacherProfileService, offers OfferService, lessons LessonService, teacherPhotos http.Handler, authConfig config.AuthConfig, log *slog.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.Recover(log))

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	})
	auth := newAuthController(authService, authConfig.CookieSecure)
	teachers := teacherProfileController{service: teacherProfiles}
	offerController := offerController{service: offers}
	lessonController := lessonController{service: lessons}
	router.Route("/v1", func(router chi.Router) {
		if teacherPhotos != nil {
			router.Handle("/media/teacher-photos/*", http.StripPrefix("/v1/media/teacher-photos/", teacherPhotos))
		}
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
		router.With(middleware.RequireAuth(authService, log)).Put("/auth/password", auth.changePassword)
		router.Route("/teacher", func(router chi.Router) {
			router.Use(middleware.RequireAuth(authService, log))
			router.Use(middleware.RequireRole(entity.RoleTeacher))
			router.Get("/profile", teachers.getMine)
			router.Put("/profile", teachers.updateMine)
			router.Put("/profile/photo", teachers.replaceMinePhoto)
			router.Delete("/profile/photo", teachers.removeMinePhoto)
			if offers != nil {
				router.Get("/offers", offerController.listMine)
			}
			if lessons != nil {
				router.Get("/lessons", lessonController.listMine)
				router.Post("/lessons", lessonController.createMine)
				router.Put("/lessons/{id}", lessonController.updateMine)
				router.Delete("/lessons/{id}", lessonController.archiveMine)
			}
		})
		router.Get("/teachers", teachers.listPublic)
		if offers != nil {
			router.Get("/offers", offerController.listPublic)
		}
		if lessons != nil {
			router.Get("/lessons", lessonController.listPublic)
		}
		router.Route("/admin/teachers", func(router chi.Router) {
			router.Use(middleware.RequireAuth(authService, log))
			router.Use(middleware.RequireRole(entity.RoleAdmin))
			router.Get("/", teachers.listAll)
			router.Post("/", teachers.create)
			router.Put("/{id}", teachers.update)
			router.Put("/{id}/photo", teachers.replacePhoto)
			router.Delete("/{id}/photo", teachers.removePhoto)
			router.Post("/{id}/account", teachers.createAccount)
			router.Put("/{id}/password", teachers.resetPassword)
			router.Delete("/{id}", teachers.archive)
		})
		if lessons != nil {
			router.Route("/admin/lessons", func(router chi.Router) {
				router.Use(middleware.RequireAuth(authService, log))
				router.Use(middleware.RequireRole(entity.RoleAdmin))
				router.Get("/", lessonController.listAll)
				router.Post("/", lessonController.createAdmin)
				router.Put("/{id}", lessonController.updateAdmin)
				router.Delete("/{id}", lessonController.archiveAdmin)
			})
		}
		if offers != nil {
			router.Route("/admin/offers", func(router chi.Router) {
				router.Use(middleware.RequireAuth(authService, log))
				router.Use(middleware.RequireRole(entity.RoleAdmin))
				router.Get("/", offerController.listAll)
				router.Post("/", offerController.create)
				router.Put("/{id}", offerController.update)
				router.Delete("/{id}", offerController.archive)
				router.Post("/{id}/teachers", offerController.assignTeacher)
			})
			router.Route("/admin/teacher-offers", func(router chi.Router) {
				router.Use(middleware.RequireAuth(authService, log))
				router.Use(middleware.RequireRole(entity.RoleAdmin))
				router.Put("/{id}", offerController.updateAssignment)
				router.Delete("/{id}", offerController.removeAssignment)
			})
		}
	})

	return router
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	apiresponse.Write(w, status, value)
}
