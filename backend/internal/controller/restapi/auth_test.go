package restapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/middleware"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestLoginSetsSecureSessionCookie(t *testing.T) {
	service := &httpAuthStub{session: entity.Session{Token: "secret-token", ExpiresAt: time.Now().Add(time.Hour)}}
	router := testRouter(service, true)
	request := httptest.NewRequest(http.MethodPost, "/v1/auth/sessions", strings.NewReader(`{"email":"teacher@example.com","password":"password"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != middleware.SessionCookieName || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Fatalf("session cookie = %#v", cookies)
	}
}

func TestMeRequiresAuthentication(t *testing.T) {
	router := testRouter(&httpAuthStub{authenticateError: usecase.ErrUnauthorized}, false)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
}

func TestMeReturnsRoles(t *testing.T) {
	service := &httpAuthStub{user: entity.User{ID: 7, Email: "teacher@example.com", Roles: []entity.Role{entity.RoleAdmin, entity.RoleTeacher}, IsActive: true}}
	router := testRouter(service, false)
	request := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "secret-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"roles":["admin","teacher"]`) {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func testRouter(service AuthService, cookieSecure bool) http.Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(nil, service, nil, nil, nil, config.AuthConfig{CookieSecure: cookieSecure}, log)
}

type httpAuthStub struct {
	session           entity.Session
	loginError        error
	user              entity.User
	authenticateError error
}

func (stub *httpAuthStub) Login(context.Context, string, string) (entity.Session, error) {
	return stub.session, stub.loginError
}

func (stub *httpAuthStub) Authenticate(context.Context, string) (entity.User, error) {
	return stub.user, stub.authenticateError
}

func (stub *httpAuthStub) ChangePassword(context.Context, int64, string, string) error { return nil }

func (stub *httpAuthStub) Logout(context.Context, string) error { return nil }
