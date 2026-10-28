package ziwei

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestComputeChartAtNormalizesOffset(t *testing.T) {
	atZ, err := time.Parse(time.RFC3339Nano, "1990-05-20T04:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	atOffset, err := time.Parse(time.RFC3339Nano, "1990-05-20T12:00:00+08:00")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	canonical, err := ComputeChartAt(tianwen.StandardLocalTime(atZ, loc), 116.407526, ganzhi.Male)
	if err != nil {
		t.Fatalf("canonical chart: %v", err)
	}
	offset, err := ComputeChartAt(tianwen.StandardLocalTime(atOffset, loc), 116.407526, ganzhi.Male)
	if err != nil {
		t.Fatalf("offset chart: %v", err)
	}

	if !reflect.DeepEqual(canonical, offset) {
		t.Fatal("offset serialization changed ziwei chart")
	}
}
