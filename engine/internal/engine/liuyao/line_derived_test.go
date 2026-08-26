package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// computeLineDerived — 每爻确定性状态（月破/发动/动爻生克）命理锚点测试。
// 命理依据：《增删卜易》月建冲爻为月破；老阴/老阳发动；动爻生克决定爻的受生受克方向。

func makeDerivedChart(lines [6]Line, yueZhi ganzhi.Zhi, dongYao []int) Chart {
	return Chart{
		Lines:   lines,
		YueZhi:  yueZhi,
		DongYao: dongYao,
	}
}

// 子月（YueZhi=子）：初爻午火发动（老阳）。
// 午 被 子 冲 → 月破；初爻发动 → dong_self；午火克酉金、生未土。
func TestComputeLineDerived_YuePo_DongSelf_ShengKe(t *testing.T) {
	lines := [6]Line{
		{Position: 1, Type: LaoYang, Zhi: ganzhi.ZhiWu, Wuxing: ganzhi.WxHuo},   // 初爻午火，发动
		{Position: 2, Type: ShaoYang, Zhi: ganzhi.ZhiYou, Wuxing: ganzhi.WxJin}, // 二爻酉金
		{Position: 3, Type: ShaoYang, Zhi: ganzhi.ZhiWei, Wuxing: ganzhi.WxTu},  // 三爻未土
		{Position: 4, Type: ShaoYang, Zhi: ganzhi.ZhiMao, Wuxing: ganzhi.WxMu},  // 四爻卯木
		{Position: 5, Type: ShaoYang, Zhi: ganzhi.ZhiChen, Wuxing: ganzhi.WxTu}, // 五爻辰土
		{Position: 6, Type: ShaoYang, Zhi: ganzhi.ZhiSi, Wuxing: ganzhi.WxHuo},  // 上爻巳火
	}
	chart := makeDerivedChart(lines, ganzhi.ZhiZi, []int{1})
	computeLineDerived(&chart)

	// 月破：子冲午（初爻午、上爻巳不受冲；辰/未/酉/卯不受子冲）
	if !chart.Lines[0].YuePo {
		t.Error("line1 午火在子月应月破(yue_po=true)")
	}
	if chart.Lines[1].YuePo || chart.Lines[2].YuePo || chart.Lines[3].YuePo || chart.Lines[4].YuePo || chart.Lines[5].YuePo {
		t.Error("line2-6 不应月破")
	}

	// 发动：初爻老阳发动
	if !chart.Lines[0].DongSelf {
		t.Error("line1 老阳应发动(dong_self=true)")
	}
	for i := 1; i < 6; i++ {
		if chart.Lines[i].DongSelf {
			t.Errorf("line%d 少阳不应发动", i+1)
		}
	}

	// 动爻生克：午火克酉金、生未土/辰土；与卯木/巳火无生克
	if !chart.Lines[1].DongKe {
		t.Error("line2 酉金应被动爻午火克(dong_ke=true)")
	}
	if !chart.Lines[2].DongSheng {
		t.Error("line3 未土应被动爻午火生(dong_sheng=true)")
	}
	if chart.Lines[3].DongSheng || chart.Lines[3].DongKe {
		t.Error("line4 卯木与午火无生克，不应有标记")
	}
	if chart.Lines[5].DongSheng || chart.Lines[5].DongKe {
		t.Error("line6 巳火与午火比和，不应有生克标记")
	}
}

// 静卦：全少阳无动爻 → 无发动/生克标记；但月破仍按月建判定。
func TestComputeLineDerived_JingGua(t *testing.T) {
	lines := [6]Line{
		{Position: 1, Type: ShaoYang, Zhi: ganzhi.ZhiWu, Wuxing: ganzhi.WxHuo}, // 初爻午火
		{Position: 2, Type: ShaoYang, Zhi: ganzhi.ZhiChen, Wuxing: ganzhi.WxTu},
		{Position: 3, Type: ShaoYang, Zhi: ganzhi.ZhiYin, Wuxing: ganzhi.WxMu},
		{Position: 4, Type: ShaoYang, Zhi: ganzhi.ZhiHai, Wuxing: ganzhi.WxShui},
		{Position: 5, Type: ShaoYang, Zhi: ganzhi.ZhiYou, Wuxing: ganzhi.WxJin},
		{Position: 6, Type: ShaoYang, Zhi: ganzhi.ZhiSi, Wuxing: ganzhi.WxHuo},
	}
	chart := makeDerivedChart(lines, ganzhi.ZhiZi, nil)
	computeLineDerived(&chart)

	for i := 0; i < 6; i++ {
		if chart.Lines[i].DongSelf || chart.Lines[i].DongSheng || chart.Lines[i].DongKe {
			t.Errorf("静卦 line%d 不应有发动/生克标记", i+1)
		}
	}
	// 子月冲午：初爻午火月破
	if !chart.Lines[0].YuePo {
		t.Error("静卦中 line1 午火在子月仍应月破")
	}
}

// 旬空：日柱己卯属甲戌旬（SixtyCycleIndex(己,卯)=15，旬1），空申酉。
// 命理依据：《增删卜易》用神旬空则事虚、出空填实方应——爻地支值日旬空则标 xun_kong。
func TestComputeLineDerived_XunKong(t *testing.T) {
	if got := ganzhi.XunKong(ganzhi.GanJi, ganzhi.ZhiMao); got != [2]ganzhi.Zhi{9, 10} {
		t.Fatalf("XunKong(己卯) = %v, want [申 酉]", got)
	}
	lines := [6]Line{
		{Position: 1, Type: ShaoYang, Zhi: ganzhi.ZhiZi, Wuxing: ganzhi.WxShui},
		{Position: 2, Type: ShaoYang, Zhi: ganzhi.ZhiYin, Wuxing: ganzhi.WxMu},
		{Position: 3, Type: ShaoYang, Zhi: ganzhi.ZhiChen, Wuxing: ganzhi.WxTu},
		{Position: 4, Type: ShaoYang, Zhi: ganzhi.ZhiWu, Wuxing: ganzhi.WxHuo},
		{Position: 5, Type: ShaoYang, Zhi: ganzhi.ZhiShen, Wuxing: ganzhi.WxJin}, // 申 → 旬空
		{Position: 6, Type: ShaoYang, Zhi: ganzhi.ZhiXu, Wuxing: ganzhi.WxTu},
	}
	chart := makeDerivedChart(lines, ganzhi.ZhiYin, nil)
	chart.XunKong = [2]ganzhi.Zhi{9, 10} // 申酉
	computeLineDerived(&chart)

	for i := 0; i < 6; i++ {
		want := i == 4 // 五爻申值旬空
		if chart.Lines[i].XunKong != want {
			t.Errorf("line%d XunKong = %v, want %v", i+1, chart.Lines[i].XunKong, want)
		}
	}
}
