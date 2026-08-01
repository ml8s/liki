package qimen

import (
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// YingQi holds timing prediction information.
type YingQi struct {
	MaXing   string `json:"ma_xing_dir"`   // 马星应期方向
	KongWang string `json:"kong_wang_fill"` // 空亡填实时机
	DutyMove string `json:"zhi_fu_shi_dong"`      // 值符值使推动
	Summary  string `json:"summary"`        // 综合应期判断
}

// computeYingQi computes 应期 timing from the pan.
func computeYingQi(pan pan) YingQi {
	var yq YingQi

	// 马星应期: 冲马星的地支年份/月份/时辰.
	mataBranches := palaceBranches(pan.MaXing)
	var chongStrs []string
	for _, mz := range mataBranches {
		chongStrs = append(chongStrs, ganzhi.ZhiName(chongBranch(mz)))
	}
	yq.MaXing = "马星在" + ganzhi.ZhiName(mataBranches[0]) + "，冲则动，应期在" + strings.Join(chongStrs, "、") + "（年月日时）"

	// 空亡填实: 空亡宫被填实时应事.
	var kwBranches []string
	for _, kw := range pan.KongWang {
		for _, z := range palaceBranches(kw) {
			kwBranches = append(kwBranches, ganzhi.ZhiName(z))
		}
	}
	if len(kwBranches) > 0 {
		yq.KongWang = "空亡在" + strings.Join(kwBranches, " ") + "，填实或冲空之时应事"
	}

	// 值符值使推动.
	yq.DutyMove = "值符" + pan.DutyStar.String() + "加" + pan.DutyDoor.String() + "，以时干" + ganzhi.GanName(pan.DriveGan) + "为应"

	yq.Summary = yq.MaXing + "；" + yq.KongWang

	return yq
}

// chongBranch returns the opposing branch (冲).
func chongBranch(z ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.ChongZhi(z)
}
