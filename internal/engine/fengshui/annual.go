package fengshui

// AnnualFlyingStar holds one star's position in the 紫白流年飞星 board.
type AnnualFlyingStar struct {
	GongNum  int    `json:"gong_num"`
	Xing     int    `json:"xing"`
	XingName string `json:"xing_name"`
	Wuxing   string `json:"wuxing"`
	Rating   string `json:"rating"`
	RuZhong  bool   `json:"ru_zhong"`
}

// AnnualBoard is the 紫白流年飞星 board for one year.
type AnnualBoard struct {
	Year    int                `json:"year"`
	RuZhong string             `json:"ru_zhong"`
	GongWei []AnnualFlyingStar `json:"gong_wei"`
}

// StarRatings maps star number (1-9) to its 吉凶 rating (五档：大吉/吉/平/凶/大凶).
var StarRatings = [10]string{"", "吉", "凶", "凶", "平", "大凶", "吉", "凶", "大吉", "吉"}

// ComputeAnnualFlyingStars computes the 紫白流年飞星 board for a year.
//
// 口诀（《沈氏玄空》）：上元甲子(1864)一白入中，中元甲子(1924)四绿入中，
// 下元甲子(1984)七赤入中，逐年逆行。1864 之前的年份按 60 年甲子周期向前回推。
// 这是八宅「流年紫白」与玄空「流年飞星」的唯一共享实现（此前 bazhai 与
// xuankong 各有一套公式，且对甲子年入中星结果不一致，已统一为本口诀）。
func ComputeAnnualFlyingStars(year int) AnnualBoard {
	var jiaZiYear, jiaZiStar int
	switch {
	case year >= 1984:
		jiaZiYear, jiaZiStar = 1984, 7 // 下元甲子七赤
	case year >= 1924:
		jiaZiYear, jiaZiStar = 1924, 4 // 中元甲子四绿
	case year >= 1864:
		jiaZiYear, jiaZiStar = 1864, 1 // 上元甲子一白
	default:
		cycle := (1864 - year) / 60
		if (1864-year)%60 != 0 {
			cycle++
		}
		jiaZiYear = 1864 - cycle*60
		nBack := (1864 - jiaZiYear) / 60
		phase := (3 - nBack%3) % 3
		jiaZiStar = []int{1, 4, 7}[phase]
	}

	diff := year - jiaZiYear
	centerNum := (jiaZiStar - diff%9 + 9) % 9
	if centerNum == 0 {
		centerNum = 9
	}

	centerStar := StarByNumber(centerNum)
	board := AnnualBoard{
		Year:    year,
		RuZhong: centerStar.Name,
		GongWei: make([]AnnualFlyingStar, 0, 9),
	}

	board.GongWei = append(board.GongWei, AnnualFlyingStar{
		GongNum:  5,
		Xing:     centerStar.Number,
		XingName: centerStar.Name,
		Wuxing:   centerStar.Element.String(),
		Rating:   StarRatings[centerStar.Number],
		RuZhong:  true,
	})

	for i, pn := range LuoshuFlyOrder {
		starNum := (centerNum + i + 1) % 9
		if starNum == 0 {
			starNum = 9
		}
		s := StarByNumber(starNum)
		board.GongWei = append(board.GongWei, AnnualFlyingStar{
			GongNum:  pn,
			Xing:     s.Number,
			XingName: s.Name,
			Wuxing:   s.Element.String(),
			Rating:   StarRatings[s.Number],
			RuZhong:  false,
		})
	}

	return board
}
