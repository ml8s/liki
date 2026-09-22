package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter()
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
		return
	}
	if rl.entries == nil {
		t.Error("entries map should be initialized")
	}
	rl.Stop()
}

func TestRateLimiter_GetLimiter_NewIP(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	lim := rl.getLimiter("1.2.3.4", 1, 5)
	if lim == nil {
		t.Fatal("getLimiter returned nil for new IP")
	}
	if lim.Burst() != 5 {
		t.Errorf("burst = %d, want 5", lim.Burst())
	}
}

func TestRateLimiter_GetLimiter_ExistingIP(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	// Same (IP, rate, burst) returns the same limiter.
	first := rl.getLimiter("1.2.3.4", 1, 5)
	same := rl.getLimiter("1.2.3.4", 1, 5)
	if first != same {
		t.Error("getLimiter should return same limiter for same IP, rate, and burst")
	}

	// Different rate/burst creates a separate bucket.
	diff := rl.getLimiter("1.2.3.4", 10, 20)
	if first == diff {
		t.Error("getLimiter should return a different limiter when rate/burst differ")
	}
	if diff.Burst() != 20 {
		t.Errorf("burst = %d, want 20", diff.Burst())
	}
}

func TestRateLimiter_Wrap_Allow(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	called := false
	handler := rl.Wrap(100, 10, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler(w, r)

	if !called {
		t.Error("wrapped handler should be called when rate limit allows")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRateLimiter_Wrap_Blocked(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	// Allow only 1 request total, burst 0
	called := false
	handler := rl.Wrap(0, 1, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"

	// First request allowed (burst=1)
	w1 := httptest.NewRecorder()
	handler(w1, r)
	if !called {
		t.Error("wrapped handler should have been called")
	}
	if w1.Code != http.StatusOK {
		t.Fatalf("first request should be allowed, got status %d", w1.Code)
	}

	// Second request blocked (token depleted, rate=0 prevents refill)
	w2 := httptest.NewRecorder()
	handler(w2, r)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w2.Code)
	}
	if got := w2.Header().Get("Retry-After"); got != "60" {
		t.Errorf("Retry-After = %q, want 60", got)
	}
	body := w2.Body.String()
	if !strings.Contains(body, "rate_limited") {
		t.Errorf("body should contain rate_limited: %s", body)
	}
}

func TestRateLimiter_Wrap_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	count := 0
	handler := rl.Wrap(rate.Limit(0), 1, func(w http.ResponseWriter, r *http.Request) {
		count++
	})

	for _, ip := range []string{"1.1.1.1:1", "2.2.2.2:2", "3.3.3.3:3"} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler(w, r)
	}

	if count != 3 {
		t.Errorf("count = %d, want 3 (each IP gets its own bucket)", count)
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter()
	rl.Stop()
	// Verify Stop is idempotent (second Stop shouldn't panic).
	rl.Stop()
}
