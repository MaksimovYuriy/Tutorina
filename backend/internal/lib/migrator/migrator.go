package migrator

import (
	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/lib/logger"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/postgres"
	"github.com/pressly/goose/v3"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.App.Env)

	database, err := postgres.New(cfg.DB)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(database, "./migrations"); err != nil {
		return err
	}
	log.Info("Database migrations applied")
	return nil
}
