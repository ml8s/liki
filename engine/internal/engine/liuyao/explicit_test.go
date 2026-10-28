package liuyao

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

	var yaos [6]int
	for i := range yaos {
		yaos[i] = 7
	}

	canonical := ComputeChartAt(tianwen.StandardLocalTime(atZ, loc), 116.407526, YongShiYao, yaos)
	offset := ComputeChartAt(tianwen.StandardLocalTime(atOffset, loc), 116.407526, YongShiYao, yaos)
	if !reflect.DeepEqual(canonical, offset) {
		t.Fatal("offset serialization changed liuyao chart")
	}
}
