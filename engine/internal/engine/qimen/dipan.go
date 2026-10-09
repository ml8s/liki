package qimen

import "liki-engine/internal/engine/ganzhi"

// placeDiPan arranges the earth plate (地盘) — 三奇六仪 on the 9 palaces.
// 阳遁: 戊起局数宫, 顺排; 阴遁: 戊起局数宫, 逆排.
func placeDiPan(ju int, yinDun bool) [9]ganzhi.Gan {
	var dipan [9]ganzhi.Gan
	start := (ju - 1 + 9) % 9
	for i := 0; i < 9; i++ {
		var pos int
		if yinDun {
			pos = (start - i + 9) % 9
		} else {
			pos = (start + i) % 9
		}
		dipan[pos] = earthGanOrder[i]
	}
	return dipan
}
