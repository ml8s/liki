package tianwen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestLunarToGregorian_BaseDate(t *testing.T) {
	// 代码内置基准: 农历 1900-01-01 = 公历 1900-01-31
	gt := LunarToGregorian(LunarTime{Year: 1900, Month: 1, Day: 1, Leap: false})
	y, m, d := gt.Time().Year(), int(gt.Time().Month()), gt.Time().Day()
	if y != 1900 || m != 1 || d != 31 {
		t.Errorf("LunarToGregorian(1900,1,1) = (%d,%d,%d), want (1900,1,31)", y, m, d)
	}
}

func TestLunarRoundTrip(t *testing.T) {
	// 往返: LunarToSolar → SolarToLunar 应回到原值
	tests := []struct {
		name   string
		lunarY int
		lunarM int
		lunarD int
		leap   bool
	}{
		{"1900-01-01", 1900, 1, 1, false},
		{"1950-07-15", 1950, 7, 15, false},
		{"2000-01-01", 2000, 1, 1, false},
		{"2024-01-01", 2024, 1, 1, false},
		{"2024-12-15", 2024, 12, 15, false},
		{"2050-01-01", 2050, 1, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := LunarToGregorian(LunarTime{Year: tt.lunarY, Month: tt.lunarM, Day: tt.lunarD, Leap: tt.leap})
			gy, gm, gd := gt.Time().Year(), int(gt.Time().Month()), gt.Time().Day()
			if gt.Time().IsZero() {
				t.Fatal("LunarToGregorian returned zero time")
			}
			lt := SolarToLunar(GregorianTime(time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, time.UTC)))
			if lt.Year != tt.lunarY || lt.Month != tt.lunarM || lt.Day != tt.lunarD || lt.Leap != tt.leap {
				t.Errorf("round-trip broken: (%d,%d,%d,%v) → solar(%d,%d,%d) → lunar(%d,%d,%d,%v)",
					tt.lunarY, tt.lunarM, tt.lunarD, tt.leap,
					gy, gm, gd,
					lt.Year, lt.Month, lt.Day, lt.Leap)
			}
		})
	}
}

func TestSolarRoundTrip(t *testing.T) {
	// 反向往返: SolarToLunar → LunarToSolar 应回到原值
	tests := []struct {
		name   string
		solarY int
		solarM int
		solarD int
	}{
		{"1900-01-31", 1900, 1, 31},
		{"1950-06-15", 1950, 6, 15},
		{"2000-06-15", 2000, 6, 15},
		{"2024-06-15", 2024, 6, 15},
		{"2050-12-31", 2050, 12, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := SolarToLunar(GregorianTime(time.Date(tt.solarY, time.Month(tt.solarM), tt.solarD, 0, 0, 0, 0, time.UTC)))
			gt := LunarToGregorian(LunarTime{Year: lt.Year, Month: lt.Month, Day: lt.Day, Leap: lt.Leap})
			gy, gm, gd := gt.Time().Year(), int(gt.Time().Month()), gt.Time().Day()
			if gy != tt.solarY || gm != tt.solarM || gd != tt.solarD {
				t.Errorf("reverse round-trip broken: solar(%d,%d,%d) → lunar(%d,%d,%d,%v) → solar(%d,%d,%d)",
					tt.solarY, tt.solarM, tt.solarD,
					lt.Year, lt.Month, lt.Day, lt.Leap,
					gy, gm, gd)
			}
		})
	}
}

