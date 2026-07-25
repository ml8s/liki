package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 命宫/身宫 ──

func TestComputeMingShen(t *testing.T) {
	// 紫微斗数全书: 命宫以寅起正月，顺数至生月，逆数至生时
	// 身宫以寅起正月，顺数至生月，顺数至生时
	tests := []struct {
		name       string
		lunarMonth int
		hourZhi    Zhi
		wantMing   Zhi // 1=子..12=亥
		wantShen   Zhi
	}{
		// 正月子时: 命宫=寅(3), 身宫=寅(3)
		{"正月子时", 1, 1, 3, 3},
		// 正月午时: 命宫=申(9), 身宫=申(9)
		// 寅起正月顺数至正月=寅, 逆数至午: 寅→丑→子→亥→戌→酉→申, 命宫=申
		// 寅顺数至正月=寅, 顺数至午: 寅→卯→辰→巳→午→未→申, 身宫=申
		{"正月午时", 1, 7, 9, 9},
		// 七月卯时: 命宫=巳(6), 身宫=亥(12)
		{"七月卯时", 7, 4, 6, 12},

		// 十一月子时: 命宫=子(1), 身宫=子(1)
		// mingZhi = ((11-1+2)%12)+1 = (12%12)+1 = 1 = 子
		{"十一月子时", 11, 1, 1, 1},
		// 十二月亥时: 命宫=丑(2)
		// mingZhi = ((12-12+2)%12)+1 = (2%12)+1 = 3 = 寅?
		// Wait, hourZhi for 亥 = 12
		// ((12-12+2)%12+12)%12+1 = (2+12)%12+1 = 2+1 = 3 = 寅
		// But should be: 寅起正月→十二月=丑(2), 逆数至亥:
		// 子→丑, 丑→子, 寅→亥, 卯→戌, 辰→酉, 巳→申, 午→未, 未→午, 申→巳, 酉→辰, 戌→卯, 亥→寅
		// 命宫=寅(3). ✓
		{"十二月亥时", 12, 12, 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ming, shen := computeMingShen(tt.lunarMonth, tt.hourZhi)
			if ming != tt.wantMing || shen != tt.wantShen {
				t.Errorf("computeMingShen(%d,%d) = (ming=%d,shen=%d), want (ming=%d,shen=%d)",
					tt.lunarMonth, tt.hourZhi,
					int(ming), int(shen),
					int(tt.wantMing), int(tt.wantShen))
			}
		})
	}
}

// ── 十二宫地支排列 ──

func TestArrangePalaceZhis(t *testing.T) {
	// 紫微十二宫从命宫起逆时针排列地支
	tests := []struct {
		name     string
		mingZhi  Zhi
		wantZhis [12]Zhi
	}{
		{
			"命宫在寅→逆排寅丑子亥戌酉申未午巳辰卯",
			3,
			[12]Zhi{3, 2, 1, 12, 11, 10, 9, 8, 7, 6, 5, 4},
		},
		{
			"命宫在子→逆排子亥戌酉申未午巳辰卯寅丑",
			1,
			[12]Zhi{1, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2},
		},
		{
			"命宫在午→逆排午巳辰卯寅丑子亥戌酉申未",
			7,
			[12]Zhi{7, 6, 5, 4, 3, 2, 1, 12, 11, 10, 9, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := arrangePalaceZhis(tt.mingZhi)
			for i := 0; i < 12; i++ {
				if got[i] != tt.wantZhis[i] {
					t.Errorf("palace[%d].Zhi = %d, want %d", i, int(got[i]), int(tt.wantZhis[i]))
				}
			}
		})
	}
}

// ── 局数 ──

func TestDetermineJuShu(t *testing.T) {
	// 局数由命宫干支纳音决定
	tests := []struct {
		name     string
		mingGan  Gan
		mingZhi  Zhi
		wantJu   juShu
	}{
		// 甲子=海中金 → 金四局
		{"甲子→金四局", 1, 1, JuMetal},
		// 丙寅=炉中火 → 火六局
		{"丙寅→火六局", 3, 3, JuFire},
		// 戊辰=大林木 → 木三局
		{"戊辰→木三局", 5, 5, JuWood},
		// 壬申=剑锋金 → 金四局
		{"壬申→金四局", 9, 9, JuMetal},
		// 乙丑=海中金 → 金四局
		{"乙丑→金四局", 2, 2, JuMetal},
		// 庚午=路旁土 → 土五局
		{"庚午→土五局", 7, 7, JuEarth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineJuShu(tt.mingGan, tt.mingZhi)
			if got != tt.wantJu {
				t.Errorf("determineJuShu(%d,%d) = %d, want %d",
					int(tt.mingGan), int(tt.mingZhi), int(got), int(tt.wantJu))
			}
		})
	}
}

// ── 紫微星定位 ──

func TestFindZiwei(t *testing.T) {
	// 紫微星定位: 查表法, 局数除日数, 商数决定位置
	tests := []struct {
		name     string
		ju       juShu
		lunarDay int
		wantPos  palaceIndex
	}{
		// 水二局: start=2(丑), n=ceil(day/2)
		{"水二局-第1日", JuWater, 1, 2},
		{"水二局-第2日", JuWater, 2, 2},
		{"水二局-第3日", JuWater, 3, 1},
		{"水二局-第5日", JuWater, 5, 0},

		// 木三局: start=4(午)
		{"木三局-第1日", JuWood, 1, 4},
		{"木三局-第3日", JuWood, 3, 4},
		{"木三局-第4日", JuWood, 4, 3},

		// 金四局: start=11(丑... wait, start=11, which is index 11)
		// Actually start=11 means palace index 11
		{"金四局-第1日", JuMetal, 1, 11},
		{"金四局-第4日", JuMetal, 4, 11},
		{"金四局-第5日", JuMetal, 5, 10},

		// 土五局: start=6
		{"土五局-第1日", JuEarth, 1, 6},
		{"土五局-第5日", JuEarth, 5, 6},
		{"土五局-第6日", JuEarth, 6, 5},

		// 火六局: start=9
		{"火六局-第1日", JuFire, 1, 9},
		{"火六局-第6日", JuFire, 6, 9},
		{"火六局-第7日", JuFire, 7, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findZiwei(tt.ju, tt.lunarDay)
			if got != tt.wantPos {
				t.Errorf("findZiwei(%d,%d) = %d, want %d",
					int(tt.ju), tt.lunarDay, int(got), int(tt.wantPos))
			}
		})
	}
}

// ── 主星安星 ──

func TestPlaceMainStars_ZiweiChain(t *testing.T) {
	// 紫微系: 紫微→天机(隔1)→(空1)→太阳→武曲→天同→(空2)→廉贞
	// 全部逆时针(offset 递减)
	ziweiPos := palaceIndex(0)

	stars := placeMainStars(ziweiPos)

	// 紫微在0位
	if !hasStarAt(stars, 0, ZiWei) {
		t.Error("紫微应在宫位0")
	}
	// 天机在紫微逆1位 = 11
	if !hasStarAt(stars, 11, TianJi) {
		t.Error("天机应在宫位11(紫微逆1位)")
	}
	// 太阳在紫微逆3位 = 9
	if !hasStarAt(stars, 9, TaiYang) {
		t.Error("太阳应在宫位9(紫微逆3位)")
	}
	// 武曲在紫微逆4位 = 8
	if !hasStarAt(stars, 8, WuQu) {
		t.Error("武曲应在宫位8(紫微逆4位)")
	}
	// 天同在紫微逆5位 = 7
	if !hasStarAt(stars, 7, TianTong) {
		t.Error("天同应在宫位7(紫微逆5位)")
	}
	// 廉贞在紫微逆8位 = 4
	if !hasStarAt(stars, 4, LianZhen) {
		t.Error("廉贞应在宫位4(紫微逆8位)")
	}
}

func TestPlaceMainStars_TianfuChain(t *testing.T) {
	// 天府系: 天府在紫微顺2位, 其余顺时针
	// 天府→太阴(+1)→贪狼(+2)→巨门(+3)→天相(+4)→天梁(+5)→七杀(+6)→(空3)→破军(+10)
	ziweiPos := palaceIndex(0)

	stars := placeMainStars(ziweiPos)

	// 天府在2位 (ziweiPos+2 clockwise)
	if !hasStarAt(stars, 2, TianFu) {
		t.Error("天府应在宫位2(紫微顺2位)")
	}
	// 太阴在3位
	if !hasStarAt(stars, 3, TaiYin) {
		t.Error("太阴应在宫位3")
	}
	// 贪狼在4位
	if !hasStarAt(stars, 4, TanLang) {
		t.Error("贪狼应在宫位4")
	}
	// 七杀在8位
	if !hasStarAt(stars, 8, QiSha) {
		t.Error("七杀应在宫位8")
	}
	// 破军在0位 (2+10=12≡0)
	if !hasStarAt(stars, 0, PoJun) {
		t.Error("破军应在宫位0(2+10≡0)")
	}
}

