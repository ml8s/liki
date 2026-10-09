package huangli

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// --- Types ---

// riZhuInfo holds the gan-zhi for a single day.
type riZhuInfo struct {
	Gan   ganzhi.Gan `json:"gan"`
	Zhi   ganzhi.Zhi `json:"zhi"`
	NaYin string     `json:"na_yin"`
}

// --- In-memory config ---

// eventRule maps a life event to its suitable and forbidden jianchu gods.
type eventRule struct {
	Label     string
	Suitable  []string
	Forbidden []string
}

// JianChuConfig holds the parsed jianchu (建除) calendar rules: sequences,
// suitable/forbidden activities, event-type rules, and shensha (神煞) data.
type jianchuConfig struct {
	Sequence   []string
	Suitable   map[string][]string
	Forbidden  map[string][]string
	EventRules map[string]eventRule
	ShenSha    map[string]map[string][]string
}

// --- Engine Functions ---

// lookupRiZhu returns the gan-zhi and na-yin for a given date.
func lookupRiZhu(t time.Time) riZhuInfo {
	p := tianwen.RiZhu(tianwen.GregorianTime(t))
	return riZhuInfo{Gan: p.Gan, Zhi: p.Zhi, NaYin: ganzhi.NayinLabel(p.Gan, p.Zhi)}
}

// lookupJianChu returns the JianChu (建除) god for a given date.
func lookupJianChu(t time.Time) string {
	mp := yueZhuForDate(t)
	yueZhi := mp.Zhi

	dp := tianwen.RiZhu(tianwen.GregorianTime(t))

	jianIdx := int(yueZhi) - 1
	dayIdx := int(dp.Zhi) - 1

	offset := (dayIdx - jianIdx + 12) % 12
	return jianChuCfg.Sequence[offset]
}

// yueZhuForDate returns the month pillar for a given date.
func yueZhuForDate(t time.Time) ganzhi.Zhu {
	return tianwen.YueZhu(tianwen.GregorianTime(t))
}