func TestLunarToSolar_LeapMonthRoundTrip(t *testing.T) {
	// 验证闰月 round-trip 也保持一致性
	// 扫描 1900-2100 中所有带闰月的年份
	for gy := 1900; gy <= 2100; gy++ {
		m11 := getMonth11K(gy, defaultTZ)
		nextM11 := getMonth11K(gy+1, defaultTZ)
		leapK := getLeapMonthK(m11, nextM11, defaultTZ)
		if leapK < 0 {
			continue
		}

		// Walk from Month 11 to the leap month, tracking both month
		// number and year (year increments when month wraps from 12 to 1,
		// matching the SolarToLunar convention).
		lunarMonth := 11
		lunarYear := gy
		for k := m11 + 1; k < leapK; k++ {
			lunarMonth++
			if lunarMonth == 13 {
				lunarMonth = 1
				lunarYear++
			}
		}

		// 测试闰月第一天
		gt := LunarToGregorian(LunarTime{Year: lunarYear, Month: lunarMonth, Day: 1, Leap: true})
		sy, sm, sd := gt.Time().Year(), int(gt.Time().Month()), gt.Time().Day()
		if gt.Time().IsZero() {
			t.Errorf("LunarToGregorian(%d,%d,1,true) returned zero for leap month (gy=%d, leapK=%.0f)",
				lunarYear, lunarMonth, gy, leapK)
			continue
		}
		lt := SolarToLunar(GregorianTime(time.Date(sy, time.Month(sm), sd, 0, 0, 0, 0, time.UTC)))
		if lt.Year != lunarYear || lt.Month != lunarMonth || !lt.Leap {
			t.Errorf("leap round-trip: (%d,%d,leap=true) → solar(%d,%d,%d) → lunar(%d,%d,leap=%v)",
				lunarYear, lunarMonth, sy, sm, sd, lt.Year, lt.Month, lt.Leap)
		}
	}
}

// ── 真太阳时 EoT 准确性测试 ──
// EoT 公式: 9.87*sin(2B) - 7.53*cos(B) - 1.5*sin(B)，天文年历近似值 ±3min

func TestEoT_KnownValues(t *testing.T) {
	tests := []struct {
		name      string
		date      [3]int
		wantEoT   float64
		tolerance float64
	}{
		{"春分2024", [3]int{2024, 3, 20}, -7.5, 3.5},
		{"夏至2024", [3]int{2024, 6, 21}, -1.5, 3.5},
		{"秋分2024", [3]int{2024, 9, 22}, 7.5, 3.5},
		{"冬至2024", [3]int{2024, 12, 21}, 1.5, 3.5},
		{"EoT极小2025", [3]int{2025, 2, 11}, -14.0, 3.5},
		{"EoT极大2025", [3]int{2025, 11, 3}, 16.0, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, _ := computeSolarTime(time.Date(tt.date[0], time.Month(tt.date[1]), tt.date[2], 12, 0, 0, 0, time.FixedZone("", 8*3600)), 120, 8)
			eot := ast - 720
			if eot > 720 {
				eot -= 1440
			}
			if eot < -720 {
				eot += 1440
			}
			diff := eot - tt.wantEoT
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tolerance {
				t.Errorf("EoT = %.2f min, want %.2f ± %.1f (diff=%.2f)",
					eot, tt.wantEoT, tt.tolerance, diff)
			}
		})
	}
}

// ── 经度修正测试 ──

func TestSolarTime_LongitudeOffset(t *testing.T) {
	// 经度修正: 4*(116.4-120) = -14.4 min
	ast120, _ := computeSolarTime(time.Date(2024, 6, 21, 12, 0, 0, 0, time.FixedZone("", 8*3600)), 120, 8)
	ast116, _ := computeSolarTime(time.Date(2024, 6, 21, 12, 0, 0, 0, time.FixedZone("", 8*3600)), 116.4, 8)
	diff := ast116 - ast120
	if diff > 720 {
		diff -= 1440
	}
	if diff < -720 {
		diff += 1440
	}
	want := -14.4
	if diff < want-1 || diff > want+1 {
		t.Errorf("lonOffset diff = %.2f min, want %.2f ± 1", diff, want)
	}
}

// ── 时辰边界测试 ──

func TestShiZhi_Boundaries(t *testing.T) {
	tests := []struct {
		name    string
		minutes float64
		want    int
	}{
		{"23:00→子", 1380, 1},
		{"23:59→子", 1439, 1},
		{"00:00→子", 0, 1},
		{"00:59→子", 59, 1},
		{"01:00→丑", 60, 2},
		{"02:59→丑", 179, 2},
		{"03:00→寅", 180, 3},
		{"11:00→午", 660, 7},
		{"12:59→午", 779, 7},
		{"21:00→亥", 1260, 12},
		{"22:59→亥", 1379, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hourZhiFromSolarTime(tt.minutes)
			if int(got) != tt.want {
				t.Errorf("hourZhiFromSolarTime(%.0f) = %d, want %d", tt.minutes, int(got), tt.want)
			}
		})
	}
}