func TestPlaceMainStars_All14Stars(t *testing.T) {
	// 全部14颗主星应落在12宫中
	stars := placeMainStars(0)
	total := 0
	for _, ss := range stars {
		total += len(ss)
	}
	if total != 14 {
		t.Errorf("总共应有14颗主星, got %d", total)
	}
}

func hasStarAt(stars map[palaceIndex][]starIndex, pos palaceIndex, s starIndex) bool {
	for _, ss := range stars[pos] {
		if ss == s {
			return true
		}
	}
	return false
}

// ── 辅星 ──

func TestLuCunPos(t *testing.T) {
	// 禄存: 甲寅 乙卯 丙巳 丁午 戊巳 己午 庚申 辛酉 壬亥 癸子
	// These are absolute branch positions (zhi-1: 0=子..11=亥).
	// Callers must convert to palaceIndex via zhiToPalace, passing 命宫 branch.
	tests := []struct {
		name    string
		yearGan Gan
		want    int
	}{
		{"甲→寅", 1, 2},
		{"乙→卯", 2, 3},
		{"丙→巳", 3, 5},
		{"丁→午", 4, 6},
		{"戊→巳", 5, 5},
		{"己→午", 6, 6},
		{"庚→申", 7, 8},
		{"辛→酉", 8, 9},
		{"壬→亥", 9, 11},
		{"癸→子", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luCunPos(tt.yearGan)
			if got != tt.want {
				t.Errorf("luCunPos(%d) = %d, want %d", tt.yearGan, got, tt.want)
			}
		})
	}
}

func TestQingYangTuoLuoPos(t *testing.T) {
	// 擎羊在禄存顺数前一位, 陀罗在禄存逆数后一位
	// 禄存年干: 甲寅(2)乙卯(3)丙戊巳(5)丁己午(6)庚申(8)辛酉(9)壬亥(11)癸子(0)
	tests := []struct {
		name       string
		yearGan    Gan
		wantQingYang int
		wantTuoLuo   int
	}{
		{"甲", 1, 3, 1},    // luCun=寅(2), qingYang=卯(3), tuoLuo=丑(1)
		{"乙", 2, 4, 2},    // luCun=卯(3), qingYang=辰(4), tuoLuo=寅(2)
		{"丙", 3, 6, 4},    // luCun=巳(5), qingYang=午(6), tuoLuo=辰(4)
		{"庚", 7, 9, 7},    // luCun=申(8), qingYang=酉(9), tuoLuo=未(7)
		{"壬", 9, 0, 10},   // luCun=亥(11), qingYang=子(0), tuoLuo=戌(10)
		{"癸", 10, 1, 11},  // luCun=子(0), qingYang=丑(1), tuoLuo=亥(11)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qingYangPos(tt.yearGan); got != tt.wantQingYang {
				t.Errorf("qingYangPos(%d) = %d, want %d", tt.yearGan, got, tt.wantQingYang)
			}
			if got := tuoLuoPos(tt.yearGan); got != tt.wantTuoLuo {
				t.Errorf("tuoLuoPos(%d) = %d, want %d", tt.yearGan, got, tt.wantTuoLuo)
			}
		})
	}
}

func TestZuoFuYouBiPos(t *testing.T) {
	// 左辅: lunarMonth→(month+2)%12 → palaceIndex
	// 右弼: lunarMonth→(11-month+12)%12 → palaceIndex
	tests := []struct {
		name       string
		lunarMonth int
		wantZF     int
		wantYB     int
	}{
		{"正月", 1, 3, 10},
		{"二月", 2, 4, 9},
		{"七月", 7, 9, 4},
		{"十二月", 12, 2, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zuoFuPos(tt.lunarMonth); got != tt.wantZF {
				t.Errorf("zuoFuPos(%d) = %d, want %d", tt.lunarMonth, got, tt.wantZF)
			}
			if got := youBiPos(tt.lunarMonth); got != tt.wantYB {
				t.Errorf("youBiPos(%d) = %d, want %d", tt.lunarMonth, got, tt.wantYB)
			}
		})
	}
}

func TestWenChangWenQuPos(t *testing.T) {
	// 文昌: hourZhi→(11-hourZhi+12)%12
	// 文曲: hourZhi→(hourZhi+3)%12
	tests := []struct {
		name    string
		hourZhi Zhi
		wantWC  int
		wantWQ  int
	}{
		{"子时", 1, 10, 4},
		{"午时", 7, 4, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wenChangPos(tt.hourZhi); got != tt.wantWC {
				t.Errorf("wenChangPos(%d) = %d, want %d", int(tt.hourZhi), got, tt.wantWC)
			}
			if got := wenQuPos(tt.hourZhi); got != tt.wantWQ {
				t.Errorf("wenQuPos(%d) = %d, want %d", int(tt.hourZhi), got, tt.wantWQ)
			}
		})
	}
}

func TestHuoXingLingXingPos(t *testing.T) {
	// 火星: 寅午戌丑起, 申子辰寅起, 巳酉丑卯起, 亥卯未酉起
	// 铃星: 寅午戌卯起, 申子辰戌起, 巳酉丑戌起, 亥卯未戌起
	tests := []struct {
		name         string
		yearZhi      Zhi
		hourZhi      Zhi
		wantHuoXing  int
		wantLingXing int
	}{
		// 寅午戌(3,7,11): 火星丑(2)起, 铃星卯(4)起
		{"寅年子时", 3, 1, 2, 4},
		// 申子辰(9,1,5): 火星寅(3)起, 铃星戌(11)起
		{"申年子时", 9, 1, 3, 11},
		// 亥卯未(12,4,8): 火星酉(10)起, 铃星戌(11)起
		{"亥年子时", 12, 1, 10, 11},
		// 巳酉丑(6,10,2): 火星卯(4)起, 铃星戌(11)起
		{"巳年子时", 6, 1, 4, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := huoXingPos(tt.yearZhi, tt.hourZhi); got != tt.wantHuoXing {
				t.Errorf("huoXingPos(%d,%d) = %d, want %d",
					tt.yearZhi, tt.hourZhi, got, tt.wantHuoXing)
			}
			if got := lingXingPos(tt.yearZhi, tt.hourZhi); got != tt.wantLingXing {
				t.Errorf("lingXingPos(%d,%d) = %d, want %d",
					tt.yearZhi, tt.hourZhi, got, tt.wantLingXing)
			}
		})
	}
}

// ── 四化 ──

func TestComputeSiHua(t *testing.T) {
	// 验证四化表中的代表性年份
	tests := []struct {
		name    string
		yearGan Gan
		want    [4]starIndex // 禄权科忌
	}{
		{"甲", 1, [4]starIndex{LianZhen, PoJun, WuQu, TaiYang}},
		{"乙", 2, [4]starIndex{TianJi, TianLiang, ZiWei, TaiYin}},
		{"丙", 3, [4]starIndex{TianTong, TianJi, WenChang, LianZhen}},
		{"庚", 7, [4]starIndex{TaiYang, WuQu, TaiYin, TianTong}},
		{"癸", 10, [4]starIndex{PoJun, JuMen, TaiYin, TanLang}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeSiHua(tt.yearGan)
			if got == nil {
				t.Fatal("computeSiHua returned nil")
			}
			if got[tt.want[0]] != HuaLu {
				t.Errorf("禄: expected %s to be 禄", starName(tt.want[0]))
			}
			if got[tt.want[1]] != HuaQuan {
				t.Errorf("权: expected %s to be 权", starName(tt.want[1]))
			}
			if got[tt.want[2]] != HuaKe {
				t.Errorf("科: expected %s to be 科", starName(tt.want[2]))
			}
			if got[tt.want[3]] != HuaJi {
				t.Errorf("忌: expected %s to be 忌", starName(tt.want[3]))
			}
			// 恰好4颗星有四化
			if len(got) != 4 {
				t.Errorf("expected 4 sihua stars, got %d", len(got))
			}
		})
	}
}

// ── 流日 riGan ──

func TestRiGan(t *testing.T) {
	// riGan converts lunar→solar→day pillar.
	// Verify it returns a valid stem (1-10) and is consistent with direct solar computation.
	tests := []struct {
		name       string
		liuYear    int
		lunarMonth int
		lunarDay   int
	}{
		{"2024-01-01", 2024, 1, 1},
		{"2024-06-15", 2024, 6, 15},
		{"2025-12-01", 2025, 12, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := riGan(tt.liuYear, tt.lunarMonth, tt.lunarDay)
			if int(got) < 1 || int(got) > 10 {
				t.Errorf("riGan(%d,%d,%d) = %d, out of range [1,10]",
					tt.liuYear, tt.lunarMonth, tt.lunarDay, int(got))
			}

			// Cross-check: convert lunar→solar, compute RiZhu, compare Gan
			gt := tianwen.LunarToGregorian(tianwen.LunarTime{Year: tt.liuYear, Month: tt.lunarMonth, Day: tt.lunarDay, Leap: false})
			if gt.Time().IsZero() {
				gt = tianwen.LunarToGregorian(tianwen.LunarTime{Year: tt.liuYear - 1, Month: tt.lunarMonth, Day: tt.lunarDay, Leap: false})
			}
			if !gt.Time().IsZero() {
				dp := tianwen.RiZhu(gt)
				if dp.Gan != got {
					sy, sm, sd := gt.Time().Date()
					t.Errorf("riGan(%d,%d,%d) = %d, but RiZhu(solar=%d-%d-%d).Gan = %d",
						tt.liuYear, tt.lunarMonth, tt.lunarDay, int(got),
						sy, int(sm), sd, int(dp.Gan))
				}
			}
		})
	}
}

