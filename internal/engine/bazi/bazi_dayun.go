package bazi

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// cstLocation 八字排盘基准时区（北京时间）。
var cstLocation = time.FixedZone("CST", 8*3600)

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
	AgeEnd   int         `json:"zhi_sui"`
	Name     string      `json:"name"`
	Element  string      `json:"wuxing"`
	ShiShen   string      `json:"shi_shen"`
}

// DaYun holds the big fortune (大运) cycle for a bazi chart.
type DaYun struct {
	StartAge         int           `json:"start_age"`          // 起运虚岁（四舍五入）
	StartYearAfter   int           `json:"start_year_after"`   // 出生后整年数（对齐 lunar）
	StartMonthAfter  int           `json:"start_month_after"`  // 余月
	StartDayAfter    int           `json:"start_day_after"`    // 余日
	Direction        string        `json:"direction"`
	Steps            []DaYunStep   `json:"steps"`
	CurrentStepIndex int           `json:"current_step_index"`
	// 距下一大运剩余年数（虚岁口径，与 steps 的 qi_sui/zhi_sui 一致）；
	// 未起运/已过完所有大运 → null（omitempty 缺席）。
	NextStepInYears *int `json:"next_step_in_years,omitempty"`
}

type daYunSteps struct {
	startAge  int
	startY    int
	startM    int
	startD    int
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

	// 起运精确换算（对齐 lunar Yun sect=1）：
	// dayDiff = 整天差（顺排=出生→下一节；逆排=上一节→出生，lunar Solar.subtract 纯日期差语义），
	// hourDiff = 节时辰 - 出生时辰（负则 +12 且 dayDiff--），
	// month = dayDiff*4 + floor(hourDiff*10/30)；day = hourDiff*10 - monthDiff*30；year = month/12。
	// 注：起运年/月/日按上述 lunar 公式（3 天 = 1 年，1 天 = 4 个月）；
	// startAge（虚岁）用真实时刻差 qe.Sub(qs)（含时分），与 dayDiff 的纯日期差语义不同。
	var qs, qe time.Time
	if forward {
		qs, qe = birthTime, targetJie
	} else {
		qs, qe = targetJie, birthTime
	}
	dayDiff := calendarDays(qs, qe)
	hourDiff := hourZhiIndex(qe) - hourZhiIndex(qs)
	if hourDiff < 0 {
		hourDiff += 12
		dayDiff--
	}
	monthDiff := hourDiff * 10 / 30
	totalM := dayDiff*4 + monthDiff
	startD := hourDiff*10 - monthDiff*30
	startY := totalM / 12
	startM := totalM % 12
	days := qe.Sub(qs).Hours() / 24
	startAge := int(days/3 + 0.5)

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
		startY:    startY,
		startM:    startM,
		startD:    startD,
		direction: dir,
		steps:     steps,
	}
}

// computeDaYun computes the labeled big fortune (大运) steps.
func computeDaYun(st tianwen.SolarTime, month ganzhi.Zhu, nianGan, riGan ganzhi.Gan, gender ganzhi.Gender) *DaYun {
	bf := computeDaYunSteps(st, month, nianGan, gender)
	r := &DaYun{
		StartAge:         bf.startAge,
		StartYearAfter:   bf.startY,
		StartMonthAfter:  bf.startM,
		StartDayAfter:    bf.startD,
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

// hourZhiIndex 时辰地支索引（子=0..亥=11，23 点=11）。
func hourZhiIndex(t time.Time) int {
	// 八字时辰按北京时间（+8）判定
	local := t.In(cstLocation)
	h := local.Hour()
	if h == 23 {
		return 11
	}
	return (h + 1) / 2 % 12
}

// calendarDays 纯日期差（忽略时刻，lunar Solar.subtract 语义）。
func calendarDays(a, b time.Time) int {
	// 统一按北京时间取日期（八字排盘基准）
	cst := cstLocation
	ay, am, ad := a.In(cst).Date()
	by, bm, bd := b.In(cst).Date()
	ta := time.Date(ay, am, ad, 0, 0, 0, 0, cst)
	tb := time.Date(by, bm, bd, 0, 0, 0, 0, cst)
	return int(tb.Sub(ta).Hours() / 24)
}
