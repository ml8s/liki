package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ipLimiter holds the token bucket for a single IP.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// RateLimiter provides per-IP token-bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*ipLimiter
	cleanupT *time.Ticker
	done     chan struct{}
	stopOnce sync.Once
}

// NewRateLimiter creates a rate limiter that periodically cleans up idle entries.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		entries:  make(map[string]*ipLimiter),
		cleanupT: time.NewTicker(10 * time.Minute),
		done:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupT.C:
			rl.mu.Lock()
			for ip, e := range rl.entries {
				if time.Since(e.lastUsed) > 10*time.Minute {
					delete(rl.entries, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Stop stops the cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.done)
	})
	rl.cleanupT.Stop()
}

// Wrap applies per-IP rate limiting to an http.HandlerFunc.
// r is the sustained rate (requests per second), burst is the bucket size.
func (rl *RateLimiter) Wrap(r rate.Limit, burst int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ip := clientIP(req)
		limiter := rl.getLimiter(ip, r, burst)
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			respondError(w, http.StatusTooManyRequests, "rate_limited", "too many requests, please slow down")
			return
		}
		next(w, req)
	}
}

func (rl *RateLimiter) getLimiter(ip string, r rate.Limit, burst int) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := fmt.Sprintf("%s|%.0f|%d", ip, float64(r), burst)
	e, ok := rl.entries[key]
	if !ok {
		limiter := rate.NewLimiter(r, burst)
		rl.entries[key] = &ipLimiter{limiter: limiter, lastUsed: time.Now()}
		return limiter
	}
	e.lastUsed = time.Now()
	return e.limiter
}
