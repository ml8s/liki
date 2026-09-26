package http

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
)

var prodOrigins = map[string]bool{
	"https://liki.hk": true,
}

var devOrigins = map[string]bool{
	"http://localhost:8080": true,
	"http://localhost:8081": true,
}

// CORSMiddleware adds permissive CORS headers for allowed origins and handles OPTIONS preflight.
func CORSMiddleware(devMode bool, next http.Handler) http.Handler {
	allowedOrigins := configuredOrigins(devMode)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := allowedOrigins[origin]

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, MCP-Protocol-Version, Mcp-Method, Mcp-Name, Authorization",
			)
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions && origin != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func configuredOrigins(devMode bool) map[string]bool {
	allowed := make(map[string]bool, len(prodOrigins)+len(devOrigins))
	for origin := range prodOrigins {
		allowed[origin] = true
	}
	if devMode {
		for origin := range devOrigins {
			allowed[origin] = true
		}
	}
	if configured := strings.TrimSpace(os.Getenv("LIKI_ALLOWED_ORIGINS")); configured != "" {
		allowed = make(map[string]bool)
		for _, origin := range strings.Split(configured, ",") {
			origin = strings.TrimSpace(origin)
			if validHTTPOrigin(origin) {
				allowed[origin] = true
			}
		}
	}
	return allowed
}

func validHTTPOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return false
	}
	return (parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" && parsed.Fragment == ""
}

// SecurityHeaders adds X-Content-Type-Options: nosniff and X-Frame-Options: DENY.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CSP is managed by Caddy for static assets; API responses are JSON only.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// Recover catches panics in downstream handlers and returns 500 instead of crashing.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				slog.Error("panic recovered", "panic", rv, "stack", string(debug.Stack()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// BodyLimit limits request body size to 1 MB.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
		next.ServeHTTP(w, r)
	})
}
