package ziwei

// ── 坐标系统一辅助层 ──────────────────────────────────────────────
// 内部坐标约定（全部领域根表达）：
//   zhiIdx (int)      : 地支索引，0=子 1=丑 ... 11=亥 —— 地支历法根，安星最终落在此
//   anXingIdx (int)   : 安星索引，0=寅 1=卯 ... 11=丑 —— 紫微「定寅首」安星坐标系
//   gongIndex       : 命宫序，0=命宫 ... 11=父母 —— 仅用于输出层
// 安星/流盘内部计算统一用 zhiIdx；gongIndex 只在 Chart.GongWei 输出时生成。
// 转换关系：
//   zhiIdx → Zhi        : +1
//   Zhi → zhiIdx        : -1
//   anXingIdx → zhiIdx  : (anXingIdx + 2) % 12   （寅=0 → 子=2）
//   zhiIdx → anXingIdx  : (zhiIdx - 2 + 12) % 12
//   zhiIdx → gongIndex（命宫支锚定）: 见 palace_map.go zhiIdxToPalaceIndex

// zhiIdxToZhi converts earth zhi index (0=子) to Zhi (1-12).
func zhiIdxToZhi(zhiIdx int) Zhi {
	return Zhi((zhiIdx%12+12)%12 + 1)
}

// zhiToZhiIdx converts Zhi (1-12) to earth zhi index (0=子).
func zhiToZhiIdx(z Zhi) int {
	return (int(z) - 1 + 12) % 12
}

// anXingIdxToZhiIdx converts 安星索引 (寅=0 安星序) to 地支索引 (子=0).
func anXingIdxToZhiIdx(anXingIdx int) int {
	return (2 + anXingIdx) % 12
}

// zhiIdxToAnXingIdx converts 地支索引 (子=0) to 安星索引 (寅=0 安星序).
func zhiIdxToAnXingIdx(zhiIdx int) int {
	return ((zhiIdx-2)%12 + 12) % 12
}

// flowPalace is one gong of a flow chart (流年/流月/流日/流时盘),
// expressed in earth-zhi coordinates.
type flowPalace struct {
	Zhi    Zhi      `json:"zhi"`         // 地支
	Name   string   `json:"name"`        // 宫名（命盘标签）
	Stars  []string `json:"xing_yao"`    // 流耀星名（无则为空）
	IsMing bool     `json:"is_liu_ming"` // 是否流盘命宫
}

// buildFlowPalaces builds a 12-gong flow chart in 安星序 (寅=0 ... 丑=11).
// 真相源：宫名 = 地支 → gongIndex → 标签，顺逆由公式表达，不存宫名顺序数组。
// liuPanIdx 是流盘索引（流年/月/日/时），用于确定流盘命宫（IsMing 标记）。
// liuPanIdx 是流盘起点（安星索引），用于宫名推导。
func buildFlowPalaces(liuPanIdx int, starByAnXingIdx map[int][]string) [12]flowPalace {
	var out [12]flowPalace
	for i := 0; i < 12; i++ {
		zhiIdx := anXingIdxToZhiIdx(i)
		// 流盘宫名 = gongLabels（根）的反向索引（顺时针序）：
		//   流盘序 names[i] = PALACES[(i - liuPanIdx) % 12]
		palIdx := ((i-liuPanIdx)%12 + 12) % 12
		name := gongLabels[(12-palIdx)%12]
		stars := starByAnXingIdx[i]
		if stars == nil {
			stars = []string{}
		}
		out[i] = flowPalace{
			Zhi:    zhiIdxToZhi(zhiIdx),
			Name:   name,
			Stars:  stars,
			IsMing: name == "命宫",
		}
	}
	return out
}
