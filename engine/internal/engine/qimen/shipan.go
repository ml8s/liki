package qimen

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// Chart bundles a complete奇门盘 with all analysis layers.
type Chart struct {
	Method              ChartMethod        `json:"method"`
	Pan                 pan                `json:"pan"`
	GanInteractions     []GanInteraction   `json:"gan_interaction"`
	MenInteractions     []MenInteraction   `json:"men_interaction"`
	XingInteractions    []XingInteraction  `json:"xing_interaction"`
	XingGongWuXing      []XingGongWuXing   `json:"xing_gong_wu_xing"`
	MenPo               []GongIndex        `json:"men_po"`
	MenZhi              []GongIndex        `json:"men_zhi"`
	Patterns            []Pattern          `json:"patterns"`
	YingQi              YingQi             `json:"ying_qi"`
	RiGanPalace         GongIndex          `json:"ri_gan_gong"`  // 日干落宫（排盘固有）
	ShiGanPalace        GongIndex          `json:"shi_gan_gong"` // 时干落宫（排盘固有）
	RiGanPalaceFacts    []GanPalaceFact    `json:"ri_gan_palace_facts"`
	ShiGanPalaceFacts   []GanPalaceFact    `json:"shi_gan_palace_facts"`
	RiShiRelation       RiShiRelation      `json:"ri_shi_relation"`
	KongWangAffected    []AffectedSymbol   `json:"kong_wang_affected"`
	MaXingAffected      []AffectedSymbol   `json:"ma_xing_affected"`
	PalaceWangShuai     []PalaceWangShuai  `json:"palace_wang_shuai"`
	ShiGanGongWangShuai PalaceWangShuai    `json:"shi_gan_gong_wang_shuai"`
	DutyStarPalace      GongIndex          `json:"zhi_fu_xing_gong"`    // 值符星落宫（排盘固有）
	DutyDoorPalace      GongIndex          `json:"zhi_shi_men_gong"`    // 值使门落宫（排盘固有）
	YongShen            *YongShenResult    `json:"yong_shen,omitempty"` // 用神领域对象（求测人+事象用神）
	Specialized         SpecializedContext `json:"specialized"`
}

// GanPalaceFact records every palace layer occupied by a stem. The primary
// ri_gan_gong/shi_gan_gong keeps the heaven-first rule; these facts preserve
// the earth-stem palace that the single value previously discarded.
type GanPalaceFact struct {
	Palace GongIndex `json:"palace"`
	Layer  string    `json:"layer"` // heaven / earth
}

type RiShiRelation struct {
	Subject  string `json:"subject"`
	Object   string `json:"object"`
	Relation string `json:"relation"`
	Name     string `json:"name"`
}

type AffectedSymbol struct {
	Symbol string     `json:"symbol"`
	Role   string     `json:"role"`
	Branch ganzhi.Zhi `json:"branch"`
	Gong   GongIndex  `json:"gong"`
}

// PalaceWangShuai 表示宫位五行在当前月令下的旺相休囚死状态。
type PalaceWangShuai struct {
	Gong          GongIndex        `json:"gong"`
	Wuxing        ganzhi.Wuxing    `json:"wuxing"`
	WangShuai     ganzhi.WangShuai `json:"wang_shuai"`
	WangShuaiName string           `json:"wang_shuai_name"`
}

// computeChart computes a qimen chart with the base timing focus.
func computeChart(bz ganzhi.Bazi, chartTime time.Time, method Method) Chart {
	chart := buildChart(bz, chartTime, method)
	chart.YingQi = computeYingQi(chart, chartTime, baseYingQiFocuses(chart))
	return chart
}

// computeChartWithYongShen computes the chart and projects timing through the
// selected symbols without first building and discarding a base timing result.
func computeChartWithYongShen(
	bz ganzhi.Bazi, chartTime time.Time, method Method,
	syms []YongShenSymbol, birthDate BirthDate,
) Chart {
	chart := buildChart(bz, chartTime, method)
	var yongShen YongShenResult
	if birthDate.Has {
		yongShen = computeYongShenWithBirth(chart, syms, birthDate)
	} else {
		yongShen = computeYongShenSymbols(chart, syms)
	}
	chart.YongShen = &yongShen
	chart.YingQi = computeYingQi(chart, chartTime, yingQiFocuses(chart, yongShen))
	return chart
}

