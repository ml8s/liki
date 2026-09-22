package ziwei

import (
	"fmt"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestBond_SideBySideFacts(t *testing.T) {
	cases := []struct {
		a, b   string
		at, bt int
		ag, bg string
	}{
		{"2000-7-17", "2000-7-17", 2, 2, "女", "女"},
		{"1984-1-1", "1984-1-1", 5, 5, "男", "男"},
		{"2000-7-17", "1983-11-29", 2, 5, "女", "男"},
		{"1981-4-23", "1993-2-26", 1, 0, "女", "男"},
		{"1994-6-25", "1997-5-8", 11, 2, "男", "女"},
		{"1968-6-15", "1995-3-15", 5, 0, "女", "女"},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ca := chartFrom(t, tc.a, tc.at, tc.ag)
			cb := chartFrom(t, tc.b, tc.bt, tc.bg)
			bd := ComputeBond(ca, cb)

			if bd.AMingGong != ca.GongWei[int(ca.MingGong)].Name {
				t.Fatalf("A ming gong = %q", bd.AMingGong)
			}
			if bd.BMingGong != cb.GongWei[int(cb.MingGong)].Name {
				t.Fatalf("B ming gong = %q", bd.BMingGong)
			}
			if bd.FuQiGong == nil || bd.ZiNvGong == nil {
				t.Fatalf("spouse/children pair missing: %+v", bd)
			}
			for _, ref := range []*PairRef{bd.FuQiGong, bd.ZiNvGong} {
				if ref.AGongName == "" || ref.BGongName == "" || ref.Status == "" {
					t.Fatalf("incomplete pair reference: %+v", ref)
				}
				if ref.AGongName != ref.BGongName {
					t.Fatalf("same palace label expected: %+v", ref)
				}
			}
		})
	}
}

func TestBond_EmptySpousePalaceStatus(t *testing.T) {
	// Fixture selected so that A's spouse palace has no major star.
	ca := chartFrom(t, "1994-6-25", 11, "男")
	cb := chartFrom(t, "1997-5-8", 2, "女")
	bd := ComputeBond(ca, cb)
	if bd.FuQiGong == nil {
		t.Fatal("spouse pair missing")
	}
	if len(bd.FuQiGong.AZhuXing) != 0 && len(bd.FuQiGong.BZhuXing) != 0 {
		t.Fatalf("fixture expected at least one empty spouse palace: %+v", bd.FuQiGong)
	}
	if bd.FuQiGong.Status == "" {
		t.Fatal("empty spouse status missing")
	}
}

func chartFrom(t *testing.T, lunar string, ti int, gender string) Chart {
	t.Helper()
	var y, m, d int
	_, _ = fmt.Sscanf(lunar, "%d-%d-%d", &y, &m, &d)
	sz := ti + 1
	if ti == 12 {
		sz = 1
		d++
	}
	g := ganzhi.Female
	if gender == "男" {
		g = ganzhi.Male
	}
	chart, err := ComputeChart(tianwen.LunarTime{Year: y, Month: m, Day: d, Shichen: ganzhi.Zhi(sz)}, g)
	if err != nil {
		t.Fatalf("invalid chart fixture: %v", err)
	}
	return chart
}
