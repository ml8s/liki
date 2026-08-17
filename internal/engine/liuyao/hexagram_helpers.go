package liuyao

import "liki-engine/internal/engine/ganzhi"

// isLiuChong 判断卦体是否六冲
// 六冲卦：子午冲、丑未冲、寅申冲、卯酉冲、辰戌冲、巳亥冲
func (g guaIndex) isLiuChong() bool {
	// 获取本卦六爻地支
	lines := getHexagramLines(g)
	if len(lines) < 6 {
		return false
	}

	// 检查相邻爻是否相冲（初爻与四爻、二爻与五爻、三爻与上爻）
	for i := 0; i < 3; i++ {
		if ganzhi.IsLiuChong(lines[i], lines[i+3]) {
			return true
		}
	}
	return false
}

// isLiuHe 判断卦体是否六合
// 六合卦：子丑合、寅亥合、卯戌合、辰酉合、巳申合、午未合
func (g guaIndex) isLiuHe() bool {
	// 获取本卦六爻地支
	lines := getHexagramLines(g)
	if len(lines) < 6 {
		return false
	}

	// 检查相邻爻是否相合（初爻与四爻、二爻与五爻、三爻与上爻）
	for i := 0; i < 3; i++ {
		if ganzhi.IsZhiHe(lines[i], lines[i+3]) {
			return true
		}
	}
	return false
}

// getHexagramLines 获取卦的六爻地支
func getHexagramLines(g guaIndex) [6]ganzhi.Zhi {
	// 这里需要从卦象数据中获取六爻地支
	// 简化实现：返回空数组，实际应该从数据表中查找
	var lines [6]ganzhi.Zhi
	return lines
}