func buildChart(bz ganzhi.Bazi, chartTime time.Time, method Method) Chart {
	ju := determineJuShu(chartTime, bz, method)
	p := computePan(ju, bz)

	riGanP := findGanPalaceIdx(p, resolveJiaDunGan(bz.Ri.Gan, bz.Ri.Zhi))
	shiGanP := findGanPalaceIdx(p, resolveJiaDunGan(bz.Shi.Gan, bz.Shi.Zhi))
	riGanFacts := findGanPalaceFacts(p, resolveJiaDunGan(bz.Ri.Gan, bz.Ri.Zhi))
	shiGanFacts := findGanPalaceFacts(p, resolveJiaDunGan(bz.Shi.Gan, bz.Shi.Zhi))
	leadP := leadPillarPalace(p)

	chart := Chart{
		Method:              buildChartMethod(ju, bz),
		Pan:                 p,
		GanInteractions:     computeGanInteractions(p),
		MenInteractions:     computeMenInteractions(p),
		XingInteractions:    computeXingInteractions(p),
		XingGongWuXing:      computeXingGongWuXing(p),
		MenPo:               findMenPo(p),
		MenZhi:              findMenZhi(p),
		Patterns:            findPatterns(p),
		RiGanPalace:         riGanP,
		ShiGanPalace:        shiGanP,
		RiGanPalaceFacts:    riGanFacts,
		ShiGanPalaceFacts:   shiGanFacts,
		RiShiRelation:       analyzeRiShiRelation(riGanP, shiGanP),
		KongWangAffected:    affectedVoidSymbols(p, riGanP, shiGanP, leadP),
		MaXingAffected:      affectedHorseSymbols(p, riGanP, shiGanP, leadP),
		PalaceWangShuai:     palaceWangShuai(p, bz.Yue.Zhi),
		ShiGanGongWangShuai: palaceWangShuaiForPalace(p, bz.Yue.Zhi, shiGanP),
		DutyStarPalace:      findStarPalace(p, p.DutyStar),
		DutyDoorPalace:      findDoorPalaceIdx(p, p.DutyDoor),
	}
	chart.Specialized = computeSpecializedContext(chart, bz)
	return chart
}

func buildChartMethod(ju juShu, bz ganzhi.Bazi) ChartMethod {
	leadZhu := leadPillar(bz, ju)
	xunIdx := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi) / 10 * 10
	xunShou := ganzhi.SixtyToZhu(xunIdx)
	fu := fuTou(bz.Ri)
	duty := findDuty(leadZhu, placeDiPan(ju.Number, ju.YinDun))
	xunShouInfo := newPillarInfo(xunShou)
	xunShouInfo.LiuYi = ganzhi.GanName(findXunShou(leadZhu))
	scope := scopeMethods[ju.Method.Scope]
	school := schoolMethods[ju.Method.School]
	dingjuMethodName := dingjuMethodNames[ju.Method.Dingju]
	quarterRuleName := quarterRuleNames[ju.Method.QuarterRule]
	var tianQin *TianQinRule
	if ju.Method.School == SchoolZhuanPan {
		rule := tianQinRule
		tianQin = &rule
	}
	var quarter *ChartQuarter
	if ju.Quarter != nil {
		quarter = &ChartQuarter{
			Index:             ju.Quarter.Index,
			MinutesPerQuarter: ju.Quarter.Minutes,
			LeadPillarMode:    ju.Quarter.LeadPillarMode,
			DunSource:         ju.Quarter.DunSource,
			HourBoundary:      ju.Quarter.HourBoundary,
		}
	}
	return ChartMethod{
		Scope: scope.Scope, ScopeName: scope.Name,
		School: school.School, SchoolName: school.Name,
		DingjuMethod: string(ju.Method.Dingju), DingjuMethodName: dingjuMethodName,
		QuarterRule: string(ju.Method.QuarterRule), QuarterRuleName: quarterRuleName,
		BaseDingjuMethod:     string(ju.Method.BaseDingju),
		BaseDingjuMethodName: dingjuMethodNames[ju.Method.BaseDingju],
		MethodSource:         scope.MethodSource, DayBoundary: qimenDayBoundary,
		SpiritMode: school.SpiritMode, Geometry: school.Geometry,
		StarMode: school.StarMode, DoorMode: school.DoorMode,
		SpiritFlight: school.SpiritFlight,
		JieQi:        ju.JieQi, YongJuJieQi: ju.YongJuJieQi, Yuan: ju.Yuan,
		ZhiRunState:     ju.ZhiRunState,
		TianQin:         tianQin,
		LeadPillar:      newPillarInfo(leadZhu),
		LeadXunShou:     xunShouInfo,
		LeadXunShouGong: duty.Palace,
		DayFuTou:        newPillarInfo(fu),
		Quarter:         quarter,
	}
}

func newPillarInfo(zhu ganzhi.Zhu) PillarInfo {
	return PillarInfo{
		Name: ganzhi.GanName(zhu.Gan) + ganzhi.ZhiName(zhu.Zhi),
		Gan:  ganzhi.GanName(zhu.Gan),
		Zhi:  ganzhi.ZhiName(zhu.Zhi),
	}
}

func affectedVoidSymbols(p pan, dayPalace, hourPalace, leadPalace GongIndex) []AffectedSymbol {
	result := []AffectedSymbol{}
	for _, void := range p.KongWang {
		for _, focus := range focusSymbols(dayPalace, hourPalace, leadPalace) {
			if void.Gong == focus.palace {
				result = append(result, AffectedSymbol{
					Symbol: focus.name, Role: focus.role, Branch: void.Branch, Gong: void.Gong,
				})
			}
		}
	}
	return result
}

