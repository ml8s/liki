package ziwei

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func nianGan(year int) Gan { return Gan(((year-4)%10+10)%10 + 1) }

func hongLuanPos(zhi Zhi) int {
	// 红鸾：卯(年支1)起子(0)，顺数至生年支
	return (4 - int(zhi) + 12) % 12
}

// ── 流月 ──

type LiuYue struct {
	MingGong     gongIndex      `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"xing_yao,omitempty"`
	GongWei      [12]flowPalace `json:"gong_wei,omitempty"`
}

// ComputeLiuYue computes the flow month chart.
// 流月干支 = 目标月干支（五虎遁，与命主无关）——真相源
// 盘起点 = monthlyIndex（iztro 公式，命主相关）——只决定 gongLabels 旋转
// lunarMonth IS the lunar month number (1-12) — domain-native input.
// The agent converts via tianwen.SolarToLunar before calling this.
func ComputeLiuYue(chart Chart, liuYear, lunarMonth int) LiuYue {
	liuYearGan, _ := yearGanZhi(liuYear)
	// 流月干支 = 目标月干支（五虎遁：年干 + 农历月）——真相源，与命主无关
	yueZhi := zhiIdxToZhi((lunarMonth + 1) % 12) // 正月寅起：农历1月→zhiIdx 2(寅)
	monthGan := yueGanByWuHuDun(liuYearGan, yueZhi)
	stars := liuYueStars(monthGan, yueZhi)

	// 盘起点 = monthlyIndex（iztro 公式，命主相关）
	mi := computeMonthlyIndex(chart, flowTarget{Year: liuYear, LunarMonth: lunarMonth})
	starByAnXingIdx := make(map[int][]string)
	for k, z := range stars {
		anXingIdx := zhiIdxToAnXingIdx(zhiToZhiIdx(z))
		starByAnXingIdx[anXingIdx] = append(starByAnXingIdx[anXingIdx], k)
	}
	return LiuYue{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          yueZhi,
		SiHua:        computeSiHua(monthGan),
		Stars:        stars,
		GongWei:      buildFlowPalaces(mi, starByAnXingIdx),
	}
}

func liuYueStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiIdx int) Zhi { return Zhi(zhiIdx + 1) }
	m := map[string]Zhi{
		"月禄": toZhi(luCunPos(gan)),
		"月羊": toZhi(qingYangPos(gan)),
		"月陀": toZhi(tuoLuoPos(gan)),
		"月魁": toZhi(tianKuiPos(gan)),
		"月钺": toZhi(tianYuePos(gan)),
		"月马": toZhi(tianMaPos(zhi)),
		"月鸾": toZhi(hongLuanPos(zhi)),
		"月喜": toZhi((hongLuanPos(zhi) + 6) % 12),
		"月昌": toZhi(chg),
		"月曲": toZhi(qu),
	}
	return m
}

// ── 流日 ──

type LiuRi struct {
	MingGong     gongIndex      `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"xing_yao,omitempty"`
	GongWei      [12]flowPalace `json:"gong_wei,omitempty"`
}

// ComputeLiuRi computes the flow day chart using iztro's dailyIndex formula.
// lunarMonth/lunarDay ARE lunar (domain-native). Day pillar requires Gregorian:
// internally converts lunar→solar via search on SolarToLunar (the inverse).
func ComputeLiuRi(chart Chart, liuYear, lunarMonth, lunarDay int) LiuRi {
	// 盘起点 = dailyIndex（直接用农历）
	di := computeDailyIndex(chart, flowTarget{Year: liuYear, LunarMonth: lunarMonth, LunarDay: lunarDay})
	// 流日干支 = 目标日干支（需公历——农历转公历）
	gt := lunarToSolar(liuYear, lunarMonth, lunarDay)
	zhu := tianwen.RiZhu(gt)
	riGan, riZhi := Gan(zhu.Gan), Zhi(zhu.Zhi)
	stars := liuRiStars(riGan, riZhi)
	starByAnXingIdx := make(map[int][]string)
	for k, z := range stars {
		anXingIdx := zhiIdxToAnXingIdx(zhiToZhiIdx(z))
		starByAnXingIdx[anXingIdx] = append(starByAnXingIdx[anXingIdx], k)
	}
	return LiuRi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          riZhi,
		SiHua:        computeSiHua(riGan),
		Stars:        stars,
		GongWei:      buildFlowPalaces(di, starByAnXingIdx),
	}
}

func liuRiStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiIdx int) Zhi { return Zhi(zhiIdx + 1) }
	m := map[string]Zhi{
		"日禄": toZhi(luCunPos(gan)),
		"日羊": toZhi(qingYangPos(gan)),
		"日陀": toZhi(tuoLuoPos(gan)),
		"日魁": toZhi(tianKuiPos(gan)),
		"日钺": toZhi(tianYuePos(gan)),
		"日马": toZhi(tianMaPos(zhi)),
		"日鸾": toZhi(hongLuanPos(zhi)),
		"日喜": toZhi((hongLuanPos(zhi) + 6) % 12),
		"日昌": toZhi(chg),
		"日曲": toZhi(qu),
	}
	return m
}

// ── 流时 ──

type LiuShi struct {
	MingGong     gongIndex      `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"xing_yao,omitempty"`
	GongWei      [12]flowPalace `json:"gong_wei,omitempty"`
}

// ComputeLiuShi computes the flow hour chart using iztro's hourlyIndex formula.
// lunarMonth/lunarDay ARE lunar (domain-native); shiZhi is the target hour zhi.
func ComputeLiuShi(chart Chart, liuYear, lunarMonth, lunarDay int, shiZhi Zhi) LiuShi {
	// 盘起点 = hourlyIndex（直接用农历）
	hi := computeHourlyIndex(chart, flowTarget{Year: liuYear, LunarMonth: lunarMonth, LunarDay: lunarDay, ShiZhi: shiZhi})
	// 流时干支：日干 + 五鼠遁（需公历——农历转公历）
	gt := lunarToSolar(liuYear, lunarMonth, lunarDay)
	zhu := tianwen.RiZhu(gt)
	riGan := Gan(zhu.Gan)
	shiGan := shiGanCalc(riGan, shiZhi)
	stars := liuShiStars(shiGan, shiZhi)
	starByAnXingIdx := make(map[int][]string)
	for k, z := range stars {
		anXingIdx := zhiIdxToAnXingIdx(zhiToZhiIdx(z))
		starByAnXingIdx[anXingIdx] = append(starByAnXingIdx[anXingIdx], k)
	}
	return LiuShi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          shiZhi,
		SiHua:        computeSiHua(shiGan),
		Stars:        stars,
		GongWei:      buildFlowPalaces(hi, starByAnXingIdx),
	}
}

func liuShiStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiIdx int) Zhi { return Zhi(zhiIdx + 1) }
	m := map[string]Zhi{
		"时禄": toZhi(luCunPos(gan)),
		"时羊": toZhi(qingYangPos(gan)),
		"时陀": toZhi(tuoLuoPos(gan)),
		"时魁": toZhi(tianKuiPos(gan)),
		"时钺": toZhi(tianYuePos(gan)),
		"时马": toZhi(tianMaPos(zhi)),
		"时鸾": toZhi(hongLuanPos(zhi)),
		"时喜": toZhi((hongLuanPos(zhi) + 6) % 12),
		"时昌": toZhi(chg),
		"时曲": toZhi(qu),
	}
	return m
}

func shiGanCalc(riGan Gan, shiZhi Zhi) Gan {
	// 五鼠遁：甲己→甲子, 乙庚→丙子, 丙辛→戊子, 丁壬→庚子, 戊癸→壬子
	ziGan := (int(riGan)-1)*2%10 + 1
	return Gan(((ziGan-1+int(shiZhi)-1)%10+10)%10 + 1)
}

