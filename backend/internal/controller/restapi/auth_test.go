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
func testRouter(service AuthService) http.Handler {
	return NewRouter(nil, service, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
}
func TestKeyLoginReturnsTemporaryToken(t *testing.T) {
	s := &httpAuthStub{session: entity.Session{Token: "session-token", ExpiresAt: time.Now().Add(time.Hour)}}
	w := httptest.NewRecorder()
	testRouter(s).ServeHTTP(w, httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{"key":"access-key"}`)))
	if w.Code != 200 || s.receivedKey != "access-key" {
		t.Fatalf("status=%d", w.Code)
	}
	if len(w.Result().Cookies()) != 0 || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), "session-token") {
		t.Fatal("invalid session response")
	}
	if strings.Contains(w.Body.String(), "access-key") {
		t.Fatal("key echoed in response")
	}
}
func TestSessionRequiresAuthentication(t *testing.T) {
	router := testRouter(&httpAuthStub{authenticateError: usecase.ErrUnauthorized})
	for _, cookie := range []bool{false, true} {
		r := httptest.NewRequest("GET", "/v1/auth/session", nil)
		if cookie {
			r.Header.Set("Authorization", "Bearer bad")
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
	r.Header.Set("Authorization", "Bearer session")
	w := httptest.NewRecorder()
	testRouter(&httpAuthStub{}).ServeHTTP(w, r)
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatal(w.Code)
	}
}
func TestOldCredentialsAreRejected(t *testing.T) {
	w := httptest.NewRecorder()
	testRouter(&httpAuthStub{}).ServeHTTP(w, httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{"username":"admin","password":"secret"}`)))
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}

func TestLegacyCookieDoesNotAuthorize(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/auth/session", nil)
	r.AddCookie(&http.Cookie{Name: "tutorina_session", Value: "previous-session"})
	w := httptest.NewRecorder()
	testRouter(&httpAuthStub{}).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}

func TestLoginRateLimitIncludesMalformedRequests(t *testing.T) {
	router := testRouter(&httpAuthStub{})
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{`)))
		if w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/auth/sessions", strings.NewReader(`{"key":"key"}`))
	r.Header.Set("X-Forwarded-For", "203.0.113.10")
	router.ServeHTTP(w, r)
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatal("login limit affected health")
	}
}