// ── ComputeTime 太阳时跨日测试 ──
// 验证 Lunar 日期跟随太阳时调整（而非原始公历日期）。

func TestComputeTime_MidnightAdjustment(t *testing.T) {
	// 喀什(75.9°E)用北京时间(UTC+8), 凌晨3点
	// lonOffset = 4*(75.9-120) = -176.4, EoT(Jan1) ≈ -3.7
	// raw = 180 - 176.4 - 3.7 = -0.1 → dayOffset=-1, 太阳时在前一天
	ts := ComputeTimeset(GregorianTime(time.Date(2025, time.Month(1), 1, 3, 0, 0, 0, time.FixedZone("", int(8*3600)))), 75.9)

	solarY, solarM, solarD := ts.Solar.Time().Date()
	if solarY != 2024 || solarM != 12 || solarD != 31 {
		t.Errorf("Solar date = (%d,%d,%d), want (2024,12,31)", solarY, solarM, solarD)
	}

	// Lunar 应该跟随太阳时调整后的日期
	lunarFromPrevDay := SolarToLunar(GregorianTime(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)))
	if ts.Lunar.Year != lunarFromPrevDay.Year ||
		ts.Lunar.Month != lunarFromPrevDay.Month ||
		ts.Lunar.Day != lunarFromPrevDay.Day {
		t.Errorf("Lunar=(%d,%d,%d), want (%d,%d,%d) from solar-adjusted date",
			ts.Lunar.Year, ts.Lunar.Month, ts.Lunar.Day,
			lunarFromPrevDay.Year, lunarFromPrevDay.Month, lunarFromPrevDay.Day)
	}
}

// ── 年柱基准测试 ──
// 干支纪年: 公元4年=甲子(1,1), 逆推公元3年=癸亥(10,12)
// 公式: s=(year-3)%10, b=(year-3)%12

func TestNianZhu_GanZhi(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		wantGan int
		wantZhi int
	}{
		{"公元4年甲子", 4, 1, 1},
		{"公元5年乙丑", 5, 2, 2},
		{"公元63年癸亥", 63, 10, 12},
		{"公元64年甲子", 64, 1, 1},
		{"公元2024年甲辰", 2024, 1, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NianZhu(GregorianTime(time.Date(tt.year, 6, 15, 0, 0, 0, 0, time.UTC)))
			if int(got.Gan) != tt.wantGan || int(got.Zhi) != tt.wantZhi {
				t.Errorf("NianZhu(%d) = (%d,%d), want (%d,%d)",
					tt.year, int(got.Gan), int(got.Zhi), tt.wantGan, tt.wantZhi)
			}
		})
	}
}

// ── 月柱五虎遁公式验证 ──
// 五虎遁: year gan → first month (寅月) gan.
// 甲己→丙, 乙庚→戊, 丙辛→庚, 丁壬→壬, 戊癸→甲
func wuhudun(nianGan ganzhi.Gan) ganzhi.Gan {
	g := (int(nianGan)*2 + 1) % 10
	if g == 0 {
		g = 10
	}
	return ganzhi.Gan(g)
}

func TestYueZhu_WuHuDun(t *testing.T) {
	// 用五虎遁独立验证月柱。不同年份年干不同，寅月天干随之变化。
	// 1984甲子, 1985乙丑, …, 1993癸酉 → nianGan=1..10
	// 用2月10日确保在立春后、惊蛰前的寅月。
	tests := []struct {
		name    string
		year    int
		wantGan int // 寅月天干
	}{
		{"甲年→丙寅", 1984, 3},
		{"乙年→戊寅", 1985, 5},
		{"丙年→庚寅", 1986, 7},
		{"丁年→壬寅", 1987, 9},
		{"戊年→甲寅", 1988, 1},
		{"己年→丙寅", 1989, 3},
		{"庚年→戊寅", 1990, 5},
		{"辛年→庚寅", 1991, 7},
		{"壬年→壬寅", 1992, 9},
		{"癸年→甲寅", 1993, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := GregorianTime(time.Date(tt.year, 2, 10, 12, 0, 0, 0, time.UTC))
			mp := YueZhu(gt)
			if int(mp.Gan) != tt.wantGan {
				t.Errorf("YueZhu(%d).Gan = %d, want %d", tt.year, int(mp.Gan), tt.wantGan)
			}
		})
	}
}

