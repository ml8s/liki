package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 生产入口按 chart 消费流年/流月/流日/流时/小运；
// 本测试锁定"offset 序列化不改变结果"的性质。
func TestFlowProjectionsNormalizeOffset(t *testing.T) {
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

	chartZ := ComputeChartAt(tianwen.StandardLocalTime(atZ, loc), 116.407526, ganzhi.Male)
	chartOffset := ComputeChartAt(tianwen.StandardLocalTime(atOffset, loc), 116.407526, ganzhi.Male)
	if !reflect.DeepEqual(chartZ, chartOffset) {
		t.Fatal("offset serialization changed chart")
	}

	cases := []struct {
		name string
		run  func(chart Chart) any
	}{
		{"liunian", func(c Chart) any {
			v, err := ComputeLiuNian(c, 2026)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}},
		{"liuyue", func(c Chart) any {
			v, err := ComputeLiuYue(c, 2026, 5)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}},
		{"liuri", func(c Chart) any {
			lt, err := ResolveLiuRiTime(tianwen.NewCivilDate(2026, 5, 20))
			if err != nil {
				t.Fatal(err)
			}
			v, err := ComputeLiuRi(c, lt)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}},
		{"liushi", func(c Chart) any {
			v, err := ComputeLiuShi(c, resolveTestLiuShi(t, 2026, 5, 20, 12, 30))
			if err != nil {
				t.Fatal(err)
			}
			return v
		}},
		{"xiaoyun", func(c Chart) any {
			return ComputeXiaoYun(c, 3)
		}},
	}
	for _, tc := range cases {
		z, o := tc.run(chartZ), tc.run(chartOffset)
		if !reflect.DeepEqual(z, o) {
			t.Fatalf("offset serialization changed %s", tc.name)
		}
	}
}
