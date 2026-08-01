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

// TestBaziGolden3_AllLayers verifies 第三轮边界 golden（bazi_golden3.json，32 例）：
// 12 节换月边界（清明/立夏/芒种/小暑/立秋/白露/寒露/立冬/大雪/小寒）、
// 极端时区（UTC+13/-11/+5:45）、60 甲子日柱连续循环、大运顺逆 4 组合、五鼠遁全 10 日干。
func TestBaziGolden3_AllLayers(t *testing.T) {
	b, err := os.ReadFile("testdata/bazi_golden3.json")
	if err != nil {
		t.Fatalf("read golden3: %v", err)
	}
	var cases []baziGoldenCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse golden3: %v", err)
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
