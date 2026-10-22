package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
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

	canonical := ComputeChartAt(atZ.UTC(), loc, 116.407526, ganzhi.Male)
	offset := ComputeChartAt(atOffset.UTC(), loc, 116.407526, ganzhi.Male)
	if !reflect.DeepEqual(canonical, offset) {
		t.Fatal("offset serialization changed bazi chart")
	}
}
