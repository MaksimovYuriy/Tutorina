package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoginLimitWindow(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	var calls int
	handler := loginLimit(10, time.Minute, func() time.Time { return now })(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls++; w.WriteHeader(204) }))
	request := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", "/auth/sessions", nil))
		return w
	}
	for i := 0; i < 10; i++ {
		if w := request(); w.Code != 204 {
			t.Fatal(w.Code)
		}
	}
	now = now.Add(1500 * time.Millisecond)
	if w := request(); w.Code != 429 || w.Header().Get("Retry-After") != "59" {
		t.Fatalf("status=%d retry=%s", w.Code, w.Header().Get("Retry-After"))
	}
	if calls != 10 {
		t.Fatal("limited request reached login")
	}
	now = now.Add(58500 * time.Millisecond)
	if w := request(); w.Code != 204 {
		t.Fatal("window did not reset")
	}
}

func TestLoginLimitConcurrentRequests(t *testing.T) {
	var allowed atomic.Int32
	now := time.Now()
	handler := loginLimit(10, time.Minute, func() time.Time { return now })(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { allowed.Add(1) }))
	var group sync.WaitGroup
	for i := 0; i < 100; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/auth/sessions", nil))
		}()
	}
	group.Wait()
	if allowed.Load() != 10 {
		t.Fatal(allowed.Load())
	}
}