// ── 排盘集成测试 ──

func TestComputeChart_Integration(t *testing.T) {
	// 已知八字排盘: 2024-02-10 12:00 (春节), 男性
	// 查农历: 2024年正月初一, 午时
	st := solarTimeForDate(2024, 2, 10, 12, 0)
	chart := ComputeChart(st, ganzhi.Male)

	// 验证基本结构
	if chart.MingGong != 0 {
		t.Errorf("MingGong should be 0, got %d", chart.MingGong)
	}
	if chart.JuShu <= 0 {
		t.Errorf("JuShu should be positive, got %d", chart.JuShu)
	}
	if len(chart.Palaces) != 12 {
		t.Errorf("expected 12 palaces, got %d", len(chart.Palaces))
	}

	// 命宫应有地支
	if chart.Palaces[0].Zhi == 0 {
		t.Error("命宫地支不应为0")
	}
	// 应有主星
	hasMajorStar := false
	for _, s := range chart.Palaces[0].Stars {
		if s.IsMajor {
			hasMajorStar = true
			break
		}
	}
	if !hasMajorStar {
		t.Log("命宫无主星可能是空宫(正常)")
	}

	// SiHua 应有4颗星
	if len(chart.SiHua) != 4 {
		t.Errorf("expected 4 sihua results, got %d", len(chart.SiHua))
	}

	t.Logf("Chart: YearGan=%d, HourZhi=%d, JuShu=%s(%d), ZiweiPos=%d",
		int(chart.YearGan), int(chart.HourZhi), chart.JuShuName, int(chart.JuShu), int(chart.ZiweiPos))
	t.Logf("MingGong: Gan=%d Zhi=%d Stars=%d",
		int(chart.Palaces[0].Gan), int(chart.Palaces[0].Zhi), len(chart.Palaces[0].Stars))
}

// ── 大限方向 ──

func TestIsDaXianForward(t *testing.T) {
	// 阳男阴女→顺行, 阴男阳女→逆行
	// 阳干: 1甲,3丙,5戊,7庚,9壬
	// 阴干: 2乙,4丁,6己,8辛,10癸
	tests := []struct {
		name      string
		gender    ganzhi.Gender
		yearGan   Gan
		wantFwd   bool
	}{
		{"甲年(阳)男→顺", ganzhi.Male, 1, true},
		{"乙年(阴)男→逆", ganzhi.Male, 2, false},
		{"甲年(阳)女→逆", ganzhi.Female, 1, false},
		{"乙年(阴)女→顺", ganzhi.Female, 2, true},
		{"丙年(阳)男→顺", ganzhi.Male, 3, true},
		{"丁年(阴)女→顺", ganzhi.Female, 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDaXianForward(tt.gender, tt.yearGan)
			if got != tt.wantFwd {
				t.Errorf("isDaXianForward(%v,%d) = %v, want %v",
					tt.gender, int(tt.yearGan), got, tt.wantFwd)
			}
		})
	}
}

// ── 流年 ──

func TestLiuNianMingGong(t *testing.T) {
	// 流年命宫 = (流年 - 出生年) % 12 从命宫偏移
	tests := []struct {
		name      string
		liuYear   int
		birthYear int
		wantOff   palaceIndex
	}{
		{"同一年", 2024, 2024, 0},
		{"下一年", 2025, 2024, 1},
		{"过12年", 2036, 2024, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := liuNianMingGong(tt.liuYear, tt.birthYear)
			if got != tt.wantOff {
				t.Errorf("liuNianMingGong(%d,%d) = %d, want %d",
					tt.liuYear, tt.birthYear, int(got), int(tt.wantOff))
			}
		})
	}
}

// ── 庙旺利陷 ──

func TestMiaoWang_KnownValues(t *testing.T) {
	tests := []struct {
		name    string
		star    starIndex
		zhi     Zhi
		want    brightness
	}{
		// 紫微: 午=庙, 丑=庙, 卯=旺, 子=平
		{"紫微居午庙", ZiWei, 7, Miao},
		{"紫微居丑庙", ZiWei, 2, Miao},
		{"紫微居卯旺", ZiWei, 4, Wang},
		{"紫微居子平", ZiWei, 1, Ping},
		// 太阳: 卯=庙, 午=庙, 子=陷, 亥=陷
		{"太阳居卯庙", TaiYang, 4, Miao},
		{"太阳居午庙", TaiYang, 7, Miao},
		{"太阳居子陷", TaiYang, 1, Xian},
		{"太阳居亥陷", TaiYang, 12, Xian},
		// 太阴: 亥=庙, 子=庙, 卯=陷, 辰=陷
		{"太阴居亥庙", TaiYin, 12, Miao},
		{"太阴居子庙", TaiYin, 1, Miao},
		{"太阴居卯陷", TaiYin, 4, Xian},
		// 廉贞: 寅=庙, 申=利
		{"廉贞居寅庙", LianZhen, 3, Miao},
		{"廉贞居申利", LianZhen, 9, Li},
		// 贪狼: 丑=庙, 酉=庙
		{"贪狼居丑庙", TanLang, 2, Miao},
		{"贪狼居酉庙", TanLang, 10, Miao},
		// 破军: 子=旺, 午=旺
		{"破军居子旺", PoJun, 1, Wang},
		{"破军居午旺", PoJun, 7, Wang},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := miaoWang(tt.star, tt.zhi)
			if got != tt.want {
				t.Errorf("miaoWang(%s,%s) = %s, want %s",
					starName(tt.star), tt.zhi.String(), got, tt.want)
			}
		})
	}
}

// ── 大限 ──

func TestComputeDaXian(t *testing.T) {
	// 验证大限起始年龄、步数、方向
	chart := Chart{
		Gender:  Male,
		YearGan: 1,  // 甲=阳
		JuShu:   JuMetal, // 金四局, 起运4岁
	}
	steps := ComputeDaXian(chart)

	if len(steps) != 12 {
		t.Fatalf("expected 12 steps, got %d", len(steps))
	}
	if steps[0].StartAge != 4 {
		t.Errorf("first step start age = %d, want 4 (juShu=4)", steps[0].StartAge)
	}
	if steps[0].Palace != 0 {
		t.Errorf("first step palace = %d, want 0 (命宫)", steps[0].Palace)
	}
	// 阳男顺行, 第二步应在兄弟宫(1)
	if steps[1].Palace != 1 {
		t.Errorf("second step palace = %d, want 1 (顺行)", steps[1].Palace)
	}
	if steps[11].StartAge != 4+11*10 {
		t.Errorf("last step start age = %d, want %d", steps[11].StartAge, 4+11*10)
	}

	// 阴女→顺行 (yearGan=2 阴, female → 阴女顺行)
	chart2 := Chart{Gender: Female, YearGan: 2, JuShu: JuWater}
	steps2 := ComputeDaXian(chart2)
	if steps2[1].Palace != 1 {
		t.Errorf("阴女顺行: second step = %d, want 1 (顺行)", steps2[1].Palace)
	}

	// 阴男→逆行 (yearGan=2 阴, male)
	chart3 := Chart{Gender: Male, YearGan: 2, JuShu: JuFire}
	steps3 := ComputeDaXian(chart3)
	if steps3[1].Palace != 11 {
		t.Errorf("阴男逆行: second step = %d, want 11 (逆行)", steps3[1].Palace)
	}
}

// ── yearGan ──

func TestYearGan(t *testing.T) {
	tests := []struct {
		name string
		year int
		want Gan
	}{
		{"2024甲辰", 2024, 1},
		{"2025乙巳", 2025, 2},
		{"2026丙午", 2026, 3},
		{"2027丁未", 2027, 4},
		{"2028戊申", 2028, 5},
		{"2029己酉", 2029, 6},
		{"2030庚戌", 2030, 7},
		{"2031辛亥", 2031, 8},
		{"2032壬子", 2032, 9},
		{"2033癸丑", 2033, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yearGan(tt.year)
			if got != tt.want {
				t.Errorf("yearGan(%d) = %d, want %d", tt.year, int(got), int(tt.want))
			}
		})
	}
}

// ── 流月 ──

