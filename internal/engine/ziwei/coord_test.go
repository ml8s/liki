package ziwei

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestZhiConversions(t *testing.T) {
	// Zhi ↔ zhiMinus1 往返
	for z := Zhi(1); z <= 12; z++ {
		zm1 := zhiToZhiMinus1(z)
		if zhiMinus1ToZhi(zm1) != z {
			t.Errorf("zhiMinus1ToZhi(zhiToZhiMinus1(%d)) = %d, want %d", z, zhiMinus1ToZhi(zm1), z)
		}
	}
	// display ↔ zhiMinus1 往返
	for d := 0; d < 12; d++ {
		zm1 := displayToZhiMinus1(d)
		if zhiMinus1ToDisplay(zm1) != d {
			t.Errorf("zhiMinus1ToDisplay(displayToZhiMinus1(%d)) = %d, want %d", d, zhiMinus1ToDisplay(zm1), d)
		}
	}
	// 关键锚点：display 0=寅 → zhiMinus1 2=寅；display 4=午 → zhiMinus1 6=午
	if displayToZhiMinus1(0) != 2 { t.Errorf("display0(寅)→zhiMinus1, got %d want 2", displayToZhiMinus1(0)) }
	if displayToZhiMinus1(4) != 6 { t.Errorf("display4(午)→zhiMinus1, got %d want 6", displayToZhiMinus1(4)) }
	if zhiMinus1ToDisplay(6) != 4 { t.Errorf("zhiMinus1 6(午)→display, got %d want 4", zhiMinus1ToDisplay(6)) }
}

func TestPalaceZhiRoundTrip(t *testing.T) {
	// palaceIndex ↔ zhiMinus1 往返（锚定命宫支午）
	mingZhi := Zhi(7) // 午
	for pi := palaceIndex(0); pi < 12; pi++ {
		zm1 := palaceToZhiMinus1(pi, mingZhi)
		if zhiMinus1ToPalace(zm1, mingZhi) != pi {
			t.Errorf("roundtrip palace %d: got %d", pi, zhiMinus1ToPalace(zm1, mingZhi))
		}
	}
	// 锚点：命宫(0)=午 → zhiMinus1 6；兄弟(1)=巳 → 5
	if palaceToZhiMinus1(0, mingZhi) != 6 { t.Errorf("palace0(命宫)=午, got zm1 %d want 6", palaceToZhiMinus1(0, mingZhi)) }
	if palaceToZhiMinus1(1, mingZhi) != 5 { t.Errorf("palace1(兄弟)=巳, got zm1 %d want 5", palaceToZhiMinus1(1, mingZhi)) }
}

func TestBuildFlowPalaces(t *testing.T) {
	// 2026 午年：yearlyIndex = display(午) = 4
	// names[i] = PALACES[(i-4)%12]，[4] = 命宫
	c := ComputeChart(tianwen.LunarTime{Year: 2000, Month: 7, Day: 17, Shichen: 3}, ganzhi.Female)
	flowMing := Zhi(7) // 午
	// display 坐标：午=4 → 流羊；巳=3 → 流禄
	starByDisplay := map[int][]string{4: {"流羊"}, 3: {"流禄"}}
	flow := buildFlowPalaces(c, flowMing, starByDisplay)
	// display 序：地支 寅卯辰...；宫名 PALACES 旋转
	if flow[0].Zhi.String() != "寅" { t.Errorf("[0]支: got %s want 寅", flow[0].Zhi.String()) }
	if flow[1].Zhi.String() != "卯" { t.Errorf("[1]支: got %s want 卯", flow[1].Zhi.String()) }
	// [4] = PALACES[(4-4)%12] = PALACES[0] = 命宫（流年命宫）
	if !flow[4].IsMing || flow[4].Name != "命宫" { t.Errorf("[4]应为流年命宫: %s is_ming=%v", flow[4].Name, flow[4].IsMing) }
	// [4] 含流羊（午位）
	if len(flow[4].Stars) != 1 || flow[4].Stars[0] != "流羊" { t.Errorf("[4]流耀: %v", flow[4].Stars) }
	// [3] 含流禄（巳位）
	if len(flow[3].Stars) != 1 || flow[3].Stars[0] != "流禄" { t.Errorf("[3]流耀: %v", flow[3].Stars) }
}