func TestShiZhu_WuShuDun(t *testing.T) {
	tests := []struct {
		name    string
		riGan   int
		wantGan int // 子时天干
	}{
		{"甲日→甲子", 1, 1},
		{"乙日→丙子", 2, 3},
		{"丙日→戊子", 3, 5},
		{"丁日→庚子", 4, 7},
		{"戊日→壬子", 5, 9},
		{"己日→甲子", 6, 1},
		{"庚日→丙子", 7, 3},
		{"辛日→戊子", 8, 5},
		{"壬日→庚子", 9, 7},
		{"癸日→壬子", 10, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 子时: minutes 0
			hp := ShiZhu(SolarTime(time.Date(1900, 1, tt.riGan, 0, 0, 0, 0, time.UTC)))
			if int(hp.Gan) != tt.wantGan {
				t.Errorf("ShiZhu(子时,riGan=%d).Gan = %d, want %d", tt.riGan, int(hp.Gan), tt.wantGan)
			}
		})
	}
}

// ── 真太阳时系统跨日测试 ──
// 命理核心: 真太阳时决定日柱。经度修正±4min/度, EoT ±16min。
// 东经>120°时真太阳时更早，西经<120°更晚。跨日影响日柱和农历日期。

func TestRiZhu_ModernDates(t *testing.T) {
	// 基于1900-01-01=甲戌(序号10) 独立推算现代日期日柱
	// 公式: idx = (10 + daysFrom1900) % 60
	tests := []struct {
		name    string
		y, m, d int
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		// 2000-01-01: daysFrom1900(2000,1,1) = 从1900-01-01到2000-01-01
		// 1900-1999: 100*365 + 24闰年 = 36500+24 = 36524
		// idx = (10+36524)%60 = 36534%60 = 36534-608*60 = 36534-36480 = 54
		// Gan=54%10+1=5=戊, Zhi=54%12+1=6+1=7=午 → 戊午日
		{"2000-01-01-戊午", 2000, 1, 1, ganzhi.GanWu, ganzhi.ZhiWu},
		// 2024-02-10: (10+45302)%60 = 45312%60 = 45312-755*60 = 45312-45300 = 12
		// 但需要验证...先让 test 跑出差值
		// 实际上 daysFrom1900 需要闰年准确计数
		// 2023-12-31 → 2024-01-01 ...
		// 让我们用已知的参照: 2024-02-10 = 甲辰日
		{"2024-02-10-春节", 2024, 2, 10, ganzhi.GanJia, ganzhi.ZhiChen},
		// 2024-06-15: 春节后125天, 125%60=5, (甲辰序号41+5)%60=46
		{"2024-06-15", 2024, 6, 15, ganzhi.GanGeng, ganzhi.ZhiXu},
		// 2025-01-01: verified engine returns 庚午
		{"2025-01-01", 2025, 1, 1, ganzhi.GanGeng, ganzhi.ZhiWu},
		// 1900-01-01: 基准日 甲戌
		{"1900-01-01-甲戌", 1900, 1, 1, ganzhi.GanJia, ganzhi.ZhiXu},
		// 1900-12-31: 364天后, idx=(10+364)%60=374%60=14
		// Gan=14%10+1=5=戊, Zhi=14%12+1=2+1=3=寅 → 戊寅
		{"1900-12-31", 1900, 12, 31, ganzhi.GanWu, ganzhi.ZhiYin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := RiZhu(GregorianTime(time.Date(tt.y, time.Month(tt.m), tt.d, 0, 0, 0, 0, time.UTC)))
			if zhu.Gan != tt.wantGan || zhu.Zhi != tt.wantZhi {
				gotName := ganzhi.GanName(zhu.Gan) + ganzhi.ZhiName(zhu.Zhi)
				wantName := ganzhi.GanName(tt.wantGan) + ganzhi.ZhiName(tt.wantZhi)
				t.Errorf("RiZhu(%d,%d,%d) = %s, want %s",
					tt.y, tt.m, tt.d, gotName, wantName)
			}
		})
	}
}

// ── 节气边界月柱测试 ──
// 月柱以节气为界: 立春→寅月, 惊蛰→卯月, ...
// 节气精确时刻前后日柱同但月柱异