func TestLiuYueSiHua(t *testing.T) {
	// 流月四化: 月干从寅宫五虎遁起
	// 正月(寅) with 甲年干 → 丙寅月, monthGan=3
	// computeSiHua(3=丙): 天同禄, 天机权, 文昌科, 廉贞忌
	got := liuYueSiHua(1, 1) // lunarMonth=1(寅), yearGan=1(甲)
	if len(got) != 4 {
		t.Fatalf("expected 4 sihua, got %d", len(got))
	}
	if got[TianTong] != HuaLu {
		t.Errorf("丙年: want 天同禄, got %v", got[TianTong])
	}
	if got[TianJi] != HuaQuan {
		t.Errorf("丙年: want 天机权, got %v", got[TianJi])
	}

	// 七月(申) with 乙年干 → 甲申月, monthGan=1
	// computeSiHua(1=甲): 廉贞禄, 破军权, 武曲科, 太阳忌
	got2 := liuYueSiHua(7, 2)
	if got2[LianZhen] != HuaLu {
		t.Errorf("甲月: want 廉贞禄, got %v", got2[LianZhen])
	}
}

// ── 剩余辅星 ──

func TestTianKuiTianYuePos(t *testing.T) {
	tests := []struct {
		name     string
		yearGan  Gan
		wantKui  int
		wantYue  int // tianYue = (tianKui + 6) % 12
	}{
		{"甲", 1, 1, 7},
		{"乙", 2, 0, 6},
		{"丙", 3, 11, 5},
		{"庚", 7, 1, 7},
		{"壬", 9, 3, 9},
		{"癸", 10, 5, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := tianKuiPos(tt.yearGan)
			if tk != tt.wantKui {
				t.Errorf("tianKuiPos(%d) = %d, want %d", tt.yearGan, tk, tt.wantKui)
			}
			ty := tianYuePos(tk)
			if ty != tt.wantYue {
				t.Errorf("tianYuePos(%d) = %d, want %d", tk, ty, tt.wantYue)
			}
		})
	}
}

func TestTianMaPos(t *testing.T) {
	// 天马: 寅午戌在申(8), 申子辰在寅(2), 巳酉丑在亥(11), 亥卯未在巳(5)
	tests := []struct {
		name    string
		yearZhi Zhi
		want    int
	}{
		{"寅→申", 3, 8},
		{"午→申", 7, 8},
		{"戌→申", 11, 8},
		{"申→寅", 9, 2},
		{"子→寅", 1, 2},
		{"巳→亥", 6, 11},
		{"卯→巳", 4, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tianMaPos(tt.yearZhi); got != tt.want {
				t.Errorf("tianMaPos(%d) = %d, want %d",
					tt.yearZhi, got, tt.want)
			}
		})
	}
}

func TestDiKongDiJiePos(t *testing.T) {
	tests := []struct {
		name       string
		hourZhi    Zhi
		wantKong   int
		wantJie    int
	}{
		{"子时", 1, 11, 11},
		{"午时", 7, 5, 5},
		{"卯时", 4, 8, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diKongPos(tt.hourZhi); got != tt.wantKong {
				t.Errorf("diKongPos(%d) = %d, want %d", int(tt.hourZhi), got, tt.wantKong)
			}
			if got := diJiePos(tt.hourZhi); got != tt.wantJie {
				t.Errorf("diJiePos(%d) = %d, want %d", int(tt.hourZhi), got, tt.wantJie)
			}
		})
	}
}

// ── 格局检测 ──

func TestFindPatterns_ZiWei(t *testing.T) {
	// "紫微朝垣": 紫微在命宫
	var palaces [12]palace
	palaces[0] = palace{
		Index: 0, Zhi: 7, // 命宫在午
		Stars: []starInfo{{Star: ZiWei, Name: "紫微", IsMajor: true}},
	}
	// set zhi for all palaces
	for i := range palaces {
		palaces[i].Index = palaceIndex(i)
		if palaces[i].Zhi == 0 {
			palaces[i].Zhi = Zhi(((int(palaces[0].Zhi)-1-i)%12+12)%12 + 1)
		}
	}
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "紫微朝垣" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '紫微朝垣' not found")
	}
}

func TestFindPatterns_ShaPoLang(t *testing.T) {
	// 杀破狼: 命宫三方有七杀、破军、贪狼
	var palaces [12]palace
	palaces[0] = palace{Index: 0, Zhi: 1, Stars: []starInfo{{Star: QiSha, Name: "七杀", IsMajor: true}}}
	palaces[4] = palace{Index: 4, Zhi: 5, Stars: []starInfo{{Star: PoJun, Name: "破军", IsMajor: true}}}  // 财帛=命宫+4
	palaces[8] = palace{Index: 8, Zhi: 9, Stars: []starInfo{{Star: TanLang, Name: "贪狼", IsMajor: true}}} // 官禄=命宫+8
	for i := range palaces {
		palaces[i].Index = palaceIndex(i)
		if palaces[i].Zhi == 0 {
			palaces[i].Zhi = Zhi(((int(palaces[0].Zhi)-1-i)%12+12)%12 + 1)
		}
	}
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "杀破狼" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '杀破狼' not found")
	}
}

func TestFindPatterns_SunMoon(t *testing.T) {
	// 日月反背: 太阳+太阴都在陷
	var palaces [12]palace
	// 太阳落陷在子(Zhi=1)
	palaces[0] = palace{Index: 0, Zhi: 1, Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}}}
	// 太阴落陷在卯(Zhi=4)
	palaces[4] = palace{Index: 4, Zhi: 4, Stars: []starInfo{{Star: TaiYin, Name: "太阴", IsMajor: true}}}
	for i := range palaces {
		palaces[i].Index = palaceIndex(i)
		if palaces[i].Zhi == 0 {
			palaces[i].Zhi = Zhi(((int(palaces[0].Zhi)-1-i)%12+12)%12 + 1)
		}
	}
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "日月反背" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '日月反背' not found")
	}
}

// ── 合盘 ──

func TestComputeBond(t *testing.T) {
	// 两张完全相同命盘，宫位互入应为命宫(0)
	a := Chart{
		Palaces: [12]palace{
			{Index: 0, Zhi: 1, Stars: []starInfo{{Star: ZiWei, Name: "紫微", IsMajor: true}}},
		},
		SiHua: siHuaResult{ZiWei: HuaLu},
	}
	// init remaining palaces zhi (counter-clockwise from 0=子)
	for i := 1; i < 12; i++ {
		a.Palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(((-i)%12+12)%12 + 1)}
	}
	b := a

	bond := ComputeBond(a, b)
	if bond.AIntoB != 0 {
		t.Errorf("same chart: AIntoB = %d, want 0 (命宫)", bond.AIntoB)
	}
	if bond.BIntoA != 0 {
		t.Errorf("same chart: BIntoA = %d, want 0 (命宫)", bond.BIntoA)
	}
	if len(bond.StarCross) < 1 {
		t.Error("star cross should have at least 1 entry")
	}
	if len(bond.SiHuaCross) < 1 {
		t.Error("sihua cross should have at least 1 entry")
	}

	// 两张命宫不同地支的命盘
	a2 := Chart{Palaces: [12]palace{{Index: 0, Zhi: 1}}}  // 命宫在子
	b2 := Chart{Palaces: [12]palace{{Index: 0, Zhi: 7}}}  // 命宫在午
	for i := 1; i < 12; i++ {
		a2.Palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(((-i)%12+12)%12 + 1)}
		b2.Palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(((7-1-i)%12+12)%12 + 1)}
	}
	bond2 := ComputeBond(a2, b2)
	// A命宫在子, B's 迁移宫(6)是子 → A入B的迁移宫
	if bond2.AIntoB != 6 {
		t.Errorf("a子→b午: AIntoB = %d, want 6 (迁移)", bond2.AIntoB)
	}
	// B命宫在午, A's 迁移宫(6)是午 → B入A的迁移宫
	if bond2.BIntoA != 6 {
		t.Errorf("a子→b午: BIntoA = %d, want 6 (迁移)", bond2.BIntoA)
	}
}

// ── 流年四化 ──

func TestLiuNianSiHua(t *testing.T) {
	// 2024=甲辰年 → 甲年四化
	got := liuNianSiHua(2024)
	if len(got) != 4 {
		t.Fatalf("expected 4 sihua, got %d", len(got))
	}
	if got[LianZhen] != HuaLu {
		t.Errorf("甲年: want 廉贞禄, got %v", got[LianZhen])
	}
	if got[PoJun] != HuaQuan {
		t.Errorf("甲年: want 破军权, got %v", got[PoJun])
	}
}

// ── helpers ──

func solarTimeForDate(year, month, day, hour, minute int) tianwen.SolarTime {
	loc := time.FixedZone("test", 8*3600) // nolint
	return tianwen.SolarTime(time.Date(year, time.Month(month), day, hour, minute, 0, 0, loc))
}

