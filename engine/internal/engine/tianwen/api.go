// Package tianwen provides astronomical computation (天文) shared by
// all Chinese metaphysics packages.
//
// Public API:
//
// # Types
//
// //	time.Time     — 真太阳时
//
//	time.Time — IANA 时区下的法定日期时间（非真太阳时）
//	LunarDate     — 农历日期（不含时辰）
//
// Core functions
//
//	ComputeTimeset(gt time.Time, lon) → Timeset
//	ApparentSolarFromCivil(t time.Time, lon, tz) → time.Time
//	LunarDateAt(gt time.Time) → LunarDate
//	CivilTimeFromLunarDate(ld LunarDate) → time.Time
//
// Zhu computations
//
//	ComputeBazi(gt time.Time, st time.Time) → ganzhi.Bazi
//	RiZhu(gt time.Time) → Zhu
//	NianZhu(gt time.Time) → Zhu
//	YueZhu(gt time.Time) → Zhu
//	ShiZhu(st time.Time) → Zhu
//
// Solar terms
//
//	SolarTermTime(year, targetLon) → time.Time
//	SolarTermIndex(year, month, day) → int
//	JianYue(gt time.Time) → ganzhi.Zhi
//	JieQiLongitudes — [24]float64
package tianwen
