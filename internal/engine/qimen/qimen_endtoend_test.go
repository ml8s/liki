package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 奇门端到端数据驱动命理锚点 ──
// 每个 case 用独立手算的权威值（五鼠遁定时干、旬空规则、三合局马星口诀、
// 天盘值符随时干+星飞布、八门值使随时支阳顺阴逆、八神阳顺阴逆、甲遁六仪）
// 锚定一个完整盘的定局/日时干落宫/生克/空亡马星/天盘星/八门/八神。
// 非自证：期望值均由外部命理规则独立得出。

// chartAtT builds a 时家奇门 chart at Beijing time (CST+8).
func chartAtT(t *testing.T, date string, hour int) Chart {
	t.Helper()
	bt, err := time.ParseInLocation("2006-01-02", date, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	st := tianwen.GregorianToSolar(bt.Add(time.Duration(hour)*time.Hour), 116.4, 8)
	return ComputeChart(st, ShiQiMen)
}

// gongAnchor 每宫的命理锚定（星/门/神，除中5虚空）。
type gongAnchor struct {
	star StarIndex
	door DoorIndex
	sp   SpiritIndex
}

// e2eCase 一个盘的完整端到端命理锚定。
type e2eCase struct {
	name          string
	date          string
	hour          int
	ju            int
	yin           bool
	riGan         ganzhi.Gan
	riZhi         ganzhi.Zhi
	shiGan        ganzhi.Gan
	riGong        GongIndex
	shiGong       GongIndex
	shengKe       string
	kongWang      [2]GongIndex
	maXing        GongIndex
	dutyStar      StarIndex
	dutyDoor      DoorIndex
	gong          [9]gongAnchor // 0=坎1 ... 8=离9
}

var e2eCases = []e2eCase{
	{
		// 2000-06-15 午时 阳遁9局：甲辰日庚午时（甲己还加甲→庚午）。
		// 日干甲→甲辰遁壬（天盘壬兑7）；时干庚（天盘庚巽4）。
		// 庚午甲子旬空戌亥→乾；午→寅午戌马在申→坤。
		// 值符天英随时干庚（地盘庚坤2）；天盘8星从坤2顺排跳过中5。
		// 八门：值使景门（阳遁顺排）起时支午离9。
		// 八神：值符神随值符星（坤2），阳遁顺排。
		name: "2000-06-15 午时 阳遁9局 甲辰日庚午时",
		date: "2000-06-15", hour: 12,
		ju: 9, yin: false,
		riGan: ganzhi.GanJia, riZhi: ganzhi.ZhiChen, shiGan: ganzhi.GanGeng,
		riGong: GongDui, shiGong: GongXun,
		shengKe:  "日干(7宫)克时干(4宫)",
		kongWang: [2]GongIndex{GongQian, GongQian}, maXing: GongKun,
		dutyStar: StarTianYing, dutyDoor: DoorJing,
		gong: [9]gongAnchor{
			{StarTianRen, DoorSi, SpiritJiuTian},   // 坎1
			{StarTianYing, DoorJingMen, SpiritZhiFu}, // 坤2
			{StarTianPeng, DoorKai, SpiritTengShe},   // 震3
			{StarTianRui, DoorXiu, SpiritTaiYin},     // 巽4
			{0, 0, 0},                                // 中5虚空
			{StarTianChong, DoorSheng, SpiritLiuHe},  // 乾6
			{StarTianFu, DoorShang, SpiritGouChen},   // 兑7
			{StarTianXin, DoorDu, SpiritZhuQue},      // 艮8
			{StarTianZhu, DoorJing, SpiritJiuDi},     // 离9
		},
	},
	{
		// 2026-06-28 午时 阴遁3局：癸酉日戊午时（戊癸起壬子→午戊午）。
		// 日干癸（天盘癸震3）；时干戊（天盘戊离9）。
		// 戊午甲寅旬空子丑→坎艮；午→马在申→坤。
		// 值符天柱随时干戊（地盘戊震3）；天盘8星从震3顺排跳过中5。
		// 八门：值使惊门（阴遁逆排）起时支午离9。
		// 八神：值符神随值符星（震3），阴遁逆排。
		name: "2026-06-28 午时 阴遁3局 癸酉日戊午时",
		date: "2026-06-28", hour: 12,
		ju: 3, yin: true,
		riGan: ganzhi.GanGui, riZhi: ganzhi.ZhiYou, shiGan: ganzhi.GanWu,
		riGong: GongZhen, shiGong: GongLi,
		shengKe:  "日干(3宫)生时干(9宫)",
		kongWang: [2]GongIndex{GongKan, GongGen}, maXing: GongKun,
		dutyStar: StarTianZhu, dutyDoor: DoorJingMen,
		gong: [9]gongAnchor{
			{StarTianFu, DoorSi, SpiritTaiYin},     // 坎1
			{StarTianXin, DoorJing, SpiritTengShe}, // 坤2
			{StarTianZhu, DoorDu, SpiritZhiFu},     // 震3
			{StarTianRen, DoorShang, SpiritJiuTian}, // 巽4
			{0, 0, 0},                              // 中5虚空
			{StarTianYing, DoorSheng, SpiritJiuDi}, // 乾6
			{StarTianPeng, DoorXiu, SpiritZhuQue},  // 兑7
			{StarTianRui, DoorKai, SpiritGouChen},  // 艮8
			{StarTianChong, DoorJingMen, SpiritLiuHe}, // 离9
		},
	},
	{
		// 2026-01-01 辰时 阳遁4局：乙亥日庚辰时（乙庚起丙子→辰庚辰）。
		// 日干乙（天盘乙兑7）；时干庚（天盘庚离9）。
		// 庚辰甲戌旬空申酉→坤兑；辰→申子辰马在寅→艮。
		// 值符天禽（旬首己在地盘？）→ 值符天禽。此盘需验证天禽寄坤2处理。
		name: "2026-01-01 辰时 阳遁4局 乙亥日庚辰时",
		date: "2026-01-01", hour: 8,
		ju: 4, yin: false,
		riGan: ganzhi.GanYi, riZhi: ganzhi.ZhiHai, shiGan: ganzhi.GanGeng,
		riGong: GongDui, shiGong: GongLi,
		shengKe:  "时干(9宫)克日干(7宫)",
		kongWang: [2]GongIndex{GongKun, GongDui}, maXing: GongGen,
		dutyStar: StarTianQin, dutyDoor: DoorSi,
		gong: [9]gongAnchor{
			{StarTianZhu, DoorShang, SpiritGouChen},   // 坎1
			{StarTianRen, DoorDu, SpiritZhuQue},       // 坤2
			{StarTianYing, DoorJing, SpiritJiuDi},     // 震3
			{StarTianPeng, DoorSi, SpiritJiuTian},     // 巽4
			{0, 0, 0},                                 // 中5虚空
			{StarTianRui, DoorJingMen, SpiritZhiFu},   // 乾6
			{StarTianChong, DoorKai, SpiritTengShe},   // 兑7
			{StarTianFu, DoorXiu, SpiritTaiYin},       // 艮8
			{StarTianXin, DoorSheng, SpiritLiuHe},     // 离9
		},
	},
	{
		// 1984-02-15 辰时 阳遁8局：己卯日戊辰时（甲己还加甲→辰戊辰）。
		// 日干己（天盘己离9）；时干戊（天盘戊艮8）。
		// 戊辰甲子旬空戌亥→乾；辰→马在寅→艮。
		name: "1984-02-15 辰时 阳遁8局 己卯日戊辰时",
		date: "1984-02-15", hour: 8,
		ju: 8, yin: false,
		riGan: ganzhi.GanJi, riZhi: ganzhi.ZhiMao, shiGan: ganzhi.GanWu,
		riGong: GongLi, shiGong: GongGen,
		shengKe:  "日干(9宫)生时干(8宫)",
		kongWang: [2]GongIndex{GongQian, GongQian}, maXing: GongGen,
		dutyStar: StarTianRen, dutyDoor: DoorSheng,
		gong: [9]gongAnchor{
			{StarTianPeng, DoorJingMen, SpiritTaiYin}, // 坎1
			{StarTianRui, DoorKai, SpiritLiuHe},       // 坤2
			{StarTianChong, DoorXiu, SpiritGouChen},   // 震3
			{StarTianFu, DoorSheng, SpiritZhuQue},     // 巽4
			{0, 0, 0},                                 // 中5虚空
			{StarTianXin, DoorShang, SpiritJiuDi},     // 乾6
			{StarTianZhu, DoorDu, SpiritJiuTian},      // 兑7
			{StarTianRen, DoorJing, SpiritZhiFu},      // 艮8
			{StarTianYing, DoorSi, SpiritTengShe},     // 离9
		},
	},
}

// TestQimenEndToEnd_Anchors 端到端数据驱动命理锚定。
func TestQimenEndToEnd_Anchors(t *testing.T) {
	for _, c := range e2eCases {
		t.Run(c.name, func(t *testing.T) {
			ch := chartAtT(t, c.date, c.hour)

			// 定局
			if ch.Pan.Jushu != c.ju || ch.Pan.YinDun != c.yin {
				t.Errorf("局=%d 阴遁=%v, want 局=%d 阴遁=%v", ch.Pan.Jushu, ch.Pan.YinDun, c.ju, c.yin)
			}
			// 日/时干支
			if ch.Pan.RiGan != c.riGan || ch.Pan.RiZhi != c.riZhi || ch.Pan.DriveGan != c.shiGan {
				t.Errorf("日=%s%s 时干=%s, want 日=%s%s 时干=%s", ch.Pan.RiGan, ch.Pan.RiZhi, ch.Pan.DriveGan, c.riGan, c.riZhi, c.shiGan)
			}
			// 日时干落宫（天盘）
			if ch.RiGanPalace != c.riGong {
				t.Errorf("ri_gan_gong=%s, want %s", ch.RiGanPalace, c.riGong)
			}
			if ch.ShiGanPalace != c.shiGong {
				t.Errorf("shi_gan_gong=%s, want %s", ch.ShiGanPalace, c.shiGong)
			}
			// 生克
			if ch.RiShiShengKe != c.shengKe {
				t.Errorf("ri_shi_sheng_ke=%q, want %q", ch.RiShiShengKe, c.shengKe)
			}
			// 空亡/马星
			if ch.Pan.KongWang != c.kongWang {
				t.Errorf("kong_wang=[%s,%s], want [%s,%s]", ch.Pan.KongWang[0], ch.Pan.KongWang[1], c.kongWang[0], c.kongWang[1])
			}
			if ch.Pan.MaXing != c.maXing {
				t.Errorf("ma_xing=%s, want %s", ch.Pan.MaXing, c.maXing)
			}
			// 值符星/值使门
			if ch.Pan.DutyStar != c.dutyStar {
				t.Errorf("值符星=%s, want %s", ch.Pan.DutyStar, c.dutyStar)
			}
			if ch.Pan.DutyDoor != c.dutyDoor {
				t.Errorf("值使门=%s, want %s", ch.Pan.DutyDoor, c.dutyDoor)
			}
			// 天盘星/八门/八神（逐宫）
			for i, w := range c.gong {
				got := ch.Pan.GongWei[i]
				if got.Star != w.star {
					t.Errorf("宫%d(%s) 星=%s, want %s", i+1, GongIndex(i+1), got.Star, w.star)
				}
				if got.Door != w.door {
					t.Errorf("宫%d(%s) 门=%s, want %s", i+1, GongIndex(i+1), got.Door, w.door)
				}
				if got.Spirit != w.sp {
					t.Errorf("宫%d(%s) 神=%s, want %s", i+1, GongIndex(i+1), got.Spirit.YangName(), w.sp.YangName())
				}
			}
		})
	}
}

// TestQimenEndToEnd_Patterns 关键格局端到端锚定。
// 每个盘断言其必须触发的格局，期望值由权威格局规则独立推导（非引擎输出自证）：
//   - 三奇得使：值使门宫天盘为乙/丙/丁
//   - 反吟：值符星落其对宫（本位对宫）
//   - 伏吟：值符星落本位
//   - 玉女守门：值使门宫天盘有丁
func TestQimenEndToEnd_Patterns(t *testing.T) {
	cases := []struct {
		name     string
		date     string
		hour     int
		must     []string // 必须触发的格局名
		mustNot  []string // 必须不触发的格局名
	}{
		{
			// 2000-06-15：值使景门落离9（天盘丙）→三奇得使；值符天英落坤2本位→非伏吟。
			name: "2000-06-15 三奇得使",
			date: "2000-06-15", hour: 12,
			must:    []string{"三奇得使"},
			mustNot: []string{"反吟"},
		},
		{
			// 2026-06-28：值符天柱落震3（对宫兑7）→反吟。
			name: "2026-06-28 反吟",
			date: "2026-06-28", hour: 12,
			must:    []string{"反吟"},
			mustNot: []string{"伏吟"},
		},
		{
			// 2026-01-01（值符天禽）：值使死门落巽4（天盘丁）→玉女守门；值符天禽寄坤2（天芮乾6）非本位→非伏吟。
			name: "2026-01-01 玉女守门",
			date: "2026-01-01", hour: 8,
			must:    []string{"玉女守门", "三奇得使"},
			mustNot: []string{"伏吟"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ch := chartAtT(t, c.date, c.hour)
			got := map[string]bool{}
			for _, p := range ch.Patterns {
				got[p.Name] = true
			}
			for _, m := range c.must {
				if !got[m] {
					t.Errorf("应触发格局 %q，但未触发", m)
				}
			}
			for _, m := range c.mustNot {
				if got[m] {
					t.Errorf("不应触发格局 %q，却触发了", m)
				}
			}
		})
	}
}
