package bazi

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// TestBaziGolden2_EdgeLayers verifies 边界 golden（bazi_golden2.json，16 例）：
// 立春/惊蛰前后、晚子时+立春、闰月前后、大月三十、跨年、夏令时、西半球、1900/2050、子时三刻。
// golden 由 lunar-typescript 生成；分钟级节气边界用例已调整为 30min 外（避开 lunar 节气近似误差）。
func TestBaziGolden2_EdgeLayers(t *testing.T) {
	b, err := os.ReadFile("testdata/bazi_golden2.json")
	if err != nil {
		t.Fatalf("read golden2: %v", err)
	}
	var cases []baziGoldenCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse golden2: %v", err)
	}
	for _, gc := range cases {
		t.Run(gc.Name, func(t *testing.T) {
			tt, err := time.ParseInLocation("2006-01-02T15:04:05", gc.Solar, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatalf("parse solar: %v", err)
			}
			st := tianwen.SolarTime(tt)
			g := ganzhi.Male
			if gc.Gender == "female" {
				g = ganzhi.Female
			}
			c := ComputeChart(st, g)
			fc := ComputeFullChart(c)
			p := fmt.Sprintf("%s%s %s%s %s%s %s%s",
				c.Nian.Gan, c.Nian.Zhi, c.Yue.Gan, c.Yue.Zhi, c.Ri.Gan, c.Ri.Zhi, c.Shi.Gan, c.Shi.Zhi)
			want := fmt.Sprintf("%s %s %s %s", gc.Pillars.Nian, gc.Pillars.Yue, gc.Pillars.Ri, gc.Pillars.Shi)
			if p != want {
				t.Errorf("四柱 = %s, want %s", p, want)
			}
			mg := fc.SanYuan.MingGong.Gan.String() + fc.SanYuan.MingGong.Zhi.String()
			if mg != gc.SanYuan.MingGong {
				t.Errorf("命宫 = %s, want %s", mg, gc.SanYuan.MingGong)
			}
			dy := c.DaYun.Steps[0].Gan.String() + c.DaYun.Steps[0].Zhi.String()
			if dy != gc.DaYun.Steps[0] {
				t.Errorf("大运首步 = %s, want %s", dy, gc.DaYun.Steps[0])
			}
		})
	}
}
