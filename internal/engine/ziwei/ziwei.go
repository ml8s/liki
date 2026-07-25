package ziwei

import "liki-engine/internal/engine/ganzhi"

// Type aliases from ganzhi.
type (
	Gan    = ganzhi.Gan
	Zhi    = ganzhi.Zhi
	Wuxing = ganzhi.Wuxing
)

const (
	Male   = ganzhi.Male
	Female = ganzhi.Female
)

// palaceIndex identifies one of 12 palaces (0=命宫 … 11=父母).
type palaceIndex int

// palace names in order 0..11 (逆时针 from 命宫).
var PalaceNames = [12]string{
	"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
	"迁移", "交友", "官禄", "田宅", "福德", "父母",
}

// starIndex enumerates all stars (main + minor).
type starIndex int

// 14 main stars.
const (
	ZiWei    starIndex = iota // 0  紫微
	TianJi                    // 1  天机
	TaiYang                   // 2  太阳
	WuQu                      // 3  武曲
	TianTong                  // 4  天同
	LianZhen                  // 5  廉贞
	TianFu                    // 6  天府
	TaiYin                    // 7  太阴
	TanLang                   // 8  贪狼
	JuMen                     // 9  巨门
	TianXiang                 // 10 天相
	TianLiang                 // 11 天梁
	QiSha                     // 12 七杀
	PoJun                     // 13 破军
)

// Minor stars (0.6).
const (
	LuCun     starIndex = iota + 14
	TianKui
	TianYue
	ZuoFu
	YouBi
	WenChang
	WenQu
	QingYang
	TuoLuo
	TianMa
	HuoXing
	LingXing
	DiKong
	DiJie
)

var starNames = map[starIndex]string{
	ZiWei: "紫微", TianJi: "天机", TaiYang: "太阳", WuQu: "武曲",
	TianTong: "天同", LianZhen: "廉贞", TianFu: "天府", TaiYin: "太阴",
	TanLang: "贪狼", JuMen: "巨门", TianXiang: "天相", TianLiang: "天梁",
	QiSha: "七杀", PoJun: "破军",
	LuCun: "禄存", TianKui: "天魁", TianYue: "天钺",
	ZuoFu: "左辅", YouBi: "右弼", WenChang: "文昌", WenQu: "文曲",
	QingYang: "擎羊", TuoLuo: "陀罗", TianMa: "天马",
	HuoXing: "火星", LingXing: "铃星", DiKong: "地空", DiJie: "地劫",
}

// starName returns the Chinese name of a star.
func starName(s starIndex) string { return starNames[s] }

// juShu is the five-element bureau number (2/3/4/5/6).
type juShu int

const (
	JuWater  juShu = 2 // 水二局
	JuWood  juShu = 3
	JuMetal  juShu = 4 // 金四局
	JuEarth  juShu = 5 // 土五局
	JuFire  juShu = 6
)

// juShuFromWuxing converts a five-element to its bureau number.
func juShuFromWuxing(w Wuxing) juShu {
	switch w {
	case ganzhi.WxShui:
		return JuWater
	case ganzhi.WxMu:
		return JuWood
	case ganzhi.WxJin:
		return JuMetal
	case ganzhi.WxTu:
		return JuEarth
	case ganzhi.WxHuo:
		return JuFire
	}
	return 0
}

var juShuNames = map[juShu]string{
	JuWater: "水二局",
	JuWood:  "木三局",
	JuMetal: "金四局",
	JuEarth: "土五局",
	JuFire:  "火六局",
}

// juShuName returns the Chinese name of a bureau.
func juShuName(j juShu) string { return juShuNames[j] }

// palace holds all computed data for one palace.
type palace struct {
	Index        palaceIndex `json:"index"`
	Name         string      `json:"name"`
	Gan          Gan         `json:"gan"`
	Zhi          Zhi         `json:"zhi"`
	IsBodyPalace bool        `json:"is_body_palace"`
	Stars        []starInfo  `json:"stars"`
	ZiweiStar    *starIndex  `json:"ziwei_star,omitempty"`
}

// starInfo is one star entry in a palace.
type starInfo struct {
	Star    starIndex `json:"star"`
	Name    string    `json:"name"`
	IsMajor bool      `json:"is_major"`
	SiHua   string    `json:"si_hua,omitempty"`     // "禄"/"权"/"科"/"忌" or empty
	Brightness string  `json:"brightness,omitempty"` // "庙"/"旺"/"利"/"平"/"陷"
}

// siHuaType is one of the four transformations.
type siHuaType string

const (
	HuaLu  siHuaType = "禄"
	HuaQuan siHuaType = "权"
	HuaKe  siHuaType = "科"
	HuaJi  siHuaType = "忌"
)

// Chart holds the complete ziwei chart.
type Chart struct {
	Palaces     [12]palace   `json:"palaces"`
	MingGong    palaceIndex  `json:"ming_gong"`
	ShenGong    palaceIndex  `json:"shen_gong"`
	JuShu       juShu        `json:"ju_shu"`
	JuShuName   string       `json:"ju_shu_name"`
	ZiweiPos    palaceIndex  `json:"ziwei_pos"`
	SiHua       siHuaResult  `json:"si_hua"`
	YearGan     Gan          `json:"year_gan"`
	HourZhi     Zhi          `json:"hour_zhi"`
	BirthYear   int          `json:"birth_year"`
	Gender      ganzhi.Gender       `json:"gender"`
	Patterns    []pattern    `json:"patterns,omitempty"`
}

// siHuaResult maps star → transformation.
type siHuaResult map[starIndex]siHuaType

// DaXianStep records one 10-year da-xian segment.
type DaXianStep struct {
	StartAge int         `json:"start_age"`
	EndAge   int         `json:"end_age"`
	Palace   palaceIndex `json:"palace"`
	Name     string      `json:"name"`
}

// LiuNian is the annual fate analysis.
type LiuNian struct {
	MingGong     palaceIndex              `json:"ming_gong"`
	MingGongName string                   `json:"ming_gong_name"`
	SiHua        siHuaResult              `json:"si_hua"`
	SiHuaPalace  map[starIndex]palaceIndex `json:"si_hua_palace"` // where each s化 star falls
	MinorStars   map[starIndex]int         `json:"minor_stars"` // zhi-1 values; convert via zhiToPalace
}

