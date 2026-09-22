package bazi

import "liki-engine/internal/engine/ganzhi"

type zhuPairEntry struct {
	AZhu   string      `json:"a_zhu"`
	BZhu   string      `json:"b_zhu"`
	AGan   string      `json:"a_gan"`
	BGan   string      `json:"b_gan"`
	AZhi   string      `json:"a_zhi"`
	BZhi   string      `json:"b_zhi"`
	GanRel GanRelation `json:"gan_guan_xi"`
	ZhiRel ZhiRelation `json:"zhi_guan_xi"`
}
type zhuCross struct {
	Pairs []zhuPairEntry `json:"pairs"`
}
type dayMasterCross struct {
	A      crossGanParty `json:"a"`
	B      crossGanParty `json:"b"`
	GanRel GanRelation   `json:"gan_guan_xi"`
}
type crossGanParty struct {
	Gan    string `json:"gan"`
	Wuxing string `json:"wuxing"`
	Gender string `json:"gender"`
}
type spousePalaceCross struct {
	AZhi     string      `json:"a_zhi"`
	BZhi     string      `json:"b_zhi"`
	Relation ZhiRelation `json:"relation"`
}
type shiShenCross struct {
	AToB map[string]string `json:"a_to_b"`
	BToA map[string]string `json:"b_to_a"`
}
type nayinPairEntry struct {
	AZhu     string `json:"a_zhu"`
	BZhu     string `json:"b_zhu"`
	ANaYin   string `json:"a_na_yin"`
	BNaYin   string `json:"b_na_yin"`
	Relation string `json:"relation"`
}
type elementCross struct {
	A        map[string]int `json:"a"`
	B        map[string]int `json:"b"`
	Combined map[string]int `json:"combined"`
	FuYi     struct {
		A yongShenFitEntry `json:"a"`
		B yongShenFitEntry `json:"b"`
	} `json:"fu_yi_in_other"`
}
type yongShenFitEntry struct {
	Model       string `json:"model"`
	Yong        string `json:"yong"`
	Xi          string `json:"xi"`
	Ji          string `json:"ji"`
	YongInOther int    `json:"yong_in_other"`
	XiInOther   int    `json:"xi_in_other"`
	JiInOther   int    `json:"ji_in_other"`
}
type nayinCross struct {
	Pairs    []nayinPairEntry `json:"pairs"`
	Elements struct {
		A map[string]int `json:"a"`
		B map[string]int `json:"b"`
	} `json:"wuxing_counts"`
}
type spouseStarOccurrence struct {
	Pillar string `json:"pillar"`
	Branch string `json:"branch"`
	Stem   string `json:"stem"`
	Source string `json:"source"`
	IsVoid bool   `json:"is_void"`
}
type spouseStarGroup struct {
	TenGod      string                 `json:"ten_god"`
	Occurrences []spouseStarOccurrence `json:"occurrences"`
}
type spouseStarFact struct {
	Gender    string          `json:"gender"`
	Primary   spouseStarGroup `json:"primary"`
	Secondary spouseStarGroup `json:"secondary"`
}
type spouseStarCross struct {
	A spouseStarFact `json:"a"`
	B spouseStarFact `json:"b"`
}
type shenshaMutual struct {
	AInB bool `json:"a_present_in_b"`
	BInA bool `json:"b_present_in_a"`
}
type shenshaCross struct {
	TianYi   shenshaMutual `json:"tian_yi"`
	Lu       shenshaMutual `json:"lu"`
	TaoHua   shenshaMutual `json:"tao_hua"`
	YiMa     shenshaMutual `json:"yi_ma"`
	KongWang shenshaMutual `json:"kong_wang"`
	KuiGang  shenshaMutual `json:"kui_gang"`
	RiDe     shenshaMutual `json:"ri_de"`
	RiGui    shenshaMutual `json:"ri_gui"`
}

