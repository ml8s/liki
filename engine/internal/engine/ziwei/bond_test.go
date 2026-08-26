package ziwei

import (
	"fmt"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type bondCase struct {
	a, b   string // lunar dates
	at, bt int    // birth time indices
	ag, bg string // genders
	desc   string
}

func TestBond(t *testing.T) {
	cases := []bondCase{
		// 1-4: 自合
		{"2000-7-17", "2000-7-17", 2, 2, "女", "女", "自合·午命"},
		{"1984-1-1", "1984-1-1", 5, 5, "男", "男", "自合·申命"},
		{"1974-4-7", "1974-4-7", 8, 8, "男", "男", "自合·未命"},
		{"1999-1-10", "1999-1-10", 1, 1, "男", "男", "自合·寅命"},
		// 5-6: 异性
		{"2000-7-17", "1983-11-29", 2, 5, "女", "男", "异性·午申"},
		{"1981-4-23", "1993-2-26", 1, 0, "女", "男", "异性·戌辰"},
		// 7-8: 同性
		{"1995-3-15", "2004-2-1", 0, 2, "女", "女", "同性·火土"},
		{"1986-9-7", "2006-12-18", 12, 6, "男", "男", "同性·木火"},
		// 9-11: 五行相生
		{"2005-2-6", "1993-2-26", 4, 0, "女", "男", "水生木"},
		{"1974-4-7", "1995-3-15", 8, 0, "男", "女", "金生水"},
		{"1984-1-1", "1993-2-26", 5, 0, "男", "男", "土生金"},
		// 12-14: 五行相克
		{"1995-3-15", "1986-9-7", 0, 12, "女", "男", "火克金"},
		{"2006-12-18", "2005-2-6", 6, 4, "男", "女", "火克金"},
		{"1993-2-26", "2005-2-6", 0, 4, "男", "女", "木克土"},
		// 15-16: 五行比和
		{"1995-3-15", "2001-7-22", 0, 0, "女", "男", "火比和"},
		{"1986-9-7", "1993-2-26", 12, 0, "男", "男", "木比和"},
		// 17-18: 夫妻宫
		{"1981-4-23", "1987-4-15", 1, 5, "女", "男", "夫妻·天同vs紫微"},
		{"1994-6-25", "1997-5-8", 11, 2, "男", "女", "夫妻·空vs七杀"},
		// 19-20: 子女宫
		{"1982-11-30", "2003-9-5", 10, 2, "女", "男", "子女·空vs紫破"},
		{"1989-2-14", "1996-7-30", 1, 5, "女", "女", "子女·同阴vs"},
		// 21-22: 海外
		{"1974-4-28", "2000-8-16", 16, 4, "男", "女", "海外·纽约"},
		{"1982-11-30", "2003-9-5", 10, 2, "女", "男", "海外·无经"},
		// 23-24: 四化
		{"2000-7-17", "1995-8-1", 2, 6, "女", "男", "四化·禄忌"},
		{"1984-12-1", "1999-1-10", 3, 1, "男", "男", "四化·同干"},
		// 25: 晚子时
		{"1986-9-7", "2001-7-22", 12, 0, "男", "男", "晚子时"},
		// 26: 闰月
		{"1968-6-15", "1995-3-15", 5, 0, "女", "女", "闰月"},
		// 27: 早子时
		{"1993-2-26", "1995-3-15", 0, 0, "男", "女", "早子时"},
		// 28: 同年不同命
		{"2000-7-17", "2000-12-5", 2, 7, "女", "女", "同庚不同命"},
	}

	var pass, fail int
	for i, tc := range cases {
		n := tc.desc
		if len(n) > 14 {
			n = n[:14]
		}
		t.Run(fmt.Sprintf("%02d_%s", i+1, n), func(t *testing.T) {
			ca := chartFrom(t, tc.a, tc.at, tc.ag)
			cb := chartFrom(t, tc.b, tc.bt, tc.bg)
			bd := ComputeBond(ca, cb)

			if bd.GongCross.AIntoB == "" || bd.GongCross.BIntoA == "" {
				t.Error("palace_cross为空")
			}
			if bd.WuXingShengKe == "" {
				t.Error("element_fit为空")
			}
			if bd.FuQiGong == nil {
				t.Error("fu_qi_gong为nil")
			}
			if bd.ZiNvGong == nil {
				t.Error("zi_nv_gong为nil")
			}
			if tc.desc[:3] == "自合" {
				if bd.GongCross.AIntoB != ca.GongWei[0].Name {
					t.Errorf("自合A入B: %s", bd.GongCross.AIntoB)
				}
				if bd.GongCross.BIntoA != ca.GongWei[0].Name {
					t.Errorf("自合B入A: %s", bd.GongCross.BIntoA)
				}
				if !strSliceEq(bd.FuQiGong.AZhuXing, bd.FuQiGong.BZhuXing) {
					t.Errorf("自合夫妻星不同: %v vs %v", bd.FuQiGong.AZhuXing, bd.FuQiGong.BZhuXing)
				}
			}
			pass++
		})
	}
	fmt.Printf("Bond测试: %d/%d pass\n", pass, pass+fail)
}

func chartFrom(t *testing.T, lunar string, ti int, gender string) Chart {
	t.Helper()
	var y, m, d int
	_, _ = fmt.Sscanf(lunar, "%d-%d-%d", &y, &m, &d)
	sz := ti + 1
	if ti == 12 {
		sz = 1
		d++
	}
	g := ganzhi.Female
	if gender == "男" {
		g = ganzhi.Male
	}
	return ComputeChart(tianwen.LunarTime{Year: y, Month: m, Day: d, Shichen: ganzhi.Zhi(sz)}, g)
}

func strSliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int)
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		m[s]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
