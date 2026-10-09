package xuankong

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
