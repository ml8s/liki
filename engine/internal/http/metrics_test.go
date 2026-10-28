package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"liki-engine/internal/observability"
)

func TestMetricsHandlerTokenBoundaries(t *testing.T) {
	target := "http://engine.test/metrics"

	request := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	MetricsAuthMiddleware("", MetricsHandler()).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unset token: status = %d, want 404", recorder.Code)
	}

	handler := MetricsAuthMiddleware("metrics-secret", MetricsHandler())
	request = httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Authorization", "Bearer wrong")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("wrong token: status = %d, want 403", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Authorization", "Bearer metrics-secret")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("valid token: status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, "text/plain; version=0.0.4") {
		t.Fatalf("metrics content type = %q", got)
	}
}

func TestMetricsMiddlewareRecordsStatusClass(t *testing.T) {
	observability.ResetMetrics()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	wrapped := MetricsMiddleware("/mcp/{domain}", next)
	request := httptest.NewRequest(http.MethodPost, "/mcp/bazi", nil)
	recorder := httptest.NewRecorder()
	wrapped.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}

	text := string(observability.RenderMetrics())
	for _, fragment := range []string{
		`route="/mcp/{domain}"`,
		`method="POST"`,
		`status_class="5xx"`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("metrics missing %q:\n%s", fragment, text)
		}
	}
}
