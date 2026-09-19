package bazi

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

//go:embed testdata/external_8char_uniapp_19810826.json
var external8CharUniAppJSON []byte

type external8CharCase struct {
	Input struct {
		CivilTime string  `json:"civil_time"`
		City      string  `json:"city"`
		Longitude float64 `json:"longitude"`
		Gender    string  `json:"gender"`
	} `json:"input"`
	ExpectedTrueSolarTime string `json:"expected_true_solar_time"`
	Oracle                struct {
		Pillars   []string   `json:"pillars"`
		TaiXi     string     `json:"tai_xi"`
		Gods      [][]string `json:"gods"`
		StartTend struct {
			Year  int    `json:"year"`
			Month int    `json:"month"`
			Day   int    `json:"day"`
			Date  string `json:"date"`
		} `json:"start_tend"`
	} `json:"oracle"`
	ExpectedDayun struct {
		StartDate       string `json:"start_date"`
		StartYearAfter  int    `json:"start_year_after"`
		StartMonthAfter int    `json:"start_month_after"`
		StartDayAfter   int    `json:"start_day_after"`
		Direction       string `json:"direction"`
		First           string `json:"first"`
	} `json:"expected_dayun"`
	ExpectedEngine struct {
		PillarLabels []string   `json:"pillar_labels"`
		HiddenLabels [][]string `json:"hidden_labels"`
		TenGodLabels [][]string `json:"ten_god_labels"`
		TrendLabels  []string   `json:"day_master_trend_labels"`
		XunLabels    []string   `json:"xun_labels"`
		XunKong      []string   `json:"xun_kong_labels"`
		SelfSitting  []string   `json:"self_sitting_labels"`
		Void         []bool     `json:"void"`
		Shensha      [][]string `json:"shensha"`
	} `json:"expected_engine"`
}

func mustDecodeExternal8Char(t *testing.T) external8CharCase {
	t.Helper()
	var tc external8CharCase
	if err := json.Unmarshal(external8CharUniAppJSON, &tc); err != nil {
		t.Fatalf("decode external oracle fixture: %v", err)
	}
	return tc
}

func TestExternalOracle_8CharUniApp_TrueSolarChart(t *testing.T) {
	tc := mustDecodeExternal8Char(t)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}
	full := ComputeFullChart(externalCanonicalChart(bz, ganzhi.Male, 1981))
	pillars := [4]fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi}
	if got := ganzhiZhuLabel(full.TaiXi); got != tc.Oracle.TaiXi {
		t.Fatalf("tai_xi = %s, want %s", got, tc.Oracle.TaiXi)
	}

	for i, want := range tc.ExpectedEngine.PillarLabels {
		got := fmt.Sprintf("%s%s", ganzhi.GanName(pillars[i].Gan), ganzhi.ZhiName(pillars[i].Zhi))
		if got != want {
			t.Fatalf("pillar[%d] = %s, want %s", i, got, want)
		}
	}
}

func TestExternalOracle_8CharUniApp_DaYunStart(t *testing.T) {
	tc := mustDecodeExternal8Char(t)
	location := time.FixedZone("UTC+8", 8*int(time.Hour/time.Second))
	birth, err := time.ParseInLocation("2006-01-02T15:04:05", "1981-08-26T01:06:00", location)
	if err != nil {
		t.Fatal(err)
	}
	chart := ComputeChart(tianwen.SolarTime(birth), ganzhi.Male)
	dy := chart.DaYun
	if dy == nil || len(dy.Steps) == 0 {
		t.Fatal("dayun is empty")
	}
	if tc.Oracle.StartTend.Date != tc.ExpectedDayun.StartDate ||
		tc.Oracle.StartTend.Year != tc.ExpectedDayun.StartYearAfter ||
		tc.Oracle.StartTend.Month != tc.ExpectedDayun.StartMonthAfter ||
		tc.Oracle.StartTend.Day != tc.ExpectedDayun.StartDayAfter {
		t.Fatalf("oracle start_tend %+v mismatches engine expectation %+v", tc.Oracle.StartTend, tc.ExpectedDayun)
	}
	if dy.StartDate != tc.ExpectedDayun.StartDate ||
		dy.StartYearAfter != tc.ExpectedDayun.StartYearAfter ||
		dy.StartMonthAfter != tc.ExpectedDayun.StartMonthAfter ||
		dy.StartDayAfter != tc.ExpectedDayun.StartDayAfter ||
		string(dy.Direction) != tc.ExpectedDayun.Direction {
		t.Fatalf("dayun start = %s %d年%d月%d日 %s, want %s %d年%d月%d日 %s",
			dy.StartDate, dy.StartYearAfter, dy.StartMonthAfter, dy.StartDayAfter, dy.Direction,
			tc.ExpectedDayun.StartDate, tc.ExpectedDayun.StartYearAfter,
			tc.ExpectedDayun.StartMonthAfter, tc.ExpectedDayun.StartDayAfter,
			tc.ExpectedDayun.Direction,
		)
	}
	if first := ganzhiZhuLabel(ganzhi.Zhu{Gan: dy.Steps[0].Gan, Zhi: dy.Steps[0].Zhi}); first != tc.ExpectedDayun.First {
		t.Fatalf("first dayun = %s, want %s", first, tc.ExpectedDayun.First)
	}
}

