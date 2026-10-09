package tianwen

import (
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestMain(m *testing.M) {
	if err := ganzhi.Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
