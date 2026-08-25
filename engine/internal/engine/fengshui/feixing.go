package fengshui

import "liki-engine/internal/engine/ganzhi"

// FlyingStar holds a single purple-white flying star (紫白飞星).
type FlyingStar struct {
	Number     int           `json:"number"`
	Color      string        `json:"color"`
	Name       string        `json:"name"`
	Element    ganzhi.Wuxing `json:"wuxing"`
	Auspicious bool          `json:"auspicious"`
}

// StarTable maps star number (1-9) to its attributes.
var StarTable = [10]FlyingStar{
	{},
	{1, "白", "一白贪狼", ganzhi.WxShui, true},
	{2, "黑", "二黑巨门", ganzhi.WxTu, false},
	{3, "碧", "三碧禄存", ganzhi.WxMu, false},
	{4, "绿", "四绿文曲", ganzhi.WxMu, true},
	{5, "黄", "五黄廉贞", ganzhi.WxTu, false},
	{6, "白", "六白武曲", ganzhi.WxJin, true},
	{7, "赤", "七赤破军", ganzhi.WxJin, false},
	{8, "白", "八白左辅", ganzhi.WxTu, true},
	{9, "紫", "九紫右弼", ganzhi.WxHuo, true},
}

// StarByNumber returns the flying star for a given number (1-9).
func StarByNumber(n int) FlyingStar {
	if n >= 1 && n <= 9 {
		return StarTable[n]
	}
	return FlyingStar{}
}

// LuoshuFlyOrder is the standard luoshu flying order (excluding center).
var LuoshuFlyOrder = [8]int{6, 7, 8, 9, 1, 2, 3, 4}
