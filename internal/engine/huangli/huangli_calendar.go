package huangli

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// --- Types ---

// riZhuInfo holds the stem-branch for a single day.
type riZhuInfo struct {
	Gan   ganzhi.Gan    `json:"gan"`
	Zhi   ganzhi.Zhi    `json:"zhi"`
	NaYin string `json:"na_yin"`
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

// taiSui returns the year's presiding branch (太岁).
func taiSui(year int) ganzhi.Zhi {
	// Year pillar branch = (year - 3) % 12, with 1=子 through 12=亥.
	// The formula already produces 1-based results after the <=0 guard;
	// do NOT add +1 or the result shifts by one branch.
	b := (year - 3) % 12
	if b <= 0 {
		b += 12
	}
	return ganzhi.Zhi(b)
}

// lookupRiZhu returns the stem-branch and na-yin for a given date.
func lookupRiZhu(t time.Time) riZhuInfo {
	p := tianwen.RiZhu(tianwen.GregorianTime(t))
	return riZhuInfo{Gan: p.Gan, Zhi: p.Zhi, NaYin: ganzhi.NayinLabel(p.Gan, p.Zhi)}
}

// lookupJianChu returns the JianChu (建除) god for a given date.
func lookupJianChu(t time.Time) string {
	mp := yueZhuForDate(t)
	monthBranch := mp.Zhi

	dp := tianwen.RiZhu(tianwen.GregorianTime(t))

	jianIdx := int(monthBranch) - 1
	dayIdx := int(dp.Zhi) - 1

	offset := (dayIdx - jianIdx + 12) % 12
	return jianChuCfg.Sequence[offset]
}

// jianChuSuitable checks if the jianchu god is suitable for the event type.
func jianChuSuitable(jianChu, eventType string) (suitable bool, marks []string, warnings []string) {
	rule, ok := jianChuCfg.EventRules[eventType]
	if !ok {
		return true, nil, nil
	}

	for _, s := range rule.Suitable {
		if jianChu == s {
			suitable = true
			marks = append(marks, jianChu+"日宜"+rule.Label)
			break
		}
	}
	for _, f := range rule.Forbidden {
		if jianChu == f {
			suitable = false
			if f == "破" {
				warnings = append(warnings, "破日，万事不宜")
			} else {
				warnings = append(warnings, jianChu+"日忌"+rule.Label)
			}
			break
		}
	}
	return suitable, marks, warnings
}
// evaluateZhi checks the branch relationship and returns marks/warnings.
func evaluateZhi(dayZhi, refZhi ganzhi.Zhi, label string) (relation string, marks []string, warnings []string) {
	switch {
	case ganzhi.IsZhiHe(dayZhi, refZhi):
		return "六合", []string{label + "六合日"}, nil
	case ganzhi.IsTripleHe(dayZhi, refZhi):
		return "三合半", []string{label + "三合"}, nil
	case ganzhi.IsTripleHui(dayZhi, refZhi):
		return "三会半", []string{label + "三会"}, nil
	case ganzhi.IsLiuChong(dayZhi, refZhi):
		return "六冲", nil, []string{"冲" + label}
	case ganzhi.IsXing(dayZhi, refZhi):
		return "相刑", nil, []string{"刑" + label}
	case ganzhi.IsHai(dayZhi, refZhi):
		return "六害", nil, []string{"害" + label}
	}
	return "无", nil, nil
}

// yueZhuForDate returns the month pillar for a given date.
func yueZhuForDate(t time.Time) ganzhi.Zhu {
	return tianwen.YueZhu(tianwen.GregorianTime(t))
}
