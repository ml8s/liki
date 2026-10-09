package fengshui

import (
	"os"
	"testing"

	"liki-engine/internal/engine/luoshu"
)

func TestMain(m *testing.M) {
	if err := luoshu.Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
