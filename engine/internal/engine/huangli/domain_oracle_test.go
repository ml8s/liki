package huangli

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_HuangliCore(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/huangli_core.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Sequence   []string          `json:"jianchu_sequence"`
		QingLong   map[string]string `json:"qing_long_start_all_12"`
		EventRules map[string]struct {
			Suitable  []string `json:"suitable"`
			Forbidden []string `json:"forbidden"`
		} `json:"event_rules"`
		KnownDates []struct {
			Date         string `json:"date"`
			DayGan       string `json:"day_gan"`
			DayZhi       string `json:"day_zhi"`
			Mansion      string `json:"mansion"`
			JianChu      string `json:"jianchu"`
			HuangDaoName string `json:"huangdao_name"`
			HuangDaoPath string `json:"huangdao_path"`
		} `json:"known_dates"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Sequence) != 12 || len(doc.QingLong) != 12 || len(doc.EventRules) != 15 {
		t.Fatalf("oracle coverage invalid: seq=%d qinglong=%d events=%d", len(doc.Sequence), len(doc.QingLong), len(doc.EventRules))
	}
	for i, name := range doc.Sequence {
		if jianChuCfg.Sequence[i] != name {
			t.Errorf("jianchu[%d] = %s, want %s", i, jianChuCfg.Sequence[i], name)
		}
	}
	for month, want := range doc.QingLong {
		z, err := ganzhi.ParseZhi(month)
		if err != nil {
			t.Fatal(err)
		}
		wantZ, err := ganzhi.ParseZhi(want)
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := qingLongStart[z]; !ok || got != wantZ {
			t.Errorf("%s月青龙起 = %v, want %s", month, got, want)
		}
	}
	for event, want := range doc.EventRules {
		got, ok := jianChuCfg.EventRules[event]
		if !ok {
			t.Fatalf("event %s missing", event)
		}
		if len(got.Suitable) != len(want.Suitable) || len(got.Forbidden) != len(want.Forbidden) {
			t.Fatalf("event %s rules differ: %+v, want %+v", event, got, want)
		}
		for i, name := range want.Suitable {
			if got.Suitable[i] != name {
				t.Errorf("event %s suitable[%d] = %s, want %s", event, i, got.Suitable[i], name)
			}
		}
		for i, name := range want.Forbidden {
			if got.Forbidden[i] != name {
				t.Errorf("event %s forbidden[%d] = %s, want %s", event, i, got.Forbidden[i], name)
			}
		}
	}
	for _, want := range doc.KnownDates {
		got, err := QueryDate(want.Date)
		if err != nil {
			t.Fatalf("%s: %v", want.Date, err)
		}
		if ganzhi.GanName(got.RiZhu.Gan) != want.DayGan || ganzhi.ZhiName(got.RiZhu.Zhi) != want.DayZhi ||
			got.JianChu != want.JianChu || got.HuangDao.Name != want.HuangDaoName || got.HuangDao.Path != want.HuangDaoPath {
			t.Fatalf("%s hard facts = %+v, want %+v", want.Date, got, want)
		}
		if want.Mansion != "" && got.Mansion.Name != want.Mansion {
			t.Errorf("%s mansion = %s, want %s", want.Date, got.Mansion.Name, want.Mansion)
		}
	}
}

func TestQueryDate_EventClassificationComesFromEngineTable(t *testing.T) {
	got, err := QueryDate("2024-06-15", "sign")
	if err != nil {
		t.Fatalf("QueryDate event: %v", err)
	}
	if got.Event != "sign" || got.EventLabel != "签约" || got.Suitability != "recommended" {
		t.Fatalf("event classification = %s/%s/%s, want sign/签约/recommended", got.Event, got.EventLabel, got.Suitability)
	}
	if got.Reason == "" {
		t.Fatal("event reason is empty")
	}
}
