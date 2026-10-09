package http

import (
	"os"
	"testing"

	"liki-engine/internal/agent"
)

func TestMain(m *testing.M) {
	if err := agent.Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
