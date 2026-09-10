package liuyao

import "liki-engine/internal/engine/ganzhi"

// DayRelation describes 日建与爻的关系.
type DayRelation struct {
	Relation string `json:"relation"` // 生/扶/克/冲/合
	Strength string `json:"strength"` // 旺/平/衰
}

func dayInteraction(lineZhi ganzhi.Zhi, riZhi ganzhi.Zhi) DayRelation {
	le := ganzhi.ZhiWuxing(lineZhi)
	de := ganzhi.ZhiWuxing(riZhi)

	rel := DayRelation{}

	// 冲
	if ganzhi.IsLiuChong(lineZhi, riZhi) {
		rel.Relation = "冲"
		rel.Strength = "衰"
		return rel
	}

	// 合
	if ganzhi.IsZhiHe(lineZhi, riZhi) {
		rel.Relation = "合"
		rel.Strength = "旺"
		return rel
	}

	// 生: day generates line
	if ganzhi.Sheng(de, le) {
		rel.Relation = "生"
		rel.Strength = "旺"
		return rel
	}

	// 扶: same element
	if le == de {
		rel.Relation = "扶"
		rel.Strength = "旺"
		return rel
	}

	// 克: day overcomes line
	if ganzhi.Ke(de, le) {
		rel.Relation = "克"
		rel.Strength = "衰"
		return rel
	}

	rel.Relation = "平"
	rel.Strength = "平"
	return rel
}

// YingQi holds 应期 prediction.
type YingQi struct {
	YongShen     string             `json:"yong_shen"`
	DongYaoPos   int                `json:"dong_yao_pos"`        // 动爻位置
	YingTimeText string             `json:"ying_time_text"`      // 应期描述
	YingTime     *YingTimeCondition `json:"ying_time,omitempty"` // 机器可读触发条件
	Assessment   string             `json:"assessment"`          // 综合判断
}

// YingTimeCondition is the auditable trigger behind an 应期 statement. A
// trigger branch is not a date: it must be projected onto the calendar.
type YingTimeCondition struct {
	Relation      string `json:"relation"`
	TargetBranch  string `json:"target_branch,omitempty"`
	TriggerBranch string `json:"trigger_branch,omitempty"`
}

func computeYingQi(p *Chart, typ YongShen) YingQi {

	yq := YingQi{
		YongShen: typ.String(),
	}

	yongPos := p.findYongShen(typ)
	if yongPos == 0 {
		// 用神不上卦，找伏神.
		fs := p.findFuShen(typ)
		if fs != nil {
			flying := p.Lines[fs.Position-1]
			yq.Assessment = typ.String() + "不上卦，伏于" + ganzhi.ZhiName(flying.Zhi) + "之下，待冲飞神" + ganzhi.ZhiName(flying.Zhi) + "（" + ganzhi.ZhiName(chongZhi(flying.Zhi)) + "）出伏为应"
			yq.YingTime = &YingTimeCondition{
				Relation:      "冲飞出伏",
				TargetBranch:  ganzhi.ZhiName(flying.Zhi),
				TriggerBranch: ganzhi.ZhiName(chongZhi(flying.Zhi)),
			}
			return yq
		}
		yq.Assessment = typ.String() + "不上卦，问事不吉"
		return yq
	}

	yao := p.Lines[yongPos-1]

	// Check if the用神 line is a动爻.
	if yao.Type.IsChanging() {
		yq.DongYaoPos = yongPos
		yq.YingTimeText = "动爻临值之时（" + ganzhi.ZhiName(yao.Zhi) + "年月）为应"
		yq.YingTime = &YingTimeCondition{
			Relation:      "逢值",
			TargetBranch:  ganzhi.ZhiName(yao.Zhi),
			TriggerBranch: ganzhi.ZhiName(yao.Zhi),
		}
	}

	// Month旺衰.
	ws := ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(yao.Zhi), p.YueZhi)

	// Day interaction.
	di := dayInteraction(yao.Zhi, p.RiZhi)

	yq.Assessment = typ.String() + "在" + ordinal(yongPos) + "爻" +
		"，月建" + ws.String() + "，日建" + di.Relation + "(" + di.Strength + ")"
	if yq.DongYaoPos > 0 {
		yq.Assessment += "，" + yq.YingTimeText
	} else {
		yq.Assessment += "，静爻待冲。冲" + ganzhi.ZhiName(chongZhi(yao.Zhi)) + "之时为应"
		yq.YingTime = &YingTimeCondition{
			Relation:      "冲",
			TargetBranch:  ganzhi.ZhiName(yao.Zhi),
			TriggerBranch: ganzhi.ZhiName(chongZhi(yao.Zhi)),
		}
	}

	return yq
}

func ordinal(n int) string {
	names := [7]string{"", "初", "二", "三", "四", "五", "上"}
	if n >= 1 && n <= 6 {
		return names[n]
	}
	return "?"
}

func chongZhi(z ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.ChongZhi(z)
}
