// Package tianwen provides astronomical computation (天文) shared by
// all Chinese metaphysics packages.
//
// Public API:
//
// Types
//   Timeset       — 三套日历输出（公历/太阳时/农历）
//   SolarTime     — 真太阳时
//   GregorianTime — 公历时间
//   LunarTime     — 农历时间（含时辰地支）
//
// Core functions
//   ComputeTimeset(gt GregorianTime, lon) → Timeset
//   GregorianToSolar(t time.Time, lon, tz) → SolarTime
//   SolarToLunar(gt GregorianTime) → LunarTime
//   LunarToGregorian(lt LunarTime) → GregorianTime
//
// Zhu computations
//   ComputeBazi(st SolarTime) → ganzhi.Bazi
//   RiZhu(gt GregorianTime) → Zhu
//   NianZhu(gt GregorianTime) → Zhu
//   YueZhu(gt GregorianTime) → Zhu
//   ShiZhu(st SolarTime) → Zhu
//
// Solar terms
//   SolarTermTime(year, targetLon) → time.Time
//   SolarTermIndex(year, month, day) → int
//   JianYue(gt GregorianTime) → ganzhi.Zhi
//   JieQiLongitudes — [24]float64
package tianwen
