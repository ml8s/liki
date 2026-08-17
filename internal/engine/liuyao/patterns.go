package liuyao

import "liki-engine/internal/engine/ganzhi"

// PatternType 特殊格局类型
type PatternType string

const (
	PatternXunKong   PatternType = "旬空"   // 旬空
	PatternYuePo     PatternType = "月破"   // 月破
	PatternFeiFu     PatternType = "飞伏"   // 飞伏
	PatternJinTui    PatternType = "进退"   // 进退神
	PatternChongHe   PatternType = "冲合"   // 六冲/六合
	PatternFanYin    PatternType = "反吟"   // 反吟/伏吟
	PatternSuiGuiRuMu PatternType = "随鬼入墓" // 随鬼入墓
	PatternDuFa      PatternType = "独发"   // 独发
	PatternDuJing    PatternType = "独静"   // 独静
	PatternLiangXian PatternType = "两现"   // 用神两现
)

// Pattern 特殊格局
type Pattern struct {
	Type      PatternType `json:"type"`       // 格局类型
	SubType   string             `json:"sub_type"`   // 子类型（如真空/假空）
	Position  int                `json:"position"`   // 相关爻位（0=全卦）
	IsTrue    bool               `json:"is_true"`    // 是否为真格局（如真破/真空）
	Assessment string            `json:"assessment"` // 断语描述
}

// ComputePatterns 计算所有特殊格局
func ComputePatterns(p *Chart, yongShen YongShen) []Pattern {
	patterns := []Pattern{}

	// 旬空
	patterns = append(patterns, computeXunKong(p, yongShen)...)

	// 月破
	patterns = append(patterns, computeYuePo(p, yongShen)...)

	// 飞伏
	patterns = append(patterns, computeFeiFu(p, yongShen)...)

	// 进退神
	patterns = append(patterns, computeJinTui(p, yongShen)...)

	// 六冲/六合
	patterns = append(patterns, computeChongHe(p, yongShen)...)

	// 反吟/伏吟
	patterns = append(patterns, computeFanYin(p, yongShen)...)

	// 随鬼入墓
	patterns = append(patterns, computeSuiGuiRuMu(p, yongShen)...)

	// 独发/独静
	patterns = append(patterns, computeDuFaDuJing(p, yongShen)...)

	// 用神两现
	patterns = append(patterns, computeLiangXian(p, yongShen)...)

	return patterns
}

// computeXunKong 计算旬空格局
func computeXunKong(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos, _ := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	// 检查用神是否旬空
	isXunKong := false
	for _, z := range p.XunKong {
		if p.Lines[yongPos-1].Zhi == z {
			isXunKong = true
			break
		}
	}

	if !isXunKong {
		return patterns
	}

	// 判断动静
	isDong := p.Lines[yongPos-1].Type.IsChanging()

	// 判断旺衰
	wangshuai := p.WangShuai[yongPos-1]
	isWang := wangshuai == ganzhi.WSWang || wangshuai == ganzhi.WSXiang

	// 判断真假空
	isTrueVacant := false
	assessment := ""

	if isWang && isDong {
		// 旺相动爻旬空 = 假空
		assessment = "旺相动爻旬空，迟成而非不成"
	} else if !isWang && !isDong {
		// 休囚静爻旬空 = 真空
		isTrueVacant = true
		assessment = "休囚静爻旬空，事不实，出空方应"
	} else if isWang && !isDong {
		// 旺相静爻旬空 = 假空（旺不为空）
		assessment = "旺相静爻旬空，旺不为空，出空方应"
	} else {
		// 休囚动爻旬空 = 假空（动不为空）
		assessment = "休囚动爻旬空，动不为空，出空方应"
	}

	patterns = append(patterns, Pattern{
		Type:       PatternXunKong,
		SubType:    "假空",
		Position:   yongPos,
		IsTrue:     isTrueVacant,
		Assessment: assessment,
	})

	return patterns
}

