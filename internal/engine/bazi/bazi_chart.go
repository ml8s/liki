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
	NaYin      string            `json:"na_yin"`
	CangGan    cangGanOut        `json:"cang_gan"`
	ShiShens   []shiShenEntry    `json:"shi_shens"`
	ChangSheng []changShengEntry `json:"chang_sheng"`
	ShenSha    []shenShaEntry    `json:"shen_sha"`
	IsVoid     bool              `json:"is_void"`
	IsSelfHe   bool              `json:"is_self_he"`
	IsKuiGang  bool              `json:"is_kui_gang"`
	SelfHeName string            `json:"self_he_name"`
}

func (z fullZhuInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Gan        string            `json:"gan"`
		Zhi        string            `json:"zhi"`
		NaYin      string            `json:"na_yin"`
		CangGan    cangGanOut        `json:"cang_gan"`
		ShiShens   []shiShenEntry    `json:"shi_shens"`
		ChangSheng []changShengEntry `json:"chang_sheng"`
		ShenSha    []shenShaEntry    `json:"shen_sha"`
		IsVoid     bool              `json:"is_void"`
		IsSelfHe   bool              `json:"is_self_he"`
		IsKuiGang  bool              `json:"is_kui_gang"`
		SelfHeName string            `json:"self_he_name"`
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

	// 出生公历年份（bazi.chart 起附）。供 bazi.liunian/liuri 按查询年份定位当年大运。
	BirthYear int `json:"birth_year"`

	// 子时换日规则说明（信息字段——引擎按 lunar 约定：晚子时(23:00-24:00)日柱不变、
	// 时柱按次日日干起；供前端/用户核验换日口径，非计算输入）。
	ZiShiRule string `json:"zi_shi_rule,omitempty"`
}

// Chart（纯排盘）不含用神——用神三派属完整命盘（bazi.fullchart 承载，
// 见 FullChart.YongShen）。chart 参与的运算（排盘/大运/流年派生）均不依赖用神。

// FullChart is the expanded bazi chart with all fields (十神/藏干/神煞/长生/空亡...).
// Use bazi.fullchart to obtain it from a lean Chart.
type FullChart struct {
	Nian   fullZhuInfo   `json:"nian"`
	Yue    fullZhuInfo   `json:"yue"`
	Ri     fullZhuInfo   `json:"ri"`
	Shi    fullZhuInfo   `json:"shi"`
	DaYun  *DaYun        `json:"da_yun"`
	Gender ganzhi.Gender `json:"gender"`

	// 出生公历年份（透传自 lean Chart，供 liunian/liuri 按年定位大运）。
	BirthYear int `json:"birth_year"`

	// 用神三派（透传自 lean Chart）。
	YongShen YongShenResult `json:"yong_shen"`

	// 补充信息（原 bazi.chart_extra）
	SanYuan    SanYuan             `json:"san_yuan"`
	GongJia    []GongJia           `json:"gong_jia,omitempty"`
	NayinRel   []NayinRelEntry     `json:"nayin_rel"`
	ChangSheng [12]ChangShengStage `json:"chang_sheng"`
	SanQiName  string              `json:"san_qi_name,omitempty"`

	// 合会冲刑（原 bazi.hehui）
	GanHe    []GanHePair   `json:"gan_he"`
	ZhiLiuHe []ZhiPairRel  `json:"zhi_liu_he"`
	SanHe    []TripleGroup `json:"san_he"`
	SanHui   []TripleGroup `json:"san_hui"`
	LiuChong []ZhiPairRel  `json:"liu_chong"`
	LiuHai   []ZhiPairRel  `json:"liu_hai"`
	LiuXing  []ZhiPairRel  `json:"liu_xing"`

	// 旬空（按日柱所居之旬），如 "午未"（甲申旬空午未）。空亡柱另见各柱 is_void。
	XunKong string `json:"xun_kong"`
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
