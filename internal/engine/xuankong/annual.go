package xuankong

// 流年飞星入中计算.
// 年飞星入中公式(下元): (年尾两位数 + 年尾两位数/4) mod 9
// 余数0则9入中.
func computeNianRuZhong(year int) int {
	tail := year % 100
	n := (tail + tail/4) % 9
	if n == 0 {
		return 9
	}
	return n
}

// 九星顺序: 入中 → 乾 → 兑 → 艮 → 离 → 坎 → 坤 → 震 → 巽

// 八宫(无中宫): 西北=乾=1, 西=兑=2, 东北=艮=3, 南=离=4, 北=坎=5, 西南=坤=6, 东=震=7, 东南=巽=8
var palaceIndexFromFeiXu = [9]int{99, 6, 7, 8, 9, 1, 2, 3, 4} // 中(0)→无, 乾(1)→6, 兑(2)→7, 艮(3)→8, 离(4)→9, 坎(5)→1, 坤(6)→2, 震(7)→3, 巽(8)→4

type AnnualStar struct {
	Palace     int    `json:"palace"`     // 洛书宫位 1-9
	Star       int    `json:"star"`       // 飞星号 1-9
	StarName   string `json:"star_name"`
	Rating     string `json:"rating"` // 吉/凶/平
	RiZhong    bool   `json:"ru_zhong"`   // 是否入中
}

type AnnualBoard struct {
	Year       int            `json:"year"`
	RuZhong    int            `json:"ru_zhong"`    // 入中星
	Stars      []AnnualStar   `json:"stars"`
}

// 九星名称
var starNames9 = [10]string{"", "贪狼", "巨门", "禄存", "文曲", "廉贞", "武曲", "破军", "左辅", "右弼"}
var starRatings = [10]string{"", "吉", "凶", "凶", "平", "大凶", "吉", "凶", "大吉", "吉"}

// ComputeAnnual computes the annual飞星 board.
func ComputeAnnual(year int) AnnualBoard {
	rz := computeNianRuZhong(year)
	var stars []AnnualStar
	for i := 0; i < 9; i++ {
		starNum := (rz - 1 + i) % 9 + 1
		pi := palaceIndexFromFeiXu[i]
		if i == 0 {
			pi = 5 // 中宫
		}
		stars = append(stars, AnnualStar{
			Palace:   pi,
			Star:     starNum,
			StarName: starNames9[starNum],
			Rating:   starRatings[starNum],
			RiZhong:  i == 0,
		})
	}
	return AnnualBoard{
		Year:    year,
		RuZhong: rz,
		Stars:   stars,
	}
}
