package mcpcontract

import "testing"

func qimenChartParams(params map[string]any) map[string]any {
	method := map[string]any{}
	for _, field := range []string{
		"scope", "school", "dingju_method", "quarter_rule",
		"base_dingju_method", "dun_source", "hour_boundary",
	} {
		if value, exists := params[field]; exists {
			method[field] = value
			delete(params, field)
		}
	}
	if method["scope"] == nil {
		method["scope"] = "hour"
	}
	if method["school"] == nil {
		method["school"] = "zhuanpan"
	}
	scope, _ := method["scope"].(string)
	school, _ := method["school"].(string)
	if (scope == "hour" || scope == "day") &&
		method["dingju_method"] == nil && school != "mingfa_feipan" && school != "jinhan_yujing" {
		method["dingju_method"] = "chaibu"
	}
	if scope == "quarter" && method["quarter_rule"] == nil {
		method["quarter_rule"] = "ten_minute_sanyuan"
	}
	params["method"] = method
	return params
}

func callQimenChart(t *testing.T, registry *Registry, params map[string]any) map[string]any {
	t.Helper()
	return callToolDocument(t, registry, "qimen_chart", qimenChartParams(params))
}

func assertQimenChartError(t *testing.T, registry *Registry, params map[string]any, expected string) {
	t.Helper()
	assertToolError(t, registry, "qimen_chart", qimenChartParams(params), expected)
}
