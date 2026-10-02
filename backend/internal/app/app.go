package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi"
	"github.com/maksimovyuriy/tutorina/backend/internal/lib/logger"
	offerrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/offer"
	sessionrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/session"
	teacherprofilerepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/teacherprofile"
	userrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/user"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/localphotos"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/postgres"
	authusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/auth"
	offerusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/offer"
	teacherprofileusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/teacherprofile"
)

func Run() error {
	appContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.App.Env)
	slog.SetDefault(log)

	database, err := postgres.New(cfg.DB)
	if err != nil {
		return err
	}
	defer database.Close()

	users := userrepo.New(database)
	sessions := sessionrepo.New(database)
	auth := authusecase.New(users, sessions, cfg.Auth.SessionTTL)
	photos, err := localphotos.New(cfg.Media.TeacherPhotosPath)
	if err != nil {
		return err
	}
	teacherProfiles := teacherprofileusecase.New(teacherprofilerepo.New(database), photos, log)
	offers := offerusecase.New(offerrepo.New(database))
	server := restapi.NewServer(cfg.HTTP, restapi.NewRouter(database, auth, teacherProfiles, offers, photos, cfg.Auth, log), log)
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("API started", slog.String("address", server.Addr))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	case <-appContext.Done():
		log.Info("API shutting down")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return err
	}
	log.Info("API stopped")
	return nil
}