func affectedHorseSymbols(p pan, dayPalace, hourPalace, leadPalace GongIndex) []AffectedSymbol {
	result := []AffectedSymbol{}
	for _, focus := range focusSymbols(dayPalace, hourPalace, leadPalace) {
		if p.MaXing.Gong == focus.palace {
			result = append(result, AffectedSymbol{
				Symbol: focus.name, Role: focus.role, Branch: p.MaXing.Branch, Gong: p.MaXing.Gong,
			})
		}
	}
	return result
}

type focusSymbol struct {
	name   string
	role   string
	palace GongIndex
}

func focusSymbols(dayPalace, hourPalace, leadPalace GongIndex) []focusSymbol {
	return []focusSymbol{
		{name: "日干", role: "pillar", palace: dayPalace},
		{name: "时干", role: "pillar", palace: hourPalace},
		{name: "主柱", role: "lead", palace: leadPalace},
	}
}

func palaceWangShuai(p pan, yueZhi ganzhi.Zhi) []PalaceWangShuai {
	result := []PalaceWangShuai{}
	for i := range p.GongWei {
		wuxing := gongWuxingTable[i]
		wangShuai := ganzhi.WangShuaiOf(wuxing, yueZhi)
		result = append(result, PalaceWangShuai{
			Gong:          GongIndex(i + 1),
			Wuxing:        wuxing,
			WangShuai:     wangShuai,
			WangShuaiName: wangShuai.String(),
		})
	}
	return result
}

func palaceWangShuaiForPalace(p pan, yueZhi ganzhi.Zhi, palace GongIndex) PalaceWangShuai {
	if palace < GongKan || palace > GongLi {
		return PalaceWangShuai{}
	}
	for _, item := range palaceWangShuai(p, yueZhi) {
		if item.Gong == palace {
			return item
		}
	}
	return PalaceWangShuai{}
}

func leadPillarPalace(p pan) GongIndex {
	return findGanPalaceIdx(p, resolveJiaDunGan(p.LeadGan, p.LeadZhi))
}

// findGanPalaceIdx 求干（求测人日干/时干/用神干）的落宫。
// 奇门领域规则：用神落宫以天盘为核心判断依据（天盘主求测人/所问事之当下状态，
// 地盘反映过去/潜藏）。故优先取天盘干所在宫；无天盘则取地盘干所在宫。
// 甲遁：甲不露，遁于六仪（由 resolveJiaDunGan 先转成六仪）。
func findGanPalaceIdx(p pan, g ganzhi.Gan) GongIndex {
	for i := range p.GongWei {
		for _, item := range p.GongWei[i].TianPan {
			if item.Gan == g {
				return GongIndex(i + 1)
			}
		}
	}
	for i := range p.GongWei {
		if p.GongWei[i].DiPanGan == g {
			return GongIndex(i + 1)
		}
	}
	return 0
}

func findGanPalaceFacts(p pan, g ganzhi.Gan) []GanPalaceFact {
	facts := []GanPalaceFact{}
	for i := range p.GongWei {
		for _, item := range p.GongWei[i].TianPan {
			if item.Gan == g {
				facts = append(facts, GanPalaceFact{Palace: GongIndex(i + 1), Layer: "heaven"})
			}
		}
	}
	for i := range p.GongWei {
		if p.GongWei[i].DiPanGan == g {
			facts = append(facts, GanPalaceFact{Palace: GongIndex(i + 1), Layer: "earth"})
		}
	}
	return facts
}

func findStarPalace(p pan, star StarIndex) GongIndex {
	for i, palace := range p.GongWei {
		for _, item := range palace.TianPan {
			if item.Star == star {
				return GongIndex(i + 1)
			}
		}
	}
	return 0
}

func analyzeRiShiRelation(dayPalace, hourPalace GongIndex) RiShiRelation {
	relation := "same"
	if dayPalace > 0 && hourPalace > 0 {
		dayElement := palaceWuxing(dayPalace)
		hourElement := palaceWuxing(hourPalace)
		if dayElement == hourElement {
			relation = "same"
		} else if ganzhi.Sheng(dayElement, hourElement) {
			relation = "ri_generates_shi"
		} else if ganzhi.Sheng(hourElement, dayElement) {
			relation = "shi_generates_ri"
		} else if ganzhi.Ke(dayElement, hourElement) {
			relation = "ri_controls_shi"
		} else {
			relation = "shi_controls_ri"
		}
	}
	return RiShiRelation{
		Subject: "ri_gan_gong", Object: "shi_gan_gong",
		Relation: relation, Name: riShiRelations[relation],
	}
}

// findDoorPalaceIdx finds which gong a door resides in.
func findDoorPalaceIdx(p pan, d DoorIndex) GongIndex {
	for i, pg := range p.GongWei {
		if pg.Door == d {
			return GongIndex(i + 1)
		}
	}
	return 0
}
