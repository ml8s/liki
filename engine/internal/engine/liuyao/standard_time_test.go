package liuyao

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestComputeChartAtUsesStandardAndLast(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	standard := tianwen.StandardLocalTime(at, ny)
	last := tianwen.ApparentSolarFromStandard(standard, -74)
	yaos := [6]int{7, 8, 7, 8, 7, 8}
	got := ComputeChartAt(standard, -74, YongShiYao, yaos)
	want := ComputeChart(standard, last, YongShiYao, yaos)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("liuyao standard/LAST composition differs from ComputeChartAt")
	}
}