func makeChart() Chart {
	var chart Chart
	chart.YearGan = 1  // 甲
	chart.HourZhi = 7  // 午
	chart.BirthYear = 2000
	chart.Gender = Male
	chart.JuShu = JuMetal
	chart.ZiweiPos = 0
	// Build palace zhis counter-clockwise from mingZhi=子(1).
	var palaces [12]palace
	for i := 0; i < 12; i++ {
		palaces[i] = palace{
			Index: palaceIndex(i),
			Name:  PalaceNames[i],
			Gan:   Gan(((int(chart.YearGan)-1-i)%10+10)%10 + 1),
			Zhi:   Zhi(((-i)%12+12)%12 + 1),
			Stars: []starInfo{
				{Star: ZiWei, Name: "紫微", IsMajor: true},
				{Star: TianJi, Name: "天机", IsMajor: true},
			},
		}
	}
	chart.Palaces = palaces
	chart.SiHua = computeSiHua(chart.YearGan)
	return chart
}

// =============================================================================
// 流年 — ComputeLiuNian / liuNianMinors / liuNianMingGong
// =============================================================================

func TestComputeLiuNian(t *testing.T) {
	chart := makeChart()
	ln := ComputeLiuNian(chart, 2005)

	if ln.MingGong != 5 { // (2005-2000)=5, xuSui=6, (6-1)%12=5
		t.Errorf("MingGong = %d, want 5", ln.MingGong)
	}
	if ln.MingGongName == "" {
		t.Error("MingGongName is empty")
	}
	if len(ln.SiHua) != 4 {
		t.Errorf("SiHua len = %d, want 4", len(ln.SiHua))
	}
	// SiHuaPalace maps stars to palaces where they appear.
	if ln.SiHuaPalace == nil {
		t.Error("SiHuaPalace is nil")
	}
	// MinorStars: 擎羊,陀罗,火星,铃星 should be 4.
	if len(ln.MinorStars) != 4 {
		t.Errorf("MinorStars len = %d, want 4", len(ln.MinorStars))
	}
}

func TestLiuNianMinors(t *testing.T) {
	// 甲年(1), 辰支(5), 午时(7)
	m := liuNianMinors(ganzhi.Zhu{Gan: 1, Zhi: 5}, 7)
	if len(m) != 4 {
		t.Fatalf("liuNianMinors len = %d, want 4", len(m))
	}
	if _, ok := m[QingYang]; !ok {
		t.Error("missing QingYang")
	}
	if _, ok := m[TuoLuo]; !ok {
		t.Error("missing TuoLuo")
	}
	if _, ok := m[HuoXing]; !ok {
		t.Error("missing HuoXing")
	}
	if _, ok := m[LingXing]; !ok {
		t.Error("missing LingXing")
	}
}

// =============================================================================
// 流月 — ComputeLiuYue / liuYueMingGong
// =============================================================================

func TestComputeLiuYue(t *testing.T) {
	chart := makeChart()
	ly := ComputeLiuYue(chart, 2005, 1) // 2005年正月

	if ly.MingGong >= 12 {
		t.Errorf("MingGong out of range: %d", ly.MingGong)
	}
	if ly.MingGongName == "" {
		t.Error("MingGongName is empty")
	}
	if len(ly.SiHua) != 4 {
		t.Errorf("SiHua len = %d, want 4", len(ly.SiHua))
	}
}

func TestLiuYueMingGong(t *testing.T) {
	// liuYueMingGong = (liuNianMing + lunarMonth - 1) % 12
	tests := []struct {
		name     string
		lunarMonth int
		liuNianMing palaceIndex
		want     palaceIndex
	}{
		{"正月", 1, 0, 0},
		{"二月", 2, 0, 1},
		{"十二月", 12, 0, 11},
		{"正月偏移", 1, 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := liuYueMingGong(tt.lunarMonth, tt.liuNianMing)
			if got != tt.want {
				t.Errorf("liuYueMingGong(%d,%d) = %d, want %d",
					tt.lunarMonth, tt.liuNianMing, got, tt.want)
			}
		})
	}
}

// =============================================================================
// 流日 — ComputeLiuRi / liuRiMingGong / liuRiSiHua
// =============================================================================

func TestComputeLiuRi(t *testing.T) {
	chart := makeChart()
	lr := ComputeLiuRi(chart, 2005, 1, 1)

	if lr.MingGong >= 12 {
		t.Errorf("MingGong out of range: %d", lr.MingGong)
	}
	if lr.MingGongName == "" {
		t.Error("MingGongName is empty")
	}
	if len(lr.SiHua) != 4 {
		t.Errorf("SiHua len = %d, want 4", len(lr.SiHua))
	}
}

func TestLiuRiMingGong(t *testing.T) {
	got := liuRiMingGong(1, 0) // day 1, yue ming=0
	if got != 0 {
		t.Errorf("liuRiMingGong(1,0) = %d, want 0", got)
	}
	got2 := liuRiMingGong(5, 3) // (3+4)%12 = 7
	if got2 != 7 {
		t.Errorf("liuRiMingGong(5,3) = %d, want 7", got2)
	}
}

func TestLiuRiSiHua(t *testing.T) {
	// liuRiSiHua is computeSiHua(dayGan) wrapper
	got := liuRiSiHua(1) // 甲
	if len(got) != 4 {
		t.Errorf("len = %d, want 4", len(got))
	}
}

// =============================================================================
// 补充 riGan 边缘情况 — fallback 路径
// =============================================================================

func TestRiGan_Fallback(t *testing.T) {
	// Test riGan for lunar dates near year boundary.
	// 农历2024年正月初一 = 2024-02-10 solar, day pillar = 甲辰 → riGan=甲(1)
	if g := riGan(2024, 1, 1); int(g) != 1 {
		t.Errorf("riGan(2024,1,1) = %d, want 1 (甲)", int(g))
	}
	// 农历2024年十二月三十 = 2025-01-29 solar → riGan=癸(10)
	if g := riGan(2024, 12, 30); int(g) != 10 {
		t.Errorf("riGan(2024,12,30) = %d, want 10 (癸)", int(g))
	}
	// 农历2023年正月初一: tests year-crossing in LunarToSolar
	if g := riGan(2023, 1, 1); int(g) != 7 {
		t.Errorf("riGan(2023,1,1) = %d, want 7 (庚)", int(g))
	}
}

// =============================================================================
// 补充 miaoWang — 无效星/无效地支
// =============================================================================

func TestMiaoWang_InvalidInput(t *testing.T) {
	// Invalid star (<0 or >=14) → Ping.
	if got := miaoWang(-1, 1); got != Ping {
		t.Errorf("miaoWang(-1,1) = %s, want Ping", got)
	}
	if got := miaoWang(14, 1); got != Ping {
		t.Errorf("miaoWang(14,1) = %s, want Ping", got)
	}
	// Invalid zhi (<1 or >12) → Ping.
	if got := miaoWang(ZiWei, 0); got != Ping {
		t.Errorf("miaoWang(ZiWei,0) = %s, want Ping", got)
	}
	if got := miaoWang(ZiWei, 13); got != Ping {
		t.Errorf("miaoWang(ZiWei,13) = %s, want Ping", got)
	}
}

// =============================================================================
// isXian / qingYangMiao — 亮度辅助
// =============================================================================

func TestIsXian(t *testing.T) {
	// 太阳居子=陷 → isXian true
	if !isXian(1, TaiYang) {
		t.Error("TaiYang at 子 should be Xian")
	}
	// 太阳居午=庙 → isXian false
	if isXian(7, TaiYang) {
		t.Error("TaiYang at 午 should NOT be Xian")
	}
}

func TestQingYangMiao(t *testing.T) {
	// 擎羊入庙: 辰戌丑未 (5,11,2,8)
	for _, z := range []Zhi{5, 11, 2, 8} {
		if !qingYangMiao(z) {
			t.Errorf("qingYangMiao(%d)=false, want true (辰戌丑未)", int(z))
		}
	}
	for _, z := range []Zhi{1, 3, 4, 6, 7, 9, 10, 12} {
		if qingYangMiao(z) {
			t.Errorf("qingYangMiao(%d)=true, want false", int(z))
		}
	}
}

// =============================================================================
// 补充 sunMoonBright — false case
// =============================================================================

func TestSunMoonBright_False(t *testing.T) {
	// 所有宫位都没有庙旺的太阳/太阴.
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: 1} // all at 子
	}
	if sunMoonBright(palaces) {
		t.Error("expected false when no bright sun/moon")
	}
}

// =============================================================================
// 补充 sfSiHuaCount — hit case
// =============================================================================

func TestSfSiHuaCount_Hit(t *testing.T) {
	var palaces [12]palace
	palaces[0] = palace{
		Index: 0, Zhi: 1,
		Stars: []starInfo{{Star: ZiWei, Name: "紫微", IsMajor: true, SiHua: "禄"}},
	}
	palaces[4] = palace{
		Index: 4, Zhi: 5,
		Stars: []starInfo{{Star: TianJi, Name: "天机", IsMajor: true, SiHua: "禄"}},
	}
	sf := sanFang(0)                       // 命财官迁
	c := sfSiHuaCount(palaces, sf, HuaLu) // 应找到2颗化禄
	if c != 2 {
		t.Errorf("sfSiHuaCount = %d, want 2", c)
	}
}

