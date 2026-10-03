package restapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
)

type slotServiceStub struct {
	public bool
	saved  bool
}

func (s *slotServiceStub) List(_ context.Context, public bool) ([]entity.Slot, error) {
	s.public = public
	return []entity.Slot{{ID: 1, Title: "Математика", StartsAt: time.Now().Add(time.Hour), EndsAt: time.Now().Add(2 * time.Hour), Capacity: 4, Occupied: 1, Published: true, Status: "planned"}}, nil
}
func (s *slotServiceStub) Save(_ context.Context, v entity.Slot) (entity.Slot, error) {
	s.saved = true
	return v, nil
}
func (s *slotServiceStub) Delete(context.Context, int64) error { return nil }
func slotRouter(auth AuthService, s SlotService) http.Handler {
	return NewRouter(nil, auth, s, config.AuthConfig{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}
func TestPublicSlotsExposeOnlyAvailability(t *testing.T) {
	s := &slotServiceStub{}
	w := httptest.NewRecorder()
	slotRouter(&httpAuthStub{}, s).ServeHTTP(w, httptest.NewRequest("GET", "/v1/slots", nil))
	if w.Code != 200 || !s.public {
		t.Fatalf("status=%d public=%v", w.Code, s.public)
	}
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) != 1 || body.Data[0]["freePlaces"] != float64(3) {
		t.Fatal(w.Body.String())
	}
	for _, key := range []string{"occupied", "published", "status", "capacity"} {
		if _, exists := body.Data[0][key]; exists {
			t.Fatalf("exposed %s", key)
		}
	}
}
func TestOnlyAdministratorCanChangeSlots(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cookie bool
		roles  []entity.Role
		want   int
	}{
		{"visitor", false, nil, 401}, {"without role", true, nil, 403}, {"admin", true, []entity.Role{entity.RoleAdmin}, 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &slotServiceStub{}
			auth := &httpAuthStub{user: entity.User{ID: 1, Roles: tc.roles}}
			r := httptest.NewRequest("POST", "/v1/admin/slots/", strings.NewReader(`{"title":"Математика"}`))
			if tc.cookie {
				r.AddCookie(&http.Cookie{Name: "tutorina_session", Value: "test"})
			}
			w := httptest.NewRecorder()
			slotRouter(auth, s).ServeHTTP(w, r)
			if w.Code != tc.want || s.saved != (tc.want == 201) {
				t.Fatalf("status=%d saved=%v", w.Code, s.saved)
			}
		})
	}
}
func TestRemovedSchoolRoutesReturnNotFound(t *testing.T) {
	router := slotRouter(&httpAuthStub{}, &slotServiceStub{})
	for _, path := range []string{"/v1/teachers", "/v1/offers", "/v1/lessons", "/v1/teacher/profile", "/v1/admin/applications/"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/v1/applications", strings.NewReader(`{}`)))
	if w.Code != 404 {
		t.Fatalf("applications: %d", w.Code)
	}
}
