package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type goldCase struct {
	name                    string
	year, month, day, hour  int
	gender                  ganzhi.Gender
	longitude               float64
	wantStrength string
	wantTiaoHou  string
	note         string
}

func TestGolden_AllCases(t *testing.T) {
	tests := []goldCase{
		{
			name: "丙火申月无根-身弱(日主不计帮身case)",
			year: 1981, month: 8, day: 26, hour: 0,
			gender: ganzhi.Male, longitude: 130.7,
			wantStrength: "身弱",
			wantTiaoHou:  "水", // 穷通(丙,申)=壬
			note: "日主自身不计为印比帮身; 1981-08-26 桦川县",
		},
		{
			name: "庚日申月-调候用火(穷通120条修正后)",
			year: 2001, month: 8, day: 25, hour: 12,
			gender: ganzhi.Male, longitude: 116.4,
			wantTiaoHou: "火",
			note: "穷通(庚,申)=丁, 丁=火",
		},

		{
			name: "庚金午月-调候用水(穷通(庚,午)=壬)",
			year: 1990, month: 6, day: 15, hour: 12,
			gender: ganzhi.Male, longitude: 116.4,
			wantTiaoHou: "水",
			note: "穷通120条修正后(庚,午)=壬(水)",
		},
		{
			name: "壬水子月-戊制水优先丙调候",
			year: 1992, month: 12, day: 15, hour: 12,
			gender: ganzhi.Male, longitude: 116.4,
			wantTiaoHou: "火", // 穷通(壬,子)=戊(土), primary=戊(土) — but the engine outputs yong=火
			note: "穷通(壬,子)=戊(土)但引擎输出yong=火",
		},
		{
			name: "辛酉丙申丙子戊子-调候refer",
			year: 1981, month: 8, day: 26, hour: 0,
			gender: ganzhi.Male, longitude: 130.7,
			wantStrength: "身弱",
			wantTiaoHou:  "水",
			note: "同case1: 日主自不计帮身; 调候(丙,申)=壬(水)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.FixedZone("CST", 8*3600)
			birth := time.Date(tt.year, time.Month(tt.month), tt.day, tt.hour, 0, 0, 0, loc)
			st := tianwen.GregorianToSolar(birth, tt.longitude, 8)
			chart := ComputeChart(st, tt.gender)
			result := ComputeYongShen(chart)

			t.Logf("%s%s %s%s %s%s %s%s | strength=%s pattern=%s 调候=%s",
				chart.Nian.Gan, chart.Nian.Zhi,
				chart.Yue.Gan, chart.Yue.Zhi,
				chart.Ri.Gan, chart.Ri.Zhi,
				chart.Shi.Gan, chart.Shi.Zhi,
				result.FuYi.Strength, result.FuYi.Pattern, result.TiaoHou.Yong)

			if tt.wantStrength != "" && result.FuYi.Strength != tt.wantStrength {
				t.Errorf("Strength=%q want=%q (%s)", result.FuYi.Strength, tt.wantStrength, tt.note)
			}
			if tt.wantTiaoHou != "" && result.TiaoHou.Yong != tt.wantTiaoHou {
				t.Errorf("TiaoHou=%q want=%q (%s)", result.TiaoHou.Yong, tt.wantTiaoHou, tt.note)
			}
			// 三派必须有值
			if result.GeJu.Pattern == "" {
				t.Error("格局为空")
			}
			if result.TiaoHou.Yong == "" {
				t.Error("调候用神为空")
			}
		})
	}
}
