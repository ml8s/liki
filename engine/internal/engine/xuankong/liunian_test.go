package xuankong

import (
	"testing"

	"liki-engine/internal/engine/fengshui"
)

// TestComputeLiuNian_NoChart：只给年份 → 纯流年飞星盘，无宅盘叠加。
func TestComputeLiuNian_NoChart(t *testing.T) {
	res := ComputeLiuNian(2024, nil)
	if res.Year != 2024 {
		t.Errorf("year = %d, want 2024", res.Year)
	}
	if res.RuZhong != "三碧禄存" {
		t.Errorf("ru_zhong = %s, want 三碧禄存", res.RuZhong)
	}
	if len(res.GongWei) != 9 {
		t.Errorf("gong_wei len = %d, want 9", len(res.GongWei))
	}
	if len(res.HouseOverlay) != 0 {
		t.Errorf("house_overlay len = %d, want 0 (no chart)", len(res.HouseOverlay))
	}
}

// TestComputeLiuNian_HouseOverlay：流年星落宫全量叠加，不预设星曜固有吉凶。
func TestComputeLiuNian_HouseOverlay(t *testing.T) {
	chart := &Chart{}
	for i := 0; i < 9; i++ {
		n := i + 1
		chart.Palaces[i] = xuanKongStar{
			PalaceNum:    n,
			PeriodStar:   fengshui.StarByNumber(n),
			MountainStar: fengshui.StarByNumber(n),
			FacingStar:   fengshui.StarByNumber(n),
		}
	}
	chart.refreshDigest()

	res := ComputeLiuNian(2024, chart)

	if len(res.HouseOverlay) != 9 {
		t.Fatalf("house_overlay len = %d, want 9", len(res.HouseOverlay))
	}
	annualByGong := map[int]fengshui.AnnualFlyingStar{}
	for _, annual := range res.GongWei {
		annualByGong[annual.GongNum] = annual
	}
	for _, o := range res.HouseOverlay {
		annual := annualByGong[o.GongNum]
		if annual.GongNum != o.GongNum || annual.XingName != o.Star {
			t.Errorf("overlay gong %d does not match annual board: %+v", o.GongNum, o)
		}
		// PalaceStars 应含宅盘该宫三星（构造时均为同名星）
		if o.PalaceStars == "" {
			t.Errorf("gong %d: palace_stars empty", o.GongNum)
		}
	}
}

// TestComputeLiuNian_HouseOverlay_2027：另一个年份锚点。
func TestComputeLiuNian_HouseOverlay_2027(t *testing.T) {
	chart := &Chart{}
	for i := 0; i < 9; i++ {
		n := i + 1
		chart.Palaces[i] = xuanKongStar{
			PalaceNum:    n,
			PeriodStar:   fengshui.StarByNumber(n),
			MountainStar: fengshui.StarByNumber(n),
			FacingStar:   fengshui.StarByNumber(n),
		}
	}
	chart.refreshDigest()
	res := ComputeLiuNian(2027, chart)

	if len(res.HouseOverlay) != 9 {
		t.Fatalf("2027 house_overlay len = %d, want 9", len(res.HouseOverlay))
	}
	annualByGong := map[int]fengshui.AnnualFlyingStar{}
	for _, annual := range res.GongWei {
		annualByGong[annual.GongNum] = annual
	}
	for _, o := range res.HouseOverlay {
		annual := annualByGong[o.GongNum]
		if annual.GongNum != o.GongNum || annual.XingName != o.Star {
			t.Errorf("2027 overlay gong %d does not match annual board: %+v", o.GongNum, o)
		}
	}
}

func TestComputeLiuNian_RejectsTamperedChart(t *testing.T) {
	chart := computeChart(0, 12, 2024)
	chart.SitMountain = 1
	if got := ComputeLiuNian(2026, &chart); got.Year != 0 || got.RuZhong != "" || len(got.GongWei) != 0 {
		t.Fatalf("tampered chart returned data: %+v", got)
	}
}
