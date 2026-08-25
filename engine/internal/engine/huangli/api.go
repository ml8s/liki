// Package huangli provides黄历 computation.
//
// Types
//   Day, Month, BondDay, BondMonth
//
// Functions
//   QueryDate(date string, event string) → (Day, error)
//   QueryMonth(yearMonth string, event string) → (Month, error)
//   ComputeBondDay(st SolarTime, event string, date string) → (BondDay, error)
//   ComputeBondMonth(st SolarTime, event string, yearMonth string) → (BondMonth, error)
package huangli

import "liki-engine/internal/engine/tianwen"

// ComputeBondDay evaluates a single day's compatibility with a given bazi.
func ComputeBondDay(st tianwen.SolarTime, eventType string, dateStr string) (BondDay, error) {
	return computeBondDay(tianwen.ComputeBazi(st), eventType, dateStr)
}

// ComputeBondMonth evaluates all days in a month for compatibility with a given bazi.
func ComputeBondMonth(st tianwen.SolarTime, eventType string, yearMonth string) (BondMonth, error) {
	return computeBondMonth(tianwen.ComputeBazi(st), eventType, yearMonth)
}