// computeYuePo 计算月破格局
func computeYuePo(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos, _ := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	// 检查用神是否月破
	if !p.Lines[yongPos-1].YuePo {
		return patterns
	}

	// 判断动静
	isDong := p.Lines[yongPos-1].Type.IsChanging()

	// 判断旺衰
	wangshuai := p.WangShuai[yongPos-1]
	isWang := wangshuai == ganzhi.WSWang || wangshuai == ganzhi.WSXiang

	// 判断得生
	deSheng := false
	for _, rel := range p.DayRelations {
		if rel.Relation == "生" && rel.Strength == "旺" {
			deSheng = true
			break
		}
	}

	// 判断真假破
	isTruePo := false
	assessment := ""

	if isWang && isDong {
		// 旺相动爻月破 = 假破
		assessment = "旺相动爻月破，先挫后成"
	} else if isWang && deSheng {
		// 旺相得生月破 = 假破
		assessment = "旺相得生月破，先挫后成"
	} else if !isWang && !isDong && !deSheng {
		// 休囚静爻月破 = 真破
		isTruePo = true
		assessment = "休囚静爻月破，当下无力"
	} else {
		// 其他情况 = 假破
		assessment = "月破但有救应，先挫后成"
	}

	patterns = append(patterns, Pattern{
		Type:       PatternYuePo,
		SubType:    "假破",
		Position:   yongPos,
		IsTrue:     isTruePo,
		Assessment: assessment,
	})

	return patterns
}

// computeFeiFu 计算飞伏格局
func computeFeiFu(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos, _ := p.findYongShen(yongShen)
	if yongPos != 0 {
		// 用神在本卦，无飞伏
		return patterns
	}

	// 查找伏神
	fuShen := p.findFuShen(yongShen)
	if fuShen == nil {
		return patterns
	}

	// 判断飞伏关系
	assessment := ""
	feiShengFu := false

	// 飞神生伏神
	feiWuxing := p.Lines[fuShen.Position-1].Wuxing
	// 将伏神地支转换为五行
	fuWuxingStr := fuShen.Zhi
	var fuWuxing ganzhi.Wuxing
	switch fuWuxingStr {
	case "子", "亥":
		fuWuxing = ganzhi.WxShui
	case "寅", "卯":
		fuWuxing = ganzhi.WxMu
	case "巳", "午":
		fuWuxing = ganzhi.WxHuo
	case "申", "酉":
		fuWuxing = ganzhi.WxJin
	case "辰", "戌", "丑", "未":
		fuWuxing = ganzhi.WxTu
	}
	if ganzhi.Sheng(feiWuxing, fuWuxing) {
		feiShengFu = true
		assessment = "飞神生伏神，伏神得力，待冲飞出伏"
	} else if ganzhi.Ke(feiWuxing, fuWuxing) {
		assessment = "飞神克伏神，伏神受压，待冲飞出伏"
	} else {
		assessment = "飞伏比和，伏神待出"
	}

	patterns = append(patterns, Pattern{
		Type:       PatternFeiFu,
		SubType:    "伏藏",
		Position:   fuShen.Position,
		IsTrue:     !feiShengFu,
		Assessment: assessment,
	})

	return patterns
}

// computeJinTui 计算进退神
func computeJinTui(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos, _ := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	// 检查是否动爻
	if !p.Lines[yongPos-1].Type.IsChanging() {
		return patterns
	}

	// 检查变爻
	if len(p.BianLines) == 0 {
		return patterns
	}

	bianLine := p.BianLines[yongPos-1]
	benLine := p.Lines[yongPos-1]

	// 判断进退
	isJin := false
	isTui := false
	assessment := ""

	// 化进：变爻地支在生旺库中前进
	if ganzhi.Sheng(ganzhi.ZhiWuxing(benLine.Zhi), ganzhi.ZhiWuxing(bianLine.Zhi)) {
		isJin = true
		assessment = "用神化进神，力量增长"
	} else if ganzhi.Ke(ganzhi.ZhiWuxing(bianLine.Zhi), ganzhi.ZhiWuxing(benLine.Zhi)) {
		isTui = true
		assessment = "用神化退神，力量衰败"
	}

	if isJin || isTui {
		patternType := PatternJinTui
		subType := "进神"
		if isTui {
			subType = "退神"
		}
		patterns = append(patterns, Pattern{
			Type:       patternType,
			SubType:    subType,
			Position:   yongPos,
			IsTrue:     isJin,
			Assessment: assessment,
		})
	}

	return patterns
}

