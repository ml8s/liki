package ziwei

import "liki-engine/internal/engine/ganzhi"
func ComputeDaXian(chart Chart) []DaXianStep {
	forward := isDaXianForward(chart.Gender, chart.YearGan)
	startAge := int(chart.JuShu)
	steps := make([]DaXianStep, 0, 12)
	pos := palaceIndex(0)
	for i := 0; i < 12; i++ {
		steps = append(steps, DaXianStep{
			StartAge: startAge + i*10,
			EndAge:   startAge + i*10 + 9,
			Palace:   pos,
			Name:     PalaceNames[pos],
		})
		if forward {
			pos = (pos + 1) % 12
		} else {
			pos = (pos + 11) % 12
		}
	}
	return steps
}

func isDaXianForward(gender ganzhi.Gender, yearGan Gan) bool {
	isYang := int(yearGan)%2 == 1
	isMale := gender == Male
	return (isMale && isYang) || (!isMale && !isYang)
}
