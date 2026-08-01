package qimen

import (
	"math"
	"sort"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// TimeSlot represents a recommended time window.
type TimeSlot struct {
	Time   string `json:"time"`
	Date   string `json:"date"`
	ShiChen string `json:"shi_chen"`
	Score  int    `json:"score"`
	Rating string `json:"rating"`
	Advice string `json:"advice"`
}

// shiChenNames maps branch index to the Chinese 时辰 name.
var shiChenNames = [13]string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// ComputeTimeSelection finds the best time slots for an event within a date range.
func ComputeTimeSelection(start, end time.Time, event EventKind, count int, longitude, timezone float64) []TimeSlot {
	if count <= 0 {
		count = 3
	}

	var slots []TimeSlot

	// Iterate through each day in range.
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		// Iterate through 12 时辰 (branches 1-12 = 子丑寅卯辰巳午未申酉戌亥).
		for branch := 1; branch <= 12; branch++ {
			// Midpoint of the 时辰 in minutes: branch 1=子=60min(01:00), branch 12=亥=1380min(23:00)
			midMinutes := branch*120 - 60
			hours := midMinutes / 60
			mins := midMinutes % 60

			// Skip late night hours (23:00-05:00) for practical selection.
			if hours >= 23 || hours < 5 {
				continue
			}

			t := time.Date(d.Year(), d.Month(), d.Day(), hours, mins, 0, 0, d.Location())
			st := tianwen.GregorianToSolar(t, longitude, timezone)

			chart := ComputeChart(st, ShiQiMen)
			score := scoreChart(chart, event)

			rating := ratingFromScore(score)
			advice := adviceFromRating(rating)

			slots = append(slots, TimeSlot{
				Time:    t.Format("15:04"),
				Date:    t.Format("2006-01-02"),
				ShiChen: shiChenNames[branch],
				Score:   score,
				Rating:  rating,
				Advice:  advice,
			})
		}
	}

	// Sort by score descending.
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].Score > slots[j].Score
	})

	if len(slots) > count {
		slots = slots[:count]
	}

	return slots
}

// scoreChart computes a numeric score for a chart based on event type.
func scoreChart(c Chart, event EventKind) int {
	score := 0

	// 1. Count auspicious patterns.
	for _, p := range c.Patterns {
		if p.Auspicious {
			score += 3
		}
	}

	// 2. Count auspicious stem interactions.
	for _, si := range c.GanInteractions {
		if si.Auspicious {
			score += 2
		}
	}

	// 3. Count auspicious star interactions.
	for _, si := range c.XingInteractions {
		if si.Auspicious {
			score += 2
		}
	}

	// 4. Penalize 门迫 and 门制.
	score -= len(c.MenPo) * 2
	score -= len(c.MenZhi)

	// 5. 五不遇时: 时干克日干, 降级
	if c.Pan.WuBuYuShi {
		score -= 10
	}

	// 6. Check 空亡.
	for _, k := range c.Pan.KongWang {
		for _, p := range c.Pan.GongWei {
			if p.Star == c.Pan.DutyStar {
				// 值符宫若空亡, 减分
				for i, pp := range c.Pan.GongWei {
					if pp.Star == c.Pan.DutyStar && GongIndex(i+1) == k {
						score -= 2
					}
				}
			}
		}
	}

	// Clamp to reasonable range.
	return int(math.Max(float64(score), -10))
}

// ratingFromScore converts a numeric score to a rating.
func ratingFromScore(score int) string {
	switch {
	case score >= 8:
		return "大吉"
	case score >= 4:
		return "吉"
	case score >= 0:
		return "平"
	case score >= -3:
		return "凶"
	default:
		return "大凶"
	}
}

// adviceFromRating generates advice text.
func adviceFromRating(rating string) string {
	switch rating {
	case "大吉":
		return "诸事皆宜，吉星高照"
	case "吉":
		return "较为顺利，可择此时行动"
	case "平":
		return "平平无奇，宜守不宜攻"
	case "凶":
		return "多有阻碍，宜谨慎"
	default:
		return "凶险之象，不宜妄动"
	}
}