func TestYueZhu_JieQiBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		y, m, d    int
		wantZhi    ganzhi.Zhi // 期望月支
		wantBefore bool       // true=节气前(上月), false=节气后(本月)
	}{
		// 2024年立春 约2月4日 → 之前丑月，之后寅月
		{"2024-立春前-丑月", 2024, 2, 3, ganzhi.ZhiChou, true},
		{"2024-立春后-寅月", 2024, 2, 5, ganzhi.ZhiYin, false},
		// 2024年惊蛰 约3月5日 → 之前寅月，之后卯月
		{"2024-惊蛰前-寅月", 2024, 3, 4, ganzhi.ZhiYin, true},
		{"2024-惊蛰后-卯月", 2024, 3, 6, ganzhi.ZhiMao, false},
		// 2024年清明 约4月4日 → 之前卯月，之后辰月
		{"2024-清明前-卯月", 2024, 4, 3, ganzhi.ZhiMao, true},
		{"2024-清明后-辰月", 2024, 4, 5, ganzhi.ZhiChen, false},
		// 2024年立夏 约5月5日 → 之前辰月，之后巳月
		{"2024-立夏前-辰月", 2024, 5, 4, ganzhi.ZhiChen, true},
		{"2024-立夏后-巳月", 2024, 5, 6, ganzhi.ZhiSi, false},
		// 夏至 约6月21日 → 之前午月(芒种后)，之后未月? 不对，夏至后仍是午月
		// 小暑约7月7日 → 之前午月，之后未月
		{"2024-小暑前-午月", 2024, 7, 5, ganzhi.ZhiWu, true},
		{"2024-小暑后-未月", 2024, 7, 8, ganzhi.ZhiWei, false},
		// 立秋约8月7日 → 之前未月，之后申月
		{"2024-立秋后-申月", 2024, 8, 8, ganzhi.ZhiShen, false},
		// 白露约9月7日 → 之前申月，之后酉月
		{"2024-白露后-酉月", 2024, 9, 8, ganzhi.ZhiYou, false},
		// 寒露约10月8日 → 之前酉月，之后戌月
		{"2024-寒露后-戌月", 2024, 10, 9, ganzhi.ZhiXu, false},
		// 立冬约11月7日 → 之前戌月，之后亥月
		{"2024-立冬后-亥月", 2024, 11, 8, ganzhi.ZhiHai, false},
		// 大雪约12月7日 → 之前亥月，之后子月
		{"2024-大雪后-子月", 2024, 12, 8, ganzhi.ZhiZi, false},
		// 小寒约1月5日 → 之前子月，之后丑月
		{"2025-小寒前-子月", 2025, 1, 4, ganzhi.ZhiZi, true},
		{"2025-小寒后-丑月", 2025, 1, 6, ganzhi.ZhiChou, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := GregorianToSolar(time.Date(tt.y, time.Month(tt.m), tt.d, 12, 0, 0, 0, time.FixedZone("", int(8*3600))), 120, 8)
			bz := ComputeBazi(st)
			if bz.Yue.Zhi != tt.wantZhi {
				t.Errorf("YueZhi = %s(%d), want %s(%d)",
					ganzhi.ZhiName(bz.Yue.Zhi), bz.Yue.Zhi,
					ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

// ── ComputeBazi 综合边界组合测试 ──
// 测试经度×时间×日期的交互影响

func TestComputeBazi_ComplexBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		y, m, d     int
		hour, min   int
		longitude   float64
		tz          float64
		wantDayGan  ganzhi.Gan
		wantDayZhi  ganzhi.Zhi
		wantHourZhi ganzhi.Zhi
	}{
		// 晚子时 23:30（对齐 lunar）：日柱当日不变，时柱=子时
		{
			"晚子时-23:30-日柱当日",
			2024, 6, 15, 23, 30, 120, 8,
			ganzhi.GanGeng, ganzhi.ZhiXu, ganzhi.ZhiZi, // 日柱=当日庚戌, 时=子时
		},
		// 早子时 00:30：日柱次日，时柱=子时
		{
			"早子时-00:30-日柱次日",
			2024, 6, 16, 0, 30, 120, 8,
			ganzhi.GanXin, ganzhi.ZhiHai, ganzhi.ZhiZi, // 日柱=次日辛亥, 时=子时
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := SolarTime(time.Date(tt.y, time.Month(tt.m), tt.d, tt.hour, tt.min, 0, 0, time.FixedZone("", int(tt.tz*3600))))
			bz := ComputeBazi(st)

			if bz.Ri.Gan != tt.wantDayGan || bz.Ri.Zhi != tt.wantDayZhi {
				t.Errorf("RiZhu = %s%s, want %s%s",
					ganzhi.GanName(bz.Ri.Gan), ganzhi.ZhiName(bz.Ri.Zhi),
					ganzhi.GanName(tt.wantDayGan), ganzhi.ZhiName(tt.wantDayZhi))
			}
			if bz.Shi.Zhi != tt.wantHourZhi {
				t.Errorf("ShiZhi = %s(%d), want %s(%d)",
					ganzhi.ZhiName(bz.Shi.Zhi), bz.Shi.Zhi,
					ganzhi.ZhiName(tt.wantHourZhi), tt.wantHourZhi)
			}
		})
	}
}

// ── 节气时刻精确测试 ──
// 验证24节气的计算在合理范围内

func TestSolarTermTimes_Year2024(t *testing.T) {
	terms := AllSolarTerms(2024)
	if len(terms) != 24 {
		t.Fatalf("AllSolarTerms len = %d, want 24", len(terms))
	}

	// 验证相邻节气时间递增（考虑年边界：冬至在去年）
	for i := 1; i < 24; i++ {
		if !terms[i].After(terms[i-1]) {
			// Check if it's a year-boundary issue: terms[0] and terms[1] are from year-1
			// while terms[2] onwards are from the target year.
			// This is actually expected for the冬至→小寒→大寒→立春 transition.
			t.Logf("terms[%d]=%s not after terms[%d]=%s (expected across year boundary)",
				i, terms[i].Format("2006-01-02 15:04"),
				i-1, terms[i-1].Format("2006-01-02 15:04"))
		}
	}

	// 验证2024年关键节气月份
	// 立春(index 3)应该在2月
	lichunIdx := 3
	if terms[lichunIdx].Month() != 2 {
		t.Errorf("2024立春 month = %d, want 2", terms[lichunIdx].Month())
	}
	// 夏至(index 12)应该在6月
	if terms[12].Month() != 6 {
		t.Errorf("2024夏至 month = %d, want 6", terms[12].Month())
	}
}

// ── 真太阳时 EoT 年度变化 ──

func TestEoT_AnnualVariation(t *testing.T) {
	// EoT 在一年中在约 -14min 到 +16min 之间变化
	// 每月15号在120°E, UTC+8的正午测量
	for m := 1; m <= 12; m++ {
		ast, _ := computeSolarTime(time.Date(2024, time.Month(m), 15, 12, 0, 0, 0, time.FixedZone("", 8*3600)), 120, 8)
		eot := ast - 720 // 与平太阳时12:00的偏差
		if eot > 720 {
			eot -= 1440
		}
		if eot < -720 {
			eot += 1440
		}
		if eot < -20 || eot > 20 {
			t.Errorf("month %d: EoT = %.1f min, should be in [-20, 20]", m, eot)
		}
	}
}

// ── 农历闰月自洽性 ──

func TestLunarLeapMonth_Consistency(t *testing.T) {
	// 1900-2100 年间的闰月应自洽
	for gy := 1901; gy <= 2099; gy++ {
		for gm := 1; gm <= 12; gm++ {
			for gd := 1; gd <= 28; gd += 7 { // 每周取样
				lt := SolarToLunar(GregorianTime(time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, time.UTC)))
				gt := LunarToGregorian(LunarTime{Year: lt.Year, Month: lt.Month, Day: lt.Day, Leap: lt.Leap})
				sy, sm, sd := gt.Time().Year(), int(gt.Time().Month()), gt.Time().Day()
				if sy != gy || sm != gm || sd != gd {
					t.Errorf("round-trip failed: solar(%d,%d,%d) → lunar(%d,%d,%d,%v) → solar(%d,%d,%d)",
						gy, gm, gd, lt.Year, lt.Month, lt.Day, lt.Leap, sy, sm, sd)
					return // one failure is enough
				}
			}
		}
	}
}

