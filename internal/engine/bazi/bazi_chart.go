package bazi

import (
	"encoding/json"

	"liki-engine/internal/engine/ganzhi"
)

// --- per-pillar data ---

type cangGanOut struct {
	Main  ganzhi.Gan  `json:"main"`
	Mid   *ganzhi.Gan `json:"mid"`
	Minor *ganzhi.Gan `json:"minor"`
}

type zhuInfo struct {
	ganzhi.Zhu
	NaYin string `json:"na_yin"`
}

func (z zhuInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gan   string `json:"gan"`
		Zhi   string `json:"zhi"`
		NaYin string `json:"na_yin"`
	}{Gan: ganzhi.GanName(z.Gan), Zhi: ganzhi.ZhiName(z.Zhi), NaYin: z.NaYin})
}

type fullZhuInfo struct {
	ganzhi.Zhu
	NaYin      string             `json:"na_yin"`
	CangGan    cangGanOut         `json:"cang_gan"`
	ShiShens   []shiShenEntry     `json:"shi_shens"`
	ChangSheng []changShengEntry  `json:"chang_sheng"`
	ShenSha    []shenShaEntry     `json:"shen_sha"`
	IsVoid     bool               `json:"is_void"`
	IsSelfHe   bool               `json:"is_self_he"`
	IsKuiGang  bool               `json:"is_kui_gang"`
	SelfHeName string             `json:"self_he_name"`
}

func (z fullZhuInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gan        string             `json:"gan"`
		Zhi        string             `json:"zhi"`
		NaYin      string             `json:"na_yin"`
		CangGan    cangGanOut         `json:"cang_gan"`
		ShiShens   []shiShenEntry     `json:"shi_shens"`
		ChangSheng []changShengEntry  `json:"chang_sheng"`
		ShenSha    []shenShaEntry     `json:"shen_sha"`
		IsVoid     bool               `json:"is_void"`
		IsSelfHe   bool               `json:"is_self_he"`
		IsKuiGang  bool               `json:"is_kui_gang"`
		SelfHeName string             `json:"self_he_name"`
	}{Gan: ganzhi.GanName(z.Gan), Zhi: ganzhi.ZhiName(z.Zhi), NaYin: z.NaYin, CangGan: z.CangGan, ShiShens: z.ShiShens, ChangSheng: z.ChangSheng, ShenSha: z.ShenSha, IsVoid: z.IsVoid, IsSelfHe: z.IsSelfHe, IsKuiGang: z.IsKuiGang, SelfHeName: z.SelfHeName})
}
type shiShenEntry struct {
	ShiShen ganzhi.ShiShen `json:"shi_shen"`
	Name    string         `json:"name"`
	Source  string         `json:"source"`
	Gan     ganzhi.Gan     `json:"gan"`
}
type changShengEntry struct {
	Stage string     `json:"stage"`
	Gan   ganzhi.Gan `json:"gan"`
}
// Ten god source constants.
const (
	sourceGan    = "stem"
	sourceMainQi = "main_qi"
	sourceMidQi  = "mid_qi"
	sourceMinQi  = "minor_qi"
)

// --- chart ---

// Chart holds a complete bazi chart: four pillars, dayun, and gender.
type Chart struct {
	Nian   zhuInfo       `json:"nian"`
	Yue    zhuInfo       `json:"yue"`
	Ri     zhuInfo       `json:"ri"`
	Shi    zhuInfo       `json:"shi"`
	DaYun  *DaYun        `json:"da_yun"`
	Gender ganzhi.Gender `json:"gender"`
}

// FullChart is the expanded bazi chart with all fields (十神/藏干/神煞/长生/空亡...).
// Use bazi.fullchart to obtain it from a lean Chart.
type FullChart struct {
	Nian   fullZhuInfo   `json:"nian"`
	Yue    fullZhuInfo   `json:"yue"`
	Ri     fullZhuInfo   `json:"ri"`
	Shi    fullZhuInfo   `json:"shi"`
	DaYun  *DaYun        `json:"da_yun"`
	Gender ganzhi.Gender `json:"gender"`
}

func (c FullChart) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: c.Nian.Zhu,
		Yue:  c.Yue.Zhu,
		Ri:   c.Ri.Zhu,
		Shi:  c.Shi.Zhu,
	}
}
func (c FullChart) NaYinArray() [4]string {
	return [4]string{c.Nian.NaYin, c.Yue.NaYin, c.Ri.NaYin, c.Shi.NaYin}
}
func (c FullChart) CangGanArray() [4]cangGanOut {
	return [4]cangGanOut{c.Nian.CangGan, c.Yue.CangGan, c.Ri.CangGan, c.Shi.CangGan}
}

var zhuLabels = [4]string{"nian", "yue", "ri", "shi"}

func (c Chart) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: c.Nian.Zhu,
		Yue:  c.Yue.Zhu,
		Ri:   c.Ri.Zhu,
		Shi:  c.Shi.Zhu,
	}
}
func (c Chart) NaYinArray() [4]string {
	return [4]string{c.Nian.NaYin, c.Yue.NaYin, c.Ri.NaYin, c.Shi.NaYin}
}
