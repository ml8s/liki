package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func computeChartCore(bz ganzhi.Bazi, st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	ny := computeNaYin(bz)
	ps := bz.Slice()

	makePI := func(i int) zhuInfo {
		return zhuInfo{Zhu: ps[i], NaYin: ny[i]}
	}

	cr := Chart{
		Nian:   makePI(0),
		Yue:    makePI(1),
		Ri:     makePI(2),
		Shi:    makePI(3),
		Gender: gender,
	}
	cr.DaYun = computeDaYun(st, bz.Yue, bz.Nian.Gan, bz.Ri.Gan, gender)
	cr.BirthYear = st.Time().Year()
	return cr
}

func computeFullFromCore(c Chart, bz ganzhi.Bazi) FullChart {
	hs := computeCangGan(bz)
	ny := c.NaYinArray()
	tgTable := computeShiShensTable(bz, hs)
	lsTable := computeChangShengTable(bz, hs)
	shensha := computeShenSha(bz)
	voidHits := computeKongWang(bz)
	ps := bz.Slice()

	makeFull := func(i int) fullZhuInfo {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		pi := fullZhuInfo{Zhu: ps[i], NaYin: ny[i], CangGan: hs[i], ShiShens: tgTable[i], ChangSheng: lsTable[i], ShenSha: shensha[i], IsVoid: isVoid}
		pi.IsSelfHe = isSelfHe(ps[i])
		if pi.IsSelfHe {
			pi.SelfHeName = selfHeName(ps[i])
		}
		pi.IsKuiGang = isKuiGang(ps[i])
		return pi
	}

	return FullChart{
		BirthYear: c.BirthYear,
		Nian:   makeFull(0),
		Yue:    makeFull(1),
		Ri:     makeFull(2),
		Shi:    makeFull(3),
		DaYun:  c.DaYun,
		Gender: c.Gender,
	}
}

func computeCangGan(bz ganzhi.Bazi) [4]cangGanOut {
	var hs [4]cangGanOut
	for i, z := range bz.Slice() {
		if z.Zhi == 0 {
			continue
		}
		qi := ganzhi.CangGanForZhi(z.Zhi)
		hs[i] = cangGanOut{Main: *qi.Main}
		if qi.Mid != nil {
			mid := *qi.Mid
			hs[i].Mid = &mid
		}
		if qi.Minor != nil {
			minor := *qi.Minor
			hs[i].Minor = &minor
		}
	}
	return hs
}

func computeNaYin(bz ganzhi.Bazi) [4]string {
	var ny [4]string
	for i, z := range bz.Slice() {
		ny[i] = ganzhi.NayinLabel(z.Gan, z.Zhi)
	}
	return ny
}

func computeElementCount(bz ganzhi.Bazi, hs [4]cangGanOut) map[ganzhi.Wuxing]int {
	wc := make(map[ganzhi.Wuxing]int)
	for _, z := range bz.Slice() {
		wc[ganzhi.GanWuxing(z.Gan)]++
	}
	for _, h := range hs {
		wc[ganzhi.GanWuxing(h.Main)]++
		if h.Mid != nil {
			wc[ganzhi.GanWuxing(*h.Mid)]++
		}
		if h.Minor != nil {
			wc[ganzhi.GanWuxing(*h.Minor)]++
		}
	}
	return wc
}

func computeShiShensTable(bz ganzhi.Bazi, hs [4]cangGanOut) [4][]shiShenEntry {
	dm := bz.Ri.Gan
	var table [4][]shiShenEntry
	ps := bz.Slice()
	for i := range ps {
		var entries []shiShenEntry
		entries = append(entries, shiShenEntry{
			ShiShen: ganzhi.ShiShenFromGan(dm, ps[i].Gan),
			Name:   ganzhi.GanName(ps[i].Gan),
			Source: sourceGan,
			Gan:    ps[i].Gan,
		})
		entries = append(entries, shiShenEntry{
			ShiShen: ganzhi.ShiShenFromGan(dm, hs[i].Main),
			Name:   ganzhi.GanName(hs[i].Main),
			Source: sourceMainQi,
			Gan:    hs[i].Main,
		})
		if hs[i].Mid != nil {
			entries = append(entries, shiShenEntry{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Mid),
				Name:   ganzhi.GanName(*hs[i].Mid),
				Source: sourceMidQi,
				Gan:    *hs[i].Mid,
			})
		}
		if hs[i].Minor != nil {
			entries = append(entries, shiShenEntry{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Minor),
				Name:   ganzhi.GanName(*hs[i].Minor),
				Source: sourceMinQi,
				Gan:    *hs[i].Minor,
			})
		}
		table[i] = entries
	}
	return table
}

func computeChangShengTable(bz ganzhi.Bazi, hs [4]cangGanOut) [4][]changShengEntry {
	var table [4][]changShengEntry
	for i, z := range bz.Slice() {
		stages, ok := ganzhi.ChangShengTable[z.Gan]
		if !ok {
			continue
		}
		for stageIdx, b := range stages {
			if b == z.Zhi {
				table[i] = []changShengEntry{{
					Stage: ganzhi.StageNamesZH[stageIdx],
					Gan:   z.Gan,
				}}
				break
			}
		}
	}
	return table
}
