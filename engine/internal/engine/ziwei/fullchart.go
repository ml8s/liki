package ziwei

// ComputeFullChart computes a full chart with all extended info.
func ComputeFullChart(chart Chart, riGan, riZhi int) Chart {
	mingZhi := chart.GongWei[chart.MingGong].Zhi
	nianZhi := chart.NianZhi
	shiZhi := chart.ShiZhi
	lunarMonth := chart.LunarMonth
	lunarDay := chart.LunarDay
	nianGan := chart.NianGan
	gender := chart.Gender

	// 1. XiaoXian
	ages := allPalaceXiaoXian(nianZhi, gender, 10, mingZhi)
	for i := 0; i < 12; i++ {
		chart.GongWei[i].Ages = ages[i]
	}

	// 2. ChangSheng
	cs := computeChangSheng(chart.JuShu, mingZhi, nianGan, gender)
	for i := 0; i < 12; i++ {
		chart.GongWei[i].ChangSheng = cs[i]
	}

	// 3. BoShi
	bs := computeBoShi(chart.JuShu, mingZhi, nianGan, gender, nianZhi)
	for i := 0; i < 12; i++ {
		chart.GongWei[i].BoShi = bs[i]
	}

	// 4. Adjective stars
	adjMap := computeAdjectiveStars(nianZhi, shiZhi, mingZhi, lunarMonth, lunarDay, nianGan, gender, riGan, riZhi, 0, 0)
	for palaceIdx := 0; palaceIdx < 12; palaceIdx++ {
		gongZhi := chart.GongWei[palaceIdx].Zhi
		var stars []string
		for name, zhiIdx := range adjMap {
			if zhiIdx == int(gongZhi)-1 {
				stars = append(stars, name)
			}
		}
		chart.GongWei[palaceIdx].ZaYao = stars
	}

	// 5. JiangQian / SuiQian
	jq := computeJiangQian(nianZhi, mingZhi)
	sq := computeSuiQian(nianZhi, mingZhi)
	for i := 0; i < 12; i++ {
		chart.GongWei[i].JiangQian = jq[i]
		chart.GongWei[i].SuiQian = sq[i]
	}

	return chart
}
