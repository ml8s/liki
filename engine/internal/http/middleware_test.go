package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecover_Panic(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := Recover(panicHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	// Should not crash the test.
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestRecover_NoPanic(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Recover(okHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing X-Frame-Options: DENY")
	}
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		devMode bool
		allowed bool
	}{
		{"prod origin", "https://liki.hk", false, true},
		{"unknown origin", "https://evil.com", false, false},
		{"localhost in dev", "http://localhost:8080", true, true},
		{"unknown in dev", "https://evil.com", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CORSMiddleware(tt.devMode, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			handler.ServeHTTP(w, req)

			aco := w.Header().Get("Access-Control-Allow-Origin")
			if tt.allowed && aco != tt.origin {
				t.Errorf("expected Access-Control-Allow-Origin=%q, got %q", tt.origin, aco)
			}
			if !tt.allowed && aco != "" {
				t.Errorf("expected no Access-Control-Allow-Origin, got %q", aco)
			}
		})
	}
}

func TestCORSMiddleware_OptionsPreflight(t *testing.T) {
	handler := CORSMiddleware(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS preflight")
	}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://liki.hk")
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestBodyLimit_UnderLimit(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"valid":"json"}`)
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBodyLimit_OverLimit(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("expected error reading oversized body")
		}
	}))

	body := strings.NewReader(strings.Repeat("x", 1<<20+1))
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
}
