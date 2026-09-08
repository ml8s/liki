package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func chartAt(t *testing.T, date string, hour int) Chart {
	t.Helper()
	bt, err := time.ParseInLocation("2006-01-02", date, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	st := tianwen.GregorianToSolar(bt.Add(time.Duration(hour)*time.Hour), 116.4, 8)
	return ComputeChart(st)
}

func TestChartDerived_Anchors(t *testing.T) {
	cases := []struct {
		name       string
		date       string
		hour       int
		ju         int
		yin        bool
		dayGong    string
		hourGong   string
		riShi      string
		voidBranch [2]string
		voidGong   [2]string
		horse      string
		horseGong  string
	}{
		{
			name: "2026-06-28 午时 夏至中元 阴遁3局", date: "2026-06-28", hour: 12,
			ju: 3, yin: true, dayGong: "震", hourGong: "兑", riShi: "shi_controls_ri",
			voidBranch: [2]string{"子", "丑"}, voidGong: [2]string{"坎", "艮"},
			horse: "申", horseGong: "坤",
		},
		{
			name: "2026-01-01 辰时 冬至下元 阳遁4局", date: "2026-01-01", hour: 8,
			ju: 4, yin: false, dayGong: "离", hourGong: "艮", riShi: "ri_generates_shi",
			voidBranch: [2]string{"申", "酉"}, voidGong: [2]string{"坤", "兑"},
			horse: "寅", horseGong: "艮",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			chart := chartAt(t, c.date, c.hour)
			if chart.Pan.Jushu != c.ju || chart.Pan.YinDun != c.yin {
				t.Fatalf("ju=%d yin=%v, want %d/%v", chart.Pan.Jushu, chart.Pan.YinDun, c.ju, c.yin)
			}
			if chart.RiGanPalace.String() != c.dayGong || chart.ShiGanPalace.String() != c.hourGong {
				t.Fatalf("day/hour palaces = %s/%s, want %s/%s",
					chart.RiGanPalace, chart.ShiGanPalace, c.dayGong, c.hourGong)
			}
			if chart.RiShiRelation.Relation != c.riShi {
				t.Fatalf("ri_shi_relation = %+v, want %s", chart.RiShiRelation, c.riShi)
			}
			for i, want := range c.voidBranch {
				got := chart.Pan.KongWang[i]
				if got.Branch.String() != want || got.Gong.String() != c.voidGong[i] {
					t.Errorf("void[%d] = %s@%s, want %s@%s", i, got.Branch, got.Gong, want, c.voidGong[i])
				}
			}
			if chart.Pan.MaXing.Branch.String() != c.horse || chart.Pan.MaXing.Gong.String() != c.horseGong {
				t.Errorf("horse = %s@%s, want %s@%s", chart.Pan.MaXing.Branch, chart.Pan.MaXing.Gong, c.horse, c.horseGong)
			}
		})
	}
}

func TestPalaceWangShuaiUsesMonthBranch(t *testing.T) {
	chart := chartAt(t, "2026-06-28", 12)
	if len(chart.PalaceWangShuai) != 9 {
		t.Fatalf("palace wangshuai count = %d, want 9", len(chart.PalaceWangShuai))
	}
	want := map[GongIndex]ganzhi.WangShuai{
		GongLi:    ganzhi.WSWang,
		GongKun:   ganzhi.WSXiang,
		GongGen:   ganzhi.WSXiang,
		GongZhong: ganzhi.WSXiang,
		GongZhen:  ganzhi.WSXiu,
		GongXun:   ganzhi.WSXiu,
		GongKan:   ganzhi.WSQiu,
		GongQian:  ganzhi.WSSi,
		GongDui:   ganzhi.WSSi,
	}
	for _, item := range chart.PalaceWangShuai {
		if item.WangShuai != want[item.Gong] {
			t.Fatalf("%s wangshuai = %s, want %s", item.Gong.String(), item.WangShuai.String(), want[item.Gong].String())
		}
	}
	if chart.ShiGanGongWangShuai.Gong != GongDui || chart.ShiGanGongWangShuai.WangShuai != ganzhi.WSSi {
		t.Fatalf("shi gan palace wangshuai = %+v", chart.ShiGanGongWangShuai)
	}
}
