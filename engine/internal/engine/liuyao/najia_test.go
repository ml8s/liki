package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 纳甲数据驱动锚点（京房纳甲，通行六爻教材表）：
// 天干：乾内甲外壬、坤内乙外癸，震庚巽辛坎戊离己艮丙兑丁（内外同）；
// 地支：八宫各六爻（乾子寅辰午申戌、兑巳卯丑亥酉未…）。
func TestNaJia_Gan_AllPalaces(t *testing.T) {
	want := map[string][2]string{
		"乾": {"甲", "壬"}, "兑": {"丁", "丁"}, "离": {"己", "己"},
		"震": {"庚", "庚"}, "巽": {"辛", "辛"}, "坎": {"戊", "戊"},
		"艮": {"丙", "丙"}, "坤": {"乙", "癸"},
	}
	for i, g := range naGanTable {
		name := palaceNames[i]
		w := want[name]
		if ganName(g[0]) != w[0] || ganName(g[1]) != w[1] {
			t.Errorf("naGanTable[%s] = [%s %s], want [%s %s]（京房纳甲）",
				name, ganName(g[0]), ganName(g[1]), w[0], w[1])
		}
	}
}

func TestNaJia_Zhi_AllPalaces(t *testing.T) {
	want := map[string][6]string{
		"乾": {"子", "寅", "辰", "午", "申", "戌"},
		"兑": {"巳", "卯", "丑", "亥", "酉", "未"},
		"离": {"卯", "丑", "亥", "酉", "未", "巳"},
		"震": {"子", "寅", "辰", "午", "申", "戌"},
		"巽": {"丑", "亥", "酉", "未", "巳", "卯"},
		"坎": {"寅", "辰", "午", "申", "戌", "子"},
		"艮": {"辰", "午", "申", "戌", "子", "寅"},
		"坤": {"未", "巳", "卯", "丑", "亥", "酉"},
	}
	for i, zs := range naZhiTable {
		name := palaceNames[i]
		w := want[name]
		for j := 0; j < 6; j++ {
			if zhiName(zs[j]) != w[j] {
				t.Errorf("naZhiTable[%s][%d] = %s, want %s（京房纳甲）",
					name, j+1, zhiName(zs[j]), w[j])
			}
		}
	}
}

// 装卦集成：乾为天六爻纳甲 = 甲子 甲寅 甲辰 壬午 壬申 壬戌（内甲外壬）。
func TestNaJia_ZhuangGua_QianWeiTian(t *testing.T) {
	lines := zhuangGua(guaIndex(0), ganzhi.GanJia, false, ganzhi.WxJin) // 乾为天
	wantGan := [6]string{"甲", "甲", "甲", "壬", "壬", "壬"}
	wantZhi := [6]string{"子", "寅", "辰", "午", "申", "戌"}
	for i := 0; i < 6; i++ {
		if ganName(lines[i].Gan) != wantGan[i] || zhiName(lines[i].Zhi) != wantZhi[i] {
			t.Errorf("乾为天 line%d = %s%s, want %s%s", i+1,
				ganName(lines[i].Gan), zhiName(lines[i].Zhi), wantGan[i], wantZhi[i])
		}
	}
}

func ganName(g ganzhi.Gan) string { return ganzhi.GanName(g) }
func zhiName(z ganzhi.Zhi) string { return ganzhi.ZhiName(z) }

// 装卦集成——经卦混搭卦（上下卦不同经卦）：纳甲按卦体上下经卦分别取，
// 而非按本宫。
func TestNaJia_ZhuangGua_MixedTrigrams(t *testing.T) {
	cases := []struct {
		name    string
		wantGan [6]string
		wantZhi [6]string
	}{
		// 上乾下艮：内艮丙辰丙午丙申、外乾壬午壬申壬戌
		{"天山遁", [6]string{"丙", "丙", "丙", "壬", "壬", "壬"}, [6]string{"辰", "午", "申", "午", "申", "戌"}},
		// 上乾下震：内震庚子庚寅庚辰、外乾壬午壬申壬戌
		{"天雷无妄", [6]string{"庚", "庚", "庚", "壬", "壬", "壬"}, [6]string{"子", "寅", "辰", "午", "申", "戌"}},
		// 上乾下坤：内坤乙未乙巳乙卯、外乾壬午壬申壬戌
		{"天地否", [6]string{"乙", "乙", "乙", "壬", "壬", "壬"}, [6]string{"未", "巳", "卯", "午", "申", "戌"}},
		// 上巽下坤：内坤乙未乙巳乙卯、外巽辛未辛巳辛卯
		{"风地观", [6]string{"乙", "乙", "乙", "辛", "辛", "辛"}, [6]string{"未", "巳", "卯", "未", "巳", "卯"}},
		// 上兑下离：内离己卯己丑己亥、外兑丁亥丁酉丁未
		{"泽火革", [6]string{"己", "己", "己", "丁", "丁", "丁"}, [6]string{"卯", "丑", "亥", "亥", "酉", "未"}},
		// 上坎下震：内震庚子庚寅庚辰、外坎戊申戊戌戊子
		{"水雷屯", [6]string{"庚", "庚", "庚", "戊", "戊", "戊"}, [6]string{"子", "寅", "辰", "申", "戌", "子"}},
		// 上艮下坤：内坤乙未乙巳乙卯、外艮丙戌丙子丙寅
		{"山地剥", [6]string{"乙", "乙", "乙", "丙", "丙", "丙"}, [6]string{"未", "巳", "卯", "戌", "子", "寅"}},
		// 上离下坤：内坤乙未乙巳乙卯、外离己酉己未己巳
		{"火地晋", [6]string{"乙", "乙", "乙", "己", "己", "己"}, [6]string{"未", "巳", "卯", "酉", "未", "巳"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, ok := guaIndexByName(t, c.name)
			if !ok {
				return
			}
			lines := zhuangGua(g, ganzhi.GanJia, false, ganzhi.WxJin)
			for i := 0; i < 6; i++ {
				gan, zhi := ganName(lines[i].Gan), zhiName(lines[i].Zhi)
				if gan != c.wantGan[i] || zhi != c.wantZhi[i] {
					t.Errorf("%s line%d = %s%s, want %s%s（经卦纳甲）",
						c.name, i+1, gan, zhi, c.wantGan[i], c.wantZhi[i])
				}
			}
		})
	}
}

