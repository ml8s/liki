package bazi

import "liki-engine/internal/engine/ganzhi"

// XiaoXian is one year in the 小限 (minor limit) cycle.
type XiaoXian struct {
	Age  int        `json:"age"`
	Zhi ganzhi.Zhi `json:"branch"`
}

// ComputeXiaoXian computes the 小限 branch for each year of age up to maxAge.
//
// Rule: male starts at 寅(3) and moves forward one branch per year;
// female starts at 申(9) and moves backward one branch per year.
func ComputeXiaoXian(gender ganzhi.Gender, maxAge int) []XiaoXian {
	if maxAge <= 0 {
		maxAge = 12
	}

	var start int
	dir := 1
	if gender == ganzhi.Male {
		start = 3 // 寅
	} else {
		start = 9 // 申
		dir = -1
	}

	entries := make([]XiaoXian, maxAge)
	for age := 1; age <= maxAge; age++ {
		offset := (age - 1) * dir
		branch := (start-1+offset)%12 + 1
		if branch <= 0 {
			branch += 12
		}
		entries[age-1] = XiaoXian{Age: age, Zhi: ganzhi.Zhi(branch)}
	}
	return entries
}
