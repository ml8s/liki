package bazi

import "liki-engine/internal/engine/ganzhi"

// XiaoYunZhu is a single year of minor fortune (小运).
type XiaoYunZhu struct {
	Age    int    `json:"age"`
	Gan ganzhi.Gan    `json:"gan"`
	Zhi ganzhi.Zhi    `json:"zhi"`
	Name   string `json:"name"`
	ShiShen string `json:"shi_shen"`
}

// ComputeXiaoYun computes the minor fortune (小运) zhus for each age starting from 1.
// ganzhi.Male: start from 丙寅 (stem=3, branch=3) and go forward.
// ganzhi.Female: start from 壬申 (stem=9, branch=9) and go backward.
// Returns up to maxAge zhus (typically up to 12 for childhood).
func computeXiaoYun(bz ganzhi.Bazi, gender ganzhi.Gender, maxAge int) []XiaoYunZhu {
	riYuan := bz.Ri.Gan
	if maxAge <= 0 {
		maxAge = 12
	}

	var startIdx int
	if gender == ganzhi.Male {
		startIdx = ganzhi.SixtyCycleIndex(3, 3) // 丙寅
	} else {
		startIdx = ganzhi.SixtyCycleIndex(9, 9) // 壬申
	}

	zhus := make([]XiaoYunZhu, 0, maxAge)
	for age := 1; age <= maxAge; age++ {
		var idx int
		if gender == ganzhi.Male {
			idx = (startIdx + (age - 1)) % 60
		} else {
			idx = (startIdx - (age - 1) + 60) % 60
		}
		zhu := ganzhi.SixtyToZhu(idx)
		name := ganzhi.GanName(zhu.Gan) + ganzhi.ZhiName(zhu.Zhi)

		tg := ganzhi.ShiShenFromGan(riYuan, zhu.Gan)

		zhus = append(zhus, XiaoYunZhu{
			Age:    age,
			Gan:    zhu.Gan,
			Zhi:    zhu.Zhi,
			Name:   name,
			ShiShen: tg.String(),
		})
	}
	return zhus
}
