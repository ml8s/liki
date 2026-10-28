package http

import (
	"strings"
	"testing"
)

func TestValidateTokenPolicy(t *testing.T) {
	if err := ValidateTokenPolicy("secret", ":8081", false); err != nil {
		t.Fatalf("configured token should pass: %v", err)
	}
	if err := ValidateTokenPolicy("", "127.0.0.1:8081", false); err != nil {
		t.Fatalf("loopback should pass: %v", err)
	}
	if err := ValidateTokenPolicy("", ":8081", true); err != nil {
		t.Fatalf("explicit override should pass: %v", err)
	}
	err := ValidateTokenPolicy("", ":8081", false)
	if err == nil || !strings.Contains(err.Error(), "LIKI_MCP_TOKEN is required") {
		t.Fatalf("non-loopback empty token = %v, want fail-closed", err)
	}
}
