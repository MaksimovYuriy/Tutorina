package restapi

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
)

func NewServer(cfg config.HTTPConfig, handler http.Handler, log *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Address, cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
}
