package qimen

import (
	"fmt"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// JudgmentResult holds the structured judgment for a 奇门 chart.
// 仅含参数化断事要素（主题宫/生克/格局/空亡马星）；值符宫/值使宫/日干宫已并入 qimen.chart（排盘固有）。
type JudgmentResult struct {
	SubjectPalace    GongIndex `json:"subject_palace"`
	EventPalace      GongIndex `json:"ying_qi_gong"`
	ShengKe          string   `json:"sheng_ke"`
	Patterns         []string `json:"patterns"`
	KongWangAffected bool     `json:"kong_wang_affected"`
	MaXingAffected   bool     `json:"ma_xing_affected"`
}

// eventYongShen maps event types to their quintessential palaces/stars/doors.
type eventConfig struct {
	yangShen  []ganzhi.Gan // stems that are relevant (e.g. 乙=woman, 庚=man for 婚姻)
	door      DoorIndex    // optional: specific door
	star      StarIndex    // optional: specific star
}

var eventConfigs = map[EventKind]eventConfig{
	EventGeneral:      {},
	EventCareer:       {door: DoorKai},
	EventWealth:       {yangShen: []ganzhi.Gan{ganzhi.GanWu}, door: DoorSheng},
	EventRelationship: {yangShen: []ganzhi.Gan{ganzhi.GanYi, ganzhi.GanGeng}},
	EventStudy:        {},
	EventTravel:       {},
	EventHealth:       {star: StarTianRui, door: DoorSi},
	EventLegal:        {door: DoorShang},
}

// subjectGan returns the subject stem for an event type.
func subjectGan(c pan, event EventKind) ganzhi.Gan {
	ec, ok := eventConfigs[event]
	if !ok {
		return c.RiGan
	}
	if len(ec.yangShen) > 0 {
		return ec.yangShen[0]
	}
	return c.RiGan
}

// ComputeJudgment analyzes a 奇门 chart for a given event.
func ComputeJudgment(c Chart, event EventKind) JudgmentResult {
	// Subject: 日干 gong (or event-specific stem)
	subjectG := subjectGan(c.Pan, event)
	subjectP := findGanPalace(c, subjectG)

	// Event: 时干 gong (or event-specific 用神)
	eventG := c.Pan.DriveGan
	eventP := findGanPalace(c, eventG)

	// 门宫生克分析（值符宫/值使宫从 chart 排盘固有字段读取）
	shengKe := analyzeShengKe(c, subjectP, eventP, c.DutyStarPalace, c.DutyDoorPalace)

	patterns := make([]string, 0)
	for _, p := range c.Patterns {
		patterns = append(patterns, p.Name)
	}

	// Check 空亡
	kongWangAffected := false
	for _, k := range c.Pan.KongWang {
		if k == subjectP || k == eventP {
			kongWangAffected = true
			break
		}
	}

	// Check 马星
	maXingAffected := c.Pan.MaXing == subjectP || c.Pan.MaXing == eventP

	return JudgmentResult{
		SubjectPalace:    subjectP,
		EventPalace:      eventP,
		ShengKe:          shengKe,
		Patterns:         patterns,
		KongWangAffected: kongWangAffected,
		MaXingAffected:   maXingAffected,
	}
}

// findGanPalace finds which gong a heavenly stem resides in (earth plate).
func findGanPalace(c Chart, g ganzhi.Gan) GongIndex {
	for i, p := range c.Pan.GongWei {
		if p.EarthStem == g || p.HeavenStem == g {
			return GongIndex(i + 1)
		}
	}
	return 0
}


// analyzeShengKe analyzes the sheng/ke relationships between key palaces.
func analyzeShengKe(c Chart, subjectP, eventP, dutyStarP, dutyDoorP GongIndex) string {
	var parts []string

	// 日干落宫生克关系
	if subjectP > 0 && eventP > 0 {
		sp := palaceWuxing(subjectP)
		ep := palaceWuxing(eventP)
		if sp == ep {
			parts = append(parts, fmt.Sprintf("日干(%d宫)与时干(%d宫)比和", subjectP, eventP))
		} else if ganzhi.Sheng(sp, ep) {
			parts = append(parts, fmt.Sprintf("日干(%d宫)生时干(%d宫)", subjectP, eventP))
		} else if ganzhi.Sheng(ep, sp) {
			parts = append(parts, fmt.Sprintf("时干(%d宫)生日干(%d宫)", eventP, subjectP))
		} else if ganzhi.Ke(sp, ep) {
			parts = append(parts, fmt.Sprintf("日干(%d宫)克时干(%d宫)", subjectP, eventP))
		} else {
			parts = append(parts, fmt.Sprintf("时干(%d宫)克日干(%d宫)", eventP, subjectP))
		}
	}

	if len(parts) == 0 {
		return "无显著生克关系"
	}
	return strings.Join(parts, "；")
}

