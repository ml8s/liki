package fengshui

import "liki-engine/internal/engine/ganzhi"

// Mountain24 represents one of the 24 directional mountains (二十四山).
// Each mountain spans 15° of the compass (360° / 24 = 15°).
type Mountain24 struct {
	Index    int           `json:"index"` // 0-23, starting from 子=0°
	Name     string        `json:"name"`  // e.g. "子", "癸", "丑"
	Angle    int           `json:"angle"` // degrees, 0-345 in 15° steps
	Element  ganzhi.Wuxing `json:"wuxing"`
	YinYang  string        `json:"yin_yang"`  // "阳" or "阴"
	Trigram  string        `json:"trigram"`   // Bagua trigram name
	YuanLong string        `json:"yuan_long"` // "天元龙"/"地元龙"/"人元龙"
}

// Mountains24Table holds all 24 mountains in order, starting from 子=0°(North).
var Mountains24Table = [24]Mountain24{
	{0, "子", 0, ganzhi.WxShui, "阴", "坎", "天元龙"},
	{1, "癸", 15, ganzhi.WxShui, "阴", "坎", "人元龙"},
	{2, "丑", 30, ganzhi.WxTu, "阴", "艮", "地元龙"},
	{3, "艮", 45, ganzhi.WxTu, "阳", "艮", "天元龙"},
	{4, "寅", 60, ganzhi.WxMu, "阳", "艮", "人元龙"},
	{5, "甲", 75, ganzhi.WxMu, "阳", "震", "地元龙"},
	{6, "卯", 90, ganzhi.WxMu, "阴", "震", "天元龙"},
	{7, "乙", 105, ganzhi.WxMu, "阴", "震", "人元龙"},
	{8, "辰", 120, ganzhi.WxTu, "阴", "巽", "地元龙"},
	{9, "巽", 135, ganzhi.WxMu, "阳", "巽", "天元龙"},
	{10, "巳", 150, ganzhi.WxHuo, "阳", "巽", "人元龙"},
	{11, "丙", 165, ganzhi.WxHuo, "阳", "离", "地元龙"},
	{12, "午", 180, ganzhi.WxHuo, "阴", "离", "天元龙"},
	{13, "丁", 195, ganzhi.WxHuo, "阴", "离", "人元龙"},
	{14, "未", 210, ganzhi.WxTu, "阴", "坤", "地元龙"},
	{15, "坤", 225, ganzhi.WxTu, "阳", "坤", "天元龙"},
	{16, "申", 240, ganzhi.WxJin, "阳", "坤", "人元龙"},
	{17, "庚", 255, ganzhi.WxJin, "阳", "兑", "地元龙"},
	{18, "酉", 270, ganzhi.WxJin, "阴", "兑", "天元龙"},
	{19, "辛", 285, ganzhi.WxJin, "阴", "兑", "人元龙"},
	{20, "戌", 300, ganzhi.WxTu, "阴", "乾", "地元龙"},
	{21, "乾", 315, ganzhi.WxJin, "阳", "乾", "天元龙"},
	{22, "亥", 330, ganzhi.WxShui, "阳", "乾", "人元龙"},
	{23, "壬", 345, ganzhi.WxShui, "阳", "坎", "地元龙"},
}
