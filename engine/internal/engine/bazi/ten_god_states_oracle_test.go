package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_TenGodAndElementStates(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_ten_god_states.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID                   string                    `json:"id"`
			Pillars              []string                  `json:"pillars"`
			ExpectedTenGodStates map[string]map[string]any `json:"expected_ten_god_states"`
			ExpectedElementState map[string]map[string]any `json:"expected_element_states"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 4 {
		t.Fatalf("cases = %d, want 4", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			full := ComputeFullChart(atomicChartFromStrings(t, tc.Pillars))
			tenGods := map[string]map[string]any{}
			for _, state := range full.TenGodStates {
				tenGods[state.ShiShen] = map[string]any{
					"wuxing": state.Wuxing, "transparent": state.Transparent,
					"hidden": state.Hidden, "rooted": state.Rooted,
					"timely": state.Timely, "count": state.Count,
					"strength": state.Strength,
				}
			}
			elements := map[string]map[string]any{}
			for _, state := range full.ElementStates {
				elements[state.Wuxing] = map[string]any{
					"season_strength": state.SeasonStrength, "strength": state.Strength,
					"transparent": state.Transparent, "rooted": state.Rooted,
					"controls": state.Controls, "controlled_by": state.ControlledBy,
					"controller_strength": state.ControllerStrength,
				}
			}
			tenGods = normalizeJSONMap(t, tenGods)
			elements = normalizeJSONMap(t, elements)
			for name, expected := range tc.ExpectedTenGodStates {
				if !reflect.DeepEqual(tenGods[name], expected) {
					t.Errorf("ten god %s = %#v, want %#v", name, tenGods[name], expected)
				}
			}
			for name, expected := range tc.ExpectedElementState {
				if !reflect.DeepEqual(elements[name], expected) {
					t.Errorf("element %s = %#v, want %#v", name, elements[name], expected)
				}
			}
		})
	}
}

func normalizeJSONMap(t *testing.T, value map[string]map[string]any) map[string]map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestDomainOracle_DaYunRootStates(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_ten_god_states.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		DayunCases []struct {
			ID                    string   `json:"id"`
			Pillars               []string `json:"pillars"`
			DaYun                 string   `json:"dayun"`
			ExpectedRooted        bool     `json:"expected_rooted"`
			ExpectedRootRefSource []string `json:"expected_root_ref_sources"`
		} `json:"dayun_root_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.DayunCases) != 4 {
		t.Fatalf("dayun cases = %d, want 4", len(doc.DayunCases))
	}
	for _, tc := range doc.DayunCases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			runes := []rune(tc.DaYun)
			gan, err := ganzhi.ParseGan(string(runes[0]))
			if err != nil {
				t.Fatal(err)
			}
			zhi, err := ganzhi.ParseZhi(string(runes[1]))
			if err != nil {
				t.Fatal(err)
			}
			chart.DaYun = &DaYun{Steps: []DaYunStep{{Gan: gan, Zhi: zhi}}}
			full := ComputeFullChart(chart)
			step := full.DaYun.Steps[0]
			if step.Rooted == nil || *step.Rooted != tc.ExpectedRooted {
				t.Errorf("rooted = %v, want %v", step.Rooted, tc.ExpectedRooted)
			}
			sources := make([]string, 0, len(step.RootRefs))
			for _, ref := range step.RootRefs {
				sources = append(sources, ref.Source)
			}
			if !reflect.DeepEqual(sources, tc.ExpectedRootRefSource) {
				t.Errorf("root refs = %v, want %v", sources, tc.ExpectedRootRefSource)
			}
		})
	}
}
