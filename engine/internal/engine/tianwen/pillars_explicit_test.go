package tianwen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestComputeBaziUsesExplicitPillarFunctions(t *testing.T) {
	cases := []struct {
		name  string
		civil time.Time
		solar time.Time
	}{
		{name: "normal", civil: time.Date(1990, 5, 20, 12, 0, 0, 0, time.FixedZone("", 8*3600)), solar: time.Date(1990, 5, 20, 11, 48, 0, 0, time.FixedZone("", 8*3600))},
		{name: "late-zi", civil: time.Date(1990, 5, 20, 15, 30, 0, 0, time.UTC), solar: time.Date(1990, 5, 20, 15, 24, 0, 0, time.UTC)},
		{name: "lichun-boundary", civil: time.Date(2024, 2, 4, 17, 26, 54, 0, time.UTC), solar: time.Date(2024, 2, 4, 17, 18, 0, 0, time.UTC)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := ganzhi.Bazi{
				Nian: NianZhu(time.Time(tc.civil)),
				Yue:  YueZhu(time.Time(tc.civil)),
				Ri:   RiZhu(time.Time(tc.solar)),
				Shi:  ShiZhu(time.Time(tc.solar)),
			}
			got := ComputeBazi(time.Time(tc.civil), time.Time(tc.solar))
			if got != want {
				t.Fatalf("pillars = %+v, want %+v", got, want)
			}
		})
	}
}

func ShiZhu(st time.Time) ganzhi.Zhu {
	solarMinutes := apparentSolarMinutes(st)
	zhi := shichenFromApparentSolarMinutes(solarMinutes)
	// 晚子时（23:00 后）：日柱不变，但时柱按次日日干起（lunar 约定）
	dayT := st
	if solarMinutes >= 1380 {
		dayT = dayT.AddDate(0, 0, 1)
	}
	dp := RiZhu(time.Time(dayT))
	gan := ganzhi.Gan(((int(dp.Gan)*2 + int(zhi) - 2) % 10))
	if gan == 0 {
		gan = 10
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}