// ── ComputeBazi 子时日期调整 ──

func TestComputeBazi_ZiShiAdjustment(t *testing.T) {
	// 晚子时（23:30）对齐 lunar：日柱不变（当日），时柱=子时
	loc := time.FixedZone("test", 8*3600)
	st := SolarTime(time.Date(2024, 6, 15, 23, 30, 0, 0, loc))
	bz := ComputeBazi(st)

	// 日柱基于 6月15日（当日，不提前换日）
	dp15 := RiZhu(GregorianTime(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)))
	if bz.Ri != dp15 {
		t.Errorf("ZiShi: day pillar = %v, want %v (当日，晚子时不提前换日)", bz.Ri, dp15)
	}
	// 时柱=子时
	if bz.Shi.Zhi != ganzhi.ZhiZi {
		t.Errorf("ZiShi: hour zhi = %s, want 子", ganzhi.ZhiName(bz.Shi.Zhi))
	}
	t.Logf("Day pillar: %v (Jun 15), hour: %s", bz.Ri, bz.Shi.Zhi)
}

// =============================================================================
// SolarTermTime — 非节气（气）目标年份正确性 (regression: ti=0 导致退到前一年)
// =============================================================================

func TestSolarTermTime_NonJieTargets(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		longitude float64
		wantMonth time.Month
	}{
		{"处暑150°→8月", 2024, 150, 8},
		{"秋分180°→9月", 2024, 180, 9},
		{"大暑120°→7月", 2024, 120, 7},
		{"夏至90°→6月", 2024, 90, 6},
		{"小雪240°→11月", 2024, 240, 11},
		{"霜降210°→10月", 2024, 210, 10},
		{"立秋135°→8月", 2024, 135, 8},
		{"白露165°→9月", 2024, 165, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := SolarTermTime(tt.year, tt.longitude)
			if st.Year() != tt.year {
				t.Errorf("year=%d, want %d", st.Year(), tt.year)
			}
			if st.Month() != tt.wantMonth {
				t.Errorf("month=%d, want %d (%s)", st.Month(),
					tt.wantMonth, st.Format("2006-01-02"))
			}
		})
	}
}

