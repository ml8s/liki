package fengshui

import "liki-engine/internal/engine/ganzhi"

// FlyingStar holds a single purple-white flying star (紫白飞星).
type FlyingStar struct {
	Number  int           `json:"number"`
	Color   string        `json:"color"`
	Name    string        `json:"name"`
	Element ganzhi.Wuxing `json:"wuxing"`
}

// StarTable maps star number (1-9) to its attributes.
var StarTable = [10]FlyingStar{
	{},
	{Number: 1, Color: "白", Name: "一白贪狼", Element: ganzhi.WxShui},
	{Number: 2, Color: "黑", Name: "二黑巨门", Element: ganzhi.WxTu},
	{Number: 3, Color: "碧", Name: "三碧禄存", Element: ganzhi.WxMu},
	{Number: 4, Color: "绿", Name: "四绿文曲", Element: ganzhi.WxMu},
	{Number: 5, Color: "黄", Name: "五黄廉贞", Element: ganzhi.WxTu},
	{Number: 6, Color: "白", Name: "六白武曲", Element: ganzhi.WxJin},
	{Number: 7, Color: "赤", Name: "七赤破军", Element: ganzhi.WxJin},
	{Number: 8, Color: "白", Name: "八白左辅", Element: ganzhi.WxTu},
	{Number: 9, Color: "紫", Name: "九紫右弼", Element: ganzhi.WxHuo},
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
