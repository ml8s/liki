// Package liuyao provides 六爻 (纳甲筮法) computation.
//
// Types
//
//	Chart, Line, YaoType,
//	LiuQin, LiuShou, YongShen,
//	YongShenResult, FuShen,
//	ganzhi.WangShuai, DayRelation, TimingCandidate
//
// Constants
//
//	YaoType: LaoYin, ShaoYang, ShaoYin, LaoYang
//	LiuQin: QinFumu, QinXiongDi, QinGuanGui, QinQiCai, QinZiSun
//	LiuShou: ShouQingLong, ShouZhuQue, ShouGouChen, ShouTengShe, ShouBaiHu, ShouXuanWu
//	YongShen: YongFumu, YongXiongDi, YongGuanGui, YongQiCai, YongZiSun, YongShiYao
//	ganzhi: WSWang, WSXiang, WSXiu, WSQiu, WSSi
//
// Functions
//
//	SecureQigua() → Casting          安全随机起卦收据
//	ComputeChart(st, yongShen, yaos) → Chart  装卦 + 用神 + 旺衰 + 应期
package liuyao

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func yaosToInts(y [6]YaoType) [6]int {
	var out [6]int
	for i, v := range y {
		out[i] = int(v)
	}
	return out
}

// ComputeChart builds a full 六爻 chart from civil time (年/月柱),
// true-solar time (日/时柱), question type, and yaos.
func ComputeChart(gt time.Time, st time.Time, yongShen YongShen, yaos [6]int) Chart {
	return computeChart(tianwen.ComputeBazi(gt, st), yongShen, yaos)
}

// ComputeChartAt is the explicit-time entry for new callers. at must already
// be normalized to UTC; loc is the explicit IANA location.
func ComputeChartAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	yongShen YongShen,
	yaos [6]int,
) Chart {
	civil := tianwen.StandardLocalTime(at, loc)
	solar := tianwen.ApparentSolarCivil(at, loc, longitude)
	return ComputeChart(
		time.Time(civil),
		time.Time(solar),
		yongShen,
		yaos,
	)
}