// ── riGan — calculates the day gan for a lunar date ──

func riGan(liuYear, lunarMonth, lunarDay int) Gan {
	// Try liuYear as the lunar year; fall back to liuYear-1 for months
	// before Chinese New Year (when the lunar year hasn't caught up).
	gt := tianwen.LunarToGregorian(tianwen.LunarTime{Year: liuYear, Month: lunarMonth, Day: lunarDay})
	if gt.Time().IsZero() {
		gt = tianwen.LunarToGregorian(tianwen.LunarTime{Year: liuYear - 1, Month: lunarMonth, Day: lunarDay})
	}
	if gt.Time().IsZero() {
		return 1 // fallback
	}
	dp := tianwen.RiZhu(gt)
	return dp.Gan
}

// 流昌流曲：天干定位（iztro算法）, 返回zhiIdx
func liuChangQuByGan(gan Gan) (changZhiIdx, quZhiIdx int) {
	table := [10][2]int{
		{3, 7}, {4, 6}, {6, 4}, {7, 3}, {6, 4},
		{7, 3}, {9, 1}, {10, 0}, {0, 10}, {1, 9},
	}
	idx := int(gan) - 1
	return (table[idx][0] + 2) % 12, (table[idx][1] + 2) % 12
}

func riZhi(liuYear, lunarMonth, lunarDay int) Zhi {
	gt := tianwen.LunarToGregorian(tianwen.LunarTime{Year: liuYear, Month: lunarMonth, Day: lunarDay})
	if gt.Time().IsZero() {
		gt = tianwen.LunarToGregorian(tianwen.LunarTime{Year: liuYear - 1, Month: lunarMonth, Day: lunarDay})
	}
	if gt.Time().IsZero() {
		return 1
	}
	dp := tianwen.RiZhu(gt)
	return dp.Zhi
}

// yueGanByWuHuDun computes the month gan via 五虎遁 (year gan + month zhi).
func yueGanByWuHuDun(nianGan Gan, yueZhi Zhi) Gan {
	// 寅月干: 甲己→丙, 乙庚→戊, 丙辛→庚, 丁壬→壬, 戊癸→甲
	base := [10]Gan{3, 5, 7, 9, 1, 3, 5, 7, 9, 1} // 寅月干（甲=1..癸=10）
	offset := int(yueZhi) - 3                     // 寅=0 偏移
	if offset < 0 {
		offset += 12
	}
	return Gan(((int(base[nianGan-1]) - 1 + offset) % 10) + 1)
}

// lunarToSolar finds the Gregorian date for a lunar date by searching
// SolarToLunar as an oracle (the inverse function). Lunar months are ~29-30
// days; searching ±35 days from a rough estimate is sufficient.
func lunarToSolar(lunarYear, lunarMonth, lunarDay int) tianwen.GregorianTime {
	cst := time.FixedZone("CST", 8*3600)
	// 粗估公历月 ≈ 农历月 + 1（农历四月 ≈ 公历五月）
	approx := time.Date(lunarYear, time.Month(lunarMonth+1), lunarDay, 0, 0, 0, 0, cst)

	// SolarToLunar 关于日期单调递增 → 二分搜索（O(log 70) ≈ 7 次调用）
	lo, hi := -35, 35
	match := func(offset int) int {
		lt := tianwen.SolarToLunar(tianwen.GregorianTime(approx.AddDate(0, 0, offset)))
		got := lt.Year*10000 + lt.Month*100 + lt.Day
		want := lunarYear*10000 + lunarMonth*100 + lunarDay
		if got < want {
			return -1
		}
		if got > want {
			return 1
		}
		return 0
	}
	for lo <= hi {
		mid := (lo + hi) / 2
		c := match(mid)
		if c == 0 {
			return tianwen.GregorianTime(approx.AddDate(0, 0, mid))
		}
		if c < 0 {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	// 未找到（不应该发生）——返回粗估值，日柱可能偏差
	return tianwen.GregorianTime(approx)
}
