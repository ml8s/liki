package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// ── 六爻断卦测试 ──

func TestLiuYao_Judgment_YongShenMapping(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		wantName string
	}{
		{name: "事业→官鬼", event: "career", wantName: "官鬼"},
		{name: "求财→妻财", event: "wealth", wantName: "妻财"},
		{name: "感情→官鬼", event: "relationship", wantName: "官鬼"},
		{name: "学业→父母", event: "study", wantName: "父母"},
		{name: "健康→官鬼", event: "health", wantName: "官鬼"},
		{name: "诉讼→官鬼", event: "legal", wantName: "官鬼"},
		{name: "出行→世爻(待补应爻)", event: "travel", wantName: "世爻"},
		{name: "默认→世爻", event: "", wantName: "世爻"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeJudgment(testChart(tt.event), tt.event)
			if result.YongShen.Name != tt.wantName {
				t.Errorf("YongShen.Name=%q, want=%q (event=%q)",
					result.YongShen.Name, tt.wantName, tt.event)
			}
		})
	}
}

func TestLiuYao_Judgment_Rating(t *testing.T) {
	tests := []struct {
		name   string
		ysType YongShen
		yongPos int
		monthZhi, dayZhi ganzhi.Zhi
		yaoType int
		yongZhi ganzhi.Zhi
		wantRating string
		wantRule   int
		desc   string
	}{
		{
			name: "用神旺相日建生→吉",
			ysType: YongQiCai, yongPos: 3,
			monthZhi: ganzhi.ZhiWu, dayZhi: ganzhi.ZhiSi,
			yaoType: 7, yongZhi: ganzhi.ZhiWu,
			wantRating: "吉", wantRule: 6,
			desc: "午月午火→妻财午火旺, line3持世→ rule6 持世旺相→吉",
		},
		{
			name: "死+日冲→凶(rule5)",
			ysType: YongQiCai, yongPos: 5,
			monthZhi: ganzhi.ZhiShen, dayZhi: ganzhi.ZhiZi,
			yaoType: 6, yongZhi: ganzhi.ZhiWu,
			wantRating: "凶", wantRule: 5,
			desc: "申月→妻财午火死, 子日午冲→日衰, line5非世→ rule5 囚死+日衰+非持世→凶",
		},
		{
			name: "月破→凶(rule1)",
			ysType: YongGuanGui, yongPos: 4,
			monthZhi: ganzhi.ZhiZi, dayZhi: ganzhi.ZhiChen,
			yaoType: 6, yongZhi: ganzhi.ZhiWu,
			wantRating: "凶", wantRule: 1,
			desc: "子月午冲→月破, 不论其他→ rule1",
		},
		{
			name: "持世+旺相→吉(rule6)",
			ysType: YongFumu, yongPos: 3,
			monthZhi: ganzhi.ZhiMao, dayZhi: ganzhi.ZhiChen,
			yaoType: 6, yongZhi: ganzhi.ZhiMao,
			wantRating: "吉", wantRule: 6,
			desc: "卯月卯木→父母旺, line3持世→ rule6",
		},
		{
			name: "月建休→平(rule9)",
			ysType: YongQiCai, yongPos: 2,
			monthZhi: ganzhi.ZhiSi, dayZhi: ganzhi.ZhiChen,
			yaoType: 6, yongZhi: ganzhi.ZhiYin,
			wantRating: "平", wantRule: 9,
			desc: "巳月巳火→妻财寅木休, line2非世→ rule9 休→平",
		},
		{
			name: "旺相→吉(rule6世持)",
			ysType: YongZiSun, yongPos: 5,
			monthZhi: ganzhi.ZhiChen, dayZhi: ganzhi.ZhiYou,
			yaoType: 6, yongZhi: ganzhi.ZhiChen,
			wantRating: "吉", wantRule: 6,
			desc: "辰月辰土旺, line3世爻→ rule6 持世旺相→吉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := makeJudgmentChart(tt.ysType, tt.yongPos, tt.monthZhi, tt.dayZhi, tt.yaoType, tt.yongZhi)
			for i := range c.Lines {
				c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
			}
			result := ComputeJudgment(c, eventForYS(tt.ysType))
			if result.Rating != tt.wantRating || result.Rule != tt.wantRule {
				t.Errorf("rating=%q rule=%d, want %q rule=%d\n  desc: %s",
					result.Rating, result.Rule, tt.wantRating, tt.wantRule, tt.desc)
			}
		})
	}
}

// eventForYS returns a canonical event for a given yong shen type.
// The event is used to look up the yong shen in eventYongShen.
// Not all yong shen types have a matching event — YongZiSun lacks one.
// For these, we return "general" since general→YongShiYao works reliably.
func eventForYS(ys YongShen) string {
	switch ys {
	case YongGuanGui: return "career"
	case YongQiCai: return "wealth"
	case YongFumu: return "study"
	case YongZiSun: return "general" // 无专挂事件, 用general(世爻)
	default: return "general"
	}
}

