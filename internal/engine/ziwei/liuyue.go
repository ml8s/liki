package ziwei

import (
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
	Stars        map[string]int `json:"stars,omitempty"`
}

func ComputeLiuYue(chart Chart, liuYear, lunarMonth int) LiuYue {
	liuYearGan := yearGan(liuYear)
	monthGan := Gan(((int(yinGan(liuYearGan)) - 1 + lunarMonth - 1) % 10 + 10) % 10 + 1)
	monthZhi := Zhi((lunarMonth + 1) % 12 + 1) // 正月寅起


	return LiuYue{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          monthZhi,
		SiHua:        liuYueSiHua(lunarMonth, liuYearGan),
		Stars:        liuYueStars(monthGan, monthZhi),
	}
}

func liuYueMingGong(flowYear int, chart Chart, targetMonth int) palaceIndex {
	_, flowZhi := yearStemBranch(flowYear)
	// iztro: yearlyIndex - birthMonth + hourEBI + targetMonth
	// where EBI is EARTHLY_BRANCHES index (子=0,丑=1,...,亥=11)
	yearlyIdx := (int(flowZhi) - 3 + 12) % 12   // = fixEarthlyBranchIndex(年支)
	hourIdx := int(chart.HourZhi) - 1            // = EARTHLY_BRANCHES.indexOf(出生时辰)
	birthM := int(chart.LunarMonth)
	iztroIdx := (yearlyIdx + hourIdx + targetMonth - birthM + 12) % 12
	// iztroIdx is iztro display index (寅=0). Convert to Liki palace.
	zhiM1 := (iztroIdx + 2) % 12
	return zhiToPalace(zhiM1, chart.Palaces[0].Zhi)
}

func liuYueSiHua(lunarMonth int, liuYearGan Gan) siHuaResult {
	yg := yinGan(liuYearGan)
	monthGan := Gan(((int(yg)-1+lunarMonth-1)%10+10)%10 + 1)
	return computeSiHua(monthGan)
}

func liuYueStars(gan Gan, zhi Zhi) map[string]int {
	chg, qu := liuChangQuByGan(gan)
	m := map[string]int{
		"月禄": luCunPos(gan),
		"月羊": qingYangPos(gan),
		"月陀": tuoLuoPos(gan),
		"月魁": tianKuiPos(gan),
		"月钺": tianYuePos(gan),
		"月马": tianMaPos(zhi),
		"月鸾": hongLuanPos(zhi),
		"月喜": (hongLuanPos(zhi) + 6) % 12,
		"月昌": chg,
		"月曲": qu,
	}
	return m
}

// ── 流日 ──

type LiuRi struct {
	MingGong     palaceIndex    `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]int `json:"stars,omitempty"`
}

func ComputeLiuRi(chart Chart, liuYear, lunarMonth, lunarDay int) LiuRi {
	dayGan := riGan(liuYear, lunarMonth, lunarDay)
	dayZhi := riZhi(liuYear, lunarMonth, lunarDay)

	return LiuRi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          dayZhi,
		SiHua:        liuRiSiHua(dayGan),
		Stars:        liuRiStars(dayGan, dayZhi),
	}
}

func liuRiStars(gan Gan, zhi Zhi) map[string]int {
	chg, qu := liuChangQuByGan(gan)
	m := map[string]int{
		"日禄": luCunPos(gan),
		"日羊": qingYangPos(gan),
		"日陀": tuoLuoPos(gan),
		"日魁": tianKuiPos(gan),
		"日钺": tianYuePos(gan),
		"日马": tianMaPos(zhi),
		"日鸾": hongLuanPos(zhi),
		"日喜": (hongLuanPos(zhi) + 6) % 12,
		"日昌": chg,
		"日曲": qu,
	}
	return m
}

func liuRiSiHua(dayGan Gan) siHuaResult {
	return computeSiHua(dayGan)
}

func liuRiMingGong(lunarDay int, liuYueMing palaceIndex) palaceIndex {
	return (liuYueMing + palaceIndex(lunarDay-1)) % 12
}

// ── 流时 ──

type LiuShi struct {
	MingGong     palaceIndex    `json:"ming_gong"`
	MingGongName string         `json:"ming_gong_name"`
	Zhi          Zhi            `json:"zhi"`
	SiHua        siHuaResult    `json:"si_hua"`
	Stars        map[string]int `json:"stars,omitempty"`
}

func ComputeLiuShi(chart Chart, liuYear, lunarMonth, lunarDay int, shiZhi Zhi) LiuShi {
	dayGan := riGan(liuYear, lunarMonth, lunarDay)
	shiGan := shiGanCalc(dayGan, shiZhi)

	return LiuShi{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          shiZhi,
		SiHua:        liuShiSiHua(shiGan),
		Stars:        liuShiStars(shiGan, shiZhi),
	}
}

func liuShiStars(gan Gan, zhi Zhi) map[string]int {
	chg, qu := liuChangQuByGan(gan)
	m := map[string]int{
		"时禄": luCunPos(gan),
		"时羊": qingYangPos(gan),
		"时陀": tuoLuoPos(gan),
		"时魁": tianKuiPos(gan),
		"时钺": tianYuePos(gan),
		"时马": tianMaPos(zhi),
		"时鸾": hongLuanPos(zhi),
		"时喜": (hongLuanPos(zhi) + 6) % 12,
		"时昌": chg,
		"时曲": qu,
	}
	return m
}

func liuShiSiHua(shiGan Gan) siHuaResult {
	return computeSiHua(shiGan)
}

func shiGanCalc(riGan Gan, shiZhi Zhi) Gan {
	ziGan := (int(riGan)-1)*2%10 + 1
	if int(riGan) == 6 || int(riGan) == 7 || int(riGan) == 8 {
		ziGan--
	}
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