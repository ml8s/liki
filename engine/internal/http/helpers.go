package http

import (
	"encoding/json"
	"errors"
	"strings"

	"net/http"
)

// ValidateTokenPolicy enforces secure-by-default behavior for non-loopback
// HTTP listeners. The explicit escape hatch is only for trusted edge setups.
func ValidateTokenPolicy(token, listenAddr string, allowNoToken bool) error {
	if token != "" || IsLoopbackAddr(listenAddr) || allowNoToken {
		return nil
	}
	return errors.New(
		"LIKI_MCP_TOKEN is required on a non-loopback address; " +
			"set LIKI_MCP_TOKEN or LIKI_ALLOW_NO_TOKEN=1",
	)
}

// IsLoopbackAddr reports whether a listen address binds only to loopback
// (e.g. "127.0.0.1:8081", "localhost:8081", "[::1]:8081"). Used for startup
// hygiene warnings; wildcard binds ("":8081, ":8081") are NOT loopback.
func IsLoopbackAddr(addr string) bool {
	return strings.HasPrefix(addr, "127.") ||
		strings.HasPrefix(addr, "localhost:") ||
		strings.HasPrefix(addr, "[::1]:")
}

// respondError writes a JSON error response with the given status code and error message.
func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
	_, _ = w.Write(body)
}
