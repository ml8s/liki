package ziwei

import "fmt"

type Bond struct {
	AGongName string `json:"a_gong_name"`
	BGongName string `json:"b_gong_name"`

	PalaceCross struct {
		AIntoB string `json:"a_into_b"`
		BIntoA string `json:"b_into_a"`
	} `json:"ming_gong_hu_ru"`

	FuQiGong  *PairRef      `json:"fu_qi_gong,omitempty"`
	ZiNvGong   *PairRef      `json:"zi_nv_gong,omitempty"`
	JiXing  []StarCross   `json:"ji_xing,omitempty"`
	ShaXing  []StarCross   `json:"sha_xing,omitempty"`
	LuMaIn     []StarCross   `json:"lu_ma_in,omitempty"`
	KongWang     []StarCross   `json:"kong_wang,omitempty"`
	SiHuaIn    []SiHuaCross  `json:"sihua_in,omitempty"`
	WuXingShengKe string        `json:"wu_xing_sheng_ke"`
	Summary    string        `json:"summary"`
}

type PairRef struct {
	AGongName string   `json:"a_gong_name"`
	BGongName string   `json:"b_gong_name"`
	AZhuXing  []string `json:"a_zhu_xing"`
	BZhuXing  []string `json:"b_zhu_xing"`
	PanJue    string   `json:"pan_jue"`
}

type StarCross struct {
	Name  string `json:"name"`
	FromA string `json:"from_a"`
	IntoB string `json:"into_b"`
}

type SiHuaCross struct {
	Name string `json:"name"`
	Type string `json:"type"`
	From string `json:"from"`
	Into string `json:"into"`
}

func ComputeBond(a, b Chart) Bond {
	ai, bi := mingGongHuRu(a, b)
	pc := struct {
		AIntoB string `json:"a_into_b"`
		BIntoA string `json:"b_into_a"`
	}{ai, bi}
	return Bond{
		AGongName: fmt.Sprintf("%d", a.BirthYear),
		BGongName:  fmt.Sprintf("%d", b.BirthYear),
		PalaceCross: pc,
		FuQiGong: gongWeiDuiZhao(a, b, 2, "夫妻"),
		ZiNvGong:  gongWeiDuiZhao(a, b, 3, "子女"),
		JiXing: xingYaoHuRu(a, b, beneficStars),
		ShaXing: xingYaoHuRu(a, b, maleficStars),
		LuMaIn:    luMaHuRu(a, b),
		KongWang:    kongWangHuRu(a, b),
		SiHuaIn:   siHuaHuYin(a, b),
		WuXingShengKe: wuXingShengKe(a, b),
		Summary:   "合盘分析供参考",
	}
}

func mingGongHuRu(a, b Chart) (string, string) {
	return palaceByZhi(b.Palaces, a.Palaces[0].Zhi),
		palaceByZhi(a.Palaces, b.Palaces[0].Zhi)
}

func palaceByZhi(p [12]palace, z Zhi) string {
	for i := range p { if p[i].Zhi == z { return p[i].Name } }
	return ""
}

func gongWeiDuiZhao(a, b Chart, idx int, label string) *PairRef {
	return &PairRef{
		AGongName: a.Palaces[idx].Name, BGongName: b.Palaces[idx].Name,
		AZhuXing: majorList(a.Palaces[idx].Stars),
		BZhuXing: majorList(b.Palaces[idx].Stars),
		PanJue: pairVerdict(label, a.Palaces[idx], b.Palaces[idx]),
	}
}

func majorList(stars []starInfo) []string {
	var r []string
	for _, s := range stars { if s.IsMajor { r = append(r, s.Name) } }
	return r
}

func pairVerdict(label string, a, b palace) string {
	switch label {
	case "夫妻": return spouseVerdict(majorList(a.Stars), majorList(b.Stars))
	case "子女": return childVerdict(majorList(a.Stars), majorList(b.Stars))
	}
	return ""
}

