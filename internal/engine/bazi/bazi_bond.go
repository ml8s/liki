package bazi

import "liki-engine/internal/engine/ganzhi"

// xunIndex returns the xun index (0-5) for a day pillar.
func xunIndex(riZhu ganzhi.Zhu) int {
	return ganzhi.SixtyCycleIndex(riZhu.Gan, riZhu.Zhi) / 10
}

type zhuPairEntry struct {
	AZhu     string      `json:"jia_zhu"`
	BZhu     string      `json:"yi_zhu"`
	AStem    string      `json:"jia_gan"`
	BStem    string      `json:"yi_gan"`
	ABranch  string      `json:"jia_zhi"`
	BBranch  string      `json:"yi_zhi"`
	Stem     GanRelation `json:"gan_guan_xi"`
	Branch   ZhiRelation `json:"zhi_guan_xi"`
}
type zhuCross struct {
	Pairs []zhuPairEntry `json:"pairs"`
}
type shiShenCross struct {
	AToB map[string]string `json:"jia_dui_yi"`
	BToA map[string]string `json:"yi_dui_jia"`
}
type nayinPairEntry struct {
	AZhu     string `json:"jia_zhu"`
	BZhu     string `json:"yi_zhu"`
	ANaYin   string `json:"jia_na_yin"`
	BNaYin   string `json:"yi_na_yin"`
	Relation string `json:"relation"`
}
type yongShenEntry struct {
	Yong         string `json:"yong"`
	Ji           string `json:"ji"`
	YongInOther  int    `json:"yong_in_other"`
	JiInOther    int    `json:"ji_in_other"`
}
type nayinCross struct {
	Pairs    []nayinPairEntry     `json:"pairs"`
	Elements struct {
		A map[string]int `json:"a"`
		B map[string]int `json:"b"`
	} `json:"wuxings"`
	YongShen struct {
		A yongShenEntry `json:"a"`
		B yongShenEntry `json:"b"`
	} `json:"yong_shen"`
}
type shenshaMutual struct {
	AInB bool `json:"a_in_b"`
	BInA bool `json:"b_in_a"`
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
type daYunCrossEntry struct {
	Gan     ganzhi.Gan `json:"gan"`
	Zhi     ganzhi.Zhi `json:"zhi"`
	Name    string     `json:"name"`
	ShiShen string     `json:"shi_shen"`
}
type daYunCross struct {
	ACurrent  daYunCrossEntry `json:"a_current"`
	BCurrent  daYunCrossEntry `json:"b_current"`
	StemRel   GanRelation     `json:"gan_guan_xi"`
	BranchRel ZhiRelation     `json:"zhi_guan_xi"`
}
// XunGong describes whether two charts share the same xun (旬) or palace (宫).
type XunGong struct {
	SameXun  bool `json:"same_xun"`
	SameGong bool `json:"same_gong"`
}
type structureCross struct {
	DaYun   daYunCross `json:"da_yun"`
	XunGong XunGong    `json:"xun_gong"`
}
// Bond holds the compatibility analysis between two bazi charts.
type Bond struct {
	ZhuCross     zhuCross      `json:"zhu_cross"`
	ShiShenCross shiShenCross  `json:"shi_shen_cross"`
	NayinCross   nayinCross    `json:"nayin_cross"`
	ShenshaCross shenshaCross  `json:"shensha_cross"`
	Structure    structureCross `json:"structure"`
}

func ComputeBond(aBazi, bBazi Chart) Bond {
	return Bond{
		ZhuCross:     computeZhuCross(aBazi, bBazi),
		ShiShenCross: computeShiShenCross(aBazi, bBazi),
		NayinCross:   computeNayinCross(aBazi, bBazi),
		ShenshaCross: computeShenshaCross(aBazi, bBazi),
		Structure:    computeStructureCross(aBazi, bBazi),
	}
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
				AStem: ganzhi.GanName(aG[i]), BStem: ganzhi.GanName(bG[j]),
				ABranch: ganzhi.ZhiName(aZ[i]), BBranch: ganzhi.ZhiName(bZ[j]),
				Stem:   analyzeGanRelation(aG[i], bG[j]),
				Branch: analyzeZhiRelation(aZ[i], bZ[j]),
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
		aToB[zhuLabels[i]+"_stem"] = ganzhi.ShiShenName(ganzhi.ShiShenType(aElem, aYY, ganzhi.GanWuxing(bG[i]), ganzhi.GanYinYang(bG[i])))
		bToA[zhuLabels[i]+"_stem"] = ganzhi.ShiShenName(ganzhi.ShiShenType(bElem, bYY, ganzhi.GanWuxing(aG[i]), ganzhi.GanYinYang(aG[i])))
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
	return nc
}



func computeShenshaCross(a, b Chart) shenshaCross {
	aBz, bBz := a.ToBazi(), b.ToBazi()
	aPs, bPs := aBz.Slice(), bBz.Slice()
	mut := func(aZ, bZ []ganzhi.Zhi) shenshaMutual {
		return shenshaMutual{branchesInZhus(aZ, bBz), branchesInZhus(bZ, aBz)}
	}
	return shenshaCross{
		TianYi:   mut(tianYiBranches(a), tianYiBranches(b)),
		Lu:       mut(luBranches(a), luBranches(b)),
		TaoHua:   mut(zhiLookup(a, taohuaBranchMap), zhiLookup(b, taohuaBranchMap)),
		YiMa:     mut(zhiLookup(a, yimaBranchMap), zhiLookup(b, yimaBranchMap)),
		KongWang: mut(kongwangBranches(aBz), kongwangBranches(bBz)),
		KuiGang:  mut(collectBranches(aPs, isKuiGang), collectBranches(bPs, isKuiGang)),
		RiDe:     mut(collectBranches(aPs, isRiDe), collectBranches(bPs, isRiDe)),
		RiGui:    mut(collectBranches(aPs, isRiGui), collectBranches(bPs, isRiGui)),
	}
}

func branchesInZhus(ts []ganzhi.Zhi, bz ganzhi.Bazi) bool {
	for _, t := range ts {
		for _, p := range bz.Slice() {
			if p.Zhi == t {
				return true
			}
		}
	}
	return false
}
func tianYiBranches(c Chart) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := tianYiLookup[c.Nian.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	if b, ok := tianYiLookup[c.Ri.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	return bs
}
func luBranches(c Chart) []ganzhi.Zhi {
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
func kongwangBranches(bz ganzhi.Bazi) []ganzhi.Zhi {
	ps := bz.Slice()
	vh := computeKongWang(bz)
	bs := make([]ganzhi.Zhi, 0, len(vh))
	for _, idx := range vh {
		bs = append(bs, ps[idx].Zhi)
	}
	return bs
}
func collectBranches(ps [4]ganzhi.Zhu, f func(ganzhi.Zhu) bool) []ganzhi.Zhi {
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

func computeStructureCross(a, b Chart) structureCross {
	return structureCross{DaYun: computeDaYunCross(a, b), XunGong: computeXunGong(a, b)}
}
func computeDaYunCross(a, b Chart) daYunCross {
	if a.DaYun == nil || b.DaYun == nil {
		return daYunCross{}
	}
	dc := daYunCross{ACurrent: currentDaYunEntry(a.DaYun), BCurrent: currentDaYunEntry(b.DaYun)}
	dc.StemRel = analyzeGanRelation(dc.ACurrent.Gan, dc.BCurrent.Gan)
	dc.BranchRel = analyzeZhiRelation(dc.ACurrent.Zhi, dc.BCurrent.Zhi)
	return dc
}
func currentDaYunEntry(dr *DaYun) daYunCrossEntry {
	if dr == nil || dr.CurrentStepIndex < 0 || dr.CurrentStepIndex >= len(dr.Steps) {
		return daYunCrossEntry{}
	}
	p := dr.Steps[dr.CurrentStepIndex]
	return daYunCrossEntry{Gan: p.Gan, Zhi: p.Zhi, Name: p.Name, ShiShen: p.ShiShen}
}
func computeXunGong(a, b Chart) XunGong {
	return XunGong{
		SameXun:  xunIndex(a.Ri.Zhu) == xunIndex(b.Ri.Zhu),
		SameGong: a.Ri.Zhi == b.Ri.Zhi,
	}
}
