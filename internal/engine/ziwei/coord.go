package ziwei

// ── 坐标系统一辅助层 ──────────────────────────────────────────────
// 内部坐标约定：
//   zhiMinus1 (int)  : 地支坐标，0=子 1=丑 ... 11=亥 —— 命盘天然坐标系，安星最终落在此
//   palaceIndex      : 命宫序坐标，0=命宫 ... 11=父母 —— 仅用于输出层
//   displayIdx       : iztro 坐标，0=寅 1=卯 ... 11=丑 —— iztro 内部盘序
// 安星/流盘等内部计算统一用 zhiMinus1；palaceIndex 只在 Chart.Palaces 输出时生成。
// 转换关系：
//   zhiMinus1 → Zhi   : +1
//   Zhi → zhiMinus1   : -1
//   displayIdx → zhiMinus1 : (displayIdx + 2) % 12   （寅=0 → 子=2）
//   zhiMinus1 → displayIdx : (zhiMinus1 - 2 + 12) % 12
//   zhiMinus1 → palaceIndex（以命宫支为锚）: (mingZhi-1 - zhiMinus1 + 12) % 12

// zhiMinus1ToZhi converts 0-based earth branch index to Zhi (1-12).
func zhiMinus1ToZhi(zm1 int) Zhi {
	return Zhi((zm1%12 + 12) % 12 + 1)
}

// zhiToZhiMinus1 converts Zhi (1-12) to 0-based earth branch index.
func zhiToZhiMinus1(z Zhi) int {
	return (int(z)-1+12) % 12
}

// displayToZhiMinus1 converts iztro display index (寅=0) to zhiMinus1 (子=0).
// 已有此函数（adjective.go），此处保留引用说明；避免重复定义用同包。

// zhiMinus1ToDisplay converts zhiMinus1 (子=0) to iztro display index (寅=0).
func zhiMinus1ToDisplay(zm1 int) int {
	return ((zm1-2)%12 + 12) % 12
}

// zhiMinus1ToPalace converts earth branch index to palace index, anchored at 命宫支.
// 已有 zhiToPalace(zhiMinus1, mingZhi)，语义相同；此处为显式命名。
func zhiMinus1ToPalace(zm1 int, mingZhi Zhi) palaceIndex {
	return zhiToPalace(zm1, mingZhi)
}

// palaceToZhiMinus1 converts palace index (0=命宫) back to earth branch index.
// Inverse of zhiMinus1ToPalace.
func palaceToZhiMinus1(pi palaceIndex, mingZhi Zhi) int {
	mingZM1 := zhiToZhiMinus1(mingZhi)
	return (mingZM1 - int(pi) + 12) % 12
}

// flowPalace is one palace of a flow chart (流年/流月/流日/流时盘),
// expressed in earth-branch coordinates.
type flowPalace struct {
	Zhi    Zhi      `json:"zhi"`     // 地支
	Name   string   `json:"name"`    // 宫名（命盘标签）
	Stars  []string `json:"stars"`   // 流耀星名（无则为空）
	IsMing bool     `json:"is_ming"` // 是否流盘命宫
}

// iztroPALACES is iztro's fixed palace-name order (PALACES array).
// Flow charts (流年/月/日/时盘) use this order rotated by the flow index.
var iztroPALACES = [12]string{"命宫", "父母", "福德", "田宅", "官禄", "仆役", "迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟"}

// buildFlowPalaces builds a 12-palace flow chart matching iztro's coordinate
// system. The truth source is the display index (寅=0 ... 丑=11):
//   - palace i is display position i, with fixed earth branch
//   - palace name = iztro PALACES rotated by flowIndex:
//     names[i] = PALACES[(i - flowIndex) % 12]
//   - flow stars located by display index (starByDisplay)
// flowIndex is the iztro flow index (yearlyIndex / monthlyIndex / dailyIndex / hourlyIndex).
func buildFlowPalaces(flowIndex int, starByDisplay map[int][]string) [12]flowPalace {
	var out [12]flowPalace
	for i := 0; i < 12; i++ {
		zm1 := displayToZhiMinus1(i)
		palIdx := ((i - flowIndex) % 12 + 12) % 12
		out[i] = flowPalace{
			Zhi:    zhiMinus1ToZhi(zm1),
			Name:   iztroPALACES[palIdx],
			Stars:  starByDisplay[i],
			IsMing: iztroPALACES[palIdx] == "命宫",
		}
	}
	return out
}
