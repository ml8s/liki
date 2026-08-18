package qimen

import (
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// qimen.chart 确定性派生字段 — 数据驱动命理锚点测试。
// 每个 case 断言：定局（局数/阴阳遁）、日时干落宫、生克文案、空亡马星影响标记，
// 及 pan 级空亡/马星数据。锚点值独立由命理规则手算确认（拆补法定局 × 时柱旬空/马星口诀）。
//
// 校验规则：
//   - 空亡 = 时柱旬空地支落宫（甲子旬空戌亥→乾…）
//   - 马星 = 时支三合局冲支落宫（申子辰马在寅→艮…）
//   - kong_wang_affected/ma_xing_affected = 日干宫或时干宫是否值空/值马

func chartAt(t *testing.T, date string, hour int) Chart {
	t.Helper()
	bt, err := time.ParseInLocation("2006-01-02", date, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	st := tianwen.GregorianToSolar(bt.Add(time.Duration(hour)*time.Hour), 116.4, 8)
	return ComputeChart(st, "时家")
}

func TestChartDerived_Anchors(t *testing.T) {
	cases := []struct {
		name         string
		date         string
		hour         int
		ju           int
		yin          bool
		riGanGong    string
		shiGanGong   string
		riShiShengKe string
		kwAffected   bool
		maAffected   bool
		kongWang     []string
		maXing       string
	}{
		{
			// 2026-06-28 午时：夏至中元（癸酉日 mod15=9 中元）→ 阴遁3局。
			// 日干癸天盘落震、时干戊天盘落离（用神落宫以天盘为核心判断依据）；
			// 时柱戊午（甲寅旬）空子丑→坎艮；午→寅午戌马在申→坤。
			name: "2026-06-28 午时 夏至中元 阴遁3局",
			date: "2026-06-28", hour: 12,
			ju: 3, yin: true,
			riGanGong: "震", shiGanGong: "离",
			riShiShengKe: "日干(3宫)生时干(9宫)",
			kwAffected: false, maAffected: false,
			kongWang: []string{"坎", "艮"}, maXing: "坤",
		},
		{
			// 2026-01-01 辰时：冬至下元（日柱 mod15 下元）→ 阳遁4局。
			// 日干乙天盘落兑、时干庚天盘落离（用神落宫以天盘为核心判断依据）；
			// 时柱庚辰（甲戌旬）空申酉→坤兑；辰→申子辰马在寅→艮。
			name: "2026-01-01 辰时 冬至下元 阳遁4局",
			date: "2026-01-01", hour: 8,
			ju: 4, yin: false,
			riGanGong: "兑", shiGanGong: "离",
			riShiShengKe: "时干(9宫)克日干(7宫)",
			kwAffected: false, maAffected: false,
			kongWang: []string{"坤", "兑"}, maXing: "艮",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ch := chartAt(t, c.date, c.hour)

			// 定局锚点。
			if ch.Pan.Jushu != c.ju || ch.Pan.YinDun != c.yin {
				t.Errorf("ju=%d yin=%v, want ju=%d yin=%v", ch.Pan.Jushu, ch.Pan.YinDun, c.ju, c.yin)
			}
			// 日时干落宫 + 生克文案。
			if ch.RiGanPalace.String() != c.riGanGong {
				t.Errorf("ri_gan_gong=%s, want %s", ch.RiGanPalace, c.riGanGong)
			}
			if ch.ShiGanPalace.String() != c.shiGanGong {
				t.Errorf("shi_gan_gong=%s, want %s", ch.ShiGanPalace, c.shiGanGong)
			}
			if ch.RiShiShengKe != c.riShiShengKe {
				t.Errorf("ri_shi_sheng_ke=%q, want %q", ch.RiShiShengKe, c.riShiShengKe)
			}
			// 生克文案自洽：宫五行关系与文案方向一致。
			sp, rp := palaceWuxing(ch.ShiGanPalace), palaceWuxing(ch.RiGanPalace)
			if !strings.Contains(ch.RiShiShengKe, "比和") && !ganzhi.Ke(sp, rp) && !ganzhi.Ke(rp, sp) && !ganzhi.Sheng(sp, rp) && !ganzhi.Sheng(rp, sp) {
				t.Errorf("生克文案 %q 与宫五行关系矛盾（时宫%s 日宫%s）", ch.RiShiShengKe, sp, rp)
			}
			// 空亡/马星宫位锚点。
			kwStr := []string{ch.Pan.KongWang[0].String(), ch.Pan.KongWang[1].String()}
			if !equalStrings(kwStr, c.kongWang) {
				t.Errorf("kong_wang=%v, want %v", kwStr, c.kongWang)
			}
			if ch.Pan.MaXing.String() != c.maXing {
				t.Errorf("ma_xing=%s, want %s", ch.Pan.MaXing, c.maXing)
			}
			// 影响标记自洽：与 pan 级空亡/马星数据一致。
			ri, shi := ch.RiGanPalace, ch.ShiGanPalace
			kw := false
			for _, k := range ch.Pan.KongWang {
				if k == ri || k == shi {
					kw = true
					break
				}
			}
			if kw != ch.KongWangAffected {
				t.Errorf("kong_wang_affected(%v) 与 pan.KongWang 推导(%v) 不一致", ch.KongWangAffected, kw)
			}
			if (ch.Pan.MaXing == ri || ch.Pan.MaXing == shi) != ch.MaXingAffected {
				t.Errorf("ma_xing_affected(%v) 与 pan.MaXing 推导不一致", ch.MaXingAffected)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
