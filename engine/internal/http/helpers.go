package http

import (
	"os"
	"strconv"
	"strings"

	"net/http"
)

// trustedProxyHops returns the number of reverse proxies allowed to append
// X-Forwarded-For. Zero (the default) means client-supplied forwarding headers
// are untrusted and RemoteAddr is authoritative.
func trustedProxyHops() int {
	raw := os.Getenv("LIKI_TRUSTED_PROXY_HOPS")
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// clientIP extracts the client IP. X-Forwarded-For is trusted only when the
// deployment explicitly declares the proxies in front of this service.
func clientIP(r *http.Request) string {
	if hops := trustedProxyHops(); hops > 0 {
		parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			index := len(parts) - hops
			if index < 0 {
				index = 0
			}
			return parts[index]
		}
	}
	addr := r.RemoteAddr
	if i := strings.LastIndexByte(addr, ':'); i > 0 {
		return addr[:i]
	}
	return addr
}

// respondError writes a JSON error response with the given status code and error message.
func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
	_ = err
}
