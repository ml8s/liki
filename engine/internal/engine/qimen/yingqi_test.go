package qimen

import (
	"encoding/json"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestYingQiTableDefinesDateWindow(t *testing.T) {
	var table struct {
		HorizonDays int `json:"horizon_days"`
		DateMatches map[string][]struct {
			Match string `json:"match"`
		} `json:"date_matches"`
	}
	if err := decodeJSONForTest(t, yingqiJSON, &table); err != nil {
		t.Fatal(err)
	}
	if table.HorizonDays != 60 {
		t.Fatalf("horizon_days = %d, want 60", table.HorizonDays)
	}
	if !reflectStrings(matchNames(table.DateMatches["ma_xing"]), []string{"opposite"}) {
		t.Fatalf("ma_xing date matches = %v", table.DateMatches["ma_xing"])
	}
	if !reflectStrings(matchNames(table.DateMatches["kong_wang"]), []string{"same", "opposite"}) {
		t.Fatalf("kong_wang date matches = %v", table.DateMatches["kong_wang"])
	}
	for _, ruleType := range []string{"duty_star", "duty_door"} {
		if !reflectStrings(matchNames(table.DateMatches[ruleType]), []string{"same", "opposite"}) {
			t.Fatalf("%s date matches = %v", ruleType, table.DateMatches[ruleType])
		}
	}
}

func TestYingQiComputesCivilDateWindow(t *testing.T) {
	you, err := ganzhi.ParseZhi("酉")
	if err != nil {
		t.Fatal(err)
	}
	mao, err := ganzhi.ParseZhi("卯")
	if err != nil {
		t.Fatal(err)
	}
	riGan, err := ganzhi.ParseGan("乙")
	if err != nil {
		t.Fatal(err)
	}
	chart := Chart{Pan: pan{
		RiGan:  riGan,
		RiZhi:  you,
		MaXing: BranchPalace{Branch: you, Gong: GongDui},
		KongWang: [2]BranchPalace{
			{Branch: you, Gong: GongDui},
			{Branch: mao, Gong: GongZhen},
		},
	}}
	chartTime := time.Date(2026, 9, 7, 23, 30, 0, 0, time.FixedZone("CST", 8*3600))
	result := computeYingQi(chart, chartTime, nil)
	if got := yingQiAnchorDate(chart, chartTime).Format("2006-01-02"); got != "2026-09-08" {
		t.Fatalf("late-zi date anchor = %s, want 2026-09-08", got)
	}
	if len(result.Candidates) != 3 {
		t.Fatalf("candidates = %d, want horse and two void branches", len(result.Candidates))
	}

	maWant := []string{
		"2026-09-14", "2026-09-26", "2026-10-08",
		"2026-10-20", "2026-11-01",
	}
	assertYingQiDates(t, result.Candidates[0], maWant, map[string]string{
		"2026-09-14": "卯", "2026-09-26": "卯", "2026-10-08": "卯",
		"2026-10-20": "卯", "2026-11-01": "卯",
	})
	for _, item := range result.Candidates[0].Dates {
		if item.Match != "opposite" {
			t.Fatalf("horse date match = %q, want opposite", item.Match)
		}
	}

	kongWant := []string{
		"2026-09-08", "2026-09-14", "2026-09-20", "2026-09-26",
		"2026-10-02", "2026-10-08", "2026-10-14", "2026-10-20",
		"2026-10-26", "2026-11-01",
	}
	kongBranches := map[string]string{
		"2026-09-08": "酉", "2026-09-14": "卯", "2026-09-20": "酉", "2026-09-26": "卯",
		"2026-10-02": "酉", "2026-10-08": "卯", "2026-10-14": "酉", "2026-10-20": "卯",
		"2026-10-26": "酉", "2026-11-01": "卯",
	}
	for _, candidate := range result.Candidates[1:] {
		if candidate.Branch != "酉" {
			continue
		}
		assertYingQiDates(t, candidate, kongWant, kongBranches)
		matches := map[string]int{}
		for _, item := range candidate.Dates {
			matches[item.Match]++
		}
		if matches["same"] != 5 || matches["opposite"] != 5 {
			t.Fatalf("void matches = %v, want five same and five opposite", matches)
		}
		return
	}
	t.Fatal("酉 void candidate missing")
}

func TestYingQiIncludesDutyStarAndDoorCandidates(t *testing.T) {
	chart := Chart{
		DutyStarPalace: GongDui,
		DutyDoorPalace: GongZhen,
	}
	result := computeYingQi(chart, time.Date(2026, 9, 9, 12, 0, 0, 0, time.FixedZone("CST", 8)), baseYingQiFocuses(chart))
	types := map[string]bool{}
	for _, candidate := range result.Candidates {
		types[candidate.Type] = true
		if candidate.Type != "duty_star" && candidate.Type != "duty_door" {
			continue
		}
		if candidate.Branch == "" || candidate.Gong == 0 || len(candidate.RelatedTo) == 0 {
			t.Fatalf("duty candidate = %+v", candidate)
		}
	}
	if !types["duty_star"] || !types["duty_door"] {
		t.Fatalf("candidate types = %+v, want duty_star and duty_door", types)
	}
}

func assertYingQiDates(
	t *testing.T, candidate YingQiCandidate, want []string, branches map[string]string,
) {
	t.Helper()
	if len(candidate.Dates) != len(want) {
		t.Fatalf("%s dates = %+v, want %v", candidate.Type, candidate.Dates, want)
	}
	var previous string
	for i, item := range candidate.Dates {
		if item.Date != want[i] {
			t.Fatalf("%s date[%d] = %s, want %s", candidate.Type, i, item.Date, want[i])
		}
		if item.Date < previous {
			t.Fatalf("%s dates are not ascending: %+v", candidate.Type, candidate.Dates)
		}
		previous = item.Date
		if item.Branch != branches[item.Date] || item.Reason == "" {
			t.Fatalf("invalid date fact %+v, want branch %s", item, branches[item.Date])
		}
	}
}

func decodeJSONForTest(t *testing.T, data []byte, target any) error {
	t.Helper()
	return json.Unmarshal(data, target)
}

func reflectStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func matchNames(matches []struct {
	Match string `json:"match"`
}) []string {
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.Match)
	}
	return result
}
