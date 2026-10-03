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
	keyrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/accesskey"
	directionrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/direction"
	sessionrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/session"
	slotrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/slot"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/postgres"
	authusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/auth"
	directionusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/direction"
	slotusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/slot"
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

	keys := keyrepo.New(database)
	sessions := sessionrepo.New(database)
	auth := authusecase.New(keys, sessions, cfg.Auth.SessionTTL)
	slots := slotusecase.New(slotrepo.New(database))
	directions := directionusecase.New(directionrepo.New(database))
	server := restapi.NewServer(cfg.HTTP, restapi.NewRouter(database, auth, slots, directions, cfg.Auth, log), log)
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
