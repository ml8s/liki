package ziwei

import "sort"

// fillChartAreas completes the canonical chart in domain-owned passes. Each
// function fills one independent chart area; GeJu is recalculated last because
// patterns read both major stars and miscellaneous stars.
func fillChartAreas(chart Chart) Chart {
	chart = fillXiaoXian(chart)
	chart = fillChangSheng(chart)
	chart = fillBoShi(chart)
	chart = fillZaYao(chart)
	chart = fillJiangQianSuiQian(chart)
	chart.Patterns = findPatterns(chart.GongWei)
	chart.DaXian = ComputeDaXian(chart.ZiweiCore())
	return chart
}

func fillXiaoXian(chart Chart) Chart {
	ages := allPalaceXiaoXian(chart.NianZhi, chart.Gender, 10, chart.GongWei[chart.MingGong].Zhi)
	for i := range chart.GongWei {
		chart.GongWei[i].Ages = ages[i]
	}
	return chart
}

func fillChangSheng(chart Chart) Chart {
	stages := computeChangSheng(chart.JuShu, chart.GongWei[chart.MingGong].Zhi, chart.NianGan, chart.Gender)
	for i := range chart.GongWei {
		chart.GongWei[i].ChangSheng = stages[i]
	}
	return chart
}

func fillBoShi(chart Chart) Chart {
	stages := computeBoShi(chart.JuShu, chart.GongWei[chart.MingGong].Zhi, chart.NianGan, chart.Gender, chart.NianZhi)
	for i := range chart.GongWei {
		chart.GongWei[i].BoShi = stages[i]
	}
	return chart
}

func fillZaYao(chart Chart) Chart {
	stars := computeAdjectiveStars(
		chart.NianZhi, chart.ShiZhi, chart.LunarMonth, chart.LunarDay, chart.NianGan,
	)
	for i := range chart.GongWei {
		palaceZhi := chart.GongWei[i].Zhi
		var names []string
		for name, zhiIdx := range stars {
			if zhiIdx == int(palaceZhi)-1 {
				names = append(names, name)
			}
		}
		// map iteration is unordered; the published ZaYao order is canonical.
		sort.Strings(names)
		chart.GongWei[i].ZaYao = names
	}
	return chart
}

func fillJiangQianSuiQian(chart Chart) Chart {
	jiangQian := computeJiangQian(chart.NianZhi, chart.GongWei[chart.MingGong].Zhi)
	suiQian := computeSuiQian(chart.NianZhi, chart.GongWei[chart.MingGong].Zhi)
	for i := range chart.GongWei {
		chart.GongWei[i].JiangQian = jiangQian[i]
		chart.GongWei[i].SuiQian = suiQian[i]
	}
	return chart
}
