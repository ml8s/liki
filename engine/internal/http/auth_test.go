package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMCPAuthMiddlewareDisabledAllowsRequests(t *testing.T) {
	called := false
	handler := MCPAuthMiddleware("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/mcp", nil))

	if !called {
		t.Fatal("disabled auth middleware should call downstream handler")
	}
}

func TestMCPAuthMiddlewareRejectsMissingOrInvalidBearer(t *testing.T) {
	handler := MCPAuthMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("downstream handler must not be called")
	}))

	for _, authorization := range []string{"", "Bearer wrong", "Basic secret"} {
		request := httptest.NewRequest("POST", "/mcp", nil)
		request.Header.Set("Authorization", authorization)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("Authorization %q: status = %d, want 401", authorization, recorder.Code)
		}
		if recorder.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("Authorization %q: missing WWW-Authenticate: Bearer", authorization)
		}
	}
}

func TestMCPAuthMiddlewareAcceptsBearerAndPublicProbes(t *testing.T) {
	called := 0
	handler := MCPAuthMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
	}))

	authorized := httptest.NewRequest("POST", "/mcp", nil)
	authorized.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, authorized)
	if recorder.Code != http.StatusOK || called != 1 {
		t.Fatalf("authorized status = %d, called = %d; want 200/1", recorder.Code, called)
	}

	for _, method := range []string{http.MethodGet, http.MethodOptions} {
		request := httptest.NewRequest(method, "/health", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s /health status = %d, want 200", method, recorder.Code)
		}
	}
}

func TestMCPAuthMiddlewareRejectsMissingToken(t *testing.T) {
	handler := MCPAuthMiddleware(
		"secret",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest("POST", "/mcp", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}