// =============================================================================
// 补充 computeSiHua — 无效天干
// =============================================================================

func TestComputeSiHua_InvalidGan(t *testing.T) {
	if got := computeSiHua(0); got != nil {
		t.Errorf("computeSiHua(0) = %v, want nil", got)
	}
	if got := computeSiHua(11); got != nil {
		t.Errorf("computeSiHua(11) = %v, want nil", got)
	}
}

// =============================================================================
// 补充 findZiwei — 无效局数
// =============================================================================

func TestFindZiwei_InvalidJuShu(t *testing.T) {
	if got := findZiwei(0, 1); got != 0 {
		t.Errorf("findZiwei(0,1) = %d, want 0", int(got))
	}
	if got := findZiwei(7, 1); got != 0 {
		t.Errorf("findZiwei(7,1) = %d, want 0", int(got))
	}
}

// =============================================================================
// 补充 luCunPos — 无效年干
// =============================================================================

func TestLuCunPos_InvalidGan(t *testing.T) {
	if got := luCunPos(0); got != 0 {
		t.Errorf("luCunPos(0) = %d, want 0", got)
	}
	if got := luCunPos(11); got != 0 {
		t.Errorf("luCunPos(11) = %d, want 0", got)
	}
}

// =============================================================================
// 补充 tianKuiPos — 无效年干
// =============================================================================

func TestTianKuiPos_InvalidGan(t *testing.T) {
	if got := tianKuiPos(0); got != 0 {
		t.Errorf("tianKuiPos(0) = %d, want 0", got)
	}
	if got := tianKuiPos(11); got != 0 {
		t.Errorf("tianKuiPos(11) = %d, want 0", got)
	}
}

// =============================================================================
// 补充 tianMaPos — 无效年支
// =============================================================================

func TestTianMaPos_InvalidZhi(t *testing.T) {
	if got := tianMaPos(0); got != 0 {
		t.Errorf("tianMaPos(0) = %d, want 0", got)
	}
	if got := tianMaPos(13); got != 0 {
		t.Errorf("tianMaPos(13) = %d, want 0", got)
	}
}

// =============================================================================
// 补充 huoXingIndex / lingXingIndex — 非三合组
// =============================================================================

func TestMarsIndex_Other(t *testing.T) {
	// Non-group zhi (not in any of the 4 groups) → default 0.
	// All zhi 1-12 are covered by the 4 groups. But what if value out of range?
	// The switch has cases for all 4 sanhe groups; any unmatch falls to default 0.
	// Since all 12 zhi are covered, just verify behaviour.
	if got := huoXingIndex(1, 1); got != 3 { // 子→申子辰 → (1+2)%12=3
		t.Errorf("huoXingIndex(1,1) = %d, want 3", got)
	}
}

func TestLingxingIndex_Group1(t *testing.T) {
	// 寅午戌 → (hourZhi+3)%12
	if got := lingXingIndex(3, 1); got != 4 {
		t.Errorf("lingXingIndex(3,1) = %d, want 4", got)
	}
}

// =============================================================================
// 补充 findShenGongIndex — 未找到 (返回0)
// =============================================================================

func TestFindShenGongIndex_NotFound(t *testing.T) {
	var zhis [12]Zhi
	for i := 0; i < 12; i++ {
		zhis[i] = Zhi(i + 1)
	}
	// Shen zhi doesn't match any palace zhi.
	if got := findShenGongIndex(zhis, 0); got != 0 {
		t.Errorf("findShenGongIndex(...,0) = %d, want 0", got)
	}
}

// =============================================================================
// 补充 findPalaceByZhi — 未找到 (返回0)
// =============================================================================

func TestFindPalaceByZhi_NotFound(t *testing.T) {
	var palaces [12]palace
	for i := 0; i < 12; i++ {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	if got := findPalaceByZhi(palaces, 13); got != 0 {
		t.Errorf("findPalaceByZhi(...,13) = %d, want 0", got)
	}
}

// =============================================================================
// juShuFromWuxing — 无效五行
// =============================================================================

func TestJuShuFromWuxing_Invalid(t *testing.T) {
	if got := juShuFromWuxing(0); got != 0 {
		t.Errorf("juShuFromWuxing(0) = %d, want 0", got)
	}
	if got := juShuFromWuxing(6); got != 0 { // only 1-5 valid
		t.Errorf("juShuFromWuxing(6) = %d, want 0", got)
	}
}

// =============================================================================
// 格局 — 日丽中天 / 月朗天门 / 双禄朝垣
// =============================================================================

func TestFindPatterns_RiLiZhongTian(t *testing.T) {
	// "日丽中天": 太阳居午宫 (index 6 → 迁移宫)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	palaces[6] = palace{Index: 6, Zhi: 7, Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}}}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "日丽中天" {
			return // found
		}
	}
	t.Error("expected pattern '日丽中天' not found")
}

func TestFindPatterns_YueLangTianMen(t *testing.T) {
	// "月朗天门": 太阴居亥宫 (index 11 → 父母宫)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	palaces[11] = palace{Index: 11, Zhi: 12, Stars: []starInfo{{Star: TaiYin, Name: "太阴", IsMajor: true}}}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "月朗天门" {
			return // found
		}
	}
	t.Error("expected pattern '月朗天门' not found")
}

func TestFindPatterns_HuoTanGe(t *testing.T) {
	// "火贪格": 火星+贪狼同宫, 贪狼不陷
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	// 贪狼在丑=庙 (不陷)
	palaces[0] = palace{Index: 0, Zhi: 2, // 命宫在丑
		Stars: []starInfo{
			{Star: HuoXing, Name: "火星", IsMajor: false},
			{Star: TanLang, Name: "贪狼", IsMajor: true},
		},
	}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "火贪格" {
			return
		}
	}
	t.Error("expected pattern '火贪格' not found")
}

func TestFindPatterns_ShuangLuChaoYuan(t *testing.T) {
	// "双禄朝垣": 命宫三方 >= 2 化禄
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	palaces[0] = palace{Index: 0, Zhi: 1,
		Stars: []starInfo{{Star: ZiWei, Name: "紫微", IsMajor: true, SiHua: "禄"}},
	}
	palaces[4] = palace{Index: 4, Zhi: 5,
		Stars: []starInfo{{Star: TianJi, Name: "天机", IsMajor: true, SiHua: "禄"}},
	}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "双禄朝垣" {
			return
		}
	}
	t.Error("expected pattern '双禄朝垣' not found")
}

func TestFindPatterns_XiongSuQianYuan(t *testing.T) {
	// "雄宿乾元": 破军在命宫且入庙
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	// 破军在子=旺 (miaoWang <= Wang)
	palaces[0] = palace{Index: 0, Zhi: 1,
		Stars: []starInfo{{Star: PoJun, Name: "破军", IsMajor: true}},
	}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "雄宿乾元" {
			return
		}
	}
	t.Error("expected pattern '雄宿乾元' not found")
}

// =============================================================================
// 格局 — 日月并明
// =============================================================================

func TestFindPatterns_RiYueBingMing(t *testing.T) {
	// 日月并明: both sun AND moon must be bright.
	// 太阳 at 午 (Zhi=7, 庙) in palace 6 (迁移).
	// 太阴 at 子 (Zhi=1, 庙) in palace 0 (命).
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi((i)%12 + 1)}
	}
	palaces[6] = palace{Index: 6, Zhi: 7, // 午=7, 太阳庙
		Stars: []starInfo{{Star: TaiYang, Name: "太阳", IsMajor: true}},
	}
	palaces[0] = palace{Index: 0, Zhi: 1, // 子=1, 太阴庙
		Stars: []starInfo{{Star: TaiYin, Name: "太阴", IsMajor: true}},
	}
	patterns := findPatterns(palaces)
	for _, p := range patterns {
		if p.Name == "日月并明" {
			return
		}
	}
	t.Error("expected pattern '日月并明' not found")
}

// =============================================================================
// liuNianSiHua — 不同年份
// =============================================================================

func TestLiuNianSiHua_VariousYears(t *testing.T) {
	// 2025=乙年 → 天机禄, 天梁权, 紫微科, 太阴忌
	got := liuNianSiHua(2025)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	if got[TianJi] != HuaLu {
		t.Errorf("乙年: want 天机禄, got %v", got[TianJi])
	}
}

// =============================================================================
// huoXingIndex — 火星安星，四组三合局全覆盖
// =============================================================================

