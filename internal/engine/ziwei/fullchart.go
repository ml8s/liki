package ziwei

// ComputeFullChart computes a full chart with all extended info.
func ComputeFullChart(chart Chart, riGan, riZhi int) Chart {
	mingZhi := chart.Palaces[chart.MingGong].Zhi
	nianZhi := chart.NianZhi
	shiZhi := chart.HourZhi
	lunarMonth := chart.LunarMonth
	lunarDay := chart.LunarDay
	nianGan := chart.YearGan // use existing field name
	gender := chart.Gender

	// 1. XiaoXian
	ages := allPalaceXiaoXian(nianZhi, gender, 10, mingZhi)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].Ages = ages[i]
	}

	// 2. ChangSheng
	cs := computeChangSheng(chart.JuShu, mingZhi, nianGan, gender)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].ChangSheng = cs[i]
	}

	// 3. BoShi
	bs := computeBoShi(chart.JuShu, mingZhi, nianGan, gender, nianZhi)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].BoShi = bs[i]
	}

	// 4. Adjective stars
	adjMap := computeAdjectiveStars(nianZhi, shiZhi, mingZhi, lunarMonth, lunarDay, nianGan, gender, riGan, riZhi, 0, 0)
	for palaceIdx := 0; palaceIdx < 12; palaceIdx++ {
		palaceZhi := chart.Palaces[palaceIdx].Zhi
		var stars []string
		for name, zhiMinus1 := range adjMap {
			if zhiMinus1 == int(palaceZhi)-1 {
				stars = append(stars, name)
			}
		}
		chart.Palaces[palaceIdx].AdjStars = stars
	}

	// 5. JiangQian / SuiQian
	jq := computeJiangQian(nianZhi, mingZhi)
	sq := computeSuiQian(nianZhi, mingZhi)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].JiangQian = jq[i]
		chart.Palaces[i].SuiQian = sq[i]
	}

	return chart
}
