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

import "liki-engine/internal/engine/tianwen"

func yaosToInts(y [6]YaoType) [6]int {
	var out [6]int
	for i, v := range y {
		out[i] = int(v)
	}
	return out
}

// ComputeChart builds a full 六爻 chart from solar time, question type, and yaos.
func ComputeChart(st tianwen.SolarTime, yongShen YongShen, yaos [6]int) Chart {
	return computeChart(tianwen.ComputeBazi(st), yongShen, yaos)
}