func TestMarsIndex_AllGroups(t *testing.T) {
	// 火星：寅午戌→丑宫起子时(hourZhi+1), 申子辰→寅宫起子时(hourZhi+2),
	//       巳酉丑→卯宫起子时(hourZhi+3), 亥卯未→酉宫起子时(hourZhi+9)
	tests := []struct {
		name     string
		yearZhi  Zhi
		hourZhi  Zhi
		want     int
	}{
		// 寅午戌组 → (hourZhi+1)%12
		{"寅午戌-子时", 3, 1, 2},   // (1+1)%12=2=丑
		{"寅午戌-午时", 3, 7, 8},   // (7+1)%12=8=未
		{"寅午戌-亥时", 3, 12, 1},  // (12+1)%12=1=子
		// 申子辰组 → (hourZhi+2)%12
		{"申子辰-子时", 9, 1, 3},   // (1+2)%12=3=寅
		{"申子辰-午时", 1, 7, 9},   // (7+2)%12=9=申
		{"申子辰-亥时", 5, 12, 2},  // (12+2)%12=2=丑
		// 巳酉丑组 → (hourZhi+3)%12
		{"巳酉丑-子时", 6, 1, 4},   // (1+3)%12=4=卯
		{"巳酉丑-午时", 10, 7, 10}, // (7+3)%12=10=酉
		{"巳酉丑-亥时", 2, 12, 3},  // (12+3)%12=3=寅
		// 亥卯未组 → (hourZhi+9)%12
		{"亥卯未-子时", 12, 1, 10}, // (1+9)%12=10=酉
		{"亥卯未-午时", 4, 7, 4},   // (7+9)%12=4=卯
		{"亥卯未-亥时", 8, 12, 9},  // (12+9)%12=9=申
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := huoXingIndex(tt.yearZhi, tt.hourZhi)
			if got != tt.want {
				t.Errorf("huoXingIndex(%d,%d)=%d, want %d", tt.yearZhi, tt.hourZhi, got, tt.want)
			}
		})
	}
}

// =============================================================================
// lingXingIndex — 铃星安星，四组三合局全覆盖
// =============================================================================

func TestLingxingIndex_AllGroups(t *testing.T) {
	// 铃星：寅午戌→卯宫起子时(hourZhi+3),
	//       申子辰/巳酉丑/亥卯未→戌宫起子时(hourZhi+10)
	tests := []struct {
		name    string
		yearZhi Zhi
		hourZhi Zhi
		want    int
	}{
		// 寅午戌组 → (hourZhi+3)%12
		{"寅午戌-子时", 3, 1, 4},   // (1+3)%12=4=卯
		{"寅午戌-午时", 7, 7, 10},  // (7+3)%12=10=酉
		{"寅午戌-亥时", 11, 12, 3}, // (12+3)%12=3=寅
		// 申子辰组 → (hourZhi+10)%12
		{"申子辰-子时", 9, 1, 11},  // (1+10)%12=11=戌
		{"申子辰-午时", 1, 7, 5},   // (7+10)%12=5=辰
		// 巳酉丑组 → (hourZhi+10)%12
		{"巳酉丑-子时", 6, 1, 11},  // (1+10)%12=11=戌
		{"巳酉丑-午时", 10, 7, 5},  // (7+10)%12=5=辰
		// 亥卯未组 → (hourZhi+10)%12
		{"亥卯未-子时", 12, 1, 11}, // (1+10)%12=11=戌
		{"亥卯未-午时", 4, 7, 5},   // (7+10)%12=5=辰
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lingXingIndex(tt.yearZhi, tt.hourZhi)
			if got != tt.want {
				t.Errorf("lingXingIndex(%d,%d)=%d, want %d", tt.yearZhi, tt.hourZhi, got, tt.want)
			}
		})
	}
}

// =============================================================================
// findPatterns — 补全格局覆盖
// =============================================================================

func TestFindPatterns_KuiYueJiaMing(t *testing.T) {
	// 魁钺夹命: 兄弟宫(1)天魁 + 父母宫(11)天钺
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	palaces[1].Stars = append(palaces[1].Stars, starInfo{Star: TianKui, Name: "天魁"})
	palaces[11].Stars = append(palaces[11].Stars, starInfo{Star: TianYue, Name: "天钺"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "魁钺夹命" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '魁钺夹命' not found")
	}
}

func TestFindPatterns_ZuoYouJiaMing(t *testing.T) {
	// 左右夹命: 兄弟宫(1)左辅 + 父母宫(11)右弼
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	palaces[1].Stars = append(palaces[1].Stars, starInfo{Star: ZuoFu, Name: "左辅"})
	palaces[11].Stars = append(palaces[11].Stars, starInfo{Star: YouBi, Name: "右弼"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "左右夹命" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '左右夹命' not found")
	}
}

func TestFindPatterns_LuMaJiaoChi(t *testing.T) {
	// 禄马交驰: 禄存+天马在命宫三方
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	// 禄存在命宫三方(0,4,8,6)中的8
	palaces[8].Stars = append(palaces[8].Stars, starInfo{Star: LuCun, Name: "禄存"})
	palaces[8].Stars = append(palaces[8].Stars, starInfo{Star: TianMa, Name: "天马"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "禄马交驰" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '禄马交驰' not found")
	}
}

func TestFindPatterns_CaiYinJiaYin(t *testing.T) {
	// 财荫夹印: 财帛宫(4)三方有化禄
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	// 财帛宫=4, 三方=(4,8,0,10)
	palaces[8].Stars = append(palaces[8].Stars, starInfo{Star: WuQu, Name: "武曲", SiHua: "禄"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "财荫夹印" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '财荫夹印' not found")
	}
}

func TestFindPatterns_JinCanGuangHui(t *testing.T) {
	// 金灿光辉: 官禄宫(8)太阳庙旺
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	// 午宫(7)太阳庙
	palaces[8].Zhi = 7 // 午
	palaces[8].Stars = append(palaces[8].Stars, starInfo{Star: TaiYang, Name: "太阳"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "金灿光辉" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '金灿光辉' not found")
	}
}

func TestFindPatterns_QingYangRuMiao(t *testing.T) {
	// 擎羊入庙: 命宫擎羊在辰戌丑未
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	palaces[0].Zhi = 5 // 辰
	palaces[0].Stars = append(palaces[0].Stars, starInfo{Star: QingYang, Name: "擎羊"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "擎羊入庙" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '擎羊入庙' not found")
	}
}

func TestFindPatterns_RiYueFanBei(t *testing.T) {
	// 日月反背: 太阳太阴双双落陷
	// 太阳陷: 子(1) or 亥(12), 太阴陷: 卯(4) or 辰(5)
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Zhi: Zhi(i + 1)}
	}
	// 太阳在子=陷
	palaces[3].Zhi = 1 // 子
	palaces[3].Stars = append(palaces[3].Stars, starInfo{Star: TaiYang, Name: "太阳"})
	// 太阴在卯=陷
	palaces[6].Zhi = 4 // 卯
	palaces[6].Stars = append(palaces[6].Stars, starInfo{Star: TaiYin, Name: "太阴"})
	patterns := findPatterns(palaces)
	found := false
	for _, p := range patterns {
		if p.Name == "日月反背" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '日月反背' not found")
	}
}

// =============================================================================
// 命盘组合 — 多种命宫组合交叉验证
// =============================================================================

func TestComputeChart_MultipleCombinations(t *testing.T) {
	tests := []struct {
		name         string
		year, month, day, hour int
		gender       ganzhi.Gender
		wantShenGong palaceIndex
		wantJuShu    juShu
		wantZiweiPos palaceIndex
	}{
		{"春节前后", 1990, 2, 5, 0, Male, 0, 5, 5},
		{"夏至前后", 1990, 7, 15, 12, Female, 0, 2, 3},
		{"冬至前后", 1990, 12, 22, 22, Male, 2, 2, 0},
		{"春分前后", 2000, 3, 20, 4, Female, 8, 2, 7},
		{"新年元旦", 1984, 1, 1, 8, Male, 4, 3, 7},
		{"年末除夕", 1984, 12, 30, 20, Female, 4, 6, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(time.Date(tt.year, time.Month(tt.month), tt.day, tt.hour, 0, 0, 0, time.FixedZone("", int(8*3600))), 120, 8)
			chart := ComputeChart(st, tt.gender)

			// All 12 palaces must have valid stem/zhi
			for i, p := range chart.Palaces {
				if p.Gan < 1 || p.Gan > 10 {
					t.Errorf("palace[%d](%s).Gan=%d, want [1,10]", i, p.Name, p.Gan)
				}
				if p.Zhi < 1 || p.Zhi > 12 {
					t.Errorf("palace[%d](%s).Zhi=%d, want [1,12]", i, p.Name, p.Zhi)
				}
			}

			if chart.ShenGong != tt.wantShenGong {
				t.Errorf("ShenGong=%d, want %d", chart.ShenGong, tt.wantShenGong)
			}
			if chart.JuShu != tt.wantJuShu {
				t.Errorf("JuShu=%d, want %d", chart.JuShu, tt.wantJuShu)
			}
			if chart.ZiweiPos != tt.wantZiweiPos {
				t.Errorf("ZiweiPos=%d, want %d", chart.ZiweiPos, tt.wantZiweiPos)
			}

			// DaXian must have 12 steps
			dx := ComputeDaXian(chart)
			if len(dx) != 12 {
				t.Errorf("DaXian len=%d, want 12", len(dx))
			}
		})
	}
}

// =============================================================================
// 四化验证 — 十天干全覆盖
// =============================================================================

func TestComputeSiHua_AllTenGan(t *testing.T) {
	// 十天干四化口诀
	// 甲: 廉破武阳 (禄权科忌)
	// 乙: 机梁紫阴
	// 丙: 同机昌廉
	// 丁: 阴同机巨
	// 戊: 贪阴右机
	// 己: 武贪梁曲
	// 庚: 阳武阴同
	// 辛: 巨阳曲昌
	// 壬: 梁紫左武
	// 癸: 破巨阴贪
	tests := []struct {
		name    string
		yearGan Gan
		wantLu  starIndex // star that gets 化禄
	}{
		{"甲→廉贞禄", Gan(1), LianZhen},
		{"乙→天机禄", Gan(2), TianJi},
		{"丙→天同禄", Gan(3), TianTong},
		{"丁→太阴禄", Gan(4), TaiYin},
		{"戊→贪狼禄", Gan(5), TanLang},
		{"己→武曲禄", Gan(6), WuQu},
		{"庚→太阳禄", Gan(7), TaiYang},
		{"辛→巨门禄", Gan(8), JuMen},
		{"壬→天梁禄", Gan(9), TianLiang},
		{"癸→破军禄", Gan(10), PoJun},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := computeSiHua(tt.yearGan)
			if sh[tt.wantLu] != HuaLu {
				t.Errorf("%s: star %d should be 化禄, got %s", tt.name, tt.wantLu, sh[tt.wantLu])
			}
			if len(sh) != 4 {
				t.Errorf("len(sihua)=%d, want 4", len(sh))
			}
		})
	}
}

// =============================================================================
// liuNianSiHua — 流年四化各种年份
// =============================================================================

func TestLiuNianSiHua_MoreYears(t *testing.T) {
	tests := []struct {
		name   string
		year   int
		wantLu starIndex
	}{
		// 2024=甲年: 廉贞禄
		{"2024甲→廉贞禄", 2024, LianZhen},
		// 2026=丙年: 天同禄
		{"2026丙→天同禄", 2026, TianTong},
		// 2027=丁年: 太阴禄
		{"2027丁→太阴禄", 2027, TaiYin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := liuNianSiHua(tt.year)
			if sh[tt.wantLu] != HuaLu {
				t.Errorf("%s: star %d should be 化禄", tt.name, tt.wantLu)
			}
		})
	}
}

