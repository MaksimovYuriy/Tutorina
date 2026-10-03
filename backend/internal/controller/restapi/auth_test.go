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

type httpAuthStub struct {
	session                       entity.Session
	loginError, authenticateError error
	receivedKey                   string
}

func (s *httpAuthStub) Login(_ context.Context, key string) (entity.Session, error) {
	s.receivedKey = key
	return s.session, s.loginError
}
func (s *httpAuthStub) Authenticate(context.Context, string) error { return s.authenticateError }
func (s *httpAuthStub) Logout(context.Context, string) error       { return nil }
func testRouter(service AuthService, secure bool) http.Handler {
	return NewRouter(nil, service, nil, nil, config.AuthConfig{CookieSecure: secure}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}
func TestKeyLoginSetsSecureSessionCookie(t *testing.T) {
	s := &httpAuthStub{session: entity.Session{Token: "session-token", ExpiresAt: time.Now().Add(time.Hour)}}
	w := httptest.NewRecorder()
	testRouter(s, true).ServeHTTP(w, httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{"key":"access-key"}`)))
	if w.Code != 204 || s.receivedKey != "access-key" {
		t.Fatalf("status=%d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != middleware.SessionCookieName || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatal("insecure session cookie")
	}
	if strings.Contains(w.Body.String(), "access-key") {
		t.Fatal("key echoed in response")
	}
}
func TestSessionRequiresAuthentication(t *testing.T) {
	router := testRouter(&httpAuthStub{authenticateError: usecase.ErrUnauthorized}, false)
	for _, cookie := range []bool{false, true} {
		r := httptest.NewRequest("GET", "/v1/auth/session", nil)
		if cookie {
			r.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "bad"})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatal(w.Code)
		}
	}
}
func TestAuthenticatedSessionReturnsNoIdentity(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/auth/session", nil)
	r.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "session"})
	w := httptest.NewRecorder()
	testRouter(&httpAuthStub{}, false).ServeHTTP(w, r)
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatal(w.Code)
	}
}
func TestOldCredentialsAreRejected(t *testing.T) {
	w := httptest.NewRecorder()
	testRouter(&httpAuthStub{}, false).ServeHTTP(w, httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{"username":"admin","password":"secret"}`)))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
