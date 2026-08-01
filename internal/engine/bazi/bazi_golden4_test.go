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

// TestBaziGolden4_AllLayers verifies 第四轮 golden（bazi_golden4.json，24 例）：
// 补全 12 时辰覆盖（丑/寅/卯/酉/戌）、时辰切换边界（亥→子/子→丑/午→未/未→申）、
// 农历正月初一/除夕、闰月位置（闰二月/闰四月）、立春当天起运边界、大运全步骤。
func TestBaziGolden4_AllLayers(t *testing.T) {
	b, err := os.ReadFile("testdata/bazi_golden4.json")
	if err != nil {
		t.Fatalf("read golden4: %v", err)
	}
	var cases []baziGoldenCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse golden4: %v", err)
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
			p := fmt.Sprintf("%s%s %s%s %s%s %s%s",
				c.Nian.Gan, c.Nian.Zhi, c.Yue.Gan, c.Yue.Zhi, c.Ri.Gan, c.Ri.Zhi, c.Shi.Gan, c.Shi.Zhi)
			want := fmt.Sprintf("%s %s %s %s", gc.Pillars.Nian, gc.Pillars.Yue, gc.Pillars.Ri, gc.Pillars.Shi)
			if p != want {
				t.Errorf("四柱 = %s, want %s", p, want)
			}
			if (c.DaYun.Direction == "顺排") != gc.DaYun.Forward {
				t.Errorf("大运方向顺排 = %v, want %v", c.DaYun.Direction == "顺排", gc.DaYun.Forward)
			}
			dy := c.DaYun.Steps[0].Gan.String() + c.DaYun.Steps[0].Zhi.String()
			if dy != gc.DaYun.Steps[0] {
				t.Errorf("大运首步 = %s, want %s", dy, gc.DaYun.Steps[0])
			}
			if c.DaYun.StartAge != gc.DaYun.StartYearAfter && c.DaYun.StartAge != gc.DaYun.StartYearAfter+1 {
				t.Errorf("起运岁数 = %d, want %d(±1)", c.DaYun.StartAge, gc.DaYun.StartYearAfter)
			}
			if len(c.DaYun.Steps) != len(gc.DaYun.Steps) {
				t.Errorf("大运步数 = %d, want %d", len(c.DaYun.Steps), len(gc.DaYun.Steps))
			} else {
				for si := range gc.DaYun.Steps {
					gs := c.DaYun.Steps[si].Gan.String() + c.DaYun.Steps[si].Zhi.String()
					if gs != gc.DaYun.Steps[si] {
						t.Errorf("大运[%d] = %s, want %s", si, gs, gc.DaYun.Steps[si])
						break
					}
				}
			}
		})
	}
}
