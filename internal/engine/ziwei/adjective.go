package ziwei

import "liki-engine/internal/engine/ganzhi"

type adjStarResult map[string]int

func computeAdjectiveStars(nianZhi, shiZhi Zhi, mingZhi Zhi, lunarMonth, lunarDay int, nianGan Gan, gender ganzhi.Gender, riGan, riZhi, soulIzTroIdx, shenGongIdx int) adjStarResult {
	r := make(adjStarResult)
	add := func(name string, zhiMinus1 int) { r[name] = zhiMinus1 }

	// A 年支系
	hl := (4 - int(nianZhi) + 12) % 12
	add("红鸾", hl)
	add("天喜", (hl+6)%12)
	for name, table := range nianStars {
		add(name, table[nianZhi-1])
	}

	// B 月支系
	for name, table := range yueStars {
		add(name, table[lunarMonth-1])
	}

	// C 时支系
	for name, table := range shiStars {
		add(name, table[shiZhi-1])
	}

	// D 日系
	zf := (lunarMonth + 3) % 12
	yf := (11 - lunarMonth + 12) % 12
	wc := (11 - int(shiZhi) + 12) % 12
	wq := (int(shiZhi) + 3) % 12
	add("三台", (zf+lunarDay-1)%12)
	add("八座", ((yf-lunarDay+1)%12+12)%12)
	add("恩光", (wc+lunarDay-2+12)%12)
	add("天贵", (wq+lunarDay-2+12)%12)

	// E 年干系
	for name, table := range ganStars {
		add(name, table[nianGan-1])
	}

	// soul/body for PALACE conversion
	soulIdx := (lunarMonth - 1 - earthlyIdxTable[shiZhi-1] + 12) % 12
	bodyIdx := (lunarMonth - 1 + earthlyIdxTable[shiZhi-1] + 12) % 12

	// F 复合系 — 年系旬空 (iztro公式)
	add("旬空", xunKongNianXi(nianZhi, nianGan))
	add("年解", nianJiePos(nianZhi))

	add("天才", displayToZhiMinus1((soulIdx+earthlyIdxTable[nianZhi-1])%12))
	add("天寿", displayToZhiMinus1((bodyIdx+earthlyIdxTable[nianZhi-1])%12))
	add("天伤", displayToZhiMinus1((5+soulIdx)%12))
	add("天使", displayToZhiMinus1((7+soulIdx)%12))

	return r
}

func displayToZhiMinus1(displayIdx int) int { return (2 + displayIdx) % 12 }

func nianJiePos(nianZhi Zhi) int {
	e := earthlyIdxTable[nianZhi-1]
	return (10 - e + 12) % 12
}

var liushiJiaZi = [10][12]int{
	{1, -1, 51, -1, 41, -1, 31, -1, 21, -1, 11, -1},
	{-1, 2, -1, 52, -1, 42, -1, 32, -1, 22, -1, 12},
	{13, -1, 3, -1, 53, -1, 43, -1, 33, -1, 23, -1},
	{-1, 14, -1, 4, -1, 54, -1, 44, -1, 34, -1, 24},
	{25, -1, 15, -1, 5, -1, 55, -1, 45, -1, 35, -1},
	{-1, 26, -1, 16, -1, 6, -1, 56, -1, 46, -1, 36},
	{37, -1, 27, -1, 17, -1, 7, -1, 57, -1, 47, -1},
	{-1, 38, -1, 28, -1, 18, -1, 8, -1, 58, -1, 48},
	{49, -1, 39, -1, 29, -1, 19, -1, 9, -1, 59, -1},
	{-1, 50, -1, 40, -1, 30, -1, 20, -1, 10, -1, 60},
}



// xunKongNianXi computes 旬空 using year-stem-branch formula (iztro年系).
func xunKongNianXi(nianZhi Zhi, nianGan Gan) int {
	yinIdx := (int(nianZhi) - 3 + 12) % 12 // EARTHLY_BRANCHES→寅0系
	xk := yinIdx + 9 - (int(nianGan)-1) + 1
	xk = (xk + 12) % 12
	if (int(nianZhi)-1)%2 != xk%2 { // 年支阴阳 vs 索引阴阳
		xk = (xk + 1) % 12
	}
	return (2 + xk) % 12 // displayToZhiMinus1
}
