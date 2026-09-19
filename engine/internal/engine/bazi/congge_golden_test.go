package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestCongGe_AllRules(t *testing.T) {
	tests := []struct {
		name        string
		nianGan     int
		nianZ       int
		yueGan      int
		yueZ        int
		riGan       int
		riZ         int
		shiGan      int
		shiZ        int
		wantPattern string
		wantYong    string
		wantXi      string
		wantJi      string
		note        string
	}{
		// 从旺格: 甲木寅月=旺, 壬(印)+甲(比)+壬(印)=3, 官杀=0
		// yong=木(日主), xi=火(食伤泄秀), ji=金(官杀)
		{
			name: "从旺格_甲木寅月", nianGan: 9, nianZ: 3, yueGan: 1, yueZ: 3,
			riGan: 1, riZ: 3, shiGan: 9, shiZ: 3,
			wantPattern: "从旺格", wantYong: "木", wantXi: "火", wantJi: "金",
			note: "印比满盘无官杀→从旺。用比劫(木), 喜食伤泄秀(火), 忌官杀(金)",
		},
		// 从杀格: 甲木酉月=死, root=none, 辛+辛=官杀2, 印=0
		// yong=金(官杀), xi=土(财生杀), ji=水(印)
		{
			name: "从杀格_甲木酉月", nianGan: 8, nianZ: 10, yueGan: 8, yueZ: 10,
			riGan: 1, riZ: 1, shiGan: 1, shiZ: 1,
			wantPattern: "从杀格", wantYong: "金", wantXi: "土", wantJi: "水",
			note: "无根死月官杀垄断无印→从杀。用官杀(金), 喜财(土), 忌印(水)",
		},
		// 从财格: 甲木戌月=囚, root=none, 戊×3=财, 比劫=0
		// yong=土(财), xi=火(食伤生财), ji=木(比劫)
		{
			name: "从财格_甲木戌月", nianGan: 5, nianZ: 11, yueGan: 5, yueZ: 11,
			riGan: 1, riZ: 1, shiGan: 5, shiZ: 11,
			wantPattern: "从财格", wantYong: "土", wantXi: "火", wantJi: "木",
			note: "无根囚月财垄断无比劫→从财。用财(土), 喜食伤(火), 忌比劫(木)",
		},
		// 从儿格: 甲木巳月=休, root=none, 丁×3=食伤, 印=0
		// 丁巳 丁丑 甲子 丁酉
		// yong=火(食伤), xi=土(财), ji=水(印)
		{
			name: "从儿格_甲木巳月", nianGan: 4, nianZ: 6, yueGan: 4, yueZ: 6,
			riGan: 1, riZ: 1, shiGan: 4, shiZ: 10,
			wantPattern: "从儿格", wantYong: "火", wantXi: "土", wantJi: "水",
			note: "无根休月食伤垄断无印→从儿。用食伤(火), 喜财(土), 忌印(水)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := canonicalChart(ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.Gan(tt.nianGan), Zhi: ganzhi.Zhi(tt.nianZ)},
				Yue:  ganzhi.Zhu{Gan: ganzhi.Gan(tt.yueGan), Zhi: ganzhi.Zhi(tt.yueZ)},
				Ri:   ganzhi.Zhu{Gan: ganzhi.Gan(tt.riGan), Zhi: ganzhi.Zhi(tt.riZ)},
				Shi:  ganzhi.Zhu{Gan: ganzhi.Gan(tt.shiGan), Zhi: ganzhi.Zhi(tt.shiZ)},
			}, ganzhi.Male, 2000)
			pat, yong, xi, ji := lookupCongGe(c)
			if pat != tt.wantPattern {
				bz := c.ToBazi()
				t.Errorf("pattern=%q want=%q\n  pillars: %s%s %s%s %s%s %s%s\n  note: %s",
					pat, tt.wantPattern,
					ganzhi.GanName(bz.Nian.Gan), ganzhi.ZhiName(bz.Nian.Zhi),
					ganzhi.GanName(bz.Yue.Gan), ganzhi.ZhiName(bz.Yue.Zhi),
					ganzhi.GanName(bz.Ri.Gan), ganzhi.ZhiName(bz.Ri.Zhi),
					ganzhi.GanName(bz.Shi.Gan), ganzhi.ZhiName(bz.Shi.Zhi),
					tt.note)
			}
			if tt.wantPattern != "" {
				if yong != tt.wantYong {
					t.Errorf("yong=%q want=%q (%s)", yong, tt.wantYong, tt.note)
				}
				if xi != tt.wantXi {
					t.Errorf("xi=%q want=%q (%s)", xi, tt.wantXi, tt.note)
				}
				if ji != tt.wantJi {
					t.Errorf("ji=%q want=%q (%s)", ji, tt.wantJi, tt.note)
				}
			}
		})
	}
}

// TestCongGe_NotTriggered_NormalChart ensures charts with 官杀 or 财 transparency
// do NOT accidentally match 从格.
func TestCongGe_NotTriggered_NormalChart(t *testing.T) {
	// 甲木申月: 甲克土当令→囚, but 庚(七杀)透干 → guan_sha_tou_gan>0
	// So 从旺格 should NOT trigger
	gengShen := canonicalChart(ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiZi},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiShen},
	}, ganzhi.Male, 2000)
	pat, _, _, _ := lookupCongGe(gengShen)
	if pat != "" {
		t.Errorf("甲木+庚透干 should NOT trigger 从格, got %q", pat)
	}
}
