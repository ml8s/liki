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
	Stars        map[string]Zhi `json:"stars,omitempty"`
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