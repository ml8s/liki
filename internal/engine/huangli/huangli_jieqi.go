package huangli

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// jieQiDepth describes the current solar term position for a given date.
type jieQiDepth struct {
	TermName     string `json:"term_name"`      // current solar term name
	DaysIn       int    `json:"days_in"`        // days into current term
	NextTermName string `json:"next_term_name"` // next solar term name
	DaysToNext   int    `json:"days_to_next"`   // days until next term
}

// Solar term names in order (24 terms, starting from 立春).
var jieQiNames = [24]string{
	"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
	"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
	"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
	"立冬", "小雪", "大雪", "冬至", "小寒", "大寒",
}

// jieQiOrdinal maps the output order of allSolarTerms (0=冬至, 1=小寒, …) to
// the canonical jieqi name index used by jieQiNames (0=立春, 1=雨水, …).
var jieQiOrdinal = func() [24]int {
	var o [24]int
	// allSolarTerms: [冬至(21), 小寒(22), 大寒(23), 立春(0), 雨水(1), …, 小雪(19), 大雪(20)]
	o[0] = 21  // 冬至
	o[1] = 22  // 小寒
	o[2] = 23  // 大寒
	for i := 3; i < 24; i++ {
		o[i] = i - 3 // 立春=0, 雨水=1, …, 大雪=20
	}
	return o
}()

// computeJieQiDepth returns the solar term that the given date falls in.
func computeJieQiDepth(year, month, day int) jieQiDepth {
	date := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
	terms := tianwen.AllSolarTerms(year)

	// Find the term interval containing date. Terms may have year mismatches;
	// normalize forward so terms are non-decreasing.
	for i := 0; i < 24; i++ {
		cur := terms[i]
		next := terms[(i+1)%24]
		if next.Before(cur) {
			next = next.AddDate(1, 0, 0)
		}
		if !date.Before(cur) && date.Before(next) {
			daysIn := int(date.Sub(cur).Hours() / 24)
			daysToNext := int(next.Sub(date).Hours() / 24)
			if daysIn < 0 {
				daysIn = 0
			}
			if daysToNext < 0 {
				daysToNext = 0
			}
			prevIdx := jieQiOrdinal[i]
			nextIdx := jieQiOrdinal[(i+1)%24]
			return jieQiDepth{
				TermName:     jieQiNames[prevIdx],
				DaysIn:       daysIn,
				NextTermName: jieQiNames[nextIdx],
				DaysToNext:   daysToNext,
			}
		}
	}

	// Should never reach here with valid input.
	return jieQiDepth{TermName: "未知"}
}

// renYuanSiLing describes which hidden stem governs during each portion of a month.
type renYuanSiLing struct {
	MonthBranch ganzhi.Zhi                  `json:"month_branch"`
	Phases      []ganzhi.RenYuanSiLingFenYe `json:"phases"`
	Current     *ganzhi.RenYuanSiLingFenYe  `json:"current"` // current governing stem, if date provided
}

// computeRenYuanSiLing returns the 人元司令分野 for the given solar month branch.
func computeRenYuanSiLing(solarMonthBranch ganzhi.Zhi) renYuanSiLing {
	phases := ganzhi.RenYuanSiLingFenYeForZhi(solarMonthBranch)
	if phases == nil {
		phases = []ganzhi.RenYuanSiLingFenYe{}
	}
	return renYuanSiLing{
		MonthBranch: solarMonthBranch,
		Phases:      phases,
	}
}

// computeRenYuanSiLingForDate returns the current governing phase for the given
// solar month branch. daysIn is the number of days into the current solar term,
// obtained from computeJieQiDepth.
func computeRenYuanSiLingForDate(monthBranch ganzhi.Zhi, daysIn int) renYuanSiLing {
	result := computeRenYuanSiLing(monthBranch)
	result.Current = findCurrentRenYuanSiLingFenYe(result.Phases, daysIn)
	return result
}

// findCurrentRenYuanSiLingFenYe finds the current governing phase given the phase table
// and the number of days into the solar month. Returns nil if phases is empty.
func findCurrentRenYuanSiLingFenYe(phases []ganzhi.RenYuanSiLingFenYe, daysIn int) *ganzhi.RenYuanSiLingFenYe {
	cumulative := 0
	for _, p := range phases {
		cumulative += p.Days
		if daysIn < cumulative {
			cp := p
			return &cp
		}
	}
	if len(phases) > 0 {
		last := phases[len(phases)-1]
		return &last
	}
	return nil
}