// guaIndexByName 按卦名查 guaTable 索引（测试用）。
func guaIndexByName(t *testing.T, name string) (guaIndex, bool) {
	t.Helper()
	for i, g := range guaTable {
		if g.Name == name {
			return guaIndex(i), true
		}
	}
	t.Errorf("卦 %q 不在 guaTable", name)
	return 0, false
}

// 全量 64 卦纳甲一致性：每卦按上下经卦规则计算期望干支，与 zhuangGua 输出比对。
// 表数据正确性由 TestNaJia_Gan/Zhi_AllPalaces（京房通说锚点）保证——两者合起来
// 覆盖「表对 + 逻辑对」，防止混搭卦再漏网。
func TestNaJia_All64Hexagrams(t *testing.T) {
	for g := guaIndex(0); g < 64; g++ {
		name := guaTable[g].Name
		upperTri, lowerTri := guaTrigrams(g)
		lowerPi, upperPi := trigramPalaceIdx[lowerTri], trigramPalaceIdx[upperTri]
		lines := zhuangGua(g, ganzhi.GanJia, false, ganzhi.WxJin)
		for i := 0; i < 6; i++ {
			wantGan := ganName(naGanTable[lowerPi][0])
			wantZhi := zhiName(naZhiTable[lowerPi][i])
			if i >= 3 {
				wantGan = ganName(naGanTable[upperPi][1])
				wantZhi = zhiName(naZhiTable[upperPi][i])
			}
			gotGan, gotZhi := ganName(lines[i].Gan), zhiName(lines[i].Zhi)
			if gotGan != wantGan || gotZhi != wantZhi {
				t.Errorf("%s line%d = %s%s, want %s%s（经卦纳甲）",
					name, i+1, gotGan, gotZhi, wantGan, wantZhi)
			}
		}
	}
}

// 全量 64 卦世应位置：guaTable 每宫 8 卦按京房八宫卦序排列
// （宫位序 0=本宫、1..5=一..五世、6=游魂、7=归魂），世位 = {6,1,2,3,4,5,4,3}[宫位序]，
// 应位 = 世位对冲（+3 爻）。
func TestShiYing_All64Hexagrams(t *testing.T) {
	shiBySeq := [8]int{6, 1, 2, 3, 4, 5, 4, 3}
	for g := guaIndex(0); g < 64; g++ {
		meta := guaTable[g]
		seq := int(g) % 8 // 宫位序（guaTable 每 8 卦一宫，按京房卦序）
		if meta.ShiPos != shiBySeq[seq] {
			t.Errorf("%s shi_pos = %d, want %d（京房八宫卦序宫位 %d）", meta.Name, meta.ShiPos, shiBySeq[seq], seq)
		}
		lines := zhuangGua(g, ganzhi.GanJia, false, ganzhi.WxJin)
		shi, ying := 0, 0
		for i := 0; i < 6; i++ {
			if lines[i].ShiYing == "世" {
				shi = i + 1
			}
			if lines[i].ShiYing == "应" {
				ying = i + 1
			}
		}
		if shi != shiBySeq[seq] {
			t.Errorf("%s 装卦世位 = %d, want %d", meta.Name, shi, shiBySeq[seq])
		}
		if ying != (shiBySeq[seq]+2)%6+1 {
			t.Errorf("%s 应位 = %d, want %d（世位对冲）", meta.Name, ying, (shiBySeq[seq]+2)%6+1)
		}
	}
}
