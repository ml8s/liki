package tianwen

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

type tianwenMultiOracleFixture struct {
	Version int `json:"version"`
	Policy  struct {
		LunarDayBoundary string `json:"lunar_day_boundary"`
		JianYue          string `json:"jian_yue"`
		Consensus        string `json:"consensus"`
	} `json:"policy"`
	Scope struct {
		Years     []int `json:"years"`
		CaseCount int   `json:"case_count"`
	} `json:"scope"`
	Sources []struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	} `json:"sources"`
	Cases []struct {
		Name      string   `json:"name"`
		Kind      string   `json:"kind"`
		Consensus []string `json:"consensus"`
		Input     struct {
			Gregorian string `json:"gregorian"`
		} `json:"input"`
		Expected struct {
			Lunar struct {
				Year  int  `json:"year"`
				Month int  `json:"month"`
				Day   int  `json:"day"`
				Leap  bool `json:"leap"`
			} `json:"lunar"`
			JianYue string `json:"jian_yue"`
		} `json:"expected"`
	} `json:"cases"`
}

func TestTianwenMultiOracleGolden_LunarMonthBoundaries(t *testing.T) {
	raw, err := os.ReadFile("testdata/multi_oracle_golden.json")
	if err != nil {
		t.Fatalf("read multi-oracle golden: %v", err)
	}
	var fixture tianwenMultiOracleFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode multi-oracle golden: %v", err)
	}
	if fixture.Version != 1 || fixture.Policy.Consensus != "all_oracles_must_agree" {
		t.Fatalf("unsupported fixture: %+v", fixture)
	}
	if len(fixture.Sources) < 2 || len(fixture.Cases) < 96 {
		t.Fatalf("thin fixture: sources=%d cases=%d", len(fixture.Sources), len(fixture.Cases))
	}
	for _, source := range fixture.Sources {
		if source.ID != "lunar-python" && source.ID != "sxtwl" {
			t.Fatalf("unexpected oracle %q", source.ID)
		}
	}

	seenYears := make(map[int]bool)
	for _, golden := range fixture.Cases {
		if len(golden.Consensus) != 2 {
			t.Fatalf("%s consensus count = %d, want 2", golden.Name, len(golden.Consensus))
		}
		if golden.Kind != "month_start" && golden.Kind != "month_end" {
			t.Fatalf("%s invalid kind %q", golden.Name, golden.Kind)
		}
		parsed, err := time.ParseInLocation(
			"2006-01-02", golden.Input.Gregorian, time.FixedZone("CST", 8*int(time.Hour/time.Second)),
		)
		if err != nil {
			t.Fatalf("%s parse date: %v", golden.Name, err)
		}
		seenYears[parsed.Year()] = true

		t.Run(golden.Name, func(t *testing.T) {
			gregorian := GregorianTime(parsed)
			gotLunar := SolarToLunar(gregorian)
			wantLunar := LunarTime{
				Year:  golden.Expected.Lunar.Year,
				Month: golden.Expected.Lunar.Month,
				Day:   golden.Expected.Lunar.Day,
				Leap:  golden.Expected.Lunar.Leap,
			}
			if gotLunar.Year != wantLunar.Year || gotLunar.Month != wantLunar.Month ||
				gotLunar.Day != wantLunar.Day || gotLunar.Leap != wantLunar.Leap {
				t.Fatalf("lunar = %+v, want two-oracle %+v", gotLunar, wantLunar)
			}

			gotGregorian := LunarToGregorian(LunarTime{
				Year:  wantLunar.Year,
				Month: wantLunar.Month,
				Day:   wantLunar.Day,
				Leap:  wantLunar.Leap,
			})
			if gotDate := gotGregorian.Time().In(parsed.Location()).Format("2006-01-02"); gotDate != parsed.Format("2006-01-02") {
				t.Fatalf("round trip = %s, want %s", gotDate, parsed.Format("2006-01-02"))
			}

			wantZhi, err := ganzhi.ParseZhi(golden.Expected.JianYue)
			if err != nil {
				t.Fatalf("parse jian yue: %v", err)
			}
			if gotZhi := JianYue(gregorian); gotZhi != wantZhi {
				t.Fatalf("jian_yue = %s, want %s", ganzhi.ZhiName(gotZhi), golden.Expected.JianYue)
			}
		})
	}
	for _, year := range fixture.Scope.Years {
		if !seenYears[year] {
			t.Fatalf("missing coverage year %d", year)
		}
	}
}
