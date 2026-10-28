package qimen

import (
	"reflect"
	"testing"
	"time"

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

	method, err := ResolveMethod("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	canonical, err := ComputeChartWithMethodAt(tianwen.StandardLocalTime(atZ, loc), 116.407526, method)
	if err != nil {
		t.Fatalf("canonical chart: %v", err)
	}
	offset, err := ComputeChartWithMethodAt(tianwen.StandardLocalTime(atOffset, loc), 116.407526, method)
	if err != nil {
		t.Fatalf("offset chart: %v", err)
	}
	if !reflect.DeepEqual(canonical, offset) {
		t.Fatal("offset serialization changed qimen chart")
	}
}
