package liuyao

import (
	"encoding/json"
	_ "embed"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

// ── Judgment types ──

type JudgmentResult struct {
	YongShen YongShenState `json:"yong_shen"`
	Rating   string        `json:"rating"`
	Rule     int           `json:"rule,omitempty"`
	Advice   string        `json:"advice"`
}

type YongShenState struct {
	Name  string `json:"name"`
	Month string `json:"month"`
	Day   string `json:"day"`
	DayPo string `json:"day_power"`
	// computed flags
	IsChiShi  bool `json:"is_chi_shi,omitempty"`
	IsYuePo   bool `json:"is_yue_po,omitempty"`
	DongSheng bool `json:"dong_sheng,omitempty"`
	DongKe    bool `json:"dong_ke,omitempty"`
	DongSelf  bool `json:"dong_self,omitempty"`
	Exists    bool `json:"exists"`
}

// ── Rule chain ──

//go:embed data/judgment_rules.json
var judgmentRulesJSON []byte

type judgmentRule struct {
	Rule   int              `json:"rule"`
	Rating string           `json:"rating"`
	Conds  judgmentConds    `json:"conds"`
	Source string           `json:"source"`
}

type judgmentConds struct {
	Exists    *bool    `json:"exists,omitempty"`
	YuePo     *bool    `json:"yue_po,omitempty"`
	ChiShi    *bool    `json:"chi_shi,omitempty"`
	DongSheng *bool    `json:"dong_sheng,omitempty"`
	DongKe    *bool    `json:"dong_ke,omitempty"`
	Month     []string `json:"month,omitempty"`
	DayPower  []string `json:"day_power,omitempty"`
}

var judgmentRules []judgmentRule

func init() {
	if err := json.Unmarshal(judgmentRulesJSON, &judgmentRules); err != nil {
		log.Fatalf("liuyao: load judgment rules: %v", err)
	}
}

// ── Event→用神映射 ──

var eventYongShen = map[string]YongShen{
	"general":      YongShiYao,
	"career":       YongGuanGui,
	"wealth":       YongQiCai,
	"relationship": YongGuanGui,
	"study":        YongFumu,
	"health":       YongGuanGui,
	"legal":        YongGuanGui,
	"travel":       YongShiYao,
}

// ComputeJudgment analyzes a 六爻 chart via rule chain.
func ComputeJudgment(c Chart, event string) JudgmentResult {
	if event == "" {
		event = "general"
	}
	ysType, ok := eventYongShen[event]
	if !ok {
		ysType = YongShiYao
	}

	ys := analyzeYongShen(c, ysType)
	rating, ruleID := lookupRating(ys)
	advice := generateAdvice(event, rating)

	return JudgmentResult{
		YongShen: ys,
		Rating:   rating,
		Rule:     ruleID,
		Advice:   advice,
	}
}

// analyzeYongShen computes all boolean flags for the rule chain.
func analyzeYongShen(c Chart, ysType YongShen) YongShenState {
	pos, isBian := c.findYongShen(ysType)
	if pos == 0 {
		return YongShenState{Name: ysType.String(), Exists: false}
	}

	var line Line
	if isBian {
		line = c.BianLines[pos-1]
	} else {
		line = c.Lines[pos-1]
	}

	ws := ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(line.Zhi), c.MonthZhi)
	di := dayInteraction(line.Zhi, c.DayZhi)
	yuePo := ganzhi.IsLiuChong(line.Zhi, c.MonthZhi)
	shiPos := c.findShiYao()

	// 动爻方向
	lineWx := ganzhi.ZhiWuxing(line.Zhi)
	dongSheng, dongKe := false, false
	for _, dpos := range c.DongYao {
		if dpos < 1 || dpos > 6 {
			continue
		}
		var dl Line
		if dpos == pos {
			dl = line
		} else {
			dl = c.Lines[dpos-1]
		}
		dw := ganzhi.ZhiWuxing(dl.Zhi)
		if ganzhi.Sheng(dw, lineWx) {
			dongSheng = true
		}
		if ganzhi.Ke(dw, lineWx) {
			dongKe = true
		}
	}

	return YongShenState{
		Name:      ysType.String(),
		Month:     ws.String(),
		Day:       di.Relation,
		DayPo:     di.Strength,
		IsChiShi:  shiPos > 0 && shiPos == pos,
		IsYuePo:   yuePo,
		DongSelf:  line.Type.IsChanging(),
		DongSheng: dongSheng,
		DongKe:    dongKe,
		Exists:    true,
	}
}

// lookupRating matches rule chain, returns (rating, ruleID).
func lookupRating(ys YongShenState) (string, int) {
	for _, rule := range judgmentRules {
		if matchRule(rule.Conds, ys) {
			return rule.Rating, rule.Rule
		}
	}
	return "平", 0
}

func matchRule(c judgmentConds, ys YongShenState) bool {
	if c.Exists != nil && ys.Exists != *c.Exists {
		return false
	}
	if c.YuePo != nil && ys.IsYuePo != *c.YuePo {
		return false
	}
	if c.ChiShi != nil && ys.IsChiShi != *c.ChiShi {
		return false
	}
	if c.DongSheng != nil && ys.DongSheng != *c.DongSheng {
		return false
	}
	if c.DongKe != nil && ys.DongKe != *c.DongKe {
		return false
	}
	if len(c.Month) > 0 {
		match := false
		for _, m := range c.Month {
			if m == ys.Month {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if len(c.DayPower) > 0 {
		match := false
		for _, d := range c.DayPower {
			if d == ys.DayPo {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func generateAdvice(event, rating string) string {
	prefixes := map[string]string{"吉": "卦象吉利", "凶": "卦象不利", "平": "卦象平平"}
	prefix := prefixes[rating]
	if prefix == "" {
		prefix = "卦象平平"
	}
	suffixes := map[string]map[string]string{
		"general":      {"吉": "，所问之事可成", "凶": "，宜谨慎行事", "平": "，宜守不宜攻"},
		"career":       {"吉": "，官鬼旺相，事业有利", "凶": "，官鬼衰弱，事业有阻", "平": "，事业一般"},
		"wealth":       {"吉": "，妻财旺相，财运亨通", "凶": "，妻财衰弱，财运不佳", "平": "，财运一般"},
		"relationship": {"吉": "，官鬼旺相，感情和谐", "凶": "，官鬼衰弱，感情多波折", "平": "，感情平稳"},
		"study":        {"吉": "，学业有成", "凶": "，学业有阻", "平": "，学业平平"},
		"health":       {"吉": "，病情好转", "凶": "，病情反复", "平": "，平稳"},
		"legal":        {"吉": "，诉讼有利", "凶": "，诉讼不利", "平": "，未明"},
		"travel":       {"吉": "，出行顺利", "凶": "，不宜出行", "平": "，出行平安"},
	}
	if m, ok := suffixes[event]; ok {
		if s, ok := m[rating]; ok {
			return prefix + s
		}
	}
	return prefix + "，谨慎行事"
}
