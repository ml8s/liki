package bazi

import "liki-engine/internal/engine/ganzhi"

// XiaoYunZhu is a single year of minor fortune (小运).
type XiaoYunZhu struct {
	Age     int        `json:"age"`
	Gan     ganzhi.Gan `json:"gan"`
	Zhi     ganzhi.Zhi `json:"zhi"`
	Name    string     `json:"name"`
	ShiShen string     `json:"shi_shen"`
}

// ComputeXiaoYun computes the minor fortune (小运) zhus for each age starting from 1.
// 起法（《星平会海》）：小运由时柱起，阳男阴女顺行、阴男阳女逆行，一位一年。
// Returns up to maxAge zhus (typically up to 12 for childhood).
func computeXiaoYun(bz ganzhi.Bazi, gender ganzhi.Gender, maxAge int) []XiaoYunZhu {
	riYuan := bz.Ri.Gan
	if maxAge <= 0 {
		maxAge = 12
	}

	startIdx := ganzhi.SixtyCycleIndex(bz.Shi.Gan, bz.Shi.Zhi)
	yearYang := ganzhi.GanYinYang(bz.Nian.Gan) == ganzhi.Yang
	forward := (gender == ganzhi.Male && yearYang) || (gender == ganzhi.Female && !yearYang)

	zhus := make([]XiaoYunZhu, 0, maxAge)
	for age := 1; age <= maxAge; age++ {
		var idx int
		if forward {
			idx = (startIdx + (age - 1)) % 60
		} else {
			idx = (startIdx - (age - 1) + 60) % 60
		}
		zhu := ganzhi.SixtyToZhu(idx)
		name := ganzhi.GanName(zhu.Gan) + ganzhi.ZhiName(zhu.Zhi)

		tg := ganzhi.ShiShenFromGan(riYuan, zhu.Gan)

		zhus = append(zhus, XiaoYunZhu{
			Age:     age,
			Gan:     zhu.Gan,
			Zhi:     zhu.Zhi,
			Name:    name,
			ShiShen: tg.String(),
		})
	}
	return zhus
}
