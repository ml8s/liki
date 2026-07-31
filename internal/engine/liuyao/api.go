// Package liuyao provides 六爻 (纳甲筮法) computation.
//
// Types
//   Chart, Line, YaoType,
//   LiuQin, LiuShou, YongShen,
//   YongShenResult, FuShen,
//   ganzhi.WangShuai, DayRelation, YingQi
//
// Constants
//   YaoType: LaoYin, ShaoYang, ShaoYin, LaoYang
//   LiuQin: QinFumu, QinXiongDi, QinGuanGui, QinQiCai, QinZiSun
//   LiuShou: ShouQingLong, ShouZhuQue, ShouGouChen, ShouTengShe, ShouBaiHu, ShouXuanWu
//   YongShen: YongFumu, YongXiongDi, YongGuanGui, YongQiCai, YongZiSun, YongShiYao
//   ganzhi: WSWang, WSXiang, WSXiu, WSQiu, WSSi
//
// Functions
//   Qigua() → QiguaResult           纯起卦（三枚铜钱摇六次）
//   ComputeChart(st, yongShen, yaos) → Chart  装卦 + 用神 + 旺衰 + 应期
package liuyao

import (
	"math/rand"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// QiguaResult is the bare output of a coin-toss hexagram draw.
type QiguaResult struct {
	Yaos    [6]int `json:"yaos"`     // 初爻到上爻，6/7/8/9
	DongYao []int  `json:"dong_yao,omitempty"` // 动爻位置 1-6（无动爻时省略）
}

// Qigua simulates three coins tossed six times.
func Qigua() QiguaResult {
	yaos := shakeCoins(rand.New(rand.NewSource(time.Now().UnixNano())))
	return QiguaResult{
		Yaos:    yaosToInts(yaos),
		DongYao: dongYao(yaos),
	}
}

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
