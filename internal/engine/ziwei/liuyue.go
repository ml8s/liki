package ziwei

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func yearGan(year int) Gan { return Gan(((year-4)%10+10)%10 + 1) }

func hongLuanPos(zhi Zhi) int {
	// 红鸾：卯(年支1)起子(0)，顺数至生年支
	return (4 - int(zhi) + 12) % 12
}

// ── 流月 ──

type LiuYue struct {
	MingGong     palaceIndex    `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"stars,omitempty"`
	Palaces      [12]flowPalace `json:"palaces,omitempty"`
}

// ComputeLiuYue computes the flow month chart.
// 流月干支 = 目标日期月干支（五虎遁，与命主无关）——真相源
// 盘起点 = monthlyIndex（iztro 公式，命主相关）——只决定 palaceNames 旋转
// liuYear/lunarMonth are GREGORIAN year/month (target date).
func ComputeLiuYue(chart Chart, liuYear, lunarMonth int) LiuYue {
	// 目标农历月（公历 → 农历）
	tgt := tianwen.SolarToLunar(tianwen.GregorianTime(time.Date(liuYear, time.Month(lunarMonth), 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))))
	liuYearGan, _ := yearStemBranch(liuYear)
	// 流月干支 = 目标月干支（五虎遁：年干 + 农历月）——真相源，与命主无关
	monthZhi := zhiMinus1ToZhi((tgt.Month + 1) % 12) // 正月寅起：农历1月→zhiMinus1 2(寅)
	monthGan := yueGanByWuHuDun(liuYearGan, monthZhi)
	stars := liuYueStars(monthGan, monthZhi)

	// 盘起点 = monthlyIndex（iztro 公式，命主相关）
	mi := computeMonthlyIndex(chart, flowTarget{Year: liuYear, LunarMonth: tgt.Month})
	starByDisplay := make(map[int][]string)
	for k, z := range stars {
		disp := zhiMinus1ToDisplay(zhiToZhiMinus1(z))
		starByDisplay[disp] = append(starByDisplay[disp], k)
	}
	return LiuYue{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          monthZhi,
		SiHua:        computeSiHua(monthGan),
		Stars:        stars,
		Palaces:      buildFlowPalaces(mi, starByDisplay),
	}
}

func liuYueSiHua(lunarMonth int, liuYearGan Gan) siHuaResult {
	yg := yinGan(liuYearGan)
	monthGan := Gan(((int(yg)-1+lunarMonth-1)%10+10)%10 + 1)
	return computeSiHua(monthGan)
}

func liuYueStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiMinus1 int) Zhi { return Zhi(zhiMinus1 + 1) }
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
	MingGong     palaceIndex    `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"stars,omitempty"`
	Palaces      [12]flowPalace `json:"palaces,omitempty"`
}

// ComputeLiuRi computes the flow day chart using iztro's dailyIndex formula.
// liuYear/lunarMonth/lunarDay are GREGORIAN (target date).
func ComputeLiuRi(chart Chart, liuYear, lunarMonth, lunarDay int) LiuRi {
	tgt := tianwen.SolarToLunar(tianwen.GregorianTime(time.Date(liuYear, time.Month(lunarMonth), lunarDay, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))))
	// 盘起点 = dailyIndex
	di := computeDailyIndex(chart, flowTarget{Year: liuYear, LunarMonth: tgt.Month, LunarDay: tgt.Day})
	// 流日干支 = 目标公历日干支
	zhu := tianwen.RiZhu(tianwen.GregorianTime(time.Date(liuYear, time.Month(lunarMonth), lunarDay, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))))
	dayGan, dayZhi := Gan(zhu.Gan), Zhi(zhu.Zhi)
	stars := liuRiStars(dayGan, dayZhi)
	starByDisplay := make(map[int][]string)
	for k, z := range stars {
		disp := zhiMinus1ToDisplay(zhiToZhiMinus1(z))
		starByDisplay[disp] = append(starByDisplay[disp], k)
	}
	return LiuRi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          dayZhi,
		SiHua:        computeSiHua(dayGan),
		Stars:        stars,
		Palaces:      buildFlowPalaces(di, starByDisplay),
	}
}

func liuRiStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiMinus1 int) Zhi { return Zhi(zhiMinus1 + 1) }
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

func liuRiSiHua(dayGan Gan) siHuaResult {
	return computeSiHua(dayGan)
}

// ── 流时 ──

type LiuShi struct {
	MingGong     palaceIndex    `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]Zhi `json:"stars,omitempty"`
	Palaces      [12]flowPalace `json:"palaces,omitempty"`
}

// ComputeLiuShi computes the flow hour chart using iztro's hourlyIndex formula.
// liuYear/lunarMonth/lunarDay are GREGORIAN; shiZhi is the target hour branch.
func ComputeLiuShi(chart Chart, liuYear, lunarMonth, lunarDay int, shiZhi Zhi) LiuShi {
	tgt := tianwen.SolarToLunar(tianwen.GregorianTime(time.Date(liuYear, time.Month(lunarMonth), lunarDay, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))))
	// 盘起点 = hourlyIndex
	hi := computeHourlyIndex(chart, flowTarget{Year: liuYear, LunarMonth: tgt.Month, LunarDay: tgt.Day, HourZhi: shiZhi})
	// 流时干支：日干 + 五鼠遁
	zhu := tianwen.RiZhu(tianwen.GregorianTime(time.Date(liuYear, time.Month(lunarMonth), lunarDay, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))))
	dayGan := Gan(zhu.Gan)
	shiGan := shiGanCalc(dayGan, shiZhi)
	stars := liuShiStars(shiGan, shiZhi)
	starByDisplay := make(map[int][]string)
	for k, z := range stars {
		disp := zhiMinus1ToDisplay(zhiToZhiMinus1(z))
		starByDisplay[disp] = append(starByDisplay[disp], k)
	}
	return LiuShi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          shiZhi,
		SiHua:        computeSiHua(shiGan),
		Stars:        stars,
		Palaces:      buildFlowPalaces(hi, starByDisplay),
	}
}

func liuShiStars(gan Gan, zhi Zhi) map[string]Zhi {
	chg, qu := liuChangQuByGan(gan)
	toZhi := func(zhiMinus1 int) Zhi { return Zhi(zhiMinus1 + 1) }
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

func liuShiSiHua(shiGan Gan) siHuaResult {
	return computeSiHua(shiGan)
}

func shiGanCalc(riGan Gan, shiZhi Zhi) Gan {
	// 五鼠遁：甲己→甲子, 乙庚→丙子, 丙辛→戊子, 丁壬→庚子, 戊癸→壬子
	ziGan := (int(riGan)-1)*2%10 + 1
	return Gan(((ziGan - 1 + int(shiZhi) - 1) % 10 + 10) % 10 + 1)
}

// ── riGan — calculates the day stem for a lunar date ──

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

// 流昌流曲：天干定位（iztro算法）, 返回zhiMinus1
func liuChangQuByGan(gan Gan) (changZM1, quZM1 int) {
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
	if gt.Time().IsZero() { return 1 }
	dp := tianwen.RiZhu(gt)
	return dp.Zhi
}
// yueGanByWuHuDun computes the month stem via 五虎遁 (year stem + month branch).
func yueGanByWuHuDun(yearGan Gan, monthZhi Zhi) Gan {
	// 寅月干: 甲己→丙, 乙庚→戊, 丙辛→庚, 丁壬→壬, 戊癸→甲
	base := [10]Gan{3, 5, 7, 9, 1, 3, 5, 7, 9, 1} // 寅月干（甲=1..癸=10）
	offset := int(monthZhi) - 3 // 寅=0 偏移
	if offset < 0 { offset += 12 }
	return Gan(((int(base[yearGan-1]) - 1 + offset) % 10) + 1)
}
