package xuankong

// Flying-star feng shui time system. Pure arithmetic from base year 1864.
// Each 运 (period) lasts 20 years. Three 运 = one 元 (60 years).
// 上元: 一运(1864) → 二运(1884) → 三运(1904)
// 中元: 四运(1924) → 五运(1944) → 六运(1964)
// 下元: 七运(1984) → 八运(2004) → 九运(2024)

const baseYear = 1864 // 上元一运起始年
const periodLength = 20

// SanYuanYun holds the 三元九运 classification for a given year.
type SanYuanYun struct {
	Year      int    `json:"year"`
	Yuan      string `json:"yuan"`       // "上元"/"中元"/"下元"
	YunNumber int    `json:"yun_number"` // 1-9
	YunName   string `json:"yun_name"`   // "一运"..."九运"
	StartYear int    `json:"start_year"` // this period's start year
	EndYear   int    `json:"end_year"`   // this period's end year
}

// ComputeSanYuanYun determines which 元 and 运 a given year belongs to.
func ComputeSanYuanYun(year int) SanYuanYun {
	if year < baseYear {
		year = baseYear
	}
	yunIdx := (year - baseYear) / periodLength // 0-8+
	yunNum := (yunIdx % 9) + 1

	yuanIdx := yunIdx / 3
	var yuanName string
	switch yuanIdx % 3 {
	case 0:
		yuanName = "上元"
	case 1:
		yuanName = "中元"
	case 2:
		yuanName = "下元"
	default: // beyond 180 years, cycle repeats
		yuanName = "上元"
	}

	startYear := baseYear + yunIdx*periodLength
	endYear := startYear + periodLength - 1

	yunNames := [10]string{"", "一运", "二运", "三运", "四运", "五运", "六运", "七运", "八运", "九运"}

	return SanYuanYun{
		Year:      year,
		Yuan:      yuanName,
		YunNumber: yunNum,
		YunName:   yunNames[yunNum],
		StartYear: startYear,
		EndYear:   endYear,
	}
}