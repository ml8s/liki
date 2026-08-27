package ziwei

// ── iztro 流盘 index 公式（display 坐标 寅=0）─────────────────────
//   monthlyIndex = fixEB(流年支) - 生月(含闰补) + 生时支(子=0) + 目标农历月(含闰补)
//   dailyIndex   = monthlyIndex + 目标农历日 - 1
//   hourlyIndex  = dailyIndex + 目标时支(子=0)
// 生时支(子=0) = chart.ShiZhi - 1
// 闰补: 出生或目标为闰月且日>15 → +1

// flowTarget is the target date for flow month/day/hour.
// 流月只用 Year + LunarMonth；流日需 LunarDay；流时需 ShiZhi。
type flowTarget struct {
	Year        int  // 目标公历年
	LunarMonth  int  // 目标农历月
	LunarDay    int  // 目标农历日（流日/流时用）
	IsLeapMonth bool // 目标是否闰月
	ShiZhi      Zhi  // 目标时支（流时用）
}

// computeMonthlyIndex returns iztro's monthly index (display 寅=0).
func computeMonthlyIndex(chart Chart, t flowTarget) int {
	// 流年支 = 目标年份的支 → fixEB（display 寅=0）
	_, liuZhi := yearGanZhi(t.Year)
	yi := zhiIdxToAnXingIdx(zhiToZhiIdx(liuZhi))
	// 生月 + 闰补
	birthLeap := 0
	if chart.BirthIsLeap && chart.LunarDay > 15 {
		birthLeap = 1
	}
	// 生时支(子=0)
	birthHourZhiIdx := zhiToZhiIdx(chart.ShiZhi)
	// 目标月 + 闰补
	dateLeap := 0
	if t.IsLeapMonth && t.LunarDay > 15 {
		dateLeap = 1
	}
	return ((yi-birthLeap-chart.BirthLunarMonth+birthHourZhiIdx+t.LunarMonth+dateLeap)%12 + 12) % 12
}

// computeDailyIndex returns iztro's daily index (display 寅=0).
func computeDailyIndex(chart Chart, t flowTarget) int {
	mi := computeMonthlyIndex(chart, t)
	return ((mi+t.LunarDay-1)%12 + 12) % 12
}

// computeHourlyIndex returns iztro's hourly index (display 寅=0).
func computeHourlyIndex(chart Chart, t flowTarget) int {
	di := computeDailyIndex(chart, t)
	return ((di+zhiToZhiIdx(t.ShiZhi))%12 + 12) % 12
}
