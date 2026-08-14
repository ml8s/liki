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

// TestComputeLiuNian_HouseOverlay：2024 年凶星（凶/大凶）落宫必须被叠加出来。
// 2024 三碧入中飞布：中5=3碧(凶)、兑7=5黄(大凶)、离9=7赤(凶)、巽4=2黑(凶) → 4 个凶星宫。
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

	res := ComputeLiuNian(2024, chart)

	wantGongs := map[int]string{5: "三碧禄存", 7: "五黄廉贞", 9: "七赤破军", 4: "二黑巨门"}
	if len(res.HouseOverlay) != len(wantGongs) {
		t.Fatalf("house_overlay len = %d, want %d (gongs %v)", len(res.HouseOverlay), len(wantGongs), wantGongs)
	}
	for _, o := range res.HouseOverlay {
		wantStar, ok := wantGongs[o.GongNum]
		if !ok {
			t.Errorf("unexpected overlay gong %d (star %s)", o.GongNum, o.Star)
			continue
		}
		if o.Star != wantStar {
			t.Errorf("gong %d: star = %s, want %s", o.GongNum, o.Star, wantStar)
		}
		if o.StarRating != "凶" && o.StarRating != "大凶" {
			t.Errorf("gong %d: star_rating = %s, want 凶/大凶", o.GongNum, o.StarRating)
		}
		// PalaceStars 应含宅盘该宫三星（构造时均为同名星）
		if o.PalaceStars == "" {
			t.Errorf("gong %d: palace_stars empty", o.GongNum)
		}
	}
}

// TestComputeLiuNian_HouseOverlay_2027：另一个年份锚点。
// 2027 九紫入中飞布：兑7=二黑(凶)、坎1=五黄(大凶)、震3=七赤(凶)、艮8=三碧(凶) → 4 个凶星宫。
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
	res := ComputeLiuNian(2027, chart)

	wantGongs := map[int]string{7: "二黑巨门", 1: "五黄廉贞", 3: "七赤破军", 8: "三碧禄存"}
	if len(res.HouseOverlay) != len(wantGongs) {
		t.Fatalf("2027 house_overlay len = %d, want %d (gongs %v)", len(res.HouseOverlay), len(wantGongs), wantGongs)
	}
	for _, o := range res.HouseOverlay {
		wantStar, ok := wantGongs[o.GongNum]
		if !ok {
			t.Errorf("unexpected overlay gong %d (star %s)", o.GongNum, o.Star)
			continue
		}
		if o.Star != wantStar {
			t.Errorf("gong %d: star = %s, want %s", o.GongNum, o.Star, wantStar)
		}
	}
}
