package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"liki-engine/internal/observability"
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
				"Content-Type, MCP-Protocol-Version, Mcp-Method, Mcp-Name",
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
	if configured := strings.TrimSpace(os.Getenv("LIKI_ENGINE_ALLOWED_ORIGINS")); configured != "" {
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

// UnsupportedMethods rejects protocol methods that this stateless server does
// not implement. MCP 2026 requires the JSON-RPC method error to ride on HTTP
// 404, rather than letting the SDK turn a listener into an idle SSE stream.
func UnsupportedMethods(next http.Handler, methods ...string) http.Handler {
	rejected := make(map[string]bool, len(methods))
	for _, method := range methods {
		rejected[method] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unable to read request", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		var envelope struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		_ = json.Unmarshal(body, &envelope)
		if rejected[envelope.Method] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			response, _ := json.Marshal(map[string]any{
				"jsonrpc": "2.0",
				"id":      envelope.ID,
				"error": map[string]any{
					"code":    -32601,
					"message": "method not found: " + envelope.Method,
				},
			})
			_, _ = w.Write(response)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MetricsMiddleware records route-template, method, and status-class metrics.
// Raw paths are never used as metric labels.
func MetricsMiddleware(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &metricsResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		observability.ObserveHTTP(
			route,
			r.Method,
			recorder.status,
			time.Since(started).Seconds(),
		)
	})
}

type metricsResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *metricsResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *metricsResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

// MetricsHandler renders dependency-free Prometheus text exposition.
func MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write(observability.RenderMetrics())
	})
}
