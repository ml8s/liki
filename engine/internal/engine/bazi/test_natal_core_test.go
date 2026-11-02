package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func pillarsCoreFromStrings(t *testing.T, pillars []string) NatalCore {
	t.Helper()
	if len(pillars) != 4 {
		t.Fatalf("pillars = %d, want 4", len(pillars))
	}
	var core NatalCore
	for index, pillar := range pillars {
		runes := []rune(pillar)
		gan, err := ganzhi.ParseGan(string(runes[0]))
		if err != nil {
			t.Fatal(err)
		}
		zhi, err := ganzhi.ParseZhi(string(runes[1]))
		if err != nil {
			t.Fatal(err)
		}
		core.Pillars[index] = ganzhi.Zhu{Gan: gan, Zhi: zhi}
		core.NaYin[index] = ganzhi.NayinLabel(gan, zhi)
	}
	core.Gender = ganzhi.Male
	return core
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("strings = %v, want %v", got, want)
		}
	}
}

func TestNatalCoreIsNotMutatedByDerivedMethods(t *testing.T) {
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	core := ComputeChart(stgt, st, ganzhi.Male).NatalCore()
	want := cloneNatalCoreForBondTest(core)

	liuNianTime, err := ResolveLiuNianTime(tianwen.NewCivilDate(2026, 6, 15))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ComputeLiuNianAtTime(core, liuNianTime); err != nil {
		t.Fatal(err)
	}
	liuYueTime, err := ResolveLiuYueTimeAt(tianwen.NewCivilDate(2026, 6, 15).Time(), time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ComputeLiuYueAtTime(core, liuYueTime); err != nil {
		t.Fatal(err)
	}
	if _, err := ComputeLiuRi(core, LiuRiTime{Date: tianwen.NewCivilDate(2026, 6, 15)}); err != nil {
		t.Fatal(err)
	}
	if _, err := ComputeLiuShi(core, resolveTestLiuShi(t, 2026, 6, 15, 12, 30)); err != nil {
		t.Fatal(err)
	}
	ComputeXiaoYun(core, 3)
	ComputeBond(core, core)

	if !reflect.DeepEqual(core, want) {
		t.Fatal("a derived method mutated NatalCore")
	}
}