func TestExternalOracle_8CharUniApp_FullDetailsAndShensha(t *testing.T) {
	tc := mustDecodeExternal8Char(t)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}
	full := ComputeFullChart(externalCanonicalChart(bz, ganzhi.Male, 1981))
	pillars := [4]fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi}

	for i, pillar := range pillars {
		if got := hiddenLabels(pillar.CangGan); !equalStrings(got, tc.ExpectedEngine.HiddenLabels[i]) {
			t.Errorf("pillar[%d] hidden = %v, want %v", i, got, tc.ExpectedEngine.HiddenLabels[i])
		}
		if got := hiddenTenGodLabels(pillar.ShiShens); !equalStrings(got, tc.ExpectedEngine.TenGodLabels[i]) {
			t.Errorf("pillar[%d] ten gods = %v, want %v", i, got, tc.ExpectedEngine.TenGodLabels[i])
		}
		if got := dayMasterTrendLabel(full, pillar.Zhi); got != tc.ExpectedEngine.TrendLabels[i] {
			t.Errorf("pillar[%d] day-master trend = %s, want %s", i, got, tc.ExpectedEngine.TrendLabels[i])
		}
		if pillar.DayMasterTrend != tc.ExpectedEngine.TrendLabels[i] {
			t.Errorf("pillar[%d] day_master_trend = %s, want %s", i, pillar.DayMasterTrend, tc.ExpectedEngine.TrendLabels[i])
		}
		if pillar.Xun != tc.ExpectedEngine.XunLabels[i] || pillar.XunKong != tc.ExpectedEngine.XunKong[i] {
			t.Errorf("pillar[%d] xun = %q/%q, want %q/%q", i, pillar.Xun, pillar.XunKong, tc.ExpectedEngine.XunLabels[i], tc.ExpectedEngine.XunKong[i])
		}
		if got := selfSittingLabel(pillar); got != tc.ExpectedEngine.SelfSitting[i] {
			t.Errorf("pillar[%d] self sitting = %s, want %s", i, got, tc.ExpectedEngine.SelfSitting[i])
		}
		if pillar.IsVoid != tc.ExpectedEngine.Void[i] {
			t.Errorf("pillar[%d] void = %v, want %v", i, pillar.IsVoid, tc.ExpectedEngine.Void[i])
		}
	}

	var engineShensha []string
	for _, pillar := range pillars {
		engineShensha = append(engineShensha, shenshaNames(pillar.ShenSha)...)
	}
	var oracleShensha []string
	for _, gods := range tc.Oracle.Gods {
		oracleShensha = append(oracleShensha, gods...)
	}
	if !includesOracleShensha(engineShensha, oracleShensha) {
		t.Fatalf("engine shensha = %v, must include oracle names %v", engineShensha, oracleShensha)
	}
}

// includesOracleShensha treats 空亡 as the pillar's existing is_void flag and
// normalizes naming differences retained by our structured model.
func includesOracleShensha(got, oracle []string) bool {
	gotSet := make(map[string]bool, len(got))
	for _, name := range got {
		gotSet[name] = true
	}
	for _, oracleName := range oracle {
		switch oracleName {
		case "空亡":
			continue
		case "文昌贵人":
			oracleName = "文昌"
		case "勾绞煞":
			oracleName = "绞神"
		}
		if !gotSet[oracleName] {
			return false
		}
	}
	return true
}

func externalCanonicalChart(bz ganzhi.Bazi, gender ganzhi.Gender, birthYear int) Chart {
	pillars := bz.Slice()
	zhuInfos := [4]zhuInfo{}
	for i, p := range pillars {
		zhuInfos[i] = zhuInfo{Zhu: p}
	}
	return Chart{
		Nian: zhuInfos[0], Yue: zhuInfos[1], Ri: zhuInfos[2], Shi: zhuInfos[3],
		Gender: gender, BirthYear: birthYear,
	}
}

func hiddenLabels(h cangGanOut) []string {
	out := []string{ganzhi.GanName(h.Main)}
	if h.Mid != nil {
		out = append(out, ganzhi.GanName(*h.Mid))
	}
	if h.Minor != nil {
		out = append(out, ganzhi.GanName(*h.Minor))
	}
	return out
}

func hiddenTenGodLabels(entries []shiShenEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Source != sourceGan {
			out = append(out, entry.ShiShen.String())
		}
	}
	return out
}

func dayMasterTrendLabel(full FullChart, zhi ganzhi.Zhi) string {
	for _, stage := range full.ChangSheng {
		if stage.Index == zhi {
			return stage.Name
		}
	}
	return ""
}

func selfSittingLabel(pillar fullZhuInfo) string { return pillar.SelfSitting }

func ganzhiZhuLabel(pillar ganzhi.Zhu) string {
	return ganzhi.GanName(pillar.Gan) + ganzhi.ZhiName(pillar.Zhi)
}

func shenshaNames(entries []shenShaEntry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name)
	}
	return out
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
