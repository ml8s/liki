package ziwei

import (
	"strings"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type testCaseRef struct {
	Lunar    string                 `json:"lunar"`
	Ti       int                    `json:"ti"`
	Gender   string                 `json:"gender"`
	Ju       string                 `json:"ju"`
	SiHua    map[string]string      `json:"sihua"`
	YSoul    string                 `json:"ysoul"`
	YSbody   string                 `json:"ysbody"`
	YuanGong string                 `json:"yuangong"`
	DaXian   []daxianRef            `json:"daxian"`
	Palaces  map[string]palaceRef   `json:"palaces"`
	Yzhi     string                 `json:"yZhi"`
	Ysihua   map[string]string      `json:"ySihua"`
	YFlow    []flowPalaceRef        `json:"yFlowPalaces,omitempty"`
	MFlow    []flowPalaceRef        `json:"mFlowPalaces,omitempty"`
	DFlow    []flowPalaceRef        `json:"dFlowPalaces,omitempty"`
	HFlow    []flowPalaceRef        `json:"hFlowPalaces,omitempty"`
	Mzhi     string                 `json:"mZhi"`
	Msihua   map[string]string      `json:"mSihua"`
	Mstars   map[string]int         `json:"mStars"`
	Dzhi     string                 `json:"dZhi"`
	Dsihua   map[string]string      `json:"dSihua"`
	Dstars   map[string]int         `json:"dStars"`
	Hzhi     string                 `json:"hZhi"`
	Hsihua   map[string]string      `json:"hSihua"`
	Hstars   map[string]int         `json:"hStars"`
}
type flowPalaceRef struct {
	Zhi    string   `json:"zhi"`
	Name   string   `json:"name"`
	Stars  []string `json:"stars"`
	IsMing bool     `json:"is_ming"`
}

type daxianRef struct {
	Palace string `json:"palace"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
}
type palaceRef struct {
	Zhi         string   `json:"zhi"`
	YuanGong    bool     `json:"yuangong"`
	Major       []string `json:"major"`
	MajorBright []string `json:"majorBright"`
	Minor       []string `json:"minor"`
	Cs          string   `json:"cs"`
	Bs          string   `json:"bs"`
	Jq          string   `json:"jq"`
	Sq          string   `json:"sq"`
	Ages        []int    `json:"ages"`
	Adj         []string `json:"adj"`
}

func loadCases(t *testing.T) []testCaseRef {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "complete_test.json"))
	if err != nil { t.Fatalf("read: %v", err) }
	var cases []testCaseRef
	if err := json.Unmarshal(data, &cases); err != nil { t.Fatalf("parse: %v", err) }
	return cases
}

func TestComplete(t *testing.T) {
	cases := loadCases(t)
	var pass, fail int
	for _, tc := range cases {
		t.Run(tc.Lunar+"_"+tc.Gender, func(t *testing.T) {
			lt := parseLT(tc)
			gender := ganzhi.Female
			if tc.Gender == "男" { gender = ganzhi.Male }
			chart := ComputeChart(lt, gender)
			fc := ComputeFullChart(chart, 0, 0)

			// 五行局
			if fc.JuShuName != tc.Ju { t.Error("局:", fc.JuShuName, "want", tc.Ju) }

			// 命主身主
			if fc.MingZhu != tc.YSoul { t.Error("命主:", fc.MingZhu, "want", tc.YSoul) }
			if fc.ShenZhu != tc.YSbody { t.Error("身主:", fc.ShenZhu, "want", tc.YSbody) }

			// 四化
			for sidStr, stype := range tc.SiHua {
				var sid int; fmt.Sscanf(sidStr, "%d", &sid)
				got, ok := fc.SiHua[starIndex(sid)]
				if !ok { t.Errorf("四化%d应%s但无", sid, stype); continue }
				if string(got) != stype { t.Errorf("四化%d: got %s want %s", sid, string(got), stype) }
			}

			// 大限(通过ComputeDaXian获取)
			dxResult := ComputeDaXian(fc)
			for _, dx := range tc.DaXian {
				var found bool
				for _, s := range dxResult {
					if s.Name == dx.Palace {
						found = true
						if s.QiSui != dx.Start || s.ZhiSui != dx.End {
							t.Errorf("大限%s: got %d-%d want %d-%d", dx.Palace, s.QiSui, s.ZhiSui, dx.Start, dx.End)
						}
						break
					}
				}
				if !found && dx.Palace != "" {
					t.Errorf("大限%s: not found", dx.Palace)
				}
			}

			// 流年盘（命主特定）
			ln := ComputeLiuNian(fc, 2026)
			assertFlowPalaces(t, "流年", ln.Palaces, tc.YFlow, &pass, &fail)
			// 流月/日/时盘（变量复用后续流月/日/时断言）
			ly2 := ComputeLiuYue(fc, 2026, 6)
			assertFlowPalaces(t, "流月", ly2.Palaces, tc.MFlow, &pass, &fail)
			lr2 := ComputeLiuRi(fc, 2026, 6, 4)
			assertFlowPalaces(t, "流日", lr2.Palaces, tc.DFlow, &pass, &fail)
			ls2 := ComputeLiuShi(fc, 2026, 6, 4, ganzhi.Zhi(1))
			assertFlowPalaces(t, "流时", ls2.Palaces, tc.HFlow, &pass, &fail)
			// 流年四化+zhi
			lnZhi := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}[ln.Zhi]
			if lnZhi != tc.Yzhi && tc.Yzhi != "" { t.Errorf("流年zhi: got %s want %s", lnZhi, tc.Yzhi) }
			for sidStr, stype := range tc.Ysihua {
				if v, ok := ln.SiHua[starIndex(atoi(sidStr))]; !ok || string(v) != stype {
					t.Errorf("流年四化%s: got %s want %s", sidStr, string(v), stype)
				}
			}
			// 流月四化+zhi+星
			ly := ComputeLiuYue(fc, 2026, 6)
			lyZhi := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}[ly.Zhi]
			if lyZhi != tc.Mzhi && tc.Mzhi != "" { t.Errorf("流月zhi: got %s want %s", lyZhi, tc.Mzhi) }
			for sidStr, stype := range tc.Msihua {
				if v, ok := ly.SiHua[starIndex(atoi(sidStr))]; !ok || string(v) != stype {
					t.Errorf("流月四化%s: got %s want %s", sidStr, string(v), stype)
				}
			}
			for sName, zhiM1 := range tc.Mstars {
				if v, ok := ly.Stars[sName]; !ok || int(v)-1 != zhiM1 {
					t.Errorf("流月星%s: got %d want %d", sName, v, zhiM1)
				}
			}
			// 流日四化+zhi+星
			lr := ComputeLiuRi(fc, 2026, 6, 4)
			lrZhi := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}[lr.Zhi]
			if lrZhi != tc.Dzhi && tc.Dzhi != "" { t.Errorf("流日zhi: got %s want %s", lrZhi, tc.Dzhi) }
			for sidStr, stype := range tc.Dsihua {
				if v, ok := lr.SiHua[starIndex(atoi(sidStr))]; !ok || string(v) != stype {
					t.Errorf("流日四化%s: got %s want %s", sidStr, string(v), stype)
				}
			}
			for sName, zhiM1 := range tc.Dstars {
				if v, ok := lr.Stars[sName]; !ok || int(v)-1 != zhiM1 {
					t.Errorf("流日星%s: got %d want %d", sName, v, zhiM1)
				}
			}
			// 流时四化+zhi+星
			ls := ComputeLiuShi(fc, 2026, 6, 4, ganzhi.Zhi(1))
			lsZhi := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}[ls.Zhi]
			if lsZhi != tc.Hzhi && tc.Hzhi != "" { t.Errorf("流时zhi: got %s want %s", lsZhi, tc.Hzhi) }
			for sidStr, stype := range tc.Hsihua {
				if v, ok := ls.SiHua[starIndex(atoi(sidStr))]; !ok || string(v) != stype {
					t.Errorf("流时四化%s: got %s want %s", sidStr, string(v), stype)
				}
			}
			for sName, zhiM1 := range tc.Hstars {
				if v, ok := ls.Stars[sName]; !ok || int(v)-1 != zhiM1 {
					t.Errorf("流时星%s: got %d want %d", sName, v, zhiM1)
				}
			}

			// 每宫
			for _, nm := range palaceLabels {
				ref, ok := tc.Palaces[nm]; if !ok { continue }
				got := findPalace(fc.Palaces, nm); if got == nil { t.Errorf("%s: not found", nm); continue }

				// 主星
				gotM := starsToNames(got.Stars, true)
				if !setEq(gotM, ref.Major) { t.Errorf("[%s]主星: got%v want%v", nm, gotM, ref.Major); fail++ } else { pass++ }
				// 辅星
				gotN := starsToNames(got.Stars, false)
				if !setEq(gotN, ref.Minor) { t.Errorf("[%s]辅星: got%v want%v", nm, gotN, ref.Minor); fail++ } else { pass++ }
				// 长生
				if got.ChangSheng != ref.Cs { t.Errorf("[%s]长生: got%s want%s", nm, got.ChangSheng, ref.Cs); fail++ } else { pass++ }
				// 亮度
				gotB := starBrNames(got.Stars, true)
				if !setEq(gotB, ref.MajorBright) { t.Errorf("[%s]亮度: got%v want%v", nm, gotB, ref.MajorBright); fail++ } else { pass++ }
				// 博士
				if got.BoShi != ref.Bs { t.Errorf("[%s]博士: got%s want%s", nm, got.BoShi, ref.Bs); fail++ } else { pass++ }
				// 将前/岁前
				if got.JiangQian != ref.Jq { t.Errorf("[%s]将前: got%s want%s", nm, got.JiangQian, ref.Jq); fail++ } else { pass++ }
				if got.SuiQian != ref.Sq { t.Errorf("[%s]岁前: got%s want%s", nm, got.SuiQian, ref.Sq); fail++ } else { pass++ }
				// 小限
				ga := got.Ages; if len(ga) > 3 { ga = ga[:3] }
				if !intEq(ga, ref.Ages) { t.Errorf("[%s]小限: got%v want%v", nm, ga, ref.Ages); fail++ } else { pass++ }
				// 杂曜
				if !setEq(got.ZaYao, ref.Adj) { t.Errorf("[%s]杂曜: got%v want%v", nm, got.ZaYao, ref.Adj); fail++ } else { pass++ }
			}
		})
	}
	if pass+fail > 0 {
		fmt.Printf("断言汇总: %d pass / %d fail / %d total\n", pass, fail, pass+fail)
	}
}

func assertFlowPalaces(t *testing.T, label string, got [12]flowPalace, ref []flowPalaceRef, pass, fail *int) {
	if len(ref) != 12 {
		return
	}
	for fi, fp := range got {
		r := ref[fi]
		if fp.Zhi.String() != r.Zhi { t.Errorf("%s盘[%d]支: got %s want %s", label, fi, fp.Zhi.String(), r.Zhi); *fail++; continue }
		if fp.Name != r.Name { t.Errorf("%s盘[%d]名: got %s want %s", label, fi, fp.Name, r.Name); *fail++; continue }
		if !setEq(fp.Stars, r.Stars) { t.Errorf("%s盘[%d]%s流耀: got %v want %v", label, fi, r.Name, fp.Stars, r.Stars); *fail++; continue }
		if fp.IsMing != r.IsMing { t.Errorf("%s盘[%d]IsMing: got %v want %v", label, fi, fp.IsMing, r.IsMing); *fail++; continue }
		*pass++
	}
}

func findPalace(p [12]palace, name string) *palace {
	for i := range p { if p[i].Name == name { return &p[i] } }
	return nil
}
func starBrNames(stars []starInfo, major bool) []string {
	var r []string
	for _, s := range stars {
		if s.IsMajor == major { r = append(r, s.Name+":"+s.Brightness) }
	}
	sortS(r); return r
}
func starsToNames(stars []starInfo, major bool) []string {
	var r []string
	for _, s := range stars { if s.IsMajor == major { r = append(r, s.Name) } }
	sortS(r); return r
}
func sortS(s []string) {
	for i := 0; i < len(s)-1; i++ { for j := i+1; j < len(s); j++ { if s[i] > s[j] { s[i], s[j] = s[j], s[i] } } }
}
func strEq(a, b []string) bool {
	if len(a) != len(b) { return false }
	for i := range a { if a[i] != b[i] { return false } }
	return true
}
func setEq(a, b []string) bool {
	if len(a) != len(b) { return false }
	m := make(map[string]bool, len(a))
	for _, s := range a { m[s] = true }
	for _, s := range b { if !m[s] { return false } }
	return true
}
func intEq(a, b []int) bool {
	if len(a) != len(b) { return false }
	for i := range a { if a[i] != b[i] { return false } }
	return true
}
func atoi(s string) int {
	var n int; fmt.Sscanf(s, "%d", &n); return n
}

func parseLT(tc testCaseRef) tianwen.LunarTime {
	var y, m, d int
	leap := false
	s := tc.Lunar
	// 闰月标记: "Y-M-D闰"
	if strings.HasSuffix(s, "闰") {
		leap = true
		s = strings.TrimSuffix(s, "闰")
	}
	fmt.Sscanf(s, "%d-%d-%d", &y, &m, &d)
	sz := tc.Ti + 1; day := d
	if tc.Ti == 12 { sz = 1; day++ }
	return tianwen.LunarTime{Year: y, Month: m, Day: day, Leap: leap, Shichen: ganzhi.Zhi(sz)}
}

// TestFlowYear validates流年 flow year horoscope components.
