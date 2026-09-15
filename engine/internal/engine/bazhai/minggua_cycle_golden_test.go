package bazhai

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestMingGuaCycleGolden_AllYearsAndGenders(t *testing.T) {
	raw, err := os.ReadFile("testdata/minggua_cycle_golden.json")
	if err != nil {
		t.Fatalf("read ming-gua cycle golden: %v", err)
	}
	var doc struct {
		Version int `json:"version"`
		Policy  struct {
			YearBoundary string `json:"year_boundary"`
			Coverage     string `json:"coverage"`
		} `json:"policy"`
		Scope struct {
			CaseCount int `json:"case_count"`
		} `json:"scope"`
		Cases []struct {
			Gender string `json:"gender"`
			Year   int    `json:"year"`
			Gua    string `json:"gua"`
			Group  string `json:"group"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode ming-gua cycle golden: %v", err)
	}
	if doc.Version != 1 || doc.Policy.YearBoundary != "gregorian_calendar_year" ||
		doc.Policy.Coverage != "1900-2099_all_gender" || len(doc.Cases) != doc.Scope.CaseCount {
		t.Fatalf("invalid fixture metadata: %+v", doc)
	}
	if len(doc.Cases) != 400 {
		t.Fatalf("case count = %d, want 400", len(doc.Cases))
	}

	for _, golden := range doc.Cases {
		gender := ganzhi.Male
		if golden.Gender == "female" {
			gender = ganzhi.Female
		} else if golden.Gender != "male" {
			t.Fatalf("invalid gender %q", golden.Gender)
		}
		t.Run(golden.Gender+"-"+strconv.Itoa(golden.Year), func(t *testing.T) {
			got := ComputeMingGua(gender, golden.Year)
			if got.YearBoundary != "gregorian_calendar_year" {
				t.Fatalf("year boundary = %q", got.YearBoundary)
			}
			if got.Gua.Name != golden.Gua || got.Group != golden.Group {
				t.Fatalf("%d %s = %s/%s, want %s/%s", golden.Year, golden.Gender, got.Gua.Name, got.Group, golden.Gua, golden.Group)
			}
		})
	}
}
