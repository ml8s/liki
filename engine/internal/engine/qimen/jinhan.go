package qimen

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// JinhanChartMethod projects the closed method and the domain rules that are
// specific to the Golden Mirror daily chart.
type JinhanChartMethod struct {
	Scope            Scope        `json:"scope"`
	ScopeName        string       `json:"scope_name"`
	School           School       `json:"school"`
	SchoolName       string       `json:"school_name"`
	DingjuMethod     DingjuMethod `json:"dingju_method"`
	DingjuMethodName string       `json:"dingju_method_name"`
	MethodSource     string       `json:"method_source"`
	DayBoundary      string       `json:"day_boundary"`
	Geometry         string       `json:"geometry"`
	StarMode         string       `json:"star_mode"`
	DoorMode         string       `json:"door_mode"`
	SpiritMode       string       `json:"spirit_mode"`
	SpiritFlight     string       `json:"spirit_flight"`
	DunSource        string       `json:"dun_source"`
	JieQi            string       `json:"jie_qi"`
	Basis            string       `json:"basis"`
}

// JinhanPalace is one Golden Mirror palace. It deliberately does not reuse the
// standard three-odd/six-instrument palace fields.
type JinhanPalace struct {
	Gong       PalaceIdentity `json:"gong"`
	Xing       string         `json:"xing"`
	Men        string         `json:"men,omitempty"`
	MenPresent bool           `json:"men_present"`
}

// JinhanDaySpirit maps one branch to the twelve-spirit table selected by the
// current day stem.
type JinhanDaySpirit struct {
	Zhi  ganzhi.Zhi `json:"zhi"`
	Shen string     `json:"shen"`
}

// JinhanPan is the complete Golden Mirror daily plate.
type JinhanPan struct {
	YinDun     bool                `json:"yin_dun"`
	RiGan      ganzhi.Gan          `json:"ri_gan"`
	RiZhi      ganzhi.Zhi          `json:"ri_zhi"`
	GongWei    [9]JinhanPalace     `json:"gong_wei"`
	DaySpirits [12]JinhanDaySpirit `json:"day_spirits"`
}

// JinhanChart is the public result of qimen.chart for school=jinhan_yujing.
type JinhanChart struct {
	Method JinhanChartMethod `json:"method"`
	Pan    JinhanPan         `json:"pan"`
}

// ComputeJinhanChart computes the Golden Mirror daily chart from true solar
// time. The method dimensions are closed by the catalog and do not need a
// dingju selection.
func ComputeJinhanChart(st tianwen.SolarTime) JinhanChart {
	bz, chartTime := chartInputs(st)
	term := currentSolarTermInfo(chartTime)
	yinDun := term.Index >= 12
	chart := JinhanChart{Method: buildJinhanMethod(term)}
	chart.Pan.YinDun = yinDun
	chart.Pan.RiGan, chart.Pan.RiZhi = bz.Ri.Gan, bz.Ri.Zhi
	for palace := GongKan; palace <= GongLi; palace++ {
		chart.Pan.GongWei[int(palace)-1] = JinhanPalace{
			Gong: palaceIdentity(palace), MenPresent: false,
		}
	}
	layoutJinhanStars(&chart.Pan, bz.Ri, yinDun)
	layoutJinhanDoors(&chart.Pan, bz.Ri, yinDun)
	layoutJinhanDaySpirits(&chart.Pan, bz.Ri)
	return chart
}

func buildJinhanMethod(term solarTermInfo) JinhanChartMethod {
	scope := scopeMethods[ScopeDay]
	school := schoolMethods[SchoolJinhanYuJing]
	return JinhanChartMethod{
		Scope: ScopeDay, ScopeName: scope.Name,
		School: SchoolJinhanYuJing, SchoolName: school.Name,
		DingjuMethod: DingjuNone, DingjuMethodName: dingjuMethodNames[DingjuNone],
		MethodSource: jinhanMethodSource, DayBoundary: qimenDayBoundary,
		Geometry: school.Geometry, StarMode: school.StarMode,
		DoorMode: school.DoorMode, SpiritMode: school.SpiritMode,
		SpiritFlight: school.SpiritFlight, DunSource: jinhanDunSource,
		JieQi: term.Name, Basis: jinhanBasis,
	}
}

func layoutJinhanStars(plate *JinhanPan, dayZhu ganzhi.Zhu, yinDun bool) {
	dayIndex := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)
	anchors := jinhanYangStarStarts
	flight := jinhanYangStarFlight
	if yinDun {
		anchors, flight = jinhanYinStarStarts, jinhanYinStarFlight
	}
	start := anchors[floorMod(dayIndex, len(anchors))]
	sequence := rotatePalaces(flight, start)
	for i, star := range jinhanStarOrder {
		plate.GongWei[int(sequence[i])-1].Xing = star
	}
}

func layoutJinhanDoors(plate *JinhanPan, dayZhu ganzhi.Zhu, yinDun bool) {
	dayIndex := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)
	group := floorMod(dayIndex/jinhanDoorDays, len(jinhanDoorStarts))
	startIndex := group
	if yinDun {
		startIndex = len(jinhanDoorStarts) - 1 - group
	}
	start := jinhanDoorStarts[startIndex]
	ring := jinhanDoorRing
	if yinDun {
		ring = reversePalaces(ring)
	}
	sequence := rotatePalaces(ring, start)
	for i, door := range jinhanDoorOrder {
		palace := sequence[i]
		plate.GongWei[int(palace)-1].Men = door.String()
		plate.GongWei[int(palace)-1].MenPresent = true
	}
}

func layoutJinhanDaySpirits(plate *JinhanPan, dayZhu ganzhi.Zhu) {
	spirits := jinhanDaySpirits[dayZhu.Gan]
	for branch := ganzhi.ZhiZi; branch <= ganzhi.ZhiHai; branch++ {
		plate.DaySpirits[int(branch)-1] = JinhanDaySpirit{
			Zhi: branch, Shen: spirits[int(branch)-1],
		}
	}
}

func rotatePalaces(sequence []GongIndex, start GongIndex) []GongIndex {
	startIndex := -1
	for i, palace := range sequence {
		if palace == start {
			startIndex = i
			break
		}
	}
	if startIndex < 0 {
		return nil
	}
	result := make([]GongIndex, len(sequence))
	for i := range sequence {
		result[i] = sequence[floorMod(startIndex+i, len(sequence))]
	}
	return result
}

func reversePalaces(sequence []GongIndex) []GongIndex {
	result := make([]GongIndex, len(sequence))
	for i, palace := range sequence {
		result[len(sequence)-1-i] = palace
	}
	return result
}
