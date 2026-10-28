package qimen

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestComputeChartWithMethodAtUsesStandardAndLast(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	standard := tianwen.StandardLocalTime(at, ny)
	last := tianwen.ApparentSolarFromStandard(standard, -74)
	method, err := ResolveMethod("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ComputeChartWithMethodAt(standard, ny, -74, method)
	if err != nil {
		t.Fatal(err)
	}
	want, err := ComputeChartWithMethod(standard, last, method)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("qimen standard/LAST composition differs from ComputeChartWithMethodAt")
	}
}
