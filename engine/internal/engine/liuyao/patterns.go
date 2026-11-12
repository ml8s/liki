package liuyao

import "liki-engine/internal/engine/ganzhi"

// PatternType 特殊格局类型
type PatternType string

const (
	PatternXunKong    PatternType = "旬空"   // 旬空
	PatternYuePo      PatternType = "月破"   // 月破
	PatternFeiFu      PatternType = "飞伏"   // 飞伏
	PatternJinTui     PatternType = "进退"   // 进退神
	PatternChongHe    PatternType = "冲合"   // 六冲/六合
	PatternFanYin     PatternType = "反吟"   // 动爻变支六冲
	PatternFuYin      PatternType = "伏吟"   // 动爻变支相同
	PatternSuiGuiRuMu PatternType = "随鬼入墓" // 随鬼入墓
	PatternDuFa       PatternType = "独发"   // 独发
	PatternDuJing     PatternType = "独静"   // 独静
	PatternLiangXian  PatternType = "两现"   // 用神两现
)

// Pattern 特殊格局
type Pattern struct {
	Type     PatternType `json:"pattern_kind"` // 格局类型
	SubType  string      `json:"sub_type"`     // 子类型（如真空/假空）
	Position int         `json:"position"`     // 相关爻位（0=全卦）
	IsTrue   bool        `json:"is_true"`      // 结构是否有实质效力（空破区分真假，其余表示已成立）
}

// jinShenTable 固定进神同五行地支顺行；不用五行生克近似。
var jinShenTable = map[ganzhi.Zhi]ganzhi.Zhi{
	ganzhi.ZhiHai:  ganzhi.ZhiZi,  // 亥化子：进
	ganzhi.ZhiYin:  ganzhi.ZhiMao, // 寅化卯：进
	ganzhi.ZhiSi:   ganzhi.ZhiWu,  // 巳化午：进
	ganzhi.ZhiShen: ganzhi.ZhiYou, // 申化酉：进
	ganzhi.ZhiChou: ganzhi.ZhiChen,
	ganzhi.ZhiChen: ganzhi.ZhiWei,
	ganzhi.ZhiWei:  ganzhi.ZhiXu,
	ganzhi.ZhiXu:   ganzhi.ZhiChou,
}

// tuiShenTable 固定退神同五行地支逆行；不用五行生克近似。
var tuiShenTable = map[ganzhi.Zhi]ganzhi.Zhi{
	ganzhi.ZhiZi:   ganzhi.ZhiHai,  // 子化亥：退
	ganzhi.ZhiMao:  ganzhi.ZhiYin,  // 卯化寅：退
	ganzhi.ZhiWu:   ganzhi.ZhiSi,   // 午化巳：退
	ganzhi.ZhiYou:  ganzhi.ZhiShen, // 酉化申：退
	ganzhi.ZhiChen: ganzhi.ZhiChou,
	ganzhi.ZhiWei:  ganzhi.ZhiChen,
	ganzhi.ZhiXu:   ganzhi.ZhiWei,
	ganzhi.ZhiChou: ganzhi.ZhiXu,
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
	patterns = append(patterns, computeFanFuYin(p, yongShen)...)

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
	yongPos := p.findYongShen(yongShen)
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

	line := p.Lines[yongPos-1]
	isWang := p.WangShuai[yongPos-1] == ganzhi.WSWang || p.WangShuai[yongPos-1] == ganzhi.WSXiang
	support := []string{}
	if isWang {
		support = append(support, "旺相")
	}
	if line.DongSelf {
		support = append(support, "发动")
	}
	if line.DongSheng {
		support = append(support, "动爻生扶")
	}
	if p.dayRelationHas(yongPos, "生", "扶") {
		support = append(support, "日建生扶")
	}
	isTrueVacant := len(support) == 0
	subType := "假空"
	if isTrueVacant {
		subType = "真空"
	}

	patterns = append(patterns, Pattern{
		Type:     PatternXunKong,
		SubType:  subType,
		Position: yongPos,
		IsTrue:   isTrueVacant,
	})

	return patterns
}

