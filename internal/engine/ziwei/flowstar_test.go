package ziwei

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type flowGold struct {
	Lunar     string         `json:"lunar"`
	Ti        int            `json:"ti"`
	Gender    string         `json:"gender"`
	FlowStars map[string]int `json:"flowStars"`
	FlowLM    int            `json:"flowLM"`
	FlowLD    int            `json:"flowLD"` // starName → zhiIdx
}

func TestFlowStarsAgainstIz(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "flow_golden.json"))
	if err != nil { t.Fatalf("read: %v", err) }
	var cases []flowGold
	if err := json.Unmarshal(data, &cases); err != nil { t.Fatalf("parse: %v", err) }

	flowYear := 2026

	var pass, total int
	type failT struct{ lunar, star string; got, want int }
	var fails []failT

	for _, tc := range cases {
		if tc.FlowLM == 0 { continue }

		// 流月天干地支
		liuYearGan := yearGan(flowYear)
		monthGan := Gan(((int(yinGan(liuYearGan)) - 1 + tc.FlowLM - 1) % 10 + 10) % 10 + 1)
		monthZhi := Zhi((tc.FlowLM + 1) % 12 + 1) // 正月寅起，不依赖命宫

		// 流日天干地支
		dayGan := riGan(flowYear, tc.FlowLM, tc.FlowLD)
		dayZhi := riZhi(flowYear, tc.FlowLM, tc.FlowLD)

		// 流时天干地支(iztro默认用流日时辰=子时)
		hourZhi := ganzhi.Zhi(1) // 子时
		shiGan := shiGanCalc(dayGan, hourZhi)

		// 计算Liki expected star positions
		likiStars := starZhiIdxMap(monthGan, monthZhi, dayGan, dayZhi, shiGan, hourZhi)

		// 对比iztro golden
		for sName, goldenZhiIdx := range tc.FlowStars {
			total++
			engineZhiIdx, ok := likiStars[sName]
			if !ok { t.Errorf("%s %s: missing in Liki", tc.Lunar, sName); continue }
			if engineZhiIdx != goldenZhiIdx {
				fails = append(fails, failT{tc.Lunar, sName, engineZhiIdx, goldenZhiIdx})
			} else { pass++ }
		}
	}

	if len(fails) > 0 {
		for i, f := range fails {
			if i >= 5 { break }
			t.Errorf("%s %s: got zhiM1=%d want %d", f.lunar, f.star, f.got, f.want)
		}
	}
	pct := float64(pass) / float64(total) * 100
	fmt.Printf("流月/日/时星验证: %d/%d (%.1f%%)\n", pass, total, pct)
	if len(fails) > 0 { fmt.Printf("失败: %d\n", len(fails)) }
}

func starZhiIdxMap(mg Gan, mz Zhi, dg Gan, dz Zhi, sg Gan, sz Zhi) map[string]int {
	r := make(map[string]int)
	mchg, mqu := liuChangQuByGan(mg)
	dchg, dqu := liuChangQuByGan(dg)
	schg, squ := liuChangQuByGan(sg)
	// monthly
	for _, s := range []string{"月禄","月羊","月陀","月魁","月钺","月马","月鸾","月喜","月昌","月曲"} {
		switch {
		case s == "月禄": r[s] = luCunPos(mg)
		case s == "月羊": r[s] = qingYangPos(mg)
		case s == "月陀": r[s] = tuoLuoPos(mg)
		case s == "月魁": r[s] = tianKuiPos(mg)
		case s == "月钺": r[s] = tianYuePos(mg)
		case s == "月马": r[s] = tianMaPos(mz)
		case s == "月鸾": r[s] = hongLuanPos(mz)
		case s == "月喜": r[s] = (hongLuanPos(mz) + 6) % 12
		case s == "月昌": r[s] = mchg
		case s == "月曲": r[s] = mqu
		}
	}
	// daily
	for _, s := range []string{"日禄","日羊","日陀","日魁","日钺","日马","日鸾","日喜","日昌","日曲"} {
		switch {
		case s == "日禄": r[s] = luCunPos(dg)
		case s == "日羊": r[s] = qingYangPos(dg)
		case s == "日陀": r[s] = tuoLuoPos(dg)
		case s == "日魁": r[s] = tianKuiPos(dg)
		case s == "日钺": r[s] = tianYuePos(dg)
		case s == "日马": r[s] = tianMaPos(dz)
		case s == "日鸾": r[s] = hongLuanPos(dz)
		case s == "日喜": r[s] = (hongLuanPos(dz) + 6) % 12
		case s == "日昌": r[s] = dchg
		case s == "日曲": r[s] = dqu
		}
	}
	// hourly
	for _, s := range []string{"时禄","时羊","时陀","时魁","时钺","时马","时鸾","时喜","时昌","时曲"} {
		switch {
		case s == "时禄": r[s] = luCunPos(sg)
		case s == "时羊": r[s] = qingYangPos(sg)
		case s == "时陀": r[s] = tuoLuoPos(sg)
		case s == "时魁": r[s] = tianKuiPos(sg)
		case s == "时钺": r[s] = tianYuePos(sg)
		case s == "时马": r[s] = tianMaPos(sz)
		case s == "时鸾": r[s] = hongLuanPos(sz)
		case s == "时喜": r[s] = (hongLuanPos(sz) + 6) % 12
		case s == "时昌": r[s] = schg
		case s == "时曲": r[s] = squ
		}
	}
	return r
}

func parseFG(tc flowGold) tianwen.LunarTime {
	var y, m, d int
	fmt.Sscanf(tc.Lunar, "%d-%d-%d", &y, &m, &d)
	sz := tc.Ti + 1; day := d
	if tc.Ti == 12 { sz = 1; day++ }
	return tianwen.LunarTime{Year: y, Month: m, Day: day, Shichen: ganzhi.Zhi(sz)}
}
