package ganzhi

import "fmt"

// Gender represents male or female, used in BaZi calculations.
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

// Gan is a heavenly stem (天干). 1=甲 .. 10=癸.
type Gan int

const (
	GanJia  Gan = 1  // 甲
	GanYi   Gan = 2  // 乙
	GanBing Gan = 3  // 丙
	GanDing Gan = 4  // 丁
	GanWu   Gan = 5  // 戊
	GanJi   Gan = 6  // 己
	GanGeng Gan = 7  // 庚
	GanXin  Gan = 8  // 辛
	GanRen  Gan = 9  // 壬
	GanGui  Gan = 10 // 癸
)

var ganWuxingBiao = [11]Wuxing{0, WxMu, WxMu, WxHuo, WxHuo, WxTu, WxTu, WxJin, WxJin, WxShui, WxShui}
var ganYinYangBiao = [11]bool{false, true, false, true, false, true, false, true, false, true, false}

// Zhi is an earthly branch (地支). 1=子 .. 12=亥.
type Zhi int

const (
	ZhiZi   Zhi = 1  // 子
	ZhiChou Zhi = 2  // 丑
	ZhiYin  Zhi = 3  // 寅
	ZhiMao  Zhi = 4  // 卯
	ZhiChen Zhi = 5  // 辰
	ZhiSi   Zhi = 6  // 巳
	ZhiWu   Zhi = 7  // 午
	ZhiWei  Zhi = 8  // 未
	ZhiShen Zhi = 9  // 申
	ZhiYou  Zhi = 10 // 酉
	ZhiXu   Zhi = 11 // 戌
	ZhiHai  Zhi = 12 // 亥
)

var zhiWuxingBiao = [13]Wuxing{0, WxShui, WxTu, WxMu, WxMu, WxTu, WxHuo, WxHuo, WxTu, WxJin, WxJin, WxTu, WxShui}

// Wuxing is the five-phase element (五行). 1=木 2=火 3=土 4=金 5=水.
type Wuxing int

const (
	WxMu   Wuxing = 1
	WxHuo  Wuxing = 2
	WxTu   Wuxing = 3
	WxJin  Wuxing = 4
	WxShui Wuxing = 5
)

// YinYang distinguishes yin from yang.
type YinYang bool

const (
	Yin  YinYang = false
	Yang YinYang = true
)

// Zhu is one heavenly-stem / earthly-branch pair (一柱).
type Zhu struct {
	Gan Gan `json:"gan"`
	Zhi Zhi `json:"zhi"`
}

// Bazi holds the four named pillars of a BaZi chart (八字).
type Bazi struct {
	Nian Zhu `json:"nian"`
	Yue  Zhu `json:"yue"`
	Ri   Zhu `json:"ri"`
	Shi  Zhu `json:"shi"`
}

// Slice returns the four pillars as an indexable array.
func (bz Bazi) Slice() [4]Zhu {
	return [4]Zhu{bz.Nian, bz.Yue, bz.Ri, bz.Shi}
}

// Validate checks that all four pillars have valid gan (1-10) and zhi (1-12) values.
func (bz Bazi) Validate() error {
	for i, p := range bz.Slice() {
		if p.Gan < 1 || p.Gan > 10 {
			return fmt.Errorf("ganzhi: bazi[%d].gan must be 1-10", i)
		}
		if p.Zhi < 1 || p.Zhi > 12 {
			return fmt.Errorf("ganzhi: bazi[%d].zhi must be 1-12", i)
		}
	}
	return nil
}

// -- primitive lookups --

// GanWuxing returns the five-phase element for a heavenly stem.
func GanWuxing(g Gan) Wuxing { return ganWuxingBiao[g] }

// GanYinYang returns the yin-yang classification for a heavenly stem.
func GanYinYang(g Gan) YinYang { return YinYang(ganYinYangBiao[g]) }

// ZhiWuxing returns the five-phase element for an earthly branch.
func ZhiWuxing(z Zhi) Wuxing { return zhiWuxingBiao[z] }

// -- five-phase cycle --

// Sheng returns true if wuxing `from` nourishes wuxing `to`.
func Sheng(from, to Wuxing) bool {
	return (from == WxMu && to == WxHuo) ||
		(from == WxHuo && to == WxTu) ||
		(from == WxTu && to == WxJin) ||
		(from == WxJin && to == WxShui) ||
		(from == WxShui && to == WxMu)
}

// Ke returns true if wuxing `from` controls wuxing `to`.
func Ke(from, to Wuxing) bool {
	return (from == WxMu && to == WxTu) ||
		(from == WxTu && to == WxShui) ||
		(from == WxShui && to == WxHuo) ||
		(from == WxHuo && to == WxJin) ||
		(from == WxJin && to == WxMu)
}

// -- sixty-cycle --

// SixtyCycleIndex returns the 0-based index in [0,59] for a given gan+zhi.
func SixtyCycleIndex(gan Gan, zhi Zhi) int {
	idx := (int(gan)*6 - int(zhi)*5 - 1) % 60
	if idx < 0 {
		idx += 60
	}
	return idx
}

// SixtyToZhu converts a 0-based sixty-cycle index to a Zhu.
func SixtyToZhu(idx int) Zhu {
	g := Gan((idx % 10) + 1)
	z := Zhi((idx % 12) + 1)
	return Zhu{Gan: g, Zhi: z}
}

// -- hours --

// HourRanges maps each earthly branch to its two-hour range.
var HourRanges = [12]string{
	"23:00-01:00", "01:00-03:00", "03:00-05:00", "05:00-07:00",
	"07:00-09:00", "09:00-11:00", "11:00-13:00", "13:00-15:00",
	"15:00-17:00", "17:00-19:00", "19:00-21:00", "21:00-23:00",
}

// -- name arrays (referenced by symbols.go) --

// ganNames maps Gan to Chinese character.
var ganNames = [11]string{"", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

var zhiNames = [13]string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