// computeYuePo 计算月破格局
func computeYuePo(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	// 检查用神是否月破
	if !p.Lines[yongPos-1].YuePo {
		return patterns
	}

	line := p.Lines[yongPos-1]
	isDong := line.DongSelf
	isWang := p.WangShuai[yongPos-1] == ganzhi.WSWang || p.WangShuai[yongPos-1] == ganzhi.WSXiang
	deSheng := line.DongSheng || p.dayRelationHas(yongPos, "生", "扶", "合")
	if line.DongSelf && len(p.BianLines) == 6 && ganzhi.Sheng(p.BianLines[yongPos-1].Wuxing, line.Wuxing) {
		deSheng = true
	}

	// 判断真假破：仅「休囚静爻月破」为真破，其余（旺相动爻/旺相得生/有救应）为假破。
	isTruePo := false
	subType := "假破"
	if !isWang && !isDong && !deSheng {
		isTruePo = true
		subType = "真破"
	}

	patterns = append(patterns, Pattern{
		Type:     PatternYuePo,
		SubType:  subType,
		Position: yongPos,
		IsTrue:   isTruePo,
	})

	return patterns
}

// computeFeiFu 计算飞伏格局
func computeFeiFu(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos := p.findYongShen(yongShen)
	if yongPos != 0 {
		// 用神在本卦，无飞伏
		return patterns
	}

	fuShenAll := p.findFuShenAll(yongShen)
	if len(fuShenAll) == 0 {
		return patterns
	}
	for _, fuShen := range fuShenAll {
		fuZhi, err := ganzhi.ParseZhi(fuShen.Zhi)
		if err != nil {
			continue
		}
		flying := p.Lines[fuShen.Position-1]
		relation := feiFuRelation(flying.Wuxing, ganzhi.ZhiWuxing(fuZhi))
		patterns = append(patterns, Pattern{
			Type:     PatternFeiFu,
			SubType:  relation,
			Position: fuShen.Position,
			IsTrue:   true,
		})
	}

	return patterns
}

// computeJinTui 计算进退神
func computeJinTui(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos := p.findYongShen(yongShen)
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

	// 进退神只认同一五行地支的顺行 / 逆行；其他变爻不构成候选。
	subType := ""
	if next, ok := jinShenTable[benLine.Zhi]; ok && next == bianLine.Zhi {
		subType = "进神"
	} else if prev, ok := tuiShenTable[benLine.Zhi]; ok && prev == bianLine.Zhi {
		subType = "退神"
	}

	if subType != "" {
		patterns = append(patterns, Pattern{
			Type:     PatternJinTui,
			SubType:  subType,
			Position: yongPos,
			IsTrue:   true,
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
			Type:     PatternChongHe,
			SubType:  "六冲",
			Position: 0,
			IsTrue:   true,
		})
	}

	// 本卦六合检查
	if p.BenGua.isLiuHe() {
		isLiuHe = true
		patterns = append(patterns, Pattern{
			Type:     PatternChongHe,
			SubType:  "六合",
			Position: 0,
			IsTrue:   true,
		})
	}
	if p.BianGua != 0 {
		benChong, benHe := p.BenGua.isLiuChong(), p.BenGua.isLiuHe()
		bianChong, bianHe := p.BianGua.isLiuChong(), p.BianGua.isLiuHe()
		for _, tc := range []struct {
			subType string
			matched bool
		}{
			{"六冲变六合", benChong && bianHe},
			{"六合变六冲", benHe && bianChong},
			{"六冲变六冲", benChong && bianChong},
			{"六合变六合", benHe && bianHe},
		} {
			if tc.matched {
				patterns = append(patterns, Pattern{
					Type: PatternChongHe, SubType: tc.subType, Position: 0, IsTrue: true,
				})
			}
		}
	}

	// 冲合不同时出现
	if isLiuChong && isLiuHe {
		// 冲合并见，按六冲处理
		patterns = patterns[:len(patterns)-1]
	}

	return patterns
}

