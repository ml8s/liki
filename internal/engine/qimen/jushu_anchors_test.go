package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 奇门定局数据驱动锚点（《奇门遁甲》通行定局表 + 三元符头规则）。

// 24 节气 × 上中下元 = 72 项权威定局值（阳遁/阴遁）。
func TestJushu_SolarTermBureau_All72(t *testing.T) {
	want := []struct {
		name                   string
		shang, zhong, xia, yin int
	}{
		{"冬至", 1, 7, 4, 1}, {"小寒", 2, 8, 5, 1}, {"大寒", 3, 9, 6, 1},
		{"立春", 8, 5, 2, 1}, {"雨水", 9, 6, 3, 1}, {"惊蛰", 1, 7, 4, 1},
		{"春分", 3, 9, 6, 1}, {"清明", 4, 1, 7, 1}, {"谷雨", 5, 2, 8, 1},
		{"立夏", 4, 1, 7, 1}, {"小满", 5, 2, 8, 1}, {"芒种", 6, 3, 9, 1},
		{"夏至", 9, 3, 6, 0}, {"小暑", 8, 2, 5, 0}, {"大暑", 7, 1, 4, 0},
		{"立秋", 2, 5, 8, 0}, {"处暑", 1, 4, 7, 0}, {"白露", 9, 3, 6, 0},
		{"秋分", 7, 1, 4, 0}, {"寒露", 6, 9, 3, 0}, {"霜降", 5, 8, 2, 0},
		{"立冬", 6, 9, 3, 0}, {"小雪", 5, 8, 2, 0}, {"大雪", 4, 7, 1, 0},
	}
	if len(solarTermBureau) != 24 {
		t.Fatalf("solarTermBureau 长度 = %d, want 24", len(solarTermBureau))
	}
	for i, w := range want {
		e := solarTermBureau[i]
		if e[0] != w.shang || e[1] != w.zhong || e[2] != w.xia || e[3] != w.yin {
			t.Errorf("%s: 上中下/阴阳 = [%d %d %d %d], want [%d %d %d %d]",
				w.name, e[0], e[1], e[2], e[3], w.shang, w.zhong, w.xia, w.yin)
		}
	}
}

// 三元符头（拆补法）：子午卯酉=上元、寅申巳亥=中元、辰戌丑未=下元。
func TestJushu_DetermineYuan_FuTou(t *testing.T) {
	cases := []struct {
		gan  ganzhi.Gan
		zhi  ganzhi.Zhi
		want int
	}{
		{ganzhi.GanJia, ganzhi.ZhiZi, 0},   // 甲子（子=子午卯酉）→ 上元
		{ganzhi.GanJi, ganzhi.ZhiMao, 0},   // 己卯 → 上元
		{ganzhi.GanJia, ganzhi.ZhiWu, 0},   // 甲午 → 上元
		{ganzhi.GanJi, ganzhi.ZhiYou, 0},   // 己酉 → 上元
		{ganzhi.GanJi, ganzhi.ZhiSi, 1},    // 己巳（巳=寅申巳亥）→ 中元
		{ganzhi.GanJia, ganzhi.ZhiShen, 1}, // 甲申 → 中元
		{ganzhi.GanJi, ganzhi.ZhiHai, 1},   // 己亥 → 中元
		{ganzhi.GanJia, ganzhi.ZhiYin, 1},  // 甲寅 → 中元
		{ganzhi.GanJia, ganzhi.ZhiXu, 2},   // 甲戌（戌=辰戌丑未）→ 下元
		{ganzhi.GanJi, ganzhi.ZhiChou, 2},  // 己丑 → 下元
		{ganzhi.GanJia, ganzhi.ZhiChen, 2}, // 甲辰 → 下元
		{ganzhi.GanJi, ganzhi.ZhiWei, 2},   // 己未 → 下元
		// 非符头日按 15 日循环
		{ganzhi.GanYi, ganzhi.ZhiChou, 0}, // 乙丑（甲子后1日）→ 上元
		{ganzhi.GanGeng, ganzhi.ZhiWu, 1}, // 庚午（己巳后1日）→ 中元
		{ganzhi.GanYi, ganzhi.ZhiHai, 2},  // 乙亥（甲戌后1日）→ 下元
	}
	for _, c := range cases {
		if got := determineYuan(ganzhi.Zhu{Gan: c.gan, Zhi: c.zhi}); got != c.want {
			t.Errorf("determineYuan(%s%s) = %d, want %d（符头规则）",
				ganzhi.GanName(c.gan), ganzhi.ZhiName(c.zhi), got, c.want)
		}
	}
}

// 端到端：具体日期（拆补法定局链：节气 → 符头定元 → 定局表）→ 阴阳遁 + 局数。
// 锚点由「通行定局表 × 三元符头规则」独立手算确认。
func TestJushu_EndToEnd_Dates(t *testing.T) {
	cases := []struct {
		date    string
		wantJu  int
		wantYin bool
	}{
		{"2024-12-22", 4, false}, // 庚申日 下元 冬至 → 阳遁4局
		{"2024-06-22", 3, true},  // 丁巳日 中元 夏至 → 阴遁3局
		{"2025-01-06", 5, false}, // 乙亥日 下元 小寒 → 阳遁5局
		{"2024-03-20", 3, false}, // 癸未日 上元 春分 → 阳遁3局
		{"2026-06-28", 3, true},  // 癸酉日 中元 夏至 → 阴遁3局
	}
	for _, c := range cases {
		t.Run(c.date, func(t *testing.T) {
			// 北京时区当日中午 12:00（与 golden 等其他奇门测试一致）。
			bt, err := time.ParseInLocation("2006-01-02", c.date, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatal(err)
			}
			bt = bt.Add(12 * time.Hour)
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, "时家")
			if chart.Pan.Jushu != c.wantJu || chart.Pan.YinDun != c.wantYin {
				t.Errorf("%s: ju=%d yin=%v, want ju=%d yin=%v（拆补法）",
					c.date, chart.Pan.Jushu, chart.Pan.YinDun, c.wantJu, c.wantYin)
			}
		})
	}
}
