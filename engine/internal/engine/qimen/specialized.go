package qimen

import (
	"sort"

	"liki-engine/internal/engine/ganzhi"
)

// SpecializedContext holds engine-owned facts for specialized readings.
// Python may add fixed palace presentation metadata, but never reselect facts.
type SpecializedContext struct {
	GengGe      []GengGeFact      `json:"geng_ge"`
	LostContext []LostContextFact `json:"lost_facts"`
	Thief       []ThiefFact       `json:"thief_facts"`
	Tianwang    []TianwangFact    `json:"tianwang_facts"`
}

type GengGeFact struct {
	Level    string     `json:"level"`
	Gong     GongIndex  `json:"gong"`
	Door     string     `json:"door,omitempty"`
	EarthGan ganzhi.Gan `json:"earth_gan"`
}

type LostContextFact struct {
	Gong      GongIndex `json:"gong"`
	Door      string    `json:"door,omitempty"`
	WangShuai string    `json:"wang_shuai,omitempty"`
}

type ThiefFact struct {
	Role              string    `json:"role"`
	Symbol            string    `json:"symbol"`
	Gong              GongIndex `json:"gong"`
	WangShuaiLabel    string    `json:"wang_shuai"`
	LuckyPatternNames []string  `json:"lucky_ge_ju_names"`
}

type TianwangFact struct {
	Kind           string     `json:"kind"`
	Gan            ganzhi.Gan `json:"gan"`
	Gong           GongIndex  `json:"gong"`
	HourGong       GongIndex  `json:"hour_gong"`
	RelationToHour string     `json:"relation_to_hour"`
}

type pillarLevel struct {
	Level string
	Gan   ganzhi.Gan
	Zhi   ganzhi.Zhi
}

func computeSpecializedContext(chart Chart, bz ganzhi.Bazi) SpecializedContext {
	return SpecializedContext{
		GengGe:      computeGengGeFacts(chart, bz),
		LostContext: computeLostContextFacts(chart),
		Thief:       computeThiefFacts(chart),
		Tianwang:    computeTianwangFacts(chart),
	}
}

func computeGengGeFacts(chart Chart, bz ganzhi.Bazi) []GengGeFact {
	pillars := []pillarLevel{
		{Level: "year", Gan: bz.Nian.Gan, Zhi: bz.Nian.Zhi},
		{Level: "month", Gan: bz.Yue.Gan, Zhi: bz.Yue.Zhi},
		{Level: "day", Gan: bz.Ri.Gan, Zhi: bz.Ri.Zhi},
		{Level: "hour", Gan: bz.Shi.Gan, Zhi: bz.Shi.Zhi},
	}
	result := []GengGeFact{}
	for _, interaction := range chart.GanInteractions {
		if interaction.TianPanGan != ganzhi.GanGeng {
			continue
		}
		for _, pillar := range pillars {
			target := resolveJiaDunGan(pillar.Gan, pillar.Zhi)
			if interaction.DiPanGan != target {
				continue
			}
			door := ""
			if int(interaction.Gong) >= 1 && int(interaction.Gong) <= len(chart.Board.GongWei) {
				palace := chart.Board.GongWei[int(interaction.Gong)-1]
				if palace.DoorSet && palace.Door != 0 {
					door = palace.Door.String()
				}
			}
			result = append(result, GengGeFact{
				Level: pillar.Level, Gong: interaction.Gong, Door: door,
				EarthGan: interaction.DiPanGan,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := levelRank(result[i].Level), levelRank(result[j].Level)
		if left != right {
			return left < right
		}
		return result[i].Gong < result[j].Gong
	})
	return result
}

const qimenLevelCount = 4

func levelRank(level string) int {
	switch level {
	case "year":
		return 0
	case "month":
		return 1
	case "day":
		return 2
	case "hour":
		return 3
	default:
		return qimenLevelCount
	}
}

func computeLostContextFacts(chart Chart) []LostContextFact {
	if chart.HourStemPalace.Palace == 0 {
		return []LostContextFact{}
	}
	palace := chart.HourStemPalace.Palace
	door := ""
	if int(palace) >= 1 && int(palace) <= len(chart.Board.GongWei) {
		item := chart.Board.GongWei[int(palace)-1]
		if item.DoorSet && item.Door != 0 {
			door = item.Door.String()
		}
	}
	return []LostContextFact{{
		Gong: palace, Door: door,
		WangShuai: chart.HourStemPalace.WangShuai.WangShuai.String(),
	}}
}

func luckyPatternsAt(patterns []Pattern, gong GongIndex) []string {
	names := []string{}
	for _, pattern := range patterns {
		if !pattern.Auspicious {
			continue
		}
		for _, palace := range pattern.GongWei {
			if palace == gong {
				names = append(names, pattern.Name)
				break
			}
		}
	}
	return names
}

func computeThiefFacts(chart Chart) []ThiefFact {
	starLabels := map[GongIndex]string{}
	for _, row := range chart.XingGongWuXing {
		if row.Star == StarTianPeng {
			starLabels[row.Gong] = row.TraditionalLabel
		}
	}
	palaceWang := map[GongIndex]string{}
	for _, row := range chart.GongWangShuai {
		palaceWang[row.Gong] = row.WangShuai.String()
	}

	result := []ThiefFact{}
	add := func(role, symbol string, gong GongIndex, label string) {
		lucky := luckyPatternsAt(chart.Patterns, gong)
		if lucky == nil {
			lucky = []string{}
		}
		result = append(result, ThiefFact{
			Role: role, Symbol: symbol, Gong: gong,
			WangShuaiLabel: label, LuckyPatternNames: lucky,
		})
	}

	for index, palace := range chart.Board.GongWei {
		gong := GongIndex(index + 1)
		if gong == GongZhong {
			continue
		}
		hasPeng := false
		for _, symbol := range palace.TianPan {
			if symbol.Star == StarTianPeng {
				hasPeng = true
				break
			}
		}
		if hasPeng {
			label := starLabels[gong]
			if label == "" {
				label = "未知"
			}
			add("大贼", "天蓬", gong, label)
		}
		if palace.SpiritSet && palace.Spirit == SpiritZhuQue &&
			spiritDisplayName(palace.Spirit, chart.Board.YinDun, chart.Board.School) == "玄武" {
			label := palaceWang[gong]
			if label == "" {
				label = "未知"
			}
			add("小贼", "玄武", gong, label)
		}
	}
	return result
}

func computeTianwangFacts(chart Chart) []TianwangFact {
	hourGong := chart.HourStemPalace.Palace
	if hourGong == 0 {
		return []TianwangFact{}
	}
	result := []TianwangFact{}
	for _, interaction := range chart.GanInteractions {
		if interaction.Gong != hourGong {
			continue
		}
		kind := ""
		switch interaction.TianPanGan {
		case ganzhi.GanGui:
			kind = "天网四张"
		case ganzhi.GanRen:
			kind = "地罗遮蔽"
		default:
			continue
		}
		result = append(result, TianwangFact{
			Kind: kind, Gan: interaction.TianPanGan,
			Gong: interaction.Gong, HourGong: hourGong,
			RelationToHour: elementRelationToHour(
				palaceWuxing(interaction.Gong), palaceWuxing(hourGong),
			),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Gong < result[j].Gong
	})
	return result
}

func elementRelationToHour(source, target ganzhi.Wuxing) string {
	if source == target {
		return "same"
	}
	switch {
	case ganzhi.Sheng(source, target):
		return "generates"
	case ganzhi.Ke(source, target):
		return "controls"
	case ganzhi.Sheng(target, source):
		return "generated_by"
	default:
		return "controlled_by"
	}
}
