package ziwei

import (
	"liki-engine/internal/engine/tianwen"
)

// LiuYue holds monthly fate analysis.
type LiuYue struct {
	MingGong     palaceIndex `json:"ming_gong"`
	MingGongName string      `json:"ming_gong_name"`
	SiHua        siHuaResult `json:"si_hua"`
}

func ComputeLiuYue(chart Chart, liuYear, lunarMonth int) LiuYue {
	liuNianMing := liuNianMingGong(liuYear, chart.BirthYear)
	ming := liuYueMingGong(lunarMonth, liuNianMing)
	liuYearGan := yearGan(liuYear)
	return LiuYue{
		MingGong:     ming,
		MingGongName: PalaceNames[ming],
		SiHua:        liuYueSiHua(lunarMonth, liuYearGan),
	}
}

func liuYueMingGong(lunarMonth int, liuNianMing palaceIndex) palaceIndex {
	return (liuNianMing + palaceIndex(lunarMonth-1)) % 12
}

func liuYueSiHua(lunarMonth int, liuYearGan Gan) siHuaResult {
	yg := yinGan(liuYearGan)
	monthGan := Gan(((int(yg)-1+lunarMonth-1)%10+10)%10 + 1)
	return computeSiHua(monthGan)
}

// LiuRi holds daily fate analysis.
type LiuRi struct {
	MingGong     palaceIndex `json:"ming_gong"`
	MingGongName string      `json:"ming_gong_name"`
	SiHua        siHuaResult `json:"si_hua"`
}

func ComputeLiuRi(chart Chart, liuYear, lunarMonth, lunarDay int) LiuRi {
	liuNianMing := liuNianMingGong(liuYear, chart.BirthYear)
	liuYueMing := liuYueMingGong(lunarMonth, liuNianMing)
	ming := liuRiMingGong(lunarDay, liuYueMing)
	dayGan := riGan(liuYear, lunarMonth, lunarDay)
	return LiuRi{
		MingGong:     ming,
		MingGongName: PalaceNames[ming],
		SiHua:        liuRiSiHua(dayGan),
	}
}

func liuRiMingGong(lunarDay int, liuYueMing palaceIndex) palaceIndex {
	return (liuYueMing + palaceIndex(lunarDay-1)) % 12
}

func yearGan(year int) Gan { return Gan(((year-4)%10+10)%10 + 1) }

// riGan computes the day stem for a given lunar date within a Gregorian year.
// It converts lunar→solar first, then uses the day-pillar formula.
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

func liuRiSiHua(dayGan Gan) siHuaResult { return computeSiHua(dayGan) }
