package liuyao

import "liki-engine/internal/engine/ganzhi"

// isLiuChong 判断卦体是否六冲。
// 内卦一至三爻与外卦四至六爻三组对应支全冲，才构成卦体六冲。
func (g guaIndex) isLiuChong() bool {
	// 获取本卦六爻地支
	lines := getHexagramLines(g)
	for i := 0; i < 3; i++ {
		if !ganzhi.IsLiuChong(lines[i], lines[i+3]) {
			return false
		}
	}
	return true
}

// isLiuHe 判断卦体是否六合
// 六合卦：子丑合、寅亥合、卯戌合、辰酉合、巳申合、午未合
func (g guaIndex) isLiuHe() bool {
	// 获取本卦六爻地支
	lines := getHexagramLines(g)
	for i := 0; i < 3; i++ {
		if !ganzhi.IsZhiHe(lines[i], lines[i+3]) {
			return false
		}
	}
	return true
}

// getHexagramLines 按上下经卦纳甲获取卦的六爻地支。
func getHexagramLines(g guaIndex) [6]ganzhi.Zhi {
	var lines [6]ganzhi.Zhi
	upperTri, lowerTri := guaTrigrams(g)
	for i := 0; i < 6; i++ {
		tri := lowerTri
		if i >= 3 {
			tri = upperTri
		}
		pi := trigramPalaceIdx[tri]
		lines[i] = naZhiTable[pi][i%3+3*(i/3)]
	}
	return lines
}
