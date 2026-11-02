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
type naYinPairEntry struct {
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
	ModelID     string `json:"fu_yi_model_id"`
	Yong        string `json:"yong"`
	Xi          string `json:"xi"`
	Ji          string `json:"ji"`
	YongInOther int    `json:"yong_in_other"`
	XiInOther   int    `json:"xi_in_other"`
	JiInOther   int    `json:"ji_in_other"`
}
type nayinCross struct {
	Pairs    []naYinPairEntry `json:"pairs"`
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
type SpouseStarFact = spouseStarFact
type SpouseStarGroup = spouseStarGroup
type spouseStarCross struct {
	A spouseStarFact `json:"a"`
	B spouseStarFact `json:"b"`
}
type ShenshaMutual struct {
	AInB bool `json:"a_present_in_b"`
	BInA bool `json:"b_present_in_a"`
}
type ShenshaCross struct {
	TianYi   ShenshaMutual `json:"tian_yi"`
	Lu       ShenshaMutual `json:"lu"`
	TaoHua   ShenshaMutual `json:"tao_hua"`
	YiMa     ShenshaMutual `json:"yi_ma"`
	KongWang ShenshaMutual `json:"kong_wang"`
	KuiGang  ShenshaMutual `json:"kui_gang"`
	RiDe     ShenshaMutual `json:"ri_de"`
	RiGui    ShenshaMutual `json:"ri_gui"`
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
	NaYinCross   nayinCross        `json:"na_yin_cross"`
	SpouseStar   spouseStarCross   `json:"spouse_star"`
	ShenshaCross ShenshaCross      `json:"shen_sha_cross"`
}

func ComputeBond(aCore, bCore Core) Bond {
	return computeBondFromFacts(
		aCore,
		bCore,
		computeBondChartFacts(aCore),
		computeBondChartFacts(bCore),
	)
}

// computeBondFromFacts is the pure composition layer: lean charts and their
// bond-specific projections go in, and a complete Bond value comes out. It
// neither reads nor writes request/session state.
func computeBondFromFacts(aCore, bCore Core, aFacts, bFacts bondChartFacts) Bond {
	return Bond{
		DayMaster:    computeDayMasterCross(aCore, bCore),
		SpousePalace: computeSpousePalaceCross(aCore, bCore),
		ZhuCross:     computeZhuCross(aCore, bCore),
		ShiShenCross: computeShiShenCross(aCore, bCore),
		ElementCross: computeElementCross(aFacts, bFacts),
		NaYinCross:   computeNaYinCross(aCore, bCore),
		SpouseStar:   computeSpouseStarCross(aFacts, bFacts),
		ShenshaCross: computeShenshaCross(aCore, bCore),
	}
}

type bondChartFacts struct {
	counts     WuxingCounts
	fuYi       FuYiResult
	spouseStar spouseStarFact
}

func computeBondChartFacts(core Core) bondChartFacts {
	bz := core.ToBazi()
	analysis := AnalyzePillars(bz, core.Gender)
	counts := WuxingCountsAt(bz, analysis.HiddenStems)
	strength := AnalyzeDayMasterStrength(bz, analysis.HiddenStems)
	relations := computeNatalRelations(bz)
	return bondChartFacts{
		counts:     counts,
		fuYi:       FuYi(bz, counts, SeasonStrengthAt(bz.Yue.Zhi), strength, relations),
		spouseStar: buildSpouseStarFactFromAnalysis(core, bz, analysis),
	}
}

func computeDayMasterCross(a, b Core) dayMasterCross {
	cross := dayMasterCross{
		GanRel: analyzeGanRelation(a.Pillars[zhuRi].Gan, b.Pillars[zhuRi].Gan),
	}
	cross.A = crossGanParty{
		Gan: ganzhi.GanName(a.Pillars[zhuRi].Gan), Wuxing: ganzhi.GanWuxing(a.Pillars[zhuRi].Gan).String(), Gender: a.Gender.String(),
	}
	cross.B = crossGanParty{
		Gan: ganzhi.GanName(b.Pillars[zhuRi].Gan), Wuxing: ganzhi.GanWuxing(b.Pillars[zhuRi].Gan).String(), Gender: b.Gender.String(),
	}
	return cross
}

func computeSpousePalaceCross(a, b Core) spousePalaceCross {
	return spousePalaceCross{
		AZhi: ganzhi.ZhiName(a.Pillars[zhuRi].Zhi), BZhi: ganzhi.ZhiName(b.Pillars[zhuRi].Zhi),
		Relation: analyzeZhiRelation(a.Pillars[zhuRi].Zhi, b.Pillars[zhuRi].Zhi),
	}
}

func computeElementCross(aFacts, bFacts bondChartFacts) elementCross {
	aCount := convertWuxingCount(aFacts.counts)
	bCount := convertWuxingCount(bFacts.counts)
	combined := convertWuxingCount(map[ganzhi.Wuxing]int{})
	for _, src := range []map[string]int{aCount, bCount} {
		for k, v := range src {
			combined[k] += v
		}
	}
	cross := elementCross{A: aCount, B: bCount, Combined: combined}
	cross.FuYi.A = yongShenFit(aFacts.fuYi, bCount)
	cross.FuYi.B = yongShenFit(bFacts.fuYi, aCount)
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
	inOther := func(value string) int {
		if value == NoCandidateWuxing {
			return 0
		}
		return other[value]
	}
	return yongShenFitEntry{
		ModelID: f.ModelID, Yong: f.Yong, Xi: f.Xi, Ji: f.Ji,
		YongInOther: inOther(f.Yong), XiInOther: inOther(f.Xi), JiInOther: inOther(f.Ji),
	}
}

func computeSpouseStarCross(a, b bondChartFacts) spouseStarCross {
	return spouseStarCross{
		A: a.spouseStar,
		B: b.spouseStar,
	}
}

func buildSpouseStarFactFromAnalysis(
	core Core, bz ganzhi.Bazi, analysis PillarAnalysis,
) spouseStarFact {
	primary, secondary := ganzhi.ShiShenZhengCai, ganzhi.ShiShenPianCai
	if core.Gender == ganzhi.Female {
		primary, secondary = ganzhi.ShiShenZhengGuan, ganzhi.ShiShenQiSha
	}
	out := spouseStarFact{
		Gender: core.Gender.String(),
		Primary: spouseStarGroup{
			TenGod:      primary.String(),
			Occurrences: []spouseStarOccurrence{},
		},
		Secondary: spouseStarGroup{
			TenGod:      secondary.String(),
			Occurrences: []spouseStarOccurrence{},
		},
	}
	for i, pillar := range bz.Slice() {
		isVoid := false
		for _, voidPillar := range analysis.VoidPalaces {
			if voidPillar == i {
				isVoid = true
				break
			}
		}
		for _, item := range analysis.TenGods[i] {
			occ := spouseStarOccurrence{
				Pillar: zhuLabels[i],
				Branch: ganzhi.ZhiName(pillar.Zhi),
				Stem:   item.Name,
				Source: item.Source,
				IsVoid: isVoid,
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

func pillarGans(core Core) [4]ganzhi.Gan {
	return [4]ganzhi.Gan{core.Pillars[0].Gan, core.Pillars[1].Gan, core.Pillars[2].Gan, core.Pillars[3].Gan}
}

func pillarZhis(core Core) [4]ganzhi.Zhi {
	return [4]ganzhi.Zhi{core.Pillars[0].Zhi, core.Pillars[1].Zhi, core.Pillars[2].Zhi, core.Pillars[3].Zhi}
}

func computeZhuCross(a, b Core) zhuCross {
	aG := pillarGans(a)
	bG := pillarGans(b)
	aZ := pillarZhis(a)
	bZ := pillarZhis(b)
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

func computeShiShenCross(a, b Core) shiShenCross {
	aG := pillarGans(a)
	bG := pillarGans(b)
	aElem, aYY := ganzhi.GanWuxing(a.Pillars[zhuRi].Gan), ganzhi.GanYinYang(a.Pillars[zhuRi].Gan)
	bElem, bYY := ganzhi.GanWuxing(b.Pillars[zhuRi].Gan), ganzhi.GanYinYang(b.Pillars[zhuRi].Gan)
	aToB, bToA := make(map[string]string, 4), make(map[string]string, 4)
	for i := 0; i < 4; i++ {
		aToB[zhuLabels[i]+"_gan"] = ganzhi.ShiShenName(ganzhi.ShiShenType(aElem, aYY, ganzhi.GanWuxing(bG[i]), ganzhi.GanYinYang(bG[i])))
		bToA[zhuLabels[i]+"_gan"] = ganzhi.ShiShenName(ganzhi.ShiShenType(bElem, bYY, ganzhi.GanWuxing(aG[i]), ganzhi.GanYinYang(aG[i])))
	}
	return shiShenCross{AToB: aToB, BToA: bToA}
}

func computeNaYinCross(a, b Core) nayinCross {
	aNy, bNy := a.NaYin, b.NaYin
	pairs := make([]naYinPairEntry, 0, 16)
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
			pairs = append(pairs, naYinPairEntry{AZhu: zhuLabels[i], BZhu: zhuLabels[j], ANaYin: aNy[i], BNaYin: bNy[j], Relation: rel})
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

func computeShenshaCross(a, b Core) ShenshaCross {
	aBz, bBz := a.ToBazi(), b.ToBazi()
	aPs, bPs := aBz.Slice(), bBz.Slice()
	mut := func(aZ, bZ []ganzhi.Zhi) ShenshaMutual {
		return ShenshaMutual{zhiInZhus(aZ, bBz), zhiInZhus(bZ, aBz)}
	}
	return ShenshaCross{
		TianYi:   mut(tianYiZhi(aBz), tianYiZhi(bBz)),
		Lu:       mut(luZhi(aBz), luZhi(bBz)),
		TaoHua:   mut(zhiLookup(aBz, taohuaZhiMap), zhiLookup(bBz, taohuaZhiMap)),
		YiMa:     mut(zhiLookup(aBz, yimaZhiMap), zhiLookup(bBz, yimaZhiMap)),
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
func tianYiZhi(bz ganzhi.Bazi) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := tianYiLookup[bz.Nian.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	if b, ok := tianYiLookup[bz.Ri.Gan]; ok {
		bs = append(bs, b[0], b[1])
	}
	return bs
}
func luZhi(bz ganzhi.Bazi) []ganzhi.Zhi {
	if bz.Ri.Gan < 1 || bz.Ri.Gan > 10 {
		return nil
	}
	return []ganzhi.Zhi{ganzhi.LuZhi(bz.Ri.Gan)}
}
func zhiLookup(bz ganzhi.Bazi, m map[ganzhi.Zhi]ganzhi.Zhi) []ganzhi.Zhi {
	var bs []ganzhi.Zhi
	if b, ok := m[bz.Nian.Zhi]; ok {
		bs = append(bs, b)
	}
	if b, ok := m[bz.Ri.Zhi]; ok {
		bs = append(bs, b)
	}
	return bs
}
func kongwangZhi(bz ganzhi.Bazi) []ganzhi.Zhi {
	ps := bz.Slice()
	vh := VoidPalaces(bz)
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