// computeFanFuYin 计算爻变与卦变的反吟、伏吟。
// 《增删卜易·反伏章》分别论述“内卦反伏”“外卦反伏”“内外反伏”，并以
// “反吟有冲有克”“伏吟……忧郁呻吟”区分二类：反吟取动爻变支六冲，
// 伏吟取动爻变支同支。静爻保留本支，不构成动变反伏。
func computeFanFuYin(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	if len(p.BianLines) > 0 {
		for _, position := range p.DongYao {
			line := p.Lines[position-1]
			bianLine := p.BianLines[position-1]
			switch {
			case line.Zhi == bianLine.Zhi:
				patterns = append(patterns, Pattern{
					Type:     PatternFuYin,
					SubType:  "伏吟",
					Position: 0,
					IsTrue:   true,
				})
			case ganzhi.IsLiuChong(line.Zhi, bianLine.Zhi):
				patterns = append(patterns, Pattern{
					Type:     PatternFanYin,
					SubType:  "反吟",
					Position: 0,
					IsTrue:   true,
				})
			}
		}
	}

	// 卦变反伏必须以本卦范围内的动爻为准。静爻虽在变卦里保留原支，
	// 但不能拿来补足内卦、外卦或全卦的范围。
	innerFuYin := fanFuYinScope(p, 0, 3, PatternFuYin)
	outerFuYin := fanFuYinScope(p, 3, 6, PatternFuYin)
	innerFanYin := fanFuYinScope(p, 0, 3, PatternFanYin)
	outerFanYin := fanFuYinScope(p, 3, 6, PatternFanYin)

	if innerFuYin && outerFuYin {
		patterns = append(patterns, Pattern{Type: PatternFuYin, SubType: "全卦伏吟", Position: 0, IsTrue: true})
	} else {
		if innerFuYin {
			patterns = append(patterns, Pattern{Type: PatternFuYin, SubType: "内卦伏吟", Position: 0, IsTrue: true})
		}
		if outerFuYin {
			patterns = append(patterns, Pattern{Type: PatternFuYin, SubType: "外卦伏吟", Position: 0, IsTrue: true})
		}
	}

	if innerFanYin && outerFanYin {
		patterns = append(patterns, Pattern{Type: PatternFanYin, SubType: "全卦反吟", Position: 0, IsTrue: true})
	} else {
		if innerFanYin {
			patterns = append(patterns, Pattern{Type: PatternFanYin, SubType: "内卦反吟", Position: 0, IsTrue: true})
		}
		if outerFanYin {
			patterns = append(patterns, Pattern{Type: PatternFanYin, SubType: "外卦反吟", Position: 0, IsTrue: true})
		}
	}

	return patterns
}

// fanFuYinScope 判断一个经卦是否整段发动且统一构成伏吟或反吟。
func fanFuYinScope(p *Chart, start, end int, kind PatternType) bool {
	if len(p.BianLines) == 0 {
		return false
	}
	moving := make(map[int]bool, len(p.DongYao))
	for _, position := range p.DongYao {
		moving[position] = true
	}
	for i := start; i < end; i++ {
		if !moving[i+1] {
			return false
		}
		line, bianLine := p.Lines[i], p.BianLines[i]
		same := line.Zhi == bianLine.Zhi
		chong := ganzhi.IsLiuChong(line.Zhi, bianLine.Zhi)
		if (kind == PatternFuYin && !same) || (kind == PatternFanYin && !chong) {
			return false
		}
	}
	return true
}

// computeSuiGuiRuMu 计算随鬼入墓
func computeSuiGuiRuMu(p *Chart, yongShen YongShen) []Pattern {
	var patterns []Pattern

	// 用神爻位
	yongPos := p.findYongShen(yongShen)
	if yongPos == 0 {
		return patterns
	}

	line := p.Lines[yongPos-1]
	if !line.MuKu || len(line.MuKuTypes) == 0 {
		return patterns
	}
	for _, source := range line.MuKuTypes {
		patterns = append(patterns, Pattern{
			Type:     PatternSuiGuiRuMu,
			SubType:  map[string]string{"day": "随鬼入日墓", "moving": "随鬼入动墓", "transformed": "随鬼化墓"}[source],
			Position: yongPos,
			IsTrue:   true,
		})
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
			Type:     PatternDuFa,
			SubType:  "独发",
			Position: p.DongYao[0],
			IsTrue:   true,
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
					Type:     PatternDuJing,
					SubType:  "独静",
					Position: i,
					IsTrue:   true,
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
			Type:     PatternLiangXian,
			SubType:  "用神两现",
			Position: 0,
			IsTrue:   true,
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
