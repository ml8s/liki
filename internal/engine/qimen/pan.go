package qimen

import (

	"liki-engine/internal/engine/ganzhi"
)

// computePan builds a pan from bureau info and driving pillar.
func computePan(ju juShu, driveZhu ganzhi.Zhu, riGan ganzhi.Gan) pan {
	dipan := placeDiPan(ju.Number, ju.YinDun)
	duty := findDuty(driveZhu, dipan)
	tianStars, tianStems := placeTianPan(driveZhu, duty.Star, dipan)
	renDoors := placeRenPan(driveZhu.Zhi, duty.Door, ju.YinDun)

	var dutyStarPalace int
	// 值符星为天禽时按天芮处理（天禽寄坤2，与天芮同宫，天盘只列天芮）。
	searchStar := duty.Star
	if duty.Star == StarTianQin {
		searchStar = StarTianRui
	}
	for i, s := range tianStars {
		if s == searchStar {
			dutyStarPalace = i
			break
		}
	}
	shenSpirits := placeShenPan(ju.YinDun, GongIndex(dutyStarPalace+1))

	var dutyDoorPalace int
	for i, d := range renDoors {
		if d == duty.Door {
			dutyDoorPalace = i
			break
		}
	}
	angans := placeAnGan(driveZhu, dutyDoorPalace)

	mata := findMaXing(driveZhu.Zhi)
	kongWang := findKongWang(driveZhu)

	pan := pan{
		Jushu:    ju.Number,
		YinDun:   ju.YinDun,
		DutyStar: duty.Star,
		DutyDoor: duty.Door,
		MaXing:   mata,
		DriveGan:  driveZhu.Gan,
		DriveZhi: driveZhu.Zhi,
		KongWang: kongWang,
		WuBuYuShi: isWuBuYuShi(riGan, driveZhu.Gan),
	}
	for i := 0; i < 9; i++ {
		pan.GongWei[i] = Gong{
			EarthStem:  dipan[i],
			HeavenStem: tianStems[i],
			Star:       tianStars[i],
			Door:       renDoors[i],
			Spirit:     shenSpirits[i],
			HiddenStem: angans[i],
		}
	}
	return pan
}


// isWuBuYuShi checks if the hour stem (时干) controls the day stem (日干)
// with the same yin-yang polarity. If true, it is 五不遇时 — an inauspicious time.
// 五不遇时 = 时干克日干, 阴克阴/阳克阳
// List: 甲庚、乙辛、丙壬、丁癸、戊甲、己乙、庚丙、辛丁、壬戊、癸己
func isWuBuYuShi(riGan, shiGan ganzhi.Gan) bool {
	riWx := ganzhi.GanWuxing(riGan)
	shiWx := ganzhi.GanWuxing(shiGan)
	if !ganzhi.Ke(shiWx, riWx) {
		return false // 时干不克日干
	}
	// Same yin-yang polarity
	return ganzhi.GanYinYang(shiGan) == ganzhi.GanYinYang(riGan)
}
