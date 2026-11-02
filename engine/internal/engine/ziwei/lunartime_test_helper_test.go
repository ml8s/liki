package ziwei

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	"time"
)

// lunarTimeFromSolar builds the LunarTime used by lunar-native test charts from
// a true-solar moment.
func lunarTimeFromSolar(st time.Time) tianwen.LunarTime {
	date := tianwen.ShanghaiLunarReference().SolarDateToLunar(tianwen.BeijingDateAt(st.UTC()))
	minutes := float64(st.Hour()*60 + st.Minute())
	return tianwen.LunarTime{
		Year: date.Year, Month: date.Month, Day: date.Day, Leap: date.Leap,
		Shichen: ganzhi.Zhi((int(minutes+60)/120)%12 + 1),
	}
}