func spouseVerdict(a, b []string) string {
	if len(a) == 0 && len(b) == 0 { return "双方夫妻宫皆空，需注意情感表达" }
	if len(b) == 0 { return "A方夫妻宫有主星，B方参考对宫" }
	if len(a) == 0 { return "B方夫妻宫有主星，A方参考对宫" }
	return "双方夫妻宫皆有主星"
}

func childVerdict(a, b []string) string {
	if len(a) == 0 && len(b) == 0 { return "双方子女宫皆空，参考对宫" }
	return "双方子女宫有主星"
}

var beneficStars = map[starIndex]string{
	TianKui: "天魁", TianYue: "天钺",
	ZuoFu: "左辅", YouBi: "右弼",
	WenChang: "文昌", WenQu: "文曲",
}

var maleficStars = map[starIndex]string{
	QingYang: "擎羊", TuoLuo: "陀罗",
	HuoXing: "火星", LingXing: "铃星",
}

func xingYaoHuRu(a, b Chart, stars map[starIndex]string) []StarCross {
	var r []StarCross
	for si, name := range stars {
		ai := findStarPalace(a.Palaces, si)
		if ai < 0 { continue }
		bn := palaceByZhi(b.Palaces, a.Palaces[ai].Zhi)
		r = append(r, StarCross{name, a.Palaces[ai].Name, bn})
	}
	return r
}

func luMaHuRu(a, b Chart) []StarCross {
	var r []StarCross
	for _, si := range []starIndex{LuCun, TianMa} {
		ni := starNames[si]
		ai := findStarPalace(a.Palaces, si)
		if ai < 0 { continue }
		bn := palaceByZhi(b.Palaces, a.Palaces[ai].Zhi)
		r = append(r, StarCross{ni, a.Palaces[ai].Name, bn})
	}
	return r
}

func kongWangHuRu(a, b Chart) []StarCross {
	var r []StarCross
	for _, nm := range []string{"截路", "空亡"} {
		ai := -1
		for i := range a.Palaces {
			for _, s := range a.Palaces[i].ZaYao {
				if s == nm { ai = i; break }
			}
			if ai >= 0 { break }
		}
		if ai < 0 { continue }
		bn := palaceByZhi(b.Palaces, a.Palaces[ai].Zhi)
		r = append(r, StarCross{nm, a.Palaces[ai].Name, bn})
	}
	return r
}

func findStarPalace(p [12]palace, si starIndex) int {
	for i := range p {
		for _, s := range p[i].Stars {
			if s.Star == si { return i }
		}
	}
	return -1
}

func siHuaHuYin(a, b Chart) []SiHuaCross {
	var r []SiHuaCross
	for si, ht := range a.SiHua {
		ai := findStarPalace(a.Palaces, si)
		if ai < 0 { continue }
		bn := palaceByZhi(b.Palaces, a.Palaces[ai].Zhi)
		ni := starNames[si]
		sht := string(ht)
		if bht, bok := b.SiHua[si]; bok {
			bs := string(bht)
			if sht == "化忌" && bs == "化忌" {
				sht = "化忌·双化忌注意"
			} else if sht == "化禄" && bs == "化忌" {
				sht = "化禄·化忌·两极"
			}
		}
		r = append(r, SiHuaCross{ni, sht, a.Palaces[ai].Name, bn})
	}
	return r
}

func wuXingShengKe(a, b Chart) string {
	jm := map[string]string{"水二局": "水", "木三局": "木", "金四局": "金", "土五局": "土", "火六局": "火"}
	ae := jm[a.JuShuName]
	be := jm[b.JuShuName]
	if ae == "" || be == "" { return "未知" }
	if ae == be { return "比和" }
	cl := []string{"金", "水", "木", "火", "土"}
	ai := -1
	for i, e := range cl { if e == ae { ai = i; break } }
	nx := cl[(ai+1)%5]
	pv := cl[(ai-1+5)%5]
	if be == nx { return ae + "生" + be + "·相生" }
	if be == pv { return "·相克" }
	return "·相克"
}
