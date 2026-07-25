package qimen

// spiritOrder is the standard 8-spirit clockwise order.
var spiritOrder = [8]SpiritIndex{
	SpiritZhiFu, SpiritTengShe, SpiritTaiYin, SpiritLiuHe,
	SpiritGouChen, SpiritZhuQue, SpiritJiuDi, SpiritJiuTian,
}

// placeShenPan arranges the spirit plate: 8 spirits on the 9 palaces.
// 值符神 fits to the same palace as 天盘值符星.
// 阳遁: clockwise; 阴遁: counter-clockwise.
func placeShenPan(yinDun bool, tianStarPos PalaceIndex) [9]SpiritIndex {
	var spirits [9]SpiritIndex

	// 值符星 palace → same position for 值符神.
	start := int(tianStarPos) - 1

	for i, si := 0, 0; si < 8; i++ {
		var pos int
		if yinDun {
			pos = (start - i + 9) % 9
		} else {
			pos = (start + i) % 9
		}
		if pos == 4 {
			continue
		}
		spirits[pos] = spiritOrder[si]
		si++
	}

	return spirits
}
