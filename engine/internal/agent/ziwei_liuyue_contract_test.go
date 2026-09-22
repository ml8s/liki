package agent

import "testing"

func TestRPC_ZiweiLiuYueTargetLunarContract(t *testing.T) {
	reg := NewRPCRegistry()
	chart := executeAndDecode(t, reg, "ziwei.chart",
		mustJSON(t, map[string]any{
			"lunar":  map[string]any{"year": 2000, "month": 8, "day": 23, "shichen": "辰"},
			"gender": "male",
		}),
	)["data"].(map[string]any)

	target := map[string]any{"year": 2025, "month": 6, "day": 16, "leap": true}
	result := executeAndDecode(t, reg, "ziwei.liuyue",
		mustJSON(t, map[string]any{"chart": chart, "target_lunar": target}),
	)
	data := result["data"].(map[string]any)
	period := data["resolved_period"].(map[string]any)
	flowMonth := period["flow_month"].(map[string]any)
	if flowMonth["year"] != float64(2025) || flowMonth["month"] != float64(7) || flowMonth["leap"] != false {
		t.Fatalf("flow_month = %#v, want 2025 month 7 non-leap", flowMonth)
	}

	tests := []map[string]any{
		// Old bare month input is rejected; the exact lunar day is required.
		{"chart": chart, "lunar_year": 2025, "lunar_month": 6},
		{"chart": chart, "target_lunar": map[string]any{"year": 2025, "month": 6, "leap": true}},
		{"chart": chart, "target_lunar": map[string]any{"year": 2024, "month": 6, "day": 10, "leap": true}},
		{"chart": chart, "target_lunar": map[string]any{"year": 2025, "month": 6, "day": 31, "leap": true}},
	}
	for _, params := range tests {
		assertError(t, reg, "ziwei.liuyue", mustJSON(t, params))
	}
}
