package ganzhi

// anHePairs lists 地支暗合 pairs (寅丑, 卯申, 午亥, 子戌).
var anHePairs = []BranchPair{
	{A: 3, B: 2},  // 寅丑
	{A: 4, B: 9},  // 卯申
	{A: 7, B: 12}, // 午亥
	{A: 1, B: 11}, // 子戌
}

// poPairs lists 地支相破 pairs (子酉, 寅亥, 辰丑, 午卯, 申巳, 戌未).
var poPairs = []BranchPair{
	{A: 1, B: 10}, // 子酉
	{A: 3, B: 12}, // 寅亥
	{A: 5, B: 2},  // 辰丑
	{A: 7, B: 4},  // 午卯
	{A: 9, B: 6},  // 申巳
	{A: 11, B: 8}, // 戌未
}

func inBranchList(branches []Zhi, b Zhi) bool {
	for _, x := range branches {
		if x == b {
			return true
		}
	}
	return false
}

// -- stem interactions --

// IsGanHe returns true if the two stems form a 天干五合 pair.
func IsGanHe(a, b Gan) bool {
	for _, p := range GanHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsZhiHe returns true if the two branches form a 地支六合 pair.
func IsZhiHe(a, b Zhi) bool {
	for _, p := range ZhiHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsTripleHe returns true if the two branches belong to the same 三合 group.
func IsTripleHe(a, b Zhi) bool {
	for _, tr := range TripleHeList {
		if inBranchList(tr.Branches, a) && inBranchList(tr.Branches, b) {
			return true
		}
	}
	return false
}

// IsTripleHui returns true if the two branches belong to the same 三会 group.
func IsTripleHui(a, b Zhi) bool {
	for _, tr := range TripleHuiList {
		if inBranchList(tr.Branches, a) && inBranchList(tr.Branches, b) {
			return true
		}
	}
	return false
}

// IsLiuChong returns true if the two branches form a 六冲 pair.
func IsLiuChong(a, b Zhi) bool {
	for _, p := range ChongPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsXing returns true if the two branches are in a 相刑 relation.
func IsXing(a, b Zhi) bool {
	for _, x := range XingGroups {
		if x.Type == "zi" {
			if a == b && inBranchList(x.Branches, a) {
				return true
			}
		} else if a != b {
			if inBranchList(x.Branches, a) && inBranchList(x.Branches, b) {
				return true
			}
		}
	}
	return false
}

// IsHai returns true if the two branches form a 六害 pair.
func IsHai(a, b Zhi) bool {
	for _, p := range HaiPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsAnHe returns true if the two branches form a 暗合 pair.
func IsAnHe(a, b Zhi) bool {
	for _, p := range anHePairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}

// IsPo returns true if the two branches form a 相破 pair.
func IsPo(a, b Zhi) bool {
	for _, p := range poPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return true
		}
	}
	return false
}
