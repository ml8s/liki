package qimen

import (
	"fmt"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// JudgmentResult holds the structured judgment for a 奇门 chart.
type JudgmentResult struct {
	SubjectPalace    GongIndex `json:"subject_palace"`
	EventPalace      GongIndex `json:"ying_qi_gong"`
	DutyStarPalace   GongIndex `json:"zhi_fu_xing_gong"`
	DutyDoorPalace   GongIndex `json:"zhi_shi_men_gong"`
	Rating           string   `json:"rating"`
	ShengKe          string   `json:"sheng_ke"`
	Patterns         []string `json:"patterns"`
	KongWangAffected bool     `json:"kong_wang_affected"`
	MaXingAffected   bool     `json:"ma_xing_affected"`
	Advice           string   `json:"advice"`
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
	riGan := c.Pan.RiGan
	riGanPalace := findGanPalace(c, riGan)

	// Subject: 日干 gong (or event-specific stem)
	subjectG := subjectGan(c.Pan, event)
	subjectP := findGanPalace(c, subjectG)

	// Event: 时干 gong (or event-specific 用神)
	eventG := c.Pan.DriveGan
	eventP := findGanPalace(c, eventG)

	// Duty star/door palaces
	dutyStarP := findStarPalace(c, c.Pan.DutyStar)
	dutyDoorP := findDoorPalace(c, c.Pan.DutyDoor)

	// 门宫生克分析 (door vs gong)
	shengKe := analyzeShengKe(c, subjectP, eventP, dutyStarP, dutyDoorP)

	// Count auspicious indicators
	auspiciousCount := 0
	var patterns []string
	for _, p := range c.Patterns {
		if p.Auspicious {
			auspiciousCount++
		}
		patterns = append(patterns, p.Name)
	}

	// Check interactions
	for _, si := range c.GanInteractions {
		if si.Auspicious {
			auspiciousCount++
		}
	}
	for _, si := range c.XingInteractions {
		if si.Auspicious {
			auspiciousCount++
		}
	}
	auspiciousCount -= len(c.MenPo) // 门迫减分
	auspiciousCount -= len(c.MenZhi) // 门制减分

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

	// Rating
	// 五不遇时直接降为凶
	isWuBu := c.Pan.WuBuYuShi
	rating := rateJudgment(auspiciousCount, kongWangAffected, maXingAffected, len(c.MenPo)+len(c.MenZhi), isWuBu)

	advice := generateAdvice(event, rating, riGanPalace, subjectP, eventP)

	return JudgmentResult{
		SubjectPalace:    subjectP,
		EventPalace:      eventP,
		DutyStarPalace:   dutyStarP,
		DutyDoorPalace:   dutyDoorP,
		Rating:           rating,
		ShengKe:          shengKe,
		Patterns:         patterns,
		KongWangAffected: kongWangAffected,
		MaXingAffected:   maXingAffected,
		Advice:           advice,
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

// findStarPalace finds which gong a star resides in.
func findStarPalace(c Chart, s StarIndex) GongIndex {
	for i, p := range c.Pan.GongWei {
		if p.Star == s {
			return GongIndex(i + 1)
		}
	}
	return 0
}

// findDoorPalace finds which gong a door resides in.
func findDoorPalace(c Chart, d DoorIndex) GongIndex {
	for i, p := range c.Pan.GongWei {
		if p.Door == d {
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

// rateJudgment produces a rating based on indicators.
func rateJudgment(auspiciousCount int, kongWang, maXing bool, menPoZhiCount int, wuBuYuShi bool) string {
	score := auspiciousCount
	if kongWang {
		score -= 2
	}
	if maXing {
		score += 1
	}
	score -= menPoZhiCount

	if wuBuYuShi {
		// 五不遇时: "号为日月损光明" — 无论其他条件多好, 至少降两级
		score -= 3
	}

	switch {
	case score >= 5:
		return "大吉"
	case score >= 3:
		return "吉"
	case score >= 0:
		return "平"
	case score >= -2:
		return "凶"
	default:
		return "大凶"
	}
}

// generateAdvice produces a Chinese advice string.
func generateAdvice(event EventKind, rating string, riGanPalace, subjectP, eventP GongIndex) string {
	switch rating {
	case "大吉":
		return "诸事顺利，吉星高照，宜果断行动"
	case "吉":
		return "较为顺利，虽有小的阻碍但总体有利"
	case "平":
		return "平平无奇，宜守不宜攻，等待时机"
	case "凶":
		return "多有阻碍，宜谨慎行事，避免冒进"
	default:
		return "凶险之象，不宜轻举妄动"
	}
}
