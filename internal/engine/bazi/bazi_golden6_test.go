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

// TestBaziGolden6_QiyunPrecision verifies 起运精确值密集覆盖（bazi_golden6.json，20 例）：
// 出生紧邻节气（0-1 岁）、节气当天不同时辰、跨月、顺逆 4 组合、远节气（10 岁+）。
// golden 由 lunar-typescript 生成；起运年/月/日精确断言（无 ±1）。
func TestBaziGolden6_QiyunPrecision(t *testing.T) {
	b, err := os.ReadFile("testdata/bazi_golden6.json")
	if err != nil {
		t.Fatalf("read golden6: %v", err)
	}
	var cases []baziGoldenCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse golden6: %v", err)
	}
	for _, gc := range cases {
		t.Run(gc.Name, func(t *testing.T) {
			tt, err := time.ParseInLocation("2006-01-02T15:04:05", gc.Solar, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatalf("parse solar: %v", err)
			}
			g := ganzhi.Male
			if gc.Gender == "female" {
				g = ganzhi.Female
			}
			c := ComputeChart(tianwen.SolarTime(tt), g)
			// 四柱
			p := fmt.Sprintf("%s%s %s%s %s%s %s%s",
				c.Nian.Gan, c.Nian.Zhi, c.Yue.Gan, c.Yue.Zhi, c.Ri.Gan, c.Ri.Zhi, c.Shi.Gan, c.Shi.Zhi)
			want := fmt.Sprintf("%s %s %s %s", gc.Pillars.Nian, gc.Pillars.Yue, gc.Pillars.Ri, gc.Pillars.Shi)
			if p != want {
				t.Errorf("四柱 = %s, want %s", p, want)
			}
			// 大运方向
			if (c.DaYun.Direction == "顺排") != gc.DaYun.Forward {
				t.Errorf("大运方向顺排 = %v, want %v", c.DaYun.Direction == "顺排", gc.DaYun.Forward)
			}
			// 起运精确年/月/日（无 ±1）
			if c.DaYun.StartYearAfter != gc.DaYun.StartYearAfter {
				t.Errorf("起运年 = %d, want %d", c.DaYun.StartYearAfter, gc.DaYun.StartYearAfter)
			}
			if c.DaYun.StartMonthAfter != gc.DaYun.StartMonthAfter {
				t.Errorf("起运月 = %d, want %d", c.DaYun.StartMonthAfter, gc.DaYun.StartMonthAfter)
			}
			if c.DaYun.StartDayAfter != gc.DaYun.StartDayAfter {
				t.Errorf("起运日 = %d, want %d", c.DaYun.StartDayAfter, gc.DaYun.StartDayAfter)
			}
			// 大运全步骤
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
