package ziwei

import (
	"testing"
)

// ── ComputeJudgment integration tests ──
// Each test constructs a specific constellation (same pattern as findPatterns tests)
// and validates that ComputeJudgment returns the expected patterns + rating.

func TestJudgment_ZiWeiChaoYuan_ShowsRating(t *testing.T) {
	// 紫微朝垣: 紫微在命宫. 高等级格局→ rating不应为"下"
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	palaces[0] = palace{Index: 0, Zhi: 3, Stars: []starInfo{{Star: ZiWei, Name: "紫微", IsMajor: true}}}
	// 紫微在寅=庙
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	found := false
	for _, p := range result.Patterns {
		if p.Name == "紫微朝垣" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'紫微朝垣' pattern not found in judgment result")
	}
	if result.Rating == "下" {
		t.Errorf("紫微在命宫→rating=%q, should not be '下'", result.Rating)
	}
	t.Logf("patterns=%d, rating=%q, rule=%d", len(result.Patterns), result.Rating, result.Rule)
}

func TestJudgment_RiLiZhongTian_FindsPattern(t *testing.T) {
	// 日丽中天: 太阳居午(宫位6=迁移宫)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	palaces[6] = palace{Index: 6, Zhi: 7, Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	for _, p := range result.Patterns {
		if p.Name == "日丽中天" {
			return
		}
	}
	t.Error("'日丽中天' pattern not found in judgment result")
}

func TestJudgment_HuoTanGe_FindsPattern(t *testing.T) {
	// 火贪格: 火星+贪狼在命宫, 贪狼不陷(丑)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	palaces[0] = palace{Index: 0, Zhi: 2, Stars: []starInfo{
		{Star: HuoXing, Name: "火星", IsMajor: false},
		{Star: TanLang, Name: "贪狼", IsMajor: true},
	}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	for _, p := range result.Patterns {
		if p.Name == "火贪格" {
			return
		}
	}
	t.Error("'火贪格' pattern not found in judgment result")
}

func TestJudgment_ShaPoLang_FindsPattern(t *testing.T) {
	// 杀破狼: 命宫三方(命/财/事)有七杀+破军+贪狼
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	// 命宫0: 七杀, 财帛宫4: 破军, 官禄宫8: 贪狼
	palaces[0] = palace{Index: 0, Zhi: 1, Stars: []starInfo{{Star: QiSha, Name: "七杀", IsMajor: true}}}
	palaces[4] = palace{Index: 4, Zhi: 5, Stars: []starInfo{{Star: PoJun, Name: "破军", IsMajor: true}}}
	palaces[8] = palace{Index: 8, Zhi: 9, Stars: []starInfo{{Star: TanLang, Name: "贪狼", IsMajor: true}}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	for _, p := range result.Patterns {
		if p.Name == "杀破狼" {
			return
		}
	}
	t.Error("'杀破狼' pattern not found in judgment result")
}

func TestJudgment_RiYueFanBei_FindsPattern(t *testing.T) {
	// 日月反背: 太阳+太阴分别在落陷宫位
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	// 太阳在子(陷), 太阴在卯(陷)
	palaces[1] = palace{Index: 1, Zhi: 1, Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}}}
	palaces[3] = palace{Index: 3, Zhi: 4, Stars: []starInfo{{Star: TaiYin, Name: "太阴", IsMajor: true}}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	for _, p := range result.Patterns {
		if p.Name == "日月反背" {
			return
		}
	}
	// This pattern requires both stars in陷 at ANY palace, which won't match our setup
	// since most palaces are randomly assigned Zhi. This test may not fire.
	t.Error("日月反背未触发: 太阳子陷+太阴卯陷应满足条件")
}

func TestJudgment_RiYueBingMing_FindsPattern(t *testing.T) {
	// 日月并明: 太阳太阴都在庙旺宫位(命/迁移/官禄/福德)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	// 太阳在午=庙, 太阴在亥=庙 (都在brightPalace list中)
	// 午=7, 亥=12. 宫位index: 太阳在午→哪个宫位? 需要Zhi=7的宫.
	// 宫位Zhi=7是午, 对应宫位index 6(迁移宫). 迁移宫是brightPalaces之一(6).
	// 太阴在亥(Zhi=12), 对应宫位index 11(父母宫). 父母宫不在brightPalaces列表!
	// brightPalaces list: 0(命), 6(迁移), 8(官禄), 10(福德)
	// Need to place them in correct宫位.
	// Let's put太阳在命宫(Zhi=7=午=庙), 太阴在迁移宫(Zhi=12=亥=庙)
	// 迁移宫index=6, 所以太阳在命宫(Zhi=7), 太阴在迁移宫(Zhi=12, index=6)
	palaces[0] = palace{Index: 0, Zhi: 7, Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}}}
	palaces[6] = palace{Index: 6, Zhi: 12, Stars: []starInfo{{Star: TaiYin, Name: "太阴", IsMajor: true}}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)

	for _, p := range result.Patterns {
		if p.Name == "日月并明" {
			return
		}
	}
	t.Error("日月并明未触发: 太阳午庙+太阴亥庙应满足条件")
}
