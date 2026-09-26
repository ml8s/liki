package http

import (
	"crypto/subtle"
	"net/http"
)

// MCPAuthMiddleware optionally requires `Authorization: Bearer <token>` for MCP
// requests. An empty token keeps the public hosted deployment backward
// compatible. Health/version probes and CORS preflights remain reachable.
func MCPAuthMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" || r.Method == http.MethodOptions || isPublicMCPProbe(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		expected := "Bearer " + token
		provided := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", "Bearer")
		respondError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
	})
}

func isPublicMCPProbe(path string) bool {
	return path == "/health" || path == "/version"
}
