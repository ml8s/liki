package qimen

import "liki-engine/internal/engine/ganzhi"

// computePan builds a chart from its table-driven plates and method geometry.
func computePan(ju juShu, bz ganzhi.Bazi) pan {
	dipan := placeDiPan(ju.Number, ju.YinDun)
	leadZhu := leadPillar(bz, ju)
	duty := findDuty(leadZhu, dipan)
	var tianPan [9][]TianPanSymbol
	var renDoors [9]DoorIndex
	var shenSpirits [9]SpiritIndex
	var hidden [9]ganzhi.Gan
	var hiddenPillars [9]string
	var dutyStarLanding, dutyDoorLanding GongIndex
	switch ju.Method.School {
	case SchoolLuoShuFeiPan:
		tianPan, dutyStarLanding = placeFlyTianPan(leadZhu, duty.Palace, dipan)
		renDoors, dutyDoorLanding = placeFlyRenPan(leadZhu, duty.Palace, ju.YinDun)
		shenSpirits = placeFlyShenPan(ju.YinDun, dutyStarLanding)
		hidden = placeFlyHidden(dipan, duty.Palace, dutyDoorLanding)
	case SchoolMingFaFeiPan:
		duty = findMingFaDuty(leadZhu, dipan)
		tianPan, dutyStarLanding = placeMingFaTianPan(leadZhu, duty.Palace, dipan)
		renDoors, dutyDoorLanding = placeMingFaRenPan(leadZhu, duty.Door, duty.Palace, ju.YinDun)
		shenSpirits = placeMingFaShenPan(ju.YinDun, dutyStarLanding)
		hiddenPillars, hidden = placeMingFaHidden(leadZhu, dutyDoorLanding, ju.YinDun)
	default:
		tianPan, dutyStarLanding = placeTianPan(leadZhu, duty.Star, dipan)
		renDoors, dutyDoorLanding = placeRenPan(leadZhu, duty.Door, ju.YinDun, dipan)
		visibleDutyStar := findStarPalace(pan{GongWei: palacesFromTianPan(tianPan)}, duty.Star)
		shenSpirits = placeShenPan(ju.YinDun, visibleDutyStar)
		hidden = placeHiddenPan(dipan, duty.Palace, dutyDoorLanding)
	}
	result := pan{
		Jushu:          ju.Number,
		School:         ju.Method.School,
		YinDun:         ju.YinDun,
		RiGan:          bz.Ri.Gan,
		RiZhi:          bz.Ri.Zhi,
		NianGan:        bz.Nian.Gan,
		NianZhi:        bz.Nian.Zhi,
		YueGan:         bz.Yue.Gan,
		YueZhi:         bz.Yue.Zhi,
		LeadGan:        leadZhu.Gan,
		LeadZhi:        leadZhu.Zhi,
		DutyStar:       duty.Star,
		DutyDoor:       duty.Door,
		DutyStarPalace: dutyStarLanding,
		DutyDoorPalace: dutyDoorLanding,
		MaXing:         findMaXing(leadZhu.Zhi),
		HourGan:        bz.Shi.Gan,
		HourZhi:        bz.Shi.Zhi,
		KongWang:       findKongWang(leadZhu),
		WuBuYuShi: scopeMethods[ju.Method.Scope].hasFeature("wu_bu_yu_shi") &&
			isWuBuYuShi(bz.Ri.Gan, bz.Shi.Gan),
	}
	for i := 0; i < 9; i++ {
		result.GongWei[i] = Gong{
			Gong:      palaceIdentity(GongIndex(i + 1)),
			DiPanGan:  dipan[i],
			AnGan:     hidden[i],
			TianPan:   tianPan[i],
			Door:      renDoors[i],
			DoorSet:   renDoors[i] != 0,
			Spirit:    shenSpirits[i],
			SpiritSet: shenSpirits[i] != 0,
		}
		if len(tianPan[i]) == 0 && ju.Method.School == SchoolZhuanPan {
			heavenGan := dipan[i]
			result.GongWei[i].TianPanGan = &heavenGan
		}
		result.GongWei[i].HiddenPillar = hiddenPillars[i]
	}
	return result
}

func leadPillar(bz ganzhi.Bazi, ju juShu) ganzhi.Zhu {
	if ju.Method.Scope == ScopeQuarter && ju.Quarter != nil {
		return ju.Quarter.Pillar
	}
	switch scopeMethods[ju.Method.Scope].LeadPillar {
	case "day":
		return bz.Ri
	case "month":
		return bz.Yue
	case "year":
		return bz.Nian
	case "hour":
		return bz.Shi
	default:
		return ganzhi.Zhu{}
	}
}

func palacesFromTianPan(items [9][]TianPanSymbol) [9]Gong {
	var result [9]Gong
	for i := range result {
		result[i].TianPan = items[i]
	}
	return result
}

// isWuBuYuShi marks an hour gan controlling the day gan with the same polarity.
func isWuBuYuShi(dayGan, hourGan ganzhi.Gan) bool {
	return wuBuYuShiTable[[2]ganzhi.Gan{dayGan, hourGan}]
}
