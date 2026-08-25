package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// ── 子平真诠·格局用神原著对照 — 8格全验证 ──

type geJuExpect struct {
	name        string
	chart       Chart
	wantPattern string
	wantUsage   string
	wantYong    string
	wantXi      string
	wantJi      string
	note        string
}

func TestZiPing_GeJu_AllEightPatterns(t *testing.T) {
	tests := []geJuExpect{
		{name: "正官格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiYou}},
			Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen}},
		}, wantPattern: "正官格", wantUsage: "顺用", wantYong: "土", wantXi: "水", wantJi: "火",
			note: "财生官(土), 印制伤(水), 忌伤官(火)"},
		{name: "正财格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiWei}},
		}, wantPattern: "正财格", wantUsage: "顺用", wantYong: "火", wantXi: "金", wantJi: "木",
			note: "食伤生财(火), 忌比劫(木)"},
		{name: "偏财格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen}},
		}, wantPattern: "偏财格", wantUsage: "顺用", wantYong: "火", wantXi: "金", wantJi: "木",
			note: "同正财: 食伤生财(火), 忌比劫(木)"},
		{name: "正印格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiZi}},
		}, wantPattern: "正印格", wantUsage: "顺用", wantYong: "金", wantXi: "木", wantJi: "土",
			note: "官杀生印(金), 忌财破印(土)"},
		{name: "偏印格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiHai}},
		}, wantPattern: "偏印格", wantUsage: "顺用", wantYong: "金", wantXi: "木", wantJi: "土",
			note: "同正印: 官杀生印(金), 忌财破印(土)"},
		{name: "食神格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiSi}},
		}, wantPattern: "食神格", wantUsage: "顺用", wantYong: "木", wantXi: "土", wantJi: "水",
			note: "财泄食(土), 忌印夺食(水)"},
		{name: "七杀格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}},
		}, wantPattern: "七杀格", wantUsage: "逆用", wantYong: "火", wantXi: "木", wantJi: "土",
			note: "食神制杀(火), 忌财生杀(土)"},
		{name: "伤官格", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiWu}},
		}, wantPattern: "伤官格", wantUsage: "逆用", wantYong: "水", wantXi: "金", wantJi: "木",
			note: "印制伤(水), 忌比劫生伤(木)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeGeJu(tt.chart, nil)
			if result.Pattern != tt.wantPattern {
				t.Errorf("Pattern=%q want=%q", result.Pattern, tt.wantPattern)
			}
			if result.Usage != tt.wantUsage {
				t.Errorf("Usage=%q want=%q", result.Usage, tt.wantUsage)
			}
			if result.Yong != tt.wantYong {
				t.Errorf("Yong=%q want=%q (%s)", result.Yong, tt.wantYong, tt.note)
			}
			if result.Xi != tt.wantXi {
				t.Errorf("Xi=%q want=%q (%s)", result.Xi, tt.wantXi, tt.note)
			}
			if result.Ji != tt.wantJi {
				t.Errorf("Ji=%q want=%q (%s)", result.Ji, tt.wantJi, tt.note)
			}
			if result.Usage == "顺用" && result.Xi == result.Ji {
				t.Errorf("xi==ji=%q, 顺用格喜忌不能相同", result.Xi)
			}
		})
	}
}

// ── 子平真诠·建禄/月刃/杂格(外格) ──

func TestZiPing_JianLuYueRen(t *testing.T) {
	tests := []geJuExpect{
		{name: "建禄格_甲寅", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}}, Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanXin}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen}},
		}, wantPattern: "建禄格", wantUsage: "逆用", wantYong: "金", wantXi: "土", wantJi: "水",
			note: "子平真诠·建禄: 用官杀克身(金)"},
		{name: "月刃格_甲卯", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiMao}}, Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanXin}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen}},
		}, wantPattern: "月刃格", wantUsage: "逆用", wantYong: "金", wantXi: "土", wantJi: "水",
			note: "子平真诠·月刃: 用官杀克身(金)"},
		{name: "杂格_甲辰_劫财透", chart: Chart{
			Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
			Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChen}},
			// 辰支: 乙(劫财)透月干 → 劫财不入八格 → 杂格逆用
			Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen}},
		}, wantPattern: "杂格", wantUsage: "逆用", wantYong: "金", wantXi: "土", wantJi: "水",
			note: "子平真诠: 比劫不入八格 → 杂格逆用"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeGeJu(tt.chart, nil)
			if result.Pattern != tt.wantPattern {
				t.Errorf("Pattern=%q want=%q", result.Pattern, tt.wantPattern)
			}
			if result.Usage != tt.wantUsage {
				t.Errorf("Usage=%q want=%q", result.Usage, tt.wantUsage)
			}
			if result.Yong != tt.wantYong {
				t.Errorf("Yong=%q want=%q (%s)", result.Yong, tt.wantYong, tt.note)
			}
		})
	}
}
