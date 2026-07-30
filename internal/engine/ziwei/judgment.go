package ziwei

import (
	"encoding/json"
	_ "embed"
	"log"
)

// ── Judgment types ──

type SiHuaItem struct {
	StarID   int    `json:"star_id"`
	StarName string `json:"star_name"`
	Type     string `json:"type"`
}

type SanFangInfo struct {
	Name    string   `json:"name"`
	ZhuXing []string `json:"zhu_xing"`
	FuXing  []string `json:"fu_xing"`
	SiHua   string   `json:"si_hua"`
}

type JudgmentResult struct {
	Patterns []pattern     `json:"patterns"`
	SiHua    []SiHuaItem   `json:"si_hua"`
	SanFang  []SanFangInfo `json:"san_fang"`
	Rating   string        `json:"rating"`
	Rule     int           `json:"rule,omitempty"`
	Summary  string        `json:"summary"`
}

// ── Rule chain ──

//go:embed data/judgment_rules.json
var judgmentRulesJSON []byte

type judgmentRule struct {
	Rule   int           `json:"rule"`
	Rating string         `json:"rating"`
	Conds  judgmentConds  `json:"conds"`
	Source string         `json:"source"`
}

type judgmentConds struct {
	TopScore    *int        `json:"top_score,omitempty"`
	TopCount    *countRange `json:"top_count,omitempty"`
	LowScore    *int        `json:"low_score,omitempty"`
	SiHuaCount  *countRange `json:"si_hua_count,omitempty"`
	Brightness  []string    `json:"brightness,omitempty"`
	ShaXing     []string    `json:"sha_xing,omitempty"`
	MingGong    string      `json:"ming_gong,omitempty"`
}

type countRange struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

var judgmentRules []judgmentRule

func init() {
	if err := json.Unmarshal(judgmentRulesJSON, &judgmentRules); err != nil {
		log.Fatalf("ziwei: load judgment rules: %v", err)
	}
}

// ── 三方四正宫名 ──

var palaceNames = [12]string{
	"命宫", "兄弟", "夫妻", "子女",
	"财帛", "疾厄", "迁移", "交友",
	"官禄", "田宅", "福德", "父母",
}

// ── 煞星 ──

var shaStars = []starIndex{HuoXing, LingXing, QingYang, TuoLuo}

// ── ComputeJudgment ──

func ComputeJudgment(c Chart) JudgmentResult {
	patterns := findPatterns(c.Palaces)

	topScore := 0
	lowScore := 0
	topCount := 0
	lowCount := 0
	for _, p := range patterns {
		if p.Score == 2 { topCount++; topScore = 2 }
		if p.Score == 1 && topScore < 1 { topScore = 1 }
		if p.Score == 0 { lowCount++; lowScore = 1 }
	}

	// 四化列表: 星名+类型
	var siHuaItems []SiHuaItem
	for si, ht := range c.SiHua {
		if ht == HuaLu || ht == HuaQuan || ht == HuaKe {
			siHuaItems = append(siHuaItems, SiHuaItem{
				StarID:   int(si),
				StarName: starNames[si],
				Type:     string(ht),
			})
		}
	}
	siHuaCount := len(siHuaItems)

	// 三方四正宫位星曜
	sfList := buildSanFangInfo(c, sanFang(0))

	// 煞星
	shaCount := 0
	for _, p := range c.Palaces {
		for _, s := range p.Stars {
			for _, ss := range shaStars {
				if s.Star == ss {
					shaCount++
				}
			}
		}
	}
	shaStr := "无"
	if shaCount > 2 { shaStr = "多" } else if shaCount > 0 { shaStr = "单" }

	// 命宫主星
	mingStarName := ""
	if len(c.Palaces[0].Stars) > 0 {
		mingStarName = starNames[c.Palaces[0].Stars[0].Star]
	}

	// 命宫亮度
	mingBright := ""
	if len(c.Palaces[0].Stars) > 0 {
		b := miaoWang(c.Palaces[0].Stars[0].Star, c.Palaces[0].Zhi)
		switch {
		case b <= Miao: mingBright = "庙"
		case b <= Wang: mingBright = "旺"
		case b <= Li:   mingBright = "利"
		case b <= Ping: mingBright = "平"
		default:        mingBright = "陷"
		}
	}

	rating, ruleID := lookupJudgment(topScore, topCount, lowScore, siHuaCount, shaStr, mingBright, mingStarName)
	summary := buildSummary(rating, patterns, siHuaCount)

	return JudgmentResult{
		Patterns: patterns,
		SiHua:    siHuaItems,
		SanFang:  sfList,
		Rating:   rating,
		Rule:     ruleID,
		Summary:  summary,
	}
}

func buildSanFangInfo(c Chart, sfPalaces [4]palaceIndex) []SanFangInfo {
	var result []SanFangInfo
	for _, pi := range sfPalaces {
		p := c.Palaces[pi]
		info := SanFangInfo{Name: palaceNames[pi]}
		for _, s := range p.Stars {
			if s.SiHua != "" {
				info.SiHua = s.SiHua
			}
		}
		for _, s := range p.Stars {
			if s.IsMajor {
				info.ZhuXing = append(info.ZhuXing, s.Name)
			} else {
				info.FuXing = append(info.FuXing, s.Name)
			}
		}
		result = append(result, info)
	}
	return result
}

func lookupJudgment(topScore, topCount, lowScore, siHuaCount int, shaStr, mingBright, mingStarName string) (string, int) {
	for _, rule := range judgmentRules {
		if matchJudgment(rule.Conds, topScore, topCount, lowScore, siHuaCount, shaStr, mingBright, mingStarName) {
			return rule.Rating, rule.Rule
		}
	}
	return "中", 0
}

func matchJudgment(c judgmentConds, topScore, topCount, lowScore, siHuaCount int, shaStr, mingBright, mingStarName string) bool {
	if c.TopScore != nil && topScore != *c.TopScore { return false }
	if c.TopCount != nil {
		if c.TopCount.Min != nil && topCount < *c.TopCount.Min { return false }
		if c.TopCount.Max != nil && topCount > *c.TopCount.Max { return false }
	}
	if c.LowScore != nil && lowScore < *c.LowScore { return false }
	if c.SiHuaCount != nil {
		if c.SiHuaCount.Min != nil && siHuaCount < *c.SiHuaCount.Min { return false }
		if c.SiHuaCount.Max != nil && siHuaCount > *c.SiHuaCount.Max { return false }
	}
	if len(c.Brightness) > 0 {
		ok := false
		for _, b := range c.Brightness { if b == mingBright { ok = true; break } }
		if !ok { return false }
	}
	if len(c.ShaXing) > 0 {
		ok := false
		for _, s := range c.ShaXing { if s == shaStr { ok = true; break } }
		if !ok { return false }
	}
	if c.MingGong != "" && mingStarName != c.MingGong { return false }
	return true
}

func buildSummary(rating string, patterns []pattern, siHuaCount int) string {
	switch rating {
	case "上": return "命盘格局上等，福泽深厚"
	case "中": return "命盘格局中等，平顺中求发展"
	default:  return "命盘格局下等，宜读命以修心"
	}
}
