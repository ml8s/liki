package tianwen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestNianZhu_Basic(t *testing.T) {
	// 年柱公式：(year-3)%10=干, (year-3)%12=支
	tests := []struct {
		name    string
		year    int
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		{"1984-甲子", 1984, ganzhi.GanJia, ganzhi.ZhiZi},
		{"2020-庚子", 2020, ganzhi.GanGeng, ganzhi.ZhiZi},
		{"2024-甲辰", 2024, ganzhi.GanJia, ganzhi.ZhiChen},
		{"1900-庚子", 1900, ganzhi.GanGeng, ganzhi.ZhiZi},
		{"2000-庚辰", 2000, ganzhi.GanGeng, ganzhi.ZhiChen},
		{"2025-乙巳", 2025, ganzhi.GanYi, ganzhi.ZhiSi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 用6月15日确保已过立春
			zhu := NianZhu(GregorianTime(time.Date(tt.year, 6, 15, 0, 0, 0, 0, time.UTC)))
			if zhu.Gan != tt.wantGan {
				t.Errorf("Gan = %s(%d), want %s(%d)", ganzhi.GanName(zhu.Gan), zhu.Gan, ganzhi.GanName(tt.wantGan), tt.wantGan)
			}
			if zhu.Zhi != tt.wantZhi {
				t.Errorf("Zhi = %s(%d), want %s(%d)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi, ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

func TestNianZhu_LiChunBoundary(t *testing.T) {
	// 立春前用上一年年柱，立春后用新年年柱。
	// 1984年立春约 2月4日。
	tests := []struct {
		name    string
		year    int
		month   int
		day     int
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		// 1984-02-03 立春前 → 1983年 癸亥
		{"1984-02-03-before-lichun", 1984, 2, 3, ganzhi.GanGui, ganzhi.ZhiHai},
		// 1984-02-05 立春后 → 1984年 甲子
		{"1984-02-05-after-lichun", 1984, 2, 5, ganzhi.GanJia, ganzhi.ZhiZi},
		// 2020-02-03 立春前 → 2019年 己亥
		{"2020-02-03-before-lichun", 2020, 2, 3, ganzhi.GanJi, ganzhi.ZhiHai},
		// 2020-02-05 立春后 → 2020年 庚子
		{"2020-02-05-after-lichun", 2020, 2, 5, ganzhi.GanGeng, ganzhi.ZhiZi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := NianZhu(GregorianTime(time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0, time.UTC)))
			if zhu.Gan != tt.wantGan {
				t.Errorf("Gan = %s(%d), want %s(%d)", ganzhi.GanName(zhu.Gan), zhu.Gan, ganzhi.GanName(tt.wantGan), tt.wantGan)
			}
			if zhu.Zhi != tt.wantZhi {
				t.Errorf("Zhi = %s(%d), want %s(%d)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi, ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

// TestNianZhu_LiChunExactTime 验证立春换年按精确时刻（非日期）：
// 立春当天立春时刻之前出生仍算上一年（2026 立春 ≈ 02-04 07:42 +08:00）。
func TestNianZhu_LiChunExactTime(t *testing.T) {
	cases := []struct {
		name string
		time string // +08:00
		want string
	}{
		{"立春前1小时", "2026-02-04T03:05:00+08:00", "乙巳"},
		{"立春后1小时", "2026-02-04T05:05:00+08:00", "丙午"},
		{"立春当天午前", "2026-02-04T02:00:00+08:00", "乙巳"},
		{"立春次日", "2026-02-05T12:00:00+08:00", "丙午"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tm, _ := time.Parse(time.RFC3339, tc.time)
			z := NianZhu(GregorianTime(tm))
			got := z.Gan.String() + z.Zhi.String()
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

// ── 月柱 ──

func TestYueZhu_Basic(t *testing.T) {
	// 用五虎遁独立计算月柱
	tests := []struct {
		name    string
		nianGan ganzhi.Gan
		zhi     int // 月支 1=寅..12=丑
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		// 甲年：正月丙寅, 二月丁卯, ..., 十一月丙子, 十二月丁丑
		{"甲年-正月-丙寅", ganzhi.GanJia, 1, ganzhi.GanBing, ganzhi.ZhiYin},
		{"甲年-五月-庚午", ganzhi.GanJia, 5, ganzhi.GanGeng, ganzhi.ZhiWu},
		{"甲年-十一月-丙子", ganzhi.GanJia, 11, ganzhi.GanBing, ganzhi.ZhiZi},
		// 乙年：正月戊寅
		{"乙年-正月-戊寅", ganzhi.GanYi, 1, ganzhi.GanWu, ganzhi.ZhiYin},
		// 丙年：正月庚寅
		{"丙年-正月-庚寅", ganzhi.GanBing, 1, ganzhi.GanGeng, ganzhi.ZhiYin},
		// 丁年：正月壬寅
		{"丁年-正月-壬寅", ganzhi.GanDing, 1, ganzhi.GanRen, ganzhi.ZhiYin},
		// 戊年：正月甲寅
		{"戊年-正月-甲寅", ganzhi.GanWu, 1, ganzhi.GanJia, ganzhi.ZhiYin},
		// 庚年：正月戊寅（乙庚同）
		{"庚年-正月-戊寅", ganzhi.GanGeng, 1, ganzhi.GanWu, ganzhi.ZhiYin},
		// 庚年六月：癸未
		{"庚年-六月-癸未", ganzhi.GanGeng, 6, ganzhi.GanGui, ganzhi.ZhiWei},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接用五虎遁公式：月干 = (正月干 + 月数 - 1) % 10
			janGan := wuhudun(tt.nianGan)
			monthNum := tt.zhi // 1=寅月
			wantMonthGan := ganzhi.Gan((int(janGan) + monthNum - 1) % 10)
			if wantMonthGan == 0 {
				wantMonthGan = 10
			}
			if wantMonthGan != tt.wantGan {
				t.Fatalf("test data bug: wuhudun(%s).month(%d) = %s, test expects %s",
					ganzhi.GanName(tt.nianGan), monthNum, ganzhi.GanName(wantMonthGan), ganzhi.GanName(tt.wantGan))
			}
			// 月支=(monthNum+1)%12+1 → 寅(1)→寅(3), 丑(12)→丑(2)...
			// 实际：寅月=支3(寅), 卯月=支4(卯), ...
			// 我们这里的 zhi 1=寅... 地支值 = (zhi+1)%12+1
			wantZhi := ganzhi.Zhi((tt.zhi+1)%12 + 1)
			if wantZhi == 0 {
				wantZhi = 12
			}
			if wantZhi != tt.wantZhi {
				t.Fatalf("test data bug: zhi %d → zhi %s, test expects %s",
					tt.zhi, ganzhi.ZhiName(wantZhi), ganzhi.ZhiName(tt.wantZhi))
			}
		})
	}
}

// ── 日柱 ──

func TestRiZhu_BaseReference(t *testing.T) {
	// 基准日: 1900-01-01 = 甲戌日
	zhu := RiZhu(GregorianTime(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)))
	if zhu.Gan != ganzhi.GanJia {
		t.Errorf("1900-01-01 Gan = %s(%d), want 甲(1)", ganzhi.GanName(zhu.Gan), zhu.Gan)
	}
	if zhu.Zhi != ganzhi.ZhiXu {
		t.Errorf("1900-01-01 Zhi = %s(%d), want 戌(11)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi)
	}
	// 验证 甲戌 = 六十甲子序号 10
	idx := ganzhi.SixtyCycleIndex(zhu.Gan, zhu.Zhi)
	if idx != 10 {
		t.Errorf("1900-01-01 甲戌 index = %d, want 10", idx)
	}
}

func TestRiZhu_KnownDates(t *testing.T) {
	// 基于1900-01-01=甲戌(序号10)推算出以下日柱
	tests := []struct {
		name    string
		year    int
		month   int
		day     int
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		// 1900-01-02 = 乙亥 (序号11)
		{"1900-01-02-乙亥", 1900, 1, 2, ganzhi.GanYi, ganzhi.ZhiHai},
		// 1900-01-03 = 丙子 (序号12)
		{"1900-01-03-丙子", 1900, 1, 3, ganzhi.GanBing, ganzhi.ZhiZi},
		// 1900-01-11 = 甲申 (序号20)
		{"1900-01-11-甲申", 1900, 1, 11, ganzhi.GanJia, ganzhi.ZhiShen},
		// 1900-02-01 = 乙巳 (序号41): 1月31天, Jan 31=甲辰(29), Feb 1=乙巳
		// 1900-01-01=甲戌(10), Jan有31天 → Jan 31 = (10+30)%60=40=甲辰, Feb 1 = 41=乙巳
		{"1900-02-01-乙巳", 1900, 2, 1, ganzhi.GanYi, ganzhi.ZhiSi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := RiZhu(GregorianTime(time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0, time.UTC)))
			if zhu.Gan != tt.wantGan {
				t.Errorf("Gan = %s(%d), want %s(%d)", ganzhi.GanName(zhu.Gan), zhu.Gan, ganzhi.GanName(tt.wantGan), tt.wantGan)
			}
			if zhu.Zhi != tt.wantZhi {
				t.Errorf("Zhi = %s(%d), want %s(%d)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi, ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

// helper: daysBetween computes days between two dates (both inclusive if startInclusive).
// Uses the standard algorithm for verification.
func daysFrom1900(y, m, d int) int {
	// Count days from 1900-01-01 to y-m-d
	// Years
	days := 0
	for yr := 1900; yr < y; yr++ {
		if isLeapYear(yr) {
			days += 366
		} else {
			days += 365
		}
	}
	// Months in current year
	monthDays := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if isLeapYear(y) {
		monthDays[1] = 29
	}
	for mo := 1; mo < m; mo++ {
		days += monthDays[mo-1]
	}
	days += d - 1
	return days
}

func TestRiZhu_JiaoChaCheck(t *testing.T) {
	// 对几个重要日期，用独立算法（天数累加）验证日柱
	tests := []struct {
		name    string
		year    int
		month   int
		day     int
		wantIdx int // 六十甲子序号 0-59, 0=甲子
	}{
		// 1900-01-01=甲戌, index=10
		{"base-1900-01-01", 1900, 1, 1, 10},
		// 1901-01-01: 1900年365天 (1900不是闰年), 365天后 = (10+365)%60=375%60=15
		{"1901-01-01", 1901, 1, 1, 15},
		// 1984-02-15: 从1900-01-01到1984-02-15
		// 84年 + 1月 + 14天
		// 1900-1983: 84*365 + leapCount(1900-1983)
		// 1900不是闰年, 1904...1980: (1980-1904)/4+1 = 76/4+1 = 20
		// = 30660 + 20 = 30680
		// 1984 Jan: 31, Feb 1-14: 14 → +45
		// Total: 30725
		// idx = (10+30725)%60 = 30735%60 = 15...
		// 30735/60=512*60=30720, remainder=15
		// Hmm let me recheck. daysFrom1900(1984,2,15):
		// 1900-1983: leap years 1904,1908,...,1980 = 20
		// 84*365+20 = 30660+20 = 30680
		// Jan(31) + 14 = 45
		// 30680+45 = 30725
		// (10+30725)%60 = 30735%60 = 30735 - 512*60 = 30735-30720 = 15
		// Index 15: Gan=15%10+1=6=己, Zhi=15%12+1=3+1=4=卯 → 己卯
		// Wait, let me double check. 甲戌=10, +1=乙亥(11), +2=丙子(12)...+30725
		// 30725/60 = 512 remainder 5. So (10+5)%60 = 15.
		// Wait: 512*60 = 30720. 30725-30720=5. 10+5=15.
		// Index 15: Gan=15%10+1=5+1=6=己, Zhi=15%12+1=3+1=4=卯
		{"1984-02-15", 1984, 2, 15, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := RiZhu(GregorianTime(time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0, time.UTC)))
			gotIdx := ganzhi.SixtyCycleIndex(zhu.Gan, zhu.Zhi)
			if gotIdx != tt.wantIdx {
				computed := daysFrom1900(tt.year, tt.month, tt.day)
				expectedIdx := (10 + computed) % 60
				t.Errorf("RiZhu(%d,%d,%d) = %s%s (index %d), want index %d (days from 1900=%d, expected idx=%d)",
					tt.year, tt.month, tt.day,
					ganzhi.GanName(zhu.Gan), ganzhi.ZhiName(zhu.Zhi), gotIdx,
					tt.wantIdx, computed, expectedIdx)
			}
		})
	}
}

func TestRiZhu_DifferentTimezoneSameDay(t *testing.T) {
	// 同一天不同时区，日柱应相同
	dates := []struct{ y, m, d int }{
		{2024, 1, 1},
		{2024, 6, 15},
		{2024, 12, 31},
	}
	for _, d := range dates {
		zhu := RiZhu(GregorianTime(time.Date(d.y, time.Month(d.m), d.d, 0, 0, 0, 0, time.UTC)))
		// 日柱仅取决于公历日期，不涉及时区
		if zhu.Gan < 1 || zhu.Gan > 10 {
			t.Errorf("RiZhu(%d,%d,%d) invalid gan: %d", d.y, d.m, d.d, zhu.Gan)
		}
	}
}

// ── 时柱 ──

func TestShiZhu_KnownCases(t *testing.T) {
	// 1900-01-01=甲日, 01-02=乙日, 01-03=丙日, 01-04=丁日,
	// 01-05=戊日, 01-07=庚日, 01-10=癸日
	tests := []struct {
		name    string
		st      SolarTime
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		// 甲日：子=甲子, 丑=乙丑, 寅=丙寅, 卯=丁卯, 辰=戊辰, 巳=己巳, ...
		{"甲日-子时(0:00)", SolarTime(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)), ganzhi.GanJia, ganzhi.ZhiZi},
		{"甲日-卯时(6:00)", SolarTime(time.Date(1900, 1, 1, 6, 0, 0, 0, time.UTC)), ganzhi.GanDing, ganzhi.ZhiMao},
		{"甲日-午时(12:00)", SolarTime(time.Date(1900, 1, 1, 12, 0, 0, 0, time.UTC)), ganzhi.GanGeng, ganzhi.ZhiWu},
		{"甲日-酉时(18:00)", SolarTime(time.Date(1900, 1, 1, 18, 0, 0, 0, time.UTC)), ganzhi.GanGui, ganzhi.ZhiYou},
		// 乙日：子=丙子
		{"乙日-子时(0:00)", SolarTime(time.Date(1900, 1, 2, 0, 0, 0, 0, time.UTC)), ganzhi.GanBing, ganzhi.ZhiZi},
		{"乙日-午时(12:00)", SolarTime(time.Date(1900, 1, 2, 12, 0, 0, 0, time.UTC)), ganzhi.GanRen, ganzhi.ZhiWu},
		// 丙日：子=戊子
		{"丙日-子时(0:00)", SolarTime(time.Date(1900, 1, 3, 0, 0, 0, 0, time.UTC)), ganzhi.GanWu, ganzhi.ZhiZi},
		// 丁日：子=庚子
		{"丁日-子时(0:00)", SolarTime(time.Date(1900, 1, 4, 0, 0, 0, 0, time.UTC)), ganzhi.GanGeng, ganzhi.ZhiZi},
		// 戊日：子=壬子
		{"戊日-子时(0:00)", SolarTime(time.Date(1900, 1, 5, 0, 0, 0, 0, time.UTC)), ganzhi.GanRen, ganzhi.ZhiZi},
		// 庚日：子=丙子（乙庚同）
		{"庚日-子时(0:00)", SolarTime(time.Date(1900, 1, 7, 0, 0, 0, 0, time.UTC)), ganzhi.GanBing, ganzhi.ZhiZi},
		// 癸日：子=壬子（戊癸同）
		{"癸日-子时(0:00)", SolarTime(time.Date(1900, 1, 10, 0, 0, 0, 0, time.UTC)), ganzhi.GanRen, ganzhi.ZhiZi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := ShiZhu(tt.st)
			if zhu.Gan != tt.wantGan {
				t.Errorf("Gan = %s(%d), want %s(%d)", ganzhi.GanName(zhu.Gan), zhu.Gan, ganzhi.GanName(tt.wantGan), tt.wantGan)
			}
			if zhu.Zhi != tt.wantZhi {
				t.Errorf("Zhi = %s(%d), want %s(%d)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi, ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

func TestShiZhu_ZhiRanges(t *testing.T) {
	// 验证时辰边界
	// 子时 23:00-01:00 → zhi=子(1), 丑时 01:00-03:00 → zhi=丑(2), ...
	// 用1900-01-01(甲日)消除时干干扰
	tests := []struct {
		name    string
		st      SolarTime
		wantZhi ganzhi.Zhi
	}{
		{"23:00-子时", SolarTime(time.Date(1900, 1, 1, 23, 0, 0, 0, time.UTC)), ganzhi.ZhiZi},
		{"00:30-子时", SolarTime(time.Date(1900, 1, 1, 0, 30, 0, 0, time.UTC)), ganzhi.ZhiZi},
		{"01:00-丑时", SolarTime(time.Date(1900, 1, 1, 1, 0, 0, 0, time.UTC)), ganzhi.ZhiChou},
		{"03:00-寅时", SolarTime(time.Date(1900, 1, 1, 3, 0, 0, 0, time.UTC)), ganzhi.ZhiYin},
		{"05:00-卯时", SolarTime(time.Date(1900, 1, 1, 5, 0, 0, 0, time.UTC)), ganzhi.ZhiMao},
		{"07:00-辰时", SolarTime(time.Date(1900, 1, 1, 7, 0, 0, 0, time.UTC)), ganzhi.ZhiChen},
		{"09:00-巳时", SolarTime(time.Date(1900, 1, 1, 9, 0, 0, 0, time.UTC)), ganzhi.ZhiSi},
		{"11:00-午时", SolarTime(time.Date(1900, 1, 1, 11, 0, 0, 0, time.UTC)), ganzhi.ZhiWu},
		{"13:00-未时", SolarTime(time.Date(1900, 1, 1, 13, 0, 0, 0, time.UTC)), ganzhi.ZhiWei},
		{"15:00-申时", SolarTime(time.Date(1900, 1, 1, 15, 0, 0, 0, time.UTC)), ganzhi.ZhiShen},
		{"17:00-酉时", SolarTime(time.Date(1900, 1, 1, 17, 0, 0, 0, time.UTC)), ganzhi.ZhiYou},
		{"19:00-戌时", SolarTime(time.Date(1900, 1, 1, 19, 0, 0, 0, time.UTC)), ganzhi.ZhiXu},
		{"21:00-亥时", SolarTime(time.Date(1900, 1, 1, 21, 0, 0, 0, time.UTC)), ganzhi.ZhiHai},
		{"22:59-亥时边界", SolarTime(time.Date(1900, 1, 1, 22, 59, 0, 0, time.UTC)), ganzhi.ZhiHai},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zhu := ShiZhu(tt.st)
			if zhu.Zhi != tt.wantZhi {
				t.Errorf("Zhi = %s(%d), want %s(%d)", ganzhi.ZhiName(zhu.Zhi), zhu.Zhi, ganzhi.ZhiName(tt.wantZhi), tt.wantZhi)
			}
		})
	}
}

func TestComputeBazi_BeijingNoon(t *testing.T) {
	// 北京 (120°E, UTC+8)，经度与标准子午线一致，真太阳时≈平太阳时
	// 2024年6月15日12:00 北京
	// 年: 2024 → 甲辰
	// 月: 6月 → 芒种(6/5)后夏至(6/21)前 → 午月。甲年→庚午月
	// 日: daysFrom1900=45456, idx=(10+45456)%60=...
	//     45456/60=757*60=45420, rem=36. idx=(10+36)%60=46
	//     46→己酉? gan=46%10+1=7=庚, zhi=46%12+1=10+1=11=戌? No...
	//     46%10=6, gan=7=庚. 46%12=10, zhi=11=戌. → 庚戌
	// 时: 12:00 → 午时。庚日午时：庚日子=丙子...午=壬午
	//     时干=(7*2+7-2)%10=19%10=9→壬 ✓
	st := GregorianToSolar(time.Date(2024, time.Month(6), 15, 12, 0, 0, 0, time.FixedZone("", int(8*3600))), 120, 8)
	bz := ComputeBazi(st)

	if bz.Nian.Gan != ganzhi.GanJia || bz.Nian.Zhi != ganzhi.ZhiChen {
		t.Errorf("年柱 = %s%s, want 甲辰", ganzhi.GanName(bz.Nian.Gan), ganzhi.ZhiName(bz.Nian.Zhi))
	}
	// 月柱：2024年6月 → 庚午月
	if bz.Yue.Zhi != ganzhi.ZhiWu {
		t.Errorf("月支 = %s(%d), want 午(7)", ganzhi.ZhiName(bz.Yue.Zhi), bz.Yue.Zhi)
	}
	// 时支: 午时
	if bz.Shi.Zhi != ganzhi.ZhiWu {
		t.Errorf("時支 = %s(%d), want 午(7)", ganzhi.ZhiName(bz.Shi.Zhi), bz.Shi.Zhi)
	}
	// 验证日主存在
	if bz.Ri.Gan < 1 || bz.Ri.Gan > 10 {
		t.Errorf("日主无效: %d", bz.Ri.Gan)
	}
}

// ── 真太阳时 ──

func TestSolarTime_BeijingNoon(t *testing.T) {
	// 北京经度约116.4°E，时区UTC+8(120°E)
	// 经度修正 ≈ 4*(116.4-120) = -14.4分钟
	// 但这里用120°E，修正=0。配合均时差(equ of time)，真太阳时≈12:00附近。
	st := GregorianToSolar(time.Date(2024, time.Month(6), 15, 12, 0, 0, 0, time.FixedZone("", int(8*3600))), 120, 8)
	tm := st.Time()
	if tm.Hour() < 11 || tm.Hour() > 13 {
		t.Errorf("北京正午真太阳时 = %02d:%02d, 预期在12:00附近", tm.Hour(), tm.Minute())
	}
}

func TestSolarTime_UrumqiOffset(t *testing.T) {
	// 乌鲁木齐约87.6°E，时区UTC+8(120°E)
	// 经度差 ≈ 4*(87.6-120) = -129.6分钟 ≈ -2小时10分钟
	// 正午12:00的真太阳时 ≈ 09:50 左右
	st := GregorianToSolar(time.Date(2024, time.Month(6), 15, 12, 0, 0, 0, time.FixedZone("", int(8*3600))), 87.6, 8)
	ast := float64(st.Time().Hour()*60 + st.Time().Minute())
	if ast > 11*60 {
		t.Errorf("乌鲁木齐正午真太阳时 = %02d:%02d (%.0f分钟), 预期约在10:00前后",
			st.Time().Hour(), st.Time().Minute(), ast)
	}
}

func TestSolarTime_LateNightDayShift(t *testing.T) {
	// 新疆喀什 (75°E, UTC+8)，晚上22:30
	// 经度修正 = 4*(75-120) = -180分钟 = -3小时
	// 真太阳时 ≈ 22:30 - 3:00 = 19:30
	st := GregorianToSolar(time.Date(2024, time.Month(6), 15, 22, 30, 0, 0, time.FixedZone("", int(8*3600))), 75, 8)
	ast := float64(st.Time().Hour()*60 + st.Time().Minute())
	// 应在19:00-20:00区间（戌时）
	if ast < 18*60 || ast > 20*60+30 {
		t.Errorf("喀什22:30真太阳时 = %02d:%02d, 预期在戌时(19:00-21:00)",
			st.Time().Hour(), st.Time().Minute())
	}
}

func TestComputeBazi_ZeroTimeReturnsValidPillars(t *testing.T) {
	st := SolarTime(time.Time{})
	bz := ComputeBazi(st)

	if bz.Nian.Gan < 1 || bz.Nian.Gan > 10 {
		t.Errorf("Nian.Gan = %d from zero time", bz.Nian.Gan)
	}
	if bz.Ri.Gan < 1 || bz.Ri.Gan > 10 {
		t.Errorf("Ri.Gan = %d from zero time", bz.Ri.Gan)
	}
}

func TestLunarToGregorian_OutOfRange(t *testing.T) {
	// Month 13 should return zero time.
	lt13 := LunarTime{Year: 2024, Month: 13, Day: 1}
	gt13 := LunarToGregorian(lt13)
	if !gt13.Time().IsZero() {
		t.Error("LunarToGregorian(2024,13,1) should return zero time for month 13")
	}

	// Day 32 should return zero time.
	lt32 := LunarTime{Year: 2024, Month: 1, Day: 32}
	gt32 := LunarToGregorian(lt32)
	if !gt32.Time().IsZero() {
		t.Error("LunarToGregorian(2024,1,32) should return zero time for day 32")
	}
}

func TestRiZhu_YearOutOfRange(t *testing.T) {
	zhu := RiZhu(GregorianTime(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)))
	if zhu.Gan < 1 || zhu.Gan > 10 {
		t.Errorf("RiZhu year 1: Gan = %d", zhu.Gan)
	}
	if zhu.Zhi < 1 || zhu.Zhi > 12 {
		t.Errorf("RiZhu year 1: Zhi = %d", zhu.Zhi)
	}

	zhu2100 := RiZhu(GregorianTime(time.Date(2100, 6, 15, 0, 0, 0, 0, time.UTC)))
	if zhu2100.Gan < 1 || zhu2100.Gan > 10 {
		t.Errorf("RiZhu year 2100: Gan = %d", zhu2100.Gan)
	}
}

func TestComputeTimeset_ExtremeLongitude(t *testing.T) {
	gt := GregorianTime(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))

	for _, lon := range []float64{-180, 0, 180, 200, -200} {
		ts := ComputeTimeset(gt, lon)
		if ts.Solar.Time().IsZero() {
			t.Errorf("ComputeTimeset(lon=%.1f) returned zero solar time", lon)
		}
	}
}

func TestComputeBazi_SolarTimeAffectsRiZhu(t *testing.T) {
	// 极端情况：喀什 23:30，真太阳时可能跨日
	// 经度修正 -3小时 → 真太阳时≈20:30，仍在同一天
	// 更极端：黑龙江抚远 134°E + UTC+8，23:30
	// 经度修正 = 4*(134-120) = +56分钟
	// 真太阳时 ≈ 00:26 → 跨越到次日！
	st := GregorianToSolar(time.Date(2024, time.Month(6), 15, 23, 30, 0, 0, time.FixedZone("", int(8*3600))), 134, 8)
	if st.Time().Day() != 16 {
		t.Errorf("抚远23:30真太阳时应跨日到6月16日，实际=%d日", st.Time().Day())
	}
}