// Bond holds natal pair facts only. Dynamic DaYun and LiuYun timing remain in
// each person's own chart and yearly tools; this type does not synthesize a
// compatibility rhythm or rating.
type Bond struct {
	DayMaster    dayMasterCross    `json:"day_master"`
	SpousePalace spousePalaceCross `json:"spouse_palace"`
	ZhuCross     zhuCross          `json:"zhu_cross"`
	ShiShenCross shiShenCross      `json:"shi_shen_cross"`
	ElementCross elementCross      `json:"wuxing_cross"`
	NayinCross   nayinCross        `json:"nayin_cross"`
	SpouseStar   spouseStarCross   `json:"spouse_star"`
	ShenshaCross shenshaCross      `json:"shensha_cross"`
}

func ComputeBond(aBazi, bBazi Chart) Bond {
	aFull, bFull := ComputeFullChart(aBazi), ComputeFullChart(bBazi)
	return Bond{
		DayMaster:    computeDayMasterCross(aBazi, bBazi),
		SpousePalace: computeSpousePalaceCross(aFull, bFull),
		ZhuCross:     computeZhuCross(aBazi, bBazi),
		ShiShenCross: computeShiShenCross(aBazi, bBazi),
		ElementCross: computeElementCross(aBazi, bBazi, aFull, bFull),
		NayinCross:   computeNayinCross(aBazi, bBazi),
		SpouseStar:   computeSpouseStarCross(aFull, bFull),
		ShenshaCross: computeShenshaCross(aBazi, bBazi),
	}
}

func computeDayMasterCross(a, b Chart) dayMasterCross {
	cross := dayMasterCross{
		GanRel: analyzeGanRelation(a.Ri.Gan, b.Ri.Gan),
	}
	cross.A = crossGanParty{
		Gan: ganzhi.GanName(a.Ri.Gan), Wuxing: ganzhi.GanWuxing(a.Ri.Gan).String(), Gender: a.Gender.String(),
	}
	cross.B = crossGanParty{
		Gan: ganzhi.GanName(b.Ri.Gan), Wuxing: ganzhi.GanWuxing(b.Ri.Gan).String(), Gender: b.Gender.String(),
	}
	return cross
}

func computeSpousePalaceCross(a, b FullChart) spousePalaceCross {
	return spousePalaceCross{
		AZhi: ganzhi.ZhiName(a.Ri.Zhi), BZhi: ganzhi.ZhiName(b.Ri.Zhi),
		Relation: analyzeZhiRelation(a.Ri.Zhi, b.Ri.Zhi),
	}
}

func computeElementCross(a, b Chart, aFull, bFull FullChart) elementCross {
	aCount := convertWuxingCount(computeElementCount(a.ToBazi(), computeCangGan(a.ToBazi())))
	bCount := convertWuxingCount(computeElementCount(b.ToBazi(), computeCangGan(b.ToBazi())))
	combined := convertWuxingCount(map[ganzhi.Wuxing]int{})
	for _, src := range []map[string]int{aCount, bCount} {
		for k, v := range src {
			combined[k] += v
		}
	}
	cross := elementCross{A: aCount, B: bCount, Combined: combined}
	cross.FuYi.A = yongShenFit(aFull.FuYi, bCount)
	cross.FuYi.B = yongShenFit(bFull.FuYi, aCount)
	return cross
}

func convertWuxingCount(src map[ganzhi.Wuxing]int) map[string]int {
	out := map[string]int{
		ganzhi.WxMu.String():   0,
		ganzhi.WxHuo.String():  0,
		ganzhi.WxTu.String():   0,
		ganzhi.WxJin.String():  0,
		ganzhi.WxShui.String(): 0,
	}
	for k, v := range src {
		out[k.String()] = v
	}
	return out
}

func yongShenFit(f FuYiResult, other map[string]int) yongShenFitEntry {
	return yongShenFitEntry{
		Model: f.Model, Yong: f.Yong, Xi: f.Xi, Ji: f.Ji,
		YongInOther: other[f.Yong], XiInOther: other[f.Xi], JiInOther: other[f.Ji],
	}
}

func computeSpouseStarCross(a, b FullChart) spouseStarCross {
	return spouseStarCross{
		A: buildSpouseStarFact(a),
		B: buildSpouseStarFact(b),
	}
}