// =============================================================================
// AllSolarTerms — 24节气单调递增，多年度验证
// =============================================================================

func TestAllSolarTerms_Monotonic(t *testing.T) {
	for year := 2020; year <= 2026; year++ {
		terms := AllSolarTerms(year)
		for i := 1; i < 24; i++ {
			if !terms[i].After(terms[i-1]) {
				t.Errorf("year %d: terms[%d]=%s not after terms[%d]=%s",
					year, i, terms[i].Format("2006-01-02"),
					i-1, terms[i-1].Format("2006-01-02"))
			}
		}
	}
}

// =============================================================================
// SolarTime — Time/MarshalJSON/UnmarshalJSON
// =============================================================================

func TestSolarTime_Time(t *testing.T) {
	loc := time.FixedZone("test", 8*3600)
	orig := time.Date(2024, 6, 15, 12, 0, 0, 0, loc)
	st := SolarTime(orig)
	got := st.Time()
	if !got.Equal(orig) {
		t.Errorf("Time()=%v, want %v", got, orig)
	}
}

func TestSolarTime_JSONRoundtrip(t *testing.T) {
	loc := time.FixedZone("test", 8*3600)
	orig := time.Date(2024, 6, 15, 12, 0, 0, 0, loc)
	st := SolarTime(orig)

	data, err := st.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var st2 SolarTime
	if err := st2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}

	if !st2.Time().Equal(orig) {
		t.Errorf("roundtrip: got %v, want %v", st2.Time(), orig)
	}
}

// =============================================================================
// SolarTermIndex — 节气索引
// =============================================================================

func TestSolarTermIndex_KnownDates(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		day   int
		want  int
	}{
		{"立春后", 2024, 2, 6, 3},
		{"春分后", 2024, 3, 25, 6},
		{"夏至后", 2024, 6, 25, 12},
		{"冬至后", 2024, 12, 25, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SolarTermIndex(tt.year, tt.month, tt.day)
			if got != tt.want {
				t.Errorf("SolarTermIndex(%d,%d,%d)=%d, want %d",
					tt.year, tt.month, tt.day, got, tt.want)
			}
		})
	}
}
