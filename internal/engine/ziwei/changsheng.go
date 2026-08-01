package ziwei

import "liki-engine/internal/engine/ganzhi"

// changShengCycle is the fixed order of 长生12神.
var changShengCycle = []string{
	"长生", "沐浴", "冠带", "临官", "帝旺", "衰",
	"病", "死", "墓", "绝", "胎", "养",
}

// juChangShengZhi returns the earthly branch (1-12) where "长生" starts for a given bureau.
func juChangShengZhi(ju juShu) Zhi {
	switch ju {
	case JuWater:
		return 9  // 申
	case JuWood:
		return 12 // 亥
	case JuMetal:
		return 6  // 巳
	case JuEarth:
		return 9  // 申
	case JuFire:
		return 3  // 寅
	}
	return 0
}

// computeChangSheng computes the 长生12神 for all 12 palaces.
// Returns an array indexed by gongIndex (0=命宫..11=父母).
// Direction: 阳男阴女→顺行, 阴男阳女→逆行 (opposite of xiaoXian).
func computeChangSheng(ju juShu, mingZhi Zhi, nianGan Gan, gender ganzhi.Gender) [12]string {
	startZhi := juChangShengZhi(ju)
	startPalace := int((int(mingZhi) - int(startZhi) + 12) % 12)

	// 阳男阴女→逆行(-1), 阴男阳女→顺行(1)
	ganYang := int(nianGan)%2 == 1
	isMale := gender == Male
	same := (ganYang && isMale) || (!ganYang && !isMale)
	dir := 1
	if same {
		dir = -1 // 阴阳相同→逆行
	}

	var result [12]string
	for i := 0; i < 12; i++ {
		var idx int
		if dir == -1 { // 逆行: 从起始宫递减
			idx = (startPalace - i + 12) % 12
		} else { // 顺行: 从起始宫递增
			idx = (startPalace + i) % 12
		}
		result[idx] = changShengCycle[i]
	}
	return result
}