func buildSpouseStarFact(fc FullChart) spouseStarFact {
	primary, secondary := ganzhi.ShiShenZhengCai, ganzhi.ShiShenPianCai
	if fc.Gender == ganzhi.Female {
		primary, secondary = ganzhi.ShiShenZhengGuan, ganzhi.ShiShenQiSha
	}
	out := spouseStarFact{
		Gender: fc.Gender.String(),
		Primary: spouseStarGroup{
			TenGod:      primary.String(),
			Occurrences: []spouseStarOccurrence{},
		},
		Secondary: spouseStarGroup{
			TenGod:      secondary.String(),
			Occurrences: []spouseStarOccurrence{},
		},
	}
	pillars := []fullZhuInfo{fc.Nian, fc.Yue, fc.Ri, fc.Shi}
	for i, pillar := range pillars {
		for _, item := range pillar.ShiShens {
			occ := spouseStarOccurrence{
				Pillar: zhuLabels[i],
				Branch: ganzhi.ZhiName(pillar.Zhi),
				Stem:   item.Name,
				Source: item.Source,
				IsVoid: pillar.IsVoid,
			}
			switch item.ShiShen {
			case primary:
				out.Primary.Occurrences = append(out.Primary.Occurrences, occ)
			case secondary:
				out.Secondary.Occurrences = append(out.Secondary.Occurrences, occ)
			}
		}
	}
	return out
}

func computeZhuCross(a, b Chart) zhuCross {
	aG := [4]ganzhi.Gan{a.Nian.Gan, a.Yue.Gan, a.Ri.Gan, a.Shi.Gan}
	bG := [4]ganzhi.Gan{b.Nian.Gan, b.Yue.Gan, b.Ri.Gan, b.Shi.Gan}
	aZ := [4]ganzhi.Zhi{a.Nian.Zhi, a.Yue.Zhi, a.Ri.Zhi, a.Shi.Zhi}
	bZ := [4]ganzhi.Zhi{b.Nian.Zhi, b.Yue.Zhi, b.Ri.Zhi, b.Shi.Zhi}
	pairs := make([]zhuPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			pairs = append(pairs, zhuPairEntry{
				AZhu: zhuLabels[i], BZhu: zhuLabels[j],
				AGan: ganzhi.GanName(aG[i]), BGan: ganzhi.GanName(bG[j]),
				AZhi: ganzhi.ZhiName(aZ[i]), BZhi: ganzhi.ZhiName(bZ[j]),
				GanRel: analyzeGanRelation(aG[i], bG[j]),
				ZhiRel: analyzeZhiRelation(aZ[i], bZ[j]),
			})
		}
	}
	return zhuCross{Pairs: pairs}
}

func computeShiShenCross(a, b Chart) shiShenCross {
	aG := [4]ganzhi.Gan{a.Nian.Gan, a.Yue.Gan, a.Ri.Gan, a.Shi.Gan}
	bG := [4]ganzhi.Gan{b.Nian.Gan, b.Yue.Gan, b.Ri.Gan, b.Shi.Gan}
	aElem, aYY := ganzhi.GanWuxing(a.Ri.Gan), ganzhi.GanYinYang(a.Ri.Gan)
	bElem, bYY := ganzhi.GanWuxing(b.Ri.Gan), ganzhi.GanYinYang(b.Ri.Gan)
	aToB, bToA := make(map[string]string, 4), make(map[string]string, 4)
	for i := 0; i < 4; i++ {
		aToB[zhuLabels[i]+"_gan"] = ganzhi.ShiShenName(ganzhi.ShiShenType(aElem, aYY, ganzhi.GanWuxing(bG[i]), ganzhi.GanYinYang(bG[i])))
		bToA[zhuLabels[i]+"_gan"] = ganzhi.ShiShenName(ganzhi.ShiShenType(bElem, bYY, ganzhi.GanWuxing(aG[i]), ganzhi.GanYinYang(aG[i])))
	}
	return shiShenCross{AToB: aToB, BToA: bToA}
}

