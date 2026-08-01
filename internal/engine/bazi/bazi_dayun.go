package bazi

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// DaYun direction constants.
const (
	DirShunPai = "顺排"
	DirNiPai   = "逆排"
)

// DaYunStep holds one 10-year fortune step in the big fortune cycle.
type DaYunStep struct {
	Gan      ganzhi.Gan  `json:"gan"`
	Zhi      ganzhi.Zhi  `json:"zhi"`
	AgeStart int         `json:"qi_sui"`
	AgeEnd   int         `json:"age_end"`
	Name     string      `json:"name"`
	Element  string      `json:"element"`
	ShiShen   string      `json:"shi_shen"`
}

// DaYun holds the big fortune (大运) cycle for a bazi chart.
type DaYun struct {
	StartAge         int           `json:"start_age"`
	Direction        string        `json:"direction"`
	Steps            []DaYunStep   `json:"steps"`
	CurrentStepIndex int           `json:"current_step_index"`
}

type daYunSteps struct {
	startAge  int
	direction string
	steps     []ganzhi.Zhu
}

func computeDaYunSteps(st tianwen.SolarTime, month ganzhi.Zhu, nianGan ganzhi.Gan, gender ganzhi.Gender) daYunSteps {
	isYang := int(nianGan)%2 == 1
	isMale := gender == ganzhi.Male
	forward := (isMale && isYang) || (!isMale && !isYang)

	birthTime := st.Time()
	birthYear := birthTime.Year()
	jz := tianwen.JianYue(tianwen.GregorianTime(birthTime))
	mi := (int(jz) + 9) % 12 // 0=寅月..11=丑月

	// The jie index in JieQiLongitudes for month mi is mi*2.
	jieIdx := mi * 2
	var targetJie time.Time
	var dir string
	if forward {
		nextIdx := ((mi + 1) % 12) * 2
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[nextIdx])
		if !targetJie.After(birthTime) {
			targetJie = tianwen.SolarTermTime(birthYear+1, tianwen.JieQiLongitudes[nextIdx])
		}
		dir = DirShunPai
	} else {
		// 逆排：目标节 = 出生时间之前最近的节（可能在前一年）
		targetJie = tianwen.SolarTermTime(birthYear, tianwen.JieQiLongitudes[jieIdx])
		if targetJie.After(birthTime) {
			// 当前年的节在出生后 → 取前一年的同名节
			// 但若出生在年初（如立春前），前一节可能在上一年年末
			targetJie = tianwen.SolarTermTime(birthYear-1, tianwen.JieQiLongitudes[jieIdx])
			if targetJie.After(birthTime) {
				// 仍在其后（极端）：向前找更早的节
				targetJie = tianwen.SolarTermTime(birthYear-1, tianwen.JieQiLongitudes[(jieIdx+22)%24])
			}
		}
		dir = DirNiPai
	}

	days := targetJie.Sub(birthTime).Hours() / 24
	if days < 0 {
		days = -days
	}
	startAge := int(days/3 + 0.5) // 3 days = 1 year, round to nearest

	// Generate 8 steps from month pillar.
	monthIdx := ganzhi.SixtyCycleIndex(month.Gan, month.Zhi)
	steps := make([]ganzhi.Zhu, 0, 9)
	for i := 1; i <= 9; i++ {
		var idx int
		if forward {
			idx = (monthIdx + i) % 60
		} else {
			idx = (monthIdx - i + 60) % 60
		}
		g := ganzhi.Gan((idx % 10) + 1)
		z := ganzhi.Zhi((idx % 12) + 1)
		steps = append(steps, ganzhi.Zhu{Gan: g, Zhi: z})
	}

	return daYunSteps{
		startAge:  startAge,
		direction: dir,
		steps:     steps,
	}
}

// computeDaYun computes the labeled big fortune (大运) steps.
func computeDaYun(st tianwen.SolarTime, month ganzhi.Zhu, nianGan, riGan ganzhi.Gan, gender ganzhi.Gender) *DaYun {
	bf := computeDaYunSteps(st, month, nianGan, gender)
	r := &DaYun{
		StartAge:  bf.startAge,
		Direction: bf.direction,
	}
	for i, step := range bf.steps {
		ageStart := bf.startAge + i*10
		r.Steps = append(r.Steps, DaYunStep{
			Gan:      step.Gan,
			Zhi:      step.Zhi,
			AgeStart: ageStart,
			AgeEnd:   ageStart + 9,
			Name:     ganzhi.GanName(step.Gan) + ganzhi.ZhiName(step.Zhi),
			Element:  ganzhi.GanWuxing(step.Gan).String(),
			ShiShen:   daYunShiShenLabel(riGan, step.Gan),
		})
	}
	return r
}

func daYunShiShenLabel(riYuan, other ganzhi.Gan) string {
	if tg := ganzhi.ShiShenFromGan(riYuan, other).String(); tg != "" {
		return tg + "运"
	}
	return "未知运"
}
