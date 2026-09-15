package huangli

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestQueryDateMatrixGolden_TwoOracleCalendarAndEventRules(t *testing.T) {
	raw, err := os.ReadFile("testdata/query_date_matrix_golden.json")
	if err != nil {
		t.Fatalf("read query-date matrix golden: %v", err)
	}
	var doc struct {
		Version int `json:"version"`
		Policy  struct {
			DayBoundary string `json:"day_boundary"`
			Consensus   string `json:"consensus"`
		} `json:"policy"`
		Scope struct {
			CaseCount     int `json:"case_count"`
			Events        int `json:"events"`
			JianchuStates int `json:"jianchu_states"`
		} `json:"scope"`
		Cases []struct {
			Date     string `json:"date"`
			Event    string `json:"event"`
			Expected struct {
				DayGan       string `json:"day_gan"`
				DayZhi       string `json:"day_zhi"`
				YueZhi       string `json:"yue_zhi"`
				JianChu      string `json:"jian_chu"`
				HuangdaoName string `json:"huangdao_name"`
				HuangdaoPath string `json:"huangdao_path"`
				Suitability  string `json:"suitability"`
			} `json:"expected"`
			Consensus []string `json:"consensus"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode query-date matrix golden: %v", err)
	}
	if doc.Version != 1 || doc.Policy.Consensus != "all_oracles_must_agree" ||
		doc.Scope.Events != 15 || doc.Scope.JianchuStates != 12 ||
		doc.Scope.CaseCount != len(doc.Cases) {
		t.Fatalf("invalid fixture metadata: %+v", doc)
	}

	seenEvents := make(map[string]bool)
	seenJianchu := make(map[string]bool)
	for _, golden := range doc.Cases {
		t.Run(golden.Date+"-"+golden.Event, func(t *testing.T) {
			got, err := QueryDate(golden.Date, golden.Event)
			if err != nil {
				t.Fatalf("QueryDate: %v", err)
			}
			wantGan, err := ganzhi.ParseGan(golden.Expected.DayGan)
			if err != nil {
				t.Fatal(err)
			}
			wantDayZhi, err := ganzhi.ParseZhi(golden.Expected.DayZhi)
			if err != nil {
				t.Fatal(err)
			}
			wantYueZhi, err := ganzhi.ParseZhi(golden.Expected.YueZhi)
			if err != nil {
				t.Fatal(err)
			}
			if got.RiZhu.Gan != wantGan || got.RiZhu.Zhi != wantDayZhi {
				t.Fatalf("ri zhu = %s%s, want %s%s", ganzhi.GanName(got.RiZhu.Gan), ganzhi.ZhiName(got.RiZhu.Zhi), golden.Expected.DayGan, golden.Expected.DayZhi)
			}
			if monthZhi := yueZhuForDate(mustParseDate(t, golden.Date)).Zhi; monthZhi != wantYueZhi {
				t.Fatalf("yue zhi = %s, want %s", ganzhi.ZhiName(monthZhi), golden.Expected.YueZhi)
			}
			if got.JianChu != golden.Expected.JianChu {
				t.Fatalf("jianchu = %s, want %s", got.JianChu, golden.Expected.JianChu)
			}
			if got.HuangDao.Name != golden.Expected.HuangdaoName || got.HuangDao.Path != golden.Expected.HuangdaoPath {
				t.Fatalf("huangdao = %s/%s, want %s/%s", got.HuangDao.Name, got.HuangDao.Path, golden.Expected.HuangdaoName, golden.Expected.HuangdaoPath)
			}
			if got.Suitability != golden.Expected.Suitability {
				t.Fatalf("suitability = %s, want %s", got.Suitability, golden.Expected.Suitability)
			}
		})
		seenEvents[golden.Event] = true
		seenJianchu[golden.Expected.JianChu] = true
	}
	if len(seenEvents) != 15 || len(seenJianchu) != 12 {
		t.Fatalf("coverage events=%d jianchu=%d, want 15/12", len(seenEvents), len(seenJianchu))
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
