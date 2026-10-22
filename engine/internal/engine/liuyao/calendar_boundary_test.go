package liuyao

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

type calendarBoundaryCases struct {
	ConclusionScope string `json:"conclusion_scope"`
	Cases           []struct {
		ID                  string `json:"id"`
		ApparentSolarInput  string `json:"solar_time"`
		ExpectedMonthBranch string `json:"expected_month_branch"`
	} `json:"cases"`
}

func TestCalendarBoundaryFixtures_MonthBranch(t *testing.T) {
	raw, err := os.ReadFile("testdata/calendar/liuyao_boundaries.json")
	if err != nil {
		t.Fatalf("read calendar fixtures: %v", err)
	}
	var cases calendarBoundaryCases
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode calendar fixtures: %v", err)
	}
	for _, tc := range cases.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			local, err := time.Parse(time.RFC3339, tc.ApparentSolarInput)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.ApparentSolarFromCivil(local, 120, 8)
			chart := ComputeChart(local, st, YongShiYao, [6]int{7, 7, 7, 7, 7, 7})
			if got := chart.YueZhi.String(); got != tc.ExpectedMonthBranch {
				t.Fatalf("month branch=%s, want %s", got, tc.ExpectedMonthBranch)
			}
		})
	}
}
