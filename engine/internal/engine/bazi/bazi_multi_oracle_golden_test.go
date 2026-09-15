package bazi

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type baziMultiOracleFixture struct {
	Version int `json:"version"`
	Policy  struct {
		YearBoundary  string `json:"year_boundary"`
		MonthBoundary string `json:"month_boundary"`
		LateZi        string `json:"late_zi"`
		Consensus     string `json:"consensus"`
	} `json:"policy"`
	Scope struct {
		Years          []int `json:"years"`
		TermsPerYear   int   `json:"terms_per_year"`
		AnchorsPerTerm int   `json:"anchors_per_term"`
		CaseCount      int   `json:"case_count"`
	} `json:"scope"`
	Sources []struct {
		ID      string   `json:"id"`
		Version string   `json:"version"`
		Fields  []string `json:"fields"`
	} `json:"sources"`
	Cases []baziMultiOracleCase `json:"cases"`
}

type baziMultiOracleCase struct {
	Name      string `json:"name"`
	Year      int    `json:"year"`
	TermIndex int    `json:"term_index"`
	JieQi     string `json:"jie_qi"`
	Side      string `json:"side"`
	Input     struct {
		Solar          string `json:"solar"`
		TimezoneOffset int    `json:"timezone_offset"`
	} `json:"input"`
	Expected struct {
		Pillars struct {
			Nian string `json:"nian"`
			Yue  string `json:"yue"`
			Ri   string `json:"ri"`
			Shi  string `json:"shi"`
		} `json:"pillars"`
	} `json:"expected"`
	Consensus []string `json:"consensus"`
}

func loadBaziMultiOracleFixture(t *testing.T) baziMultiOracleFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/multi_oracle_golden.json")
	if err != nil {
		t.Fatalf("read multi-oracle golden: %v", err)
	}
	var fixture baziMultiOracleFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse multi-oracle golden: %v", err)
	}
	if fixture.Version != 1 {
		t.Fatalf("fixture version = %d, want 1", fixture.Version)
	}
	if fixture.Policy.Consensus != "all_oracles_must_agree" {
		t.Fatalf("unexpected fixture policy: %+v", fixture.Policy)
	}
	if len(fixture.Sources) != 3 {
		t.Fatalf("source count = %d, want 3", len(fixture.Sources))
	}
	if len(fixture.Cases) != fixture.Scope.CaseCount {
		t.Fatalf("case count = %d, scope = %d", len(fixture.Cases), fixture.Scope.CaseCount)
	}
	if len(fixture.Cases) < 192 {
		t.Fatalf("case count = %d, want at least 192", len(fixture.Cases))
	}
	return fixture
}

func TestBaziMultiOracleGolden_AllTermBoundaries(t *testing.T) {
	fixture := loadBaziMultiOracleFixture(t)
	seen := make(map[string]bool, len(fixture.Cases))
	coverage := make(map[[3]int]bool)

	for i := range fixture.Cases {
		golden := &fixture.Cases[i]
		if seen[golden.Name] {
			t.Fatalf("duplicate case %q", golden.Name)
		}
		seen[golden.Name] = true
		if len(golden.Consensus) != 3 {
			t.Fatalf("case %q consensus count = %d, want 3", golden.Name, len(golden.Consensus))
		}

		t.Run(golden.Name, func(t *testing.T) {
			location := time.FixedZone(
				"UTC+08",
				golden.Input.TimezoneOffset*int(time.Hour/time.Second),
			)
			birth, err := time.ParseInLocation("2006-01-02T15:04:05", golden.Input.Solar, location)
			if err != nil {
				t.Fatalf("parse solar %q: %v", golden.Input.Solar, err)
			}
			// Gender is irrelevant to four pillars; alternate it to catch accidental
			// gender leakage from Da Yun logic into chart construction.
			gender := ganzhi.Male
			if i%2 == 1 {
				gender = ganzhi.Female
			}
			chart := ComputeChart(tianwen.SolarTime(birth), gender)

			got := fmt.Sprintf(
				"%s%s %s%s %s%s %s%s",
				chart.Nian.Gan, chart.Nian.Zhi,
				chart.Yue.Gan, chart.Yue.Zhi,
				chart.Ri.Gan, chart.Ri.Zhi,
				chart.Shi.Gan, chart.Shi.Zhi,
			)
			want := fmt.Sprintf(
				"%s %s %s %s",
				golden.Expected.Pillars.Nian,
				golden.Expected.Pillars.Yue,
				golden.Expected.Pillars.Ri,
				golden.Expected.Pillars.Shi,
			)
			if got != want {
				t.Fatalf("pillars = %q, want three-oracle consensus %q", got, want)
			}
		})

		side := 0
		if golden.Side == "after" {
			side = 1
		}
		coverage[[3]int{golden.Year, golden.TermIndex, side}] = true
	}

	for _, year := range fixture.Scope.Years {
		for term := 0; term < fixture.Scope.TermsPerYear; term++ {
			for side := 0; side < fixture.Scope.AnchorsPerTerm; side++ {
				if !coverage[[3]int{year, term, side}] {
					t.Fatalf("missing coverage: year=%d term-slot=%d side=%d", year, term, side)
				}
			}
		}
	}
}
