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

// XiaoYunSet 是一个流派（起法）的整段小运。
type XiaoYunSet struct {
	Method string       `json:"method"` // 流派标识
	Basis  string       `json:"basis"`  // 起法依据
	Zhus   []XiaoYunZhu `json:"zhus"`
}

// ComputeXiaoYun computes minor fortune (小运) by multiple schools, each as its own set
// (like 用神 by school). Returns up to maxAge zhus per school (typically 12 for childhood).
func computeXiaoYun(bz ganzhi.Bazi, gender ganzhi.Gender, maxAge int) []XiaoYunSet {
	riYuan := bz.Ri.Gan
	if maxAge <= 0 {
		maxAge = 12
	}

	build := func(startIdx int, forward bool) []XiaoYunZhu {
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

	// 流派一：《三命通会》男起丙寅顺行、女起壬申逆行（固定）
	key := "female"
	if gender == ganzhi.Male {
		key = "male"
	}
	rule := xiaoYunRules[key]
	sanMing := build(
		ganzhi.SixtyCycleIndex(ganzhi.Gan(rule.StartGan), ganzhi.Zhi(rule.StartZhi)),
		rule.Direction > 0,
	)

	// 流派二：《星平会海》由时柱起，阳男阴女顺、阴男阳女逆
	yearYang := ganzhi.GanYinYang(bz.Nian.Gan) == ganzhi.Yang
	forward := (gender == ganzhi.Male && yearYang) || (gender == ganzhi.Female && !yearYang)
	xingPing := build(ganzhi.SixtyCycleIndex(bz.Shi.Gan, bz.Shi.Zhi), forward)

	return []XiaoYunSet{
		{Method: "san_ming_tong_hui", Basis: "《三命通会》：男起丙寅顺行、女起壬申逆行", Zhus: sanMing},
		{Method: "xing_ping_hui_hai", Basis: "《星平会海》：由时柱起，阳男阴女顺、阴男阳女逆", Zhus: xingPing},
	}
}
