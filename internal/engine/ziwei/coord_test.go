package ziwei

import (
	"testing"
)

func TestZhiConversions(t *testing.T) {
	// Zhi ↔ zhiIdx 往返
	for z := Zhi(1); z <= 12; z++ {
		zm1 := zhiToZhiIdx(z)
		if zhiIdxToZhi(zm1) != z {
			t.Errorf("zhiIdxToZhi(zhiToZhiIdx(%d)) = %d, want %d", z, zhiIdxToZhi(zm1), z)
		}
	}
	// display ↔ zhiIdx 往返
	for d := 0; d < 12; d++ {
		zm1 := anXingIdxToZhiIdx(d)
		if zhiIdxToAnXingIdx(zm1) != d {
			t.Errorf("zhiIdxToAnXingIdx(anXingIdxToZhiIdx(%d)) = %d, want %d", d, zhiIdxToAnXingIdx(zm1), d)
		}
	}
	// 关键锚点：display 0=寅 → zhiIdx 2=寅；display 4=午 → zhiIdx 6=午
	if anXingIdxToZhiIdx(0) != 2 { t.Errorf("display0(寅)→zhiIdx, got %d want 2", anXingIdxToZhiIdx(0)) }
	if anXingIdxToZhiIdx(4) != 6 { t.Errorf("display4(午)→zhiIdx, got %d want 6", anXingIdxToZhiIdx(4)) }
	if zhiIdxToAnXingIdx(6) != 4 { t.Errorf("zhiIdx 6(午)→display, got %d want 4", zhiIdxToAnXingIdx(6)) }
}

func TestPalaceZhiRoundTrip(t *testing.T) {
	// palaceIndex ↔ zhiIdx 往返（锚定命宫支午 = zhiIdx 6）
	mingZM1 := 6 // 午
	for pi := palaceIndex(0); pi < 12; pi++ {
		zm1 := palaceIndexToZhiIdx(mingZM1, pi)
		if zhiIdxToPalaceIndex(mingZM1, zm1) != pi {
			t.Errorf("roundtrip palace %d: got %d", pi, zhiIdxToPalaceIndex(mingZM1, zm1))
		}
	}
	// 锚点：命宫(0)=午 → zhiIdx 6；兄弟(1)=巳 → 5
	if palaceIndexToZhiIdx(mingZM1, 0) != 6 { t.Errorf("palace0(命宫)=午, got zm1 %d want 6", palaceIndexToZhiIdx(mingZM1, 0)) }
	if palaceIndexToZhiIdx(mingZM1, 1) != 5 { t.Errorf("palace1(兄弟)=巳, got zm1 %d want 5", palaceIndexToZhiIdx(mingZM1, 1)) }
}

func TestBuildFlowPalaces(t *testing.T) {
	// flowIndex=4（午年 yearlyIndex）：names[i] = PALACES[(i-4)%12]，[4] = 命宫
	starByDisplay := map[int][]string{4: {"流羊"}, 3: {"流禄"}}
	flow := buildFlowPalaces(6, 4, starByDisplay) // 命宫午(zhiIdx 6), flowIndex 4
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
