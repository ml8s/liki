package bazi

import _ "embed"

import (
	"encoding/json"
	"fmt"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

// ── Data-driven strength & congge rule lookup ────────────────

// strengthRule represents one strength classification rule.
type strengthRule struct {
	Rule      int      `json:"rule"`
	Root      string   `json:"root"`
	Season    []string `json:"season"`
	Strength  string   `json:"strength"`
	Upgrade   *string  `json:"upgrade,omitempty"`
	NeedYinBi int      `json:"need_yinbi,omitempty"`
	Note      string   `json:"note"`
}

// congGeRule represents one 从格 detection rule.
type congGeRule struct {
	Rule       int              `json:"rule"`
	Type       string           `json:"type"`
	Conditions congGeConditions `json:"conditions"`
	Yong       string           `json:"yong"`
	Xi         string           `json:"xi"`
	Ji         string           `json:"ji"`
	Source     string           `json:"source"`
}

type congGeConditions struct {
	Root       string         `json:"root"`
	Season     []string       `json:"season"`
	Stems      stemConditions `json:"stems"`
	RootDetail []string       `json:"root_detail,omitempty"`
}

type stemConditions struct {
	YinBiMin      *int `json:"yin_bi_min,omitempty"`
	GuanShaTouGan *int `json:"guan_sha_tou_gan,omitempty"`
	GuanShaMin    *int `json:"guan_sha_min,omitempty"`
	YinTouGan     *int `json:"yin_tou_gan,omitempty"` // 印透干数量 (excluding 比劫)
	CaiMin        *int `json:"cai_min,omitempty"`
	BiJieTouGan   *int `json:"bi_jie_tou_gan,omitempty"`
	ShiShangMin   *int `json:"shi_shang_min,omitempty"`
}

// ── Embed and load ──

//go:embed data/strength_rules.json
var strengthRulesJSON []byte

//go:embed data/congge_rules.json
var congGeRulesJSON []byte

var (
	strengthRules []strengthRule
	congGeRules   []congGeRule
)

func init() {
	if err := loadStrengthRules(); err != nil {
		log.Fatalf("bazi: load strength rules: %v", err)
	}
	if err := loadCongGeRules(); err != nil {
		log.Fatalf("bazi: load congge rules: %v", err)
	}
}

func loadStrengthRules() error {
	if err := json.Unmarshal(strengthRulesJSON, &strengthRules); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if len(strengthRules) == 0 {
		return fmt.Errorf("empty strength rules")
	}
	return nil
}

func loadCongGeRules() error {
	if err := json.Unmarshal(congGeRulesJSON, &congGeRules); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if len(congGeRules) == 0 {
		return fmt.Errorf("empty congge rules")
	}
	return nil
}

// ── Root type classification ──

// classifyRoot determines the root type for the day master.
func classifyRoot(riYuan ganzhi.Gan, monthBranch ganzhi.Zhi, cangGan [4]cangGanOut) string {
	dmElem := ganzhi.GanWuxing(riYuan)

	monthCG := cangGan[1]
	if ganzhi.GanWuxing(monthCG.Main) == dmElem {
		return "month_main"
	}
	if (monthCG.Mid != nil && ganzhi.GanWuxing(*monthCG.Mid) == dmElem) ||
		(monthCG.Minor != nil && ganzhi.GanWuxing(*monthCG.Minor) == dmElem) {
		return "month_mid"
	}

	for i, hs := range cangGan {
		if i == 1 {
			continue
		}
		if ganzhi.GanWuxing(hs.Main) == dmElem {
			return "branch_main"
		}
		if (hs.Mid != nil && ganzhi.GanWuxing(*hs.Mid) == dmElem) ||
			(hs.Minor != nil && ganzhi.GanWuxing(*hs.Minor) == dmElem) {
			return "branch_mid"
		}
	}
	return "none"
}

// classifySeason returns the season label (旺相休囚死) for the day master.
func classifySeason(riYuan ganzhi.Gan, monthBranch ganzhi.Zhi) string {
	dmElem := ganzhi.GanWuxing(riYuan)
	monthElem := ganzhi.ZhiWuxing(monthBranch)
	genElem := elementThatGenerates(dmElem)
	ctrlElem := elementThatControls(dmElem)

	switch {
	case monthElem == dmElem:
		return "旺"
	case monthElem == genElem:
		return "相"
	case monthElem == ctrlElem:
		return "死"
	case elementThatControls(monthElem) == dmElem:
		return "囚"
	default:
		return "休"
	}
}

// countYinBi returns印比 count (比劫同五行+印生我) from the 4 stems.
func countYinBi(c Chart) int {
	// Count 印(生我) and 比劫(同我) from年/月/时干 only (excluding日主自身).
	dmElem := ganzhi.GanWuxing(c.Ri.Gan)
	genElem := elementThatGenerates(dmElem)
	count := 0
	for _, p := range []struct{ gan ganzhi.Gan }{
		{c.Nian.Gan}, {c.Yue.Gan}, {c.Shi.Gan},
	} {
		e := ganzhi.GanWuxing(p.gan)
		if e == dmElem || e == genElem {
			count++
		}
	}
	return count
}

// ── Lookup methods ──

// lookupStrength returns the day master's strength by matching
// root type, season, and 印比 against the strength rules JSON.
func lookupStrength(rootType, season string, yinBiCount int) string {
	for _, rule := range strengthRules {
		if rule.Root != rootType {
			continue
		}
		seasonMatch := false
		for _, s := range rule.Season {
			if s == season {
				seasonMatch = true
				break
			}
		}
		if !seasonMatch {
			continue
		}
		if rule.Upgrade != nil && rule.NeedYinBi > 0 && yinBiCount >= rule.NeedYinBi {
			return *rule.Upgrade
		}
		return rule.Strength
	}
	return "身弱"
}

// lookupCongGe checks the day master against 从格 rules in priority order.
func lookupCongGe(c Chart) (string, string, string, string) {
	riYuan := c.Ri.Gan
	dmElem := ganzhi.GanWuxing(riYuan)
	rootType := classifyRoot(riYuan, c.Yue.Zhi, computeCangGan(c.ToBazi()))
	season := classifySeason(riYuan, c.Yue.Zhi)
	stems := countStems(c)

	for _, rule := range congGeRules {
		if !matchRoot(rule.Conditions.Root, rootType, rule.Conditions.RootDetail) {
			continue
		}
		if !matchSeasons(rule.Conditions.Season, season) {
			continue
		}
		if !matchStemConditions(rule.Conditions.Stems, stems) {
			continue
		}
		yong, xi, ji := resolveCongGeYongXiJi(rule, dmElem)
		return rule.Type, yong, xi, ji
	}
	return "", "", "", ""
}

// ── Helpers ──

type stemCounts struct {
	biBi     int // 比劫 (same as dmElem)
	yin      int // 印 (generates dmElem)
	guanSha  int // 官杀 (controls dmElem)
	cai      int // 财 (controlled by dmElem)
	shiShang int // 食伤 (drains dmElem)
}

func (sc stemCounts) yinBi() int { return sc.biBi + sc.yin }

func countStems(c Chart) stemCounts {
	dmElem := ganzhi.GanWuxing(c.Ri.Gan)
	var sc stemCounts
	for _, p := range []struct{ gan ganzhi.Gan }{
		{c.Nian.Gan}, {c.Yue.Gan}, {c.Shi.Gan},
	} {
		e := ganzhi.GanWuxing(p.gan)
		switch {
		case e == dmElem:
			sc.biBi++
		case e == elementThatGenerates(dmElem):
			sc.yin++
		case e == elementThatControls(dmElem):
			sc.guanSha++
		case e == elementControlledBy(dmElem):
			sc.cai++
		case e == elementThatDrains(dmElem):
			sc.shiShang++
		}
	}
	return sc
}

func matchRoot(ruleRoot, actualRoot string, rootDetail []string) bool {
	switch ruleRoot {
	case "any":
		return true
	case "none":
		return actualRoot == "none"
	case "mid":
		return actualRoot == "month_mid" || actualRoot == "branch_mid"
	default:
		return ruleRoot == actualRoot
	}
}

func matchSeasons(allowed []string, actual string) bool {
	for _, s := range allowed {
		if s == actual {
			return true
		}
	}
	return false
}

func matchStemConditions(cond stemConditions, sc stemCounts) bool {
	if cond.YinBiMin != nil && sc.yinBi() < *cond.YinBiMin {
		return false
	}
	if cond.GuanShaTouGan != nil && sc.guanSha > *cond.GuanShaTouGan {
		return false
	}
	if cond.GuanShaMin != nil && sc.guanSha < *cond.GuanShaMin {
		return false
	}
	if cond.YinTouGan != nil && sc.yin > *cond.YinTouGan {
		return false
	}
	if cond.CaiMin != nil && sc.cai < *cond.CaiMin {
		return false
	}
	if cond.BiJieTouGan != nil && sc.biBi > *cond.BiJieTouGan {
		return false
	}
	if cond.ShiShangMin != nil && sc.shiShang < *cond.ShiShangMin {
		return false
	}
	return true
}

func resolveCongGeYongXiJi(rule congGeRule, dmElem ganzhi.Wuxing) (yong, xi, ji string) {
	resolve := func(symbol string) string {
		switch symbol {
		case "dm_elem":
			return dmElem.String()
		case "ke_dm":
			return elementThatControls(dmElem).String()
		case "xi_from_dm":
			return elementThatDrains(dmElem).String()
		case "guansha":
			return elementThatControls(dmElem).String()
		case "cai":
			return elementControlledBy(dmElem).String()
		case "yin":
			return elementThatGenerates(dmElem).String()
		case "bijie":
			return dmElem.String()
		case "shishang":
			return elementThatDrains(dmElem).String()
		default:
			return ""
		}
	}
	return resolve(rule.Yong), resolve(rule.Xi), resolve(rule.Ji)
}