func TestLiuYao_Judgment_AdviceNonEmpty(t *testing.T) {
	events := []string{"general", "career", "wealth", "relationship", "study", "health", "legal", "travel"}
	for _, event := range events {
		t.Run(event, func(t *testing.T) {
			result := ComputeJudgment(testChart(event), event)
			if result.Advice == "" {
				t.Errorf("Advice empty, event=%q", event)
			}
		})
	}
}

// ── 辅助 ──

func testChart(event string) Chart {
	return makeJudgmentChart(YongShiYao, 3, ganzhi.ZhiWu, ganzhi.ZhiChen, 7, ganzhi.ZhiWu)
}

// makeJudgmentChart builds a Chart with specific配置.
func makeJudgmentChart(ysType YongShen, yongPos int, monthZhi, dayZhi ganzhi.Zhi, yaoType int, yongZhi ganzhi.Zhi) Chart {
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		z := yongZhi
		lq := setLiuQinForYS(ysType, i+1, yongPos)
		lines[i] = Line{
			Position: i + 1,
			Type:     YaoType(yaoType),
			Zhi:      z,
			LiuQin:  lq,
		}
	}

	lines[2].ShiYing = "世"
	lines[5].ShiYing = "应"
	if ysType == YongShiYao && yongPos > 0 {
		lines[yongPos-1].ShiYing = "世"
	}

	return Chart{
		Lines:    lines,
		YongShen: YongShenResult{Name: ysType.String(), Position: yongPos},
		DayGan:   ganzhi.GanJia,
		DayZhi:   dayZhi,
		MonthZhi: monthZhi,
		MonthGan: ganzhi.GanJia,
		Palace:   "离",
		PalaceWuxing: ganzhi.WxHuo,
	}
}

func setLiuQinForYS(ysType YongShen, pos, yongPos int) LiuQin {
	if pos == yongPos {
		switch ysType {
		case YongFumu:
			return QinFumu
		case YongXiongDi:
			return QinXiongDi
		case YongGuanGui:
			return QinGuanGui
		case YongQiCai:
			return QinQiCai
		case YongZiSun:
			return QinZiSun
		}
	}
	return QinXiongDi // non-matching
}

func TestLiuYao_Judgment_JingGua(t *testing.T) {
	// 静卦: 无动爻, bian_lines 的 gan/zhi 为空字符串
	// 命理: 静卦也应能正常断卦, 不报错
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		zhi := ganzhi.Zhi(i + 1)
		lines[i] = Line{
			Position: i + 1,
			Type:     6,
			Zhi:      zhi,
			LiuQin:  QinXiongDi,
		}
	}
	lines[0].ShiYing = "世"
	lines[2].ShiYing = "应"

	// Bian lines with empty gan/zhi (静卦)
	bianLines := [6]Line{}
	for i := 0; i < 6; i++ {
		bianLines[i] = Line{
			Position: i + 1,
			Type:     6,
			LiuQin:  QinXiongDi,
		}
	}

	c := Chart{
		Lines:     lines,
		BianLines: bianLines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiWu,
		MonthZhi: ganzhi.ZhiYou,
		DongYao:  []int{},
	}
	for i := range c.Lines {
		c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
	}

	result := ComputeJudgment(c, "general")
	if result.Rating == "" {
		t.Error("静卦 rating 为空")
	}
	if result.YongShen.Name == "" {
		t.Error("静卦 yongshen 为空")
	}
	t.Logf("静卦: rating=%q, yongshen=%s", result.Rating, result.YongShen.Name)
}

func TestLiuYao_Judgment_JingGuaJSON(t *testing.T) {
	// 直接用 Go struct 构造静卦, 模拟 APa 传入的场景
	lines := [6]Line{
		{Position: 1, Type: 6, LiuQin: QinZiSun, ShiYing: "世"},
		{Position: 2, Type: 6, LiuQin: QinQiCai},
		{Position: 3, Type: 6, LiuQin: QinXiongDi},
		{Position: 4, Type: 6, LiuQin: QinGuanGui},
		{Position: 5, Type: 6, LiuQin: QinFumu},
		{Position: 6, Type: 6, LiuQin: QinZiSun, ShiYing: "应"},
	}
	// 静卦: BianLines 全零值(gan/zhi为空字符串)
	var bianLines [6]Line

	chart := Chart{
		Lines:     lines,
		BianLines: bianLines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiZi,
		MonthZhi: ganzhi.ZhiYou,
		BenGua:   21,
		DongYao:  []int{},
	}
	// 初始化 WangShuai
	for i := range chart.Lines {
		chart.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(chart.Lines[i].Zhi), chart.MonthZhi)
	}

	// 不应崩溃
	result := ComputeJudgment(chart, "general")
	if result.Rating == "" {
		t.Error("静卦 rating 为空")
	}
	t.Logf("静卦JSON: rating=%q, yongshen=%s", result.Rating, result.YongShen.Name)
}
