package ziwei

import "liki-engine/internal/engine/ganzhi"

func daXianQiSui(ju juShu) int {
	return int(ju)
}

func ComputeDaXian(chart Chart) []DaXianStep {
	forward := isDaXianForward(chart.Gender, chart.NianGan)
	qiSui := daXianQiSui(chart.JuShu)
	steps := make([]DaXianStep, 0, 12)
	pos := gongIndex(0)
	// 大限起于出生年：虚岁 QiSui → 公历年 = birthYear + QiSui − 1（虚岁 1 岁 = 出生年）。
	startY := chart.BirthYear + qiSui - 1
	for i := 0; i < 12; i++ {
		steps = append(steps, DaXianStep{
			QiSui:     qiSui + i*10,
			ZhiSui:    qiSui + i*10 + 9,
			StartYear: startY + i*10,
			EndYear:   startY + i*10 + 9,
			Gong:      pos,
			Name:      gongLabels[pos],
		})
		if forward {
			pos = (pos + 11) % 12 // 经典顺行→逆Liki序往后走
		} else {
			pos = (pos + 1) % 12 // 经典逆行→顺Liki序往前走
		}
	}
	return steps
}

func isDaXianForward(gender ganzhi.Gender, nianGan Gan) bool {
	isYang := int(nianGan)%2 == 1
	isMale := gender == Male
	// 阳男阴女→顺行（与长生方向相反）
	return (isMale && isYang) || (!isMale && !isYang)
}
