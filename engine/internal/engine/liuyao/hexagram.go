package liuyao

import "liki-engine/internal/engine/ganzhi"

// guaTable holds all 64 hexagrams, ordered by palace (八宫). Loaded from data/hexagrams.json.
// naGanTable maps palace → [内卦干, 外卦干] for 纳甲（乾甲壬、坤乙癸，余宫内外同）.
// naZhiTable maps palace → 6 branch indices（内三爻 + 外三爻）; 装卦按上下经卦取表
//（trigramPalaceIdx 映射），非按本宫。Loaded from data/hexagrams.json.

// dayGanShouOrder returns the六兽 assignment order starting from a given day stem.
func dayGanShouOrder(d ganzhi.Gan) [6]LiuShou {
	// 甲乙起青龙, 丙丁起朱雀, 戊起勾陈, 己起螣蛇, 庚辛起白虎, 壬癸起玄武
	base := int(d)          // 甲(1)乙(2)...
	start := (base - 1) / 2 // 甲乙→0, 丙丁→1, 戊→2, 己→3(shift), 庚辛→4, 壬癸→5
	if base >= 6 {          // 己起螣蛇，跳过钩陈的 pair
		start++
	}
	var order [6]LiuShou
	for i := 0; i < 6; i++ {
		order[i] = LiuShou((start + i) % 6)
	}
	return order
}
