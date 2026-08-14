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
