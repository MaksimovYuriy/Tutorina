package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/controller/restapi/apiresponse"
)

// LoginLimit bounds all login requests for this single-administrator application.
// It stores no client identifiers and is shared by requests within one process.
func LoginLimit() func(http.Handler) http.Handler {
	return loginLimit(10, time.Minute, time.Now)
}

func loginLimit(maximum int, window time.Duration, now func() time.Time) func(http.Handler) http.Handler {
	var mu sync.Mutex
	var reset time.Time
	var count int
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := now()
			mu.Lock()
			if !current.Before(reset) {
				reset = current.Add(window)
				count = 0
			}
			allowed := count < maximum
			retry := reset.Sub(current)
			if allowed {
				count++
			}
			mu.Unlock()
			if !allowed {
				seconds := int((retry + time.Second - 1) / time.Second)
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
				w.Header().Set("Cache-Control", "no-store")
				apiresponse.WriteError(w, http.StatusTooManyRequests, "login_rate_limit", "Слишком много попыток входа. Подождите минуту и попробуйте снова.", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
