package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// ── 六爻断卦规则链命理测试 ──
// 用已知卦象+世应+月日验证综合评定正确

func TestLiuYao_Judgment_MonthPo_BigJi(t *testing.T) {
	// 命理: 月破(月冲用神支)→凶, rule1优先
	// 世爻在未, 月丑冲未 → yongShen=世爻, yuePo=true
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		lines[i] = Line{Position: i + 1, Type: 6, Zhi: ganzhi.Zhi(i + 1), LiuQin: QinXiongDi}
	}
	// 月破: 未(6)与丑(2)六冲
	lines[1] = Line{Position: 2, Type: 6, Zhi: ganzhi.ZhiWei, LiuQin: QinXiongDi, ShiYing: "世"}
	lines[5] = Line{Position: 6, Type: 6, Zhi: ganzhi.ZhiChou, LiuQin: QinXiongDi, ShiYing: "应"}

	c := Chart{
		Lines:    lines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiChen,
		MonthZhi: ganzhi.ZhiChou, // 丑月→冲未
	}
	for i := range c.Lines {
		c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
	}

	result := ComputeJudgment(c, "general")
	if result.Rating != "凶" {
		t.Errorf("月破→ rating=%q, 命理应为凶", result.Rating)
	}
	if result.Rule != 1 {
		t.Logf("月破 rule=%d, 期望1", result.Rule)
	}
}

func TestLiuYao_Judgment_ChiShiWangXiang_Ji(t *testing.T) {
	// 命理: 持世+旺相→吉(rule6)
	// 午月午火旺, 世爻在午→chi_shi=true, month=旺
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		lines[i] = Line{Position: i + 1, Type: 6, Zhi: ganzhi.Zhi(i + 1), LiuQin: QinXiongDi}
	}
	lines[2] = Line{Position: 3, Type: 6, Zhi: ganzhi.ZhiWu, LiuQin: QinGuanGui, ShiYing: "世"}
	lines[5] = Line{Position: 6, Type: 6, Zhi: ganzhi.ZhiZi, LiuQin: QinXiongDi, ShiYing: "应"}

	c := Chart{
		Lines:    lines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiChen,
		MonthZhi: ganzhi.ZhiWu, // 午月→午火旺
	}
	for i := range c.Lines {
		c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
	}

	result := ComputeJudgment(c, "career")
	if result.Rating != "吉" {
		t.Errorf("持世旺相→ rating=%q, 命理应为吉", result.Rating)
	}
}

func TestLiuYao_Judgment_ShangGuaBuShang_Ji(t *testing.T) {
	// 命理: 用神不上卦→凶(rule2)
	// 卦中无官鬼爻→求官用神不上卦
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		lines[i] = Line{Position: i + 1, Type: 6, Zhi: ganzhi.Zhi(i + 1), LiuQin: QinZiSun}
	}
	lines[2].ShiYing = "世"
	lines[5].ShiYing = "应"

	c := Chart{
		Lines:    lines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiWu,
		MonthZhi: ganzhi.ZhiYou,
		BenGua:   30, // 离卦
	}
	for i := range c.Lines {
		c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
	}

	result := ComputeJudgment(c, "career")
	if result.Rating != "凶" {
		t.Errorf("用神不上卦→ rating=%q, 命理应为凶", result.Rating)
	}
}

func TestLiuYao_Judgment_Xiu_FallbackPing(t *testing.T) {
	// 命理: 月建休+无其他因素→平(rule9)
	lines := [6]Line{}
	for i := 0; i < 6; i++ {
		lines[i] = Line{Position: i + 1, Type: 6, Zhi: ganzhi.Zhi(i + 1), LiuQin: QinXiongDi}
	}
	lines[2].ShiYing = "世"
	lines[5].ShiYing = "应"
	// 用神为巳火在寅月→休
	lines[3] = Line{Position: 4, Type: 6, Zhi: ganzhi.ZhiSi, LiuQin: QinGuanGui}

	c := Chart{
		Lines:    lines,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiWu,
		MonthZhi: ganzhi.ZhiYin, // 寅月→巳火休
	}
	for i := range c.Lines {
		c.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(c.Lines[i].Zhi), c.MonthZhi)
	}

	result := ComputeJudgment(c, "career")
	if result.Rating != "平" {
		t.Errorf("月休→ rating=%q, 命理应为平", result.Rating)
	}
}
