package ziwei

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

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

func (p palaceIndex) MarshalJSON() ([]byte, error) {
	// 宫名 = palaceLabels（palaceIndex 定义的固定标签），不依赖命盘
	if int(p) < 0 || int(p) >= 12 {
		return json.Marshal("")
	}
	return json.Marshal(palaceLabels[p])
}

func (p *palaceIndex) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("palaceIndex must be a string (e.g. \"命宫\"), got %s", string(data))
	}
	for i, name := range palaceLabels {
		if name == s {
			*p = palaceIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown palace name: %q", s)
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
	HongLuan
	TianXi
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
	HongLuan: "红鸾", TianXi: "天喜",
}

// starName returns the Chinese name of a star.
func starName(s starIndex) string { return starNames[s] }

func (s starIndex) MarshalJSON() ([]byte, error) {
	name, ok := starNames[s]
	if !ok {
		return json.Marshal("")
	}
	return json.Marshal(name)
}

func (s *starIndex) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("starIndex must be a string, got %s", string(data))
	}
	return s.fromName(name)
}

// MarshalText 使 starIndex 作为 map key 时序列化为星名（Go json 的 map key 只认 TextMarshaler）。
func (s starIndex) MarshalText() ([]byte, error) {
	name, ok := starNames[s]
	if !ok {
		return []byte(""), nil
	}
	return []byte(name), nil
}

// UnmarshalText 使 starIndex 作为 map key 时可从星名反序列化。
func (s *starIndex) UnmarshalText(b []byte) error {
	return s.fromName(string(b))
}

func (s *starIndex) fromName(name string) error {
	for k, n := range starNames {
		if n == name {
			*s = k
			return nil
		}
	}
	return fmt.Errorf("unknown star: %q", name)
}

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

func (j juShu) MarshalJSON() ([]byte, error) {
	s, ok := juShuNames[j]
	if !ok {
		return json.Marshal("")
	}
	return json.Marshal(s)
}

func (j *juShu) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("juShu must be a string (e.g. \"水二局\"), got %s", string(data))
	}
	if s == "" {
		*j = 0
		return nil
	}
	for k, name := range juShuNames {
		if name == s {
			*j = k
			return nil
		}
	}
	return fmt.Errorf("unknown ju: %q", s)
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
	IsBodyPalace bool        `json:"is_shen_gong"`
	IsYuanGong   bool        `json:"is_yuan_gong,omitempty"`
	Stars        []starInfo  `json:"xing_yao"`
	ZiweiStar    *starIndex  `json:"ziwei_star,omitempty"`
	Ages         []int       `json:"ages,omitempty"`
	ChangSheng   string      `json:"chang_sheng,omitempty"`
	BoShi        string      `json:"bo_shi,omitempty"`
	JiangQian    string      `json:"jiang_qian,omitempty"`
	SuiQian      string      `json:"sui_qian,omitempty"`
	ZaYao       []string    `json:"za_yao,omitempty"`
}

// starInfo is one star entry in a palace.
type starInfo struct {
	Star    starIndex `json:"xing"`
	Name    string    `json:"name"`
	IsMajor bool      `json:"is_zhu_xing"`
	SiHua   string    `json:"si_hua,omitempty"`     // "禄"/"权"/"科"/"忌" or empty
	Brightness string  `json:"liang_du,omitempty"` // "庙"/"旺"/"利"/"平"/"陷"
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
	Palaces     [12]palace   `json:"gong_wei"`
	MingGong    palaceIndex  `json:"ming_gong"`
	ShenGong    palaceIndex  `json:"shen_gong"`
	JuShu       juShu        `json:"ju_shu"`
	JuShuName   string       `json:"ju_shu_name"`
	ZiweiPos    palaceIndex  `json:"ziwei_pos"`
	SiHua       siHuaResult  `json:"si_hua"`
	NianGan     Gan                 `json:"nian_gan"`
	NianZhi     Zhi                 `json:"nian_zhi,omitempty"`
	ShiZhi      Zhi                 `json:"shi_zhi"`
	BirthYear   int                 `json:"birth_year"`
	LunarMonth  int                 `json:"lunar_month,omitempty"`
	LunarDay    int                 `json:"lunar_day,omitempty"`
	BirthLunarMonth int                 `json:"birth_lunar_month,omitempty"`
	BirthIsLeap    bool                 `json:"birth_is_leap,omitempty"`
	Gender      ganzhi.Gender       `json:"gender"`
	MingZhu     string              `json:"ming_zhu,omitempty"`
	ShenZhu     string              `json:"shen_zhu,omitempty"`
	Patterns    []pattern           `json:"patterns,omitempty"`
}

// siHuaResult maps star → transformation.
type siHuaResult map[starIndex]siHuaType

// DaXianStep records one 10-year da-xian segment.
type DaXianStep struct {
	QiSui int         `json:"qi_sui"`
	ZhiSui int        `json:"zhi_sui"`
	Palace palaceIndex `json:"gong"`
	Name   string      `json:"name"`
}

// LiuNian is the annual fate analysis.
type LiuNian struct {
	MingGong     palaceIndex              `json:"ming_gong"`
	MingGongName string                   `json:"ming_gong_name"` // 流年命宫名
	Zhi          Zhi                      `json:"zhi"`            // 流年地支
	SiHua        siHuaResult              `json:"si_hua"`
	SiHuaPalace  map[starIndex]palaceIndex `json:"si_hua_palace"`
	FuXing       map[starIndex]Zhi         `json:"fu_xing"`
	Palaces      [12]flowPalace           `json:"gong_wei"`       // 流年盘（地支坐标 12 宫）
}

