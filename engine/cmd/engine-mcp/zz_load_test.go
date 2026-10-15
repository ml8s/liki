package main

import (
	"os"
	"testing"

	"liki-engine/internal/engine/tables"
)

func TestMain(m *testing.M) {
	if err := tables.Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
