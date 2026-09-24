package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ============================================================================
// 流时 (LiuShi) — 晚子时(23:00-24:00)时柱按次日日干起
// ============================================================================

func TestLiuShi_WanZiShi_UsesNextDayGan(t *testing.T) {
	// 1990-05-20 12:00 CST → 庚午 辛巳 乙酉（日干乙）。
	st := tianwen.GregorianToSolar(
		time.Date(1990, 5, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	bz := chart.ToBazi()
	if bz.Ri.Gan != ganzhi.GanYi {
		t.Fatalf("期望 5/20 日干乙（庚午 辛巳 乙酉），got %d", bz.Ri.Gan)
	}

	// 早子时 00:00：当天日干（乙）起五鼠遁 → 乙日子时 丙子。
	early, err := computeLiuShi(bz, 1990, 5, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if early.ShiGan != ganzhi.GanBing {
		t.Errorf("5/20 00:00 早子时 ShiGan = %d, want 丙(3)", early.ShiGan)
	}

	// 晚子时 23:00：时柱按次日（5/21 丙日）日干起 → 丙日子时 戊子。
	late, err := computeLiuShi(bz, 1990, 5, 20, 23)
	if err != nil {
		t.Fatal(err)
	}
	if late.ShiGan != ganzhi.GanWu {
		t.Errorf("5/20 23:00 晚子时 ShiGan = %d, want 戊(5)（次日日干）", late.ShiGan)
	}

	// 次日 00:00（丙日早子时）同样 戊子——与 5/20 晚子时一致（跨日同子时）。
	next, err := computeLiuShi(bz, 1990, 5, 21, 0)
	if err != nil {
		t.Fatal(err)
	}
	if next.ShiGan != ganzhi.GanWu {
		t.Errorf("5/21 00:00 ShiGan = %d, want 戊(5)", next.ShiGan)
	}
}

func TestLiuShi_OtherHours_UseSameDayGan(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1990, 5, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	bz := chart.ToBazi()

	// 卯时（6:00）用当天日干乙 → 乙日卯时 己卯。
	ls, err := computeLiuShi(bz, 1990, 5, 20, 6)
	if err != nil {
		t.Fatal(err)
	}
	if ls.ShiGan != ganzhi.GanJi {
		t.Errorf("5/20 06:00 ShiGan = %d, want 己(6)", ls.ShiGan)
	}
}