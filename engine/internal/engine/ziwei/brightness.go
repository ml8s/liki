package ziwei

// brightness level: 0=庙 1=旺 2=利 3=平 4=陷.
type brightness int

const (
	Miao brightness = iota
	Wang
	De
	Li
	Ping
	Xian
	Bu
)

var brightnessNames = [7]string{"庙", "旺", "得", "利", "平", "陷", "不"}

func (b brightness) String() string { return brightnessNames[b] }

var miaoWangTable [14][12]brightness

func brightnessFrom(s string) brightness {
	for i, name := range brightnessNames {
		if name == s {
			return brightness(i)
		}
	}
	return Ping
}

// minorBrightnessTable 文昌/文曲 × 12 地支亮度（iztro minor_star_brightness——golden 见 testdata/minor_brightness_golden.json）
// 行序：文昌, 文曲；列序：子丑寅卯辰巳午未申酉戌亥
var minorBrightnessTable = [2][12]brightness{
	{De, Miao, Xian, Li, De, Miao, Xian, Li, De, Miao, Xian, Li},       // 文昌：得庙陷利得庙陷利得庙陷利
	{De, Miao, Ping, Wang, De, Miao, Xian, Wang, De, Miao, Xian, Wang}, // 文曲：得庙平旺得庙陷旺得庙陷旺
}

func miaoWang(star starIndex, zhi Zhi) brightness {
	if star < 0 || int(star) >= 14 {
		// 文昌/文曲有亮度表（iztro minor_star_brightness）——其余辅星无表返回平
		if star == WenChang || star == WenQu {
			idx := 0
			if star == WenQu {
				idx = 1
			}
			return minorBrightnessTable[idx][zhi-1]
		}
		return Ping
	}
	// Convert Liki Zhi (子=1) to iztro 安星索引 (寅=0 安星序)
	anXingIdx := (zhi + 9) % 12
	if int(anXingIdx) < 0 || int(anXingIdx) >= 12 {
		return Ping
	}
	return miaoWangTable[star][anXingIdx]
}
