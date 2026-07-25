package fengshui

import "liki-engine/internal/engine/ganzhi"

// Palace represents one of the nine palaces in the 洛书 (Luoshu) grid.
type Palace struct {
	Number    int           `json:"number"`
	Name      string        `json:"name"`
	Direction string        `json:"direction"`
	Element   ganzhi.Wuxing `json:"wuxing"`
}

// PalaceTable holds all nine palaces indexed by palace number (1-9).
var PalaceTable = [10]Palace{
	{},
	{1, "坎", "北", ganzhi.WxShui},
	{2, "坤", "西南", ganzhi.WxTu},
	{3, "震", "东", ganzhi.WxMu},
	{4, "巽", "东南", ganzhi.WxMu},
	{5, "中", "中", ganzhi.WxTu},
	{6, "乾", "西北", ganzhi.WxJin},
	{7, "兑", "西", ganzhi.WxJin},
	{8, "艮", "东北", ganzhi.WxTu},
	{9, "离", "南", ganzhi.WxHuo},
}

// PalaceByNumber returns the palace for a given number (1-9).
func PalaceByNumber(n int) Palace {
	if n >= 1 && n <= 9 {
		return PalaceTable[n]
	}
	return Palace{}
}
