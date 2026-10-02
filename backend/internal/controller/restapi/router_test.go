package restapi

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
)

func TestHealth(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(&sql.DB{}, authStub{}, nil, nil, nil, nil, config.AuthConfig{}, log)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
}

type authStub struct{}

func (authStub) Login(context.Context, string, string) (entity.Session, error) {
	return entity.Session{}, nil
}

func (authStub) Authenticate(context.Context, string) (entity.User, error) {
	return entity.User{}, nil
}

func (authStub) ChangePassword(context.Context, int64, string, string) error { return nil }

func (authStub) Logout(context.Context, string) error { return nil }
