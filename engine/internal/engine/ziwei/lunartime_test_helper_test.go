package ziwei

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	"time"
)

// lunarTimeFromSolar builds the transitional LunarTime (lunar date + solar
// shichen) from a true-solar moment. Removed when ComputeChart takes Gregorian.
func lunarTimeFromSolar(st time.Time) tianwen.LunarTime {
	date := tianwen.LunarDateAt(st)
	minutes := float64(st.Hour()*60 + st.Minute())
	return tianwen.LunarTime{
		Year: date.Year, Month: date.Month, Day: date.Day, Leap: date.Leap,
		Shichen: ganzhi.Zhi((int(minutes+60)/120)%12 + 1),
	}
}