func computeNayinCross(a, b Chart) nayinCross {
	aNy, bNy := a.NaYinArray(), b.NaYinArray()
	pairs := make([]nayinPairEntry, 0, 16)
	for i := 0; i < 4; i++ {
		ae := ganzhi.NayinWuxing(aNy[i])
		for j := 0; j < 4; j++ {
			be := ganzhi.NayinWuxing(bNy[j])
			rel := "相同"
			if ae != be {
				if ganzhi.Sheng(ae, be) || ganzhi.Sheng(be, ae) {
					rel = "相生"
				} else {
					rel = "相克"
				}
			}
			pairs = append(pairs, nayinPairEntry{AZhu: zhuLabels[i], BZhu: zhuLabels[j], ANaYin: aNy[i], BNaYin: bNy[j], Relation: rel})
		}
	}
	nc := nayinCross{Pairs: pairs}

	// 纳音五行分布（四柱计数）
	countElems := func(ny [4]string) map[string]int {
		m := map[string]int{}
		for _, s := range ny {
			m[ganzhi.NayinWuxing(s).String()]++
		}
		return m
	}
	nc.Elements.A = countElems(aNy)
	nc.Elements.B = countElems(bNy)
	return nc
}

func computeShenshaCross(a, b Chart) shenshaCross {
	aBz, bBz := a.ToBazi(), b.ToBazi()
	aPs, bPs := aBz.Slice(), bBz.Slice()
	mut := func(aZ, bZ []ganzhi.Zhi) shenshaMutual {
		return shenshaMutual{zhiInZhus(aZ, bBz), zhiInZhus(bZ, aBz)}
	}
	return shenshaCross{
		TianYi:   mut(tianYiZhi(a), tianYiZhi(b)),
		Lu:       mut(luZhi(a), luZhi(b)),
		TaoHua:   mut(zhiLookup(a, taohuaZhiMap), zhiLookup(b, taohuaZhiMap)),
		YiMa:     mut(zhiLookup(a, yimaZhiMap), zhiLookup(b, yimaZhiMap)),
		KongWang: mut(kongwangZhi(aBz), kongwangZhi(bBz)),
		KuiGang:  mut(collectZhi(aPs, isKuiGang), collectZhi(bPs, isKuiGang)),
		RiDe:     mut(collectZhi(aPs, isRiDe), collectZhi(bPs, isRiDe)),
		RiGui:    mut(collectZhi(aPs, isRiGui), collectZhi(bPs, isRiGui)),
	}
}

func zhiInZhus(ts []ganzhi.Zhi, bz ganzhi.Bazi) bool {
	for _, t := range ts {
		for _, p := range bz.Slice() {
			if p.Zhi == t {
				return true
			}
		}
	}
	return false
}
func tianYiZhi(c Chart) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := tianYiLookup[c.Nian.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	if b, ok := tianYiLookup[c.Ri.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	return bs
}
func luZhi(c Chart) []ganzhi.Zhi {
	if c.Ri.Gan < 1 || c.Ri.Gan > 10 {
		return nil
	}
	return []ganzhi.Zhi{ganzhi.ChangShengTable[c.Ri.Gan][3]}
}
func zhiLookup(c Chart, m map[ganzhi.Zhi]ganzhi.Zhi) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := m[c.Nian.Zhi]; ok {
		bs = append(bs, b)
	}
	if b, ok := m[c.Ri.Zhi]; ok {
		bs = append(bs, b)
	}
	return bs
}
func kongwangZhi(bz ganzhi.Bazi) []ganzhi.Zhi {
	ps := bz.Slice()
	vh := computeKongWang(bz)
	bs := make([]ganzhi.Zhi, 0, len(vh))
	for _, idx := range vh {
		bs = append(bs, ps[idx].Zhi)
	}
	return bs
}
func collectZhi(ps [4]ganzhi.Zhu, f func(ganzhi.Zhu) bool) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	for _, p := range ps {
		if f(p) {
			bs = append(bs, p.Zhi)
		}
	}
	return bs
}
func isRiDe(p ganzhi.Zhu) bool  { _, ok := riDeSet[[2]int{int(p.Gan), int(p.Zhi)}]; return ok }
func isRiGui(p ganzhi.Zhu) bool { _, ok := riGuiSet[[2]int{int(p.Gan), int(p.Zhi)}]; return ok }

var riDeSet map[[2]int]bool
var riGuiSet map[[2]int]bool