// computeChongHe 计算六冲/六合
func computeChongHe(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 检查卦体六冲
	isLiuChong := false
	isLiuHe := false

	// 本卦六冲检查
	if p.BenGua.isLiuChong() {
		isLiuChong = true
		patterns = append(patterns, Pattern{
			Type:       PatternChongHe,
			SubType:    "六冲",
			Position:   0,
			IsTrue:     true,
			Assessment: "六冲卦，事多反复",
		})
	}

	// 本卦六合检查
	if p.BenGua.isLiuHe() {
		isLiuHe = true
		patterns = append(patterns, Pattern{
			Type:       PatternChongHe,
			SubType:    "六合",
			Position:   0,
			IsTrue:     true,
			Assessment: "六合卦，事易成",
		})
	}

	// 冲合不同时出现
	if isLiuChong && isLiuHe {
		// 冲合并见，按六冲处理
		patterns = patterns[:len(patterns)-1]
	}

	return patterns
}

// computeFanYin 计算反吟/伏吟
func computeFanYin(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 检查反吟（本卦与变卦地支相冲）
	if len(p.BianLines) > 0 {
		isFanYin := false
		for i, line := range p.Lines {
			bianLine := p.BianLines[i]
			if ganzhi.IsLiuChong(line.Zhi, bianLine.Zhi) {
				isFanYin = true
				break
			}
		}

		if isFanYin {
			patterns = append(patterns, Pattern{
				Type:       PatternFanYin,
				SubType:    "反吟",
				Position:   0,
				IsTrue:     true,
				Assessment: "反吟卦，事多反复",
			})
		}
	}

	// 检查伏吟（本卦与变卦地支相同）
	if len(p.BianLines) > 0 {
		isFuYin := false
		for i, line := range p.Lines {
			bianLine := p.BianLines[i]
			if line.Zhi == bianLine.Zhi {
				isFuYin = true
				break
			}
		}

		if isFuYin {
			patterns = append(patterns, Pattern{
				Type:       PatternFanYin,
				SubType:    "伏吟",
				Position:   0,
				IsTrue:     true,
				Assessment: "伏吟卦，事多停滞",
			})
		}
	}

	return patterns
}

// computeSuiGuiRuMu 计算随鬼入墓
func computeSuiGuiRuMu(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos, _ := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	// 检查用神是否入墓
	if p.Lines[yongPos-1].MuKu {
		// 检查是否随鬼（官鬼爻发动）
		for _, pos := range p.DongYao {
			if p.Lines[pos-1].LiuQin == QinGuanGui {
				patterns = append(patterns, Pattern{
					Type:       PatternSuiGuiRuMu,
					SubType:    "随鬼入墓",
					Position:   yongPos,
					IsTrue:     true,
					Assessment: "用神随鬼入墓，事多闭塞",
				})
				break
			}
		}
	}

	return patterns
}

// computeDuFaDuJing 计算独发/独静
func computeDuFaDuJing(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	dongCount := len(p.DongYao)

	// 独发：五静一动
	if dongCount == 1 {
		patterns = append(patterns, Pattern{
			Type:       PatternDuFa,
			SubType:    "独发",
			Position:   p.DongYao[0],
			IsTrue:     true,
			Assessment: "独发爻，主事之应期",
		})
	}

	// 独静：五动一静
	if dongCount == 5 {
		// 找静爻
		for i := 1; i <= 6; i++ {
			isDong := false
			for _, pos := range p.DongYao {
				if pos == i {
					isDong = true
					break
				}
			}
			if !isDong {
				patterns = append(patterns, Pattern{
					Type:       PatternDuJing,
					SubType:    "独静",
					Position:   i,
					IsTrue:     true,
					Assessment: "独静爻，主事之应期",
				})
				break
			}
		}
	}

	return patterns
}

// computeLiangXian 计算用神两现
func computeLiangXian(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 统计用神爻数
	count := 0
	for _, line := range p.Lines {
		if line.LiuQin == yongShenToLiuQinInternal(yongShen) {
			count++
		}
	}

	if count >= 2 {
		patterns = append(patterns, Pattern{
			Type:       PatternLiangXian,
			SubType:    "用神两现",
			Position:   0,
			IsTrue:     true,
			Assessment: "用神两现，取旺不取衰",
		})
	}

	return patterns
}

// yongShenToLiuQin 将用神类型转换为六亲类型（内部使用）
func yongShenToLiuQinInternal(ys YongShen) LiuQin {
	switch ys {
	case YongFumu:
		return QinFumu
	case YongXiongDi:
		return QinXiongDi
	case YongGuanGui:
		return QinGuanGui
	case YongQiCai:
		return QinQiCai
	case YongZiSun:
		return QinZiSun
	default:
		return QinFumu
	}
}