// =============================================================================
// star交叉组合 — 14主星+辅星在不同宫位的交叉验证
// =============================================================================

func TestPlaceMainStars_FullCoverage(t *testing.T) {
	// 验证每个紫微位置(0-11)都能正确安14主星
	for ziweiPos := palaceIndex(0); ziweiPos < 12; ziweiPos++ {
		m := placeMainStars(ziweiPos)
		totalStars := 0
		for _, stars := range m {
			totalStars += len(stars)
		}
		if totalStars != 14 {
			t.Errorf("紫微在%d: total stars=%d, want 14", ziweiPos, totalStars)
		}
		// Verify Ziwei is at correct position
		foundZiwei := false
		for _, s := range m[ziweiPos] {
			if s == ZiWei {
				foundZiwei = true
			}
		}
		if !foundZiwei {
			t.Errorf("紫微在%d: ZiWei not found at its position", ziweiPos)
		}
		// 天府 should be at ziweiPos+2
		tianfuExpect := (ziweiPos + 2) % 12
		foundTianfu := false
		for _, s := range m[tianfuExpect] {
			if s == TianFu {
				foundTianfu = true
			}
		}
		if !foundTianfu {
			t.Errorf("紫微在%d: TianFu not found at position %d", ziweiPos, tianfuExpect)
		}
	}
}

// =============================================================================
// leap month — 闰月排盘
// =============================================================================

func TestComputeChart_LeapMonth(t *testing.T) {
	// 2025年农历闰六月十八，午时，男性
	// 闰月按传统命理规则视同下一个月（七月）计算。
	// Verify the chart is valid and consistent.
	lt := tianwen.LunarTime{Year: 2025, Month: 6, Day: 18, Leap: true}
	gt := tianwen.LunarToGregorian(lt)
	if gt.Time().IsZero() {
		t.Fatal("LunarToGregorian returned zero time for leap month date")
	}
	st := tianwen.GregorianToSolar(gt.Time(), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)

	if len(chart.Palaces) != 12 {
		t.Errorf("expected 12 palaces, got %d", len(chart.Palaces))
	}
	if chart.JuShu < 2 || chart.JuShu > 6 {
		t.Errorf("JuShu = %d, want [2,6]", chart.JuShu)
	}
	if chart.MingGong >= 12 {
		t.Errorf("MingGong = %d, want [0,11]", chart.MingGong)
	}
	for i, p := range chart.Palaces {
		if p.Gan < 1 || p.Gan > 10 {
			t.Errorf("palace[%d].Gan = %d, want [1,10]", i, p.Gan)
		}
		if p.Zhi < 1 || p.Zhi > 12 {
			t.Errorf("palace[%d].Zhi = %d, want [1,12]", i, p.Zhi)
		}
	}

	dx := ComputeDaXian(chart)
	if len(dx) != 12 {
		t.Errorf("DaXian len = %d, want 12", len(dx))
	}

	t.Logf("Leap month chart: JuShu=%s MingGong=%d ShenGong=%d ZiweiPos=%d",
		chart.JuShuName, chart.MingGong, chart.ShenGong, chart.ZiweiPos)
}

func TestComputeChart_LeapMonthVsNormalMonth(t *testing.T) {
	// 同一年，闰六月十八 vs 七月十八 — 两个命盘应该不同（月份不同）。
	// 闰月出生 → 命宫按闰月的地支算。
	ltLeap := tianwen.LunarTime{Year: 2025, Month: 6, Day: 18, Leap: true}
	gtLeap := tianwen.LunarToGregorian(ltLeap)
	if gtLeap.Time().IsZero() {
		t.Fatal("LunarToGregorian returned zero for leap month")
	}

	ltNormal := tianwen.LunarTime{Year: 2025, Month: 7, Day: 18, Leap: false}
	gtNormal := tianwen.LunarToGregorian(ltNormal)
	if gtNormal.Time().IsZero() {
		t.Fatal("LunarToGregorian returned zero for normal month")
	}

	stLeap := tianwen.GregorianToSolar(gtLeap.Time(), 116.4, 8)
	stNormal := tianwen.GregorianToSolar(gtNormal.Time(), 116.4, 8)

	chartLeap := ComputeChart(stLeap, ganzhi.Male)
	chartNormal := ComputeChart(stNormal, ganzhi.Male)

	// Month 6 (leap) vs month 7 (normal) should produce different charts.
	if chartLeap.JuShu == chartNormal.JuShu &&
		chartLeap.MingGong == chartNormal.MingGong &&
		chartLeap.ZiweiPos == chartNormal.ZiweiPos {
		t.Error("leap month 6 and normal month 7 should produce different charts")
	}
	t.Logf("Leap JuShu=%s MingGong=%d Normal JuShu=%s MingGong=%d",
		chartLeap.JuShuName, chartLeap.MingGong, chartNormal.JuShuName, chartNormal.MingGong)
}

// =============================================================================
// miaoWang — 主星庙旺陷完整覆盖
// =============================================================================

func TestMiaoWang_StarBranchMatrix(t *testing.T) {
	tests := []struct {
		name   string
		star   starIndex
		branch Zhi
		want   brightness
	}{
		// 紫微: 表{平,庙,庙,旺,利,旺,庙,平,旺,庙,旺,平}
		{"紫微子→平", ZiWei, 1, Ping},
		{"紫微丑→庙", ZiWei, 2, Miao},
		{"紫微卯→旺", ZiWei, 4, Wang},
		// 天机: 表{平,陷,庙,旺,利,旺,庙,平,利,陷,利,平}
		{"天机子→平", TianJi, 1, Ping},
		{"天机丑→陷", TianJi, 2, Xian},
		{"天机卯→旺", TianJi, 4, Wang},
		// 太阳: 表{陷,平,旺,庙,旺,旺,庙,利,利,平,平,陷}
		{"太阳子→陷", TaiYang, 1, Xian},
		{"太阳午→庙", TaiYang, 7, Miao},
		// 太阴: 表{庙,庙,平,陷,陷,平,利,利,庙,庙,庙,庙}
		{"太阴亥→庙", TaiYin, 12, Miao},
		{"太阴卯→陷", TaiYin, 4, Xian},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := miaoWang(tt.star, tt.branch)
			if got != tt.want {
				t.Errorf("miaoWang(%s,%s)=%s, want %s",
					starName(tt.star), ganzhi.ZhiName(tt.branch), got, tt.want)
			}
		})
	}
}
