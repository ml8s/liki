package http

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresForwardedHeaderByDefault(t *testing.T) {
	t.Setenv("LIKI_TRUSTED_PROXY_HOPS", "")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.1")

	if got := clientIP(req); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want RemoteAddr", got)
	}
}

func TestClientIPUsesLastAddressWithOneTrustedProxy(t *testing.T) {
	t.Setenv("LIKI_TRUSTED_PROXY_HOPS", "1")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 203.0.113.10")

	if got := clientIP(req); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want 203.0.113.10", got)
	}
}

func TestClientIPHandlesConfiguredProxyHops(t *testing.T) {
	t.Setenv("LIKI_TRUSTED_PROXY_HOPS", "2")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 203.0.113.10, 10.0.0.1")

	if got := clientIP(req); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want 203.0.113.10", got)
	}
}
