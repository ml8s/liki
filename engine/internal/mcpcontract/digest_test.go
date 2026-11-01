package mcpcontract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestCanonicalDigestIsStableAndKeyOrderIndependent(t *testing.T) {
	left := map[string]any{"b": json.Number("1"), "a": []any{"甲", map[string]any{"z": false, "y": nil}}}
	right := map[string]any{"a": []any{"甲", map[string]any{"y": nil, "z": false}}, "b": json.Number("1")}
	first, err := canonicalJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	second, err := canonicalJSON(right)
	if err != nil {
		t.Fatal(err)
	}
	wantCanonical := `{"a":["甲",{"y":null,"z":false}],"b":1}`
	if string(first) != string(second) || string(first) != wantCanonical {
		t.Fatalf("canonical JSON = %s / %s", first, second)
	}
}

func TestChartDigestExcludesItselfAndBindsChartContent(t *testing.T) {
	registry := testEngine()
	raw, err := registry.Call(context.Background(), "bazi_chart", mustJSON(t, BaziChartInput{
		StandardTime: "1990-05-20T12:49:00+08:00", Longitude: 116.4, Gender: "male",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var chart BaziChart
	decode(t, raw, &chart)
	if chart.ChartDigest == "" {
		t.Fatal("chart has no digest")
	}
	withoutDigest := chart
	withoutDigest.ChartDigest = ""
	first, err := chartDigest(withoutDigest)
	if err != nil {
		t.Fatal(err)
	}
	if string(chart.ChartDigest) != first {
		t.Fatalf("chart digest = %s, want %s", chart.ChartDigest, first)
	}

	changed := chart
	changed.Gender = "female"
	second, err := chartDigest(changed)
	if err != nil {
		t.Fatal(err)
	}
	if second == first {
		t.Fatal("changed chart kept the same digest")
	}
}

func TestCanonicalDigestMatchesPythonReferenceVector(t *testing.T) {
	value := map[string]any{"b": json.Number("1"), "a": []any{"甲", map[string]any{"z": false, "y": nil}}}
	got, err := canonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(got)
	want := "068a8baec8a2336e95b6153e1b88d070a9b2fc0a6826888732c68685828ff275"
	if hex.EncodeToString(sum[:]) != want {
		t.Fatalf("canonical digest = %s, want %s", hex.EncodeToString(sum[:]), want)
	}
}

func TestSourceChartDigestFailsClosed(t *testing.T) {
	chart := map[string]any{"birth_year": 2000, "gender": "male"}
	if _, err := sourceChartDigest(chart); err == nil || !strings.Contains(err.Error(), "no chart_digest") {
		t.Fatalf("missing digest error = %v", err)
	}
	digest, err := chartDigest(chart)
	if err != nil {
		t.Fatal(err)
	}
	chart["chart_digest"] = digest
	if _, err := sourceChartDigest(chart); err != nil {
		t.Fatalf("valid chart rejected: %v", err)
	}
	chart["birth_year"] = 2001
	if _, err := sourceChartDigest(chart); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("tampered chart error = %v", err)
	}
}

func TestXuankongSeriesSourceIdentity(t *testing.T) {
	registry := testEngine()
	chartRaw, err := registry.Call(context.Background(), "xuankong_chart", mustJSON(t, XuankongChartInput{
		StandardTime: "2026-07-31T12:00:00+08:00", ZuoShan: 2, XiangShan: 14,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var chart map[string]any
	decode(t, chartRaw, &chart)
	want := chart["chart_digest"].(string)
	input := map[string]any{"chart": chart, "standard_time": "2026-06-15T12:00:00+08:00"}
	raw, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		SourceChartDigest ChartDigest      `json:"source_chart_digest"`
		Items             []map[string]any `json:"items"`
	}
	decode(t, raw, &output)
	if string(output.SourceChartDigest) != want || len(output.Items[0]["house_overlay"].([]any)) == 0 {
		t.Fatalf("xuankong series identity = %s, overlay=%v", output.SourceChartDigest, output.Items[0]["house_overlay"])
	}

	input["chart"].(map[string]any)["chart_digest"] = "0000000000000000000000000000000000000000000000000000000000000000"
	if _, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input)); err == nil {
		t.Fatal("tampered xuankong chart was accepted")
	}

	delete(input, "chart")
	raw, err = registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var noChart struct {
		SourceChartDigest ChartDigest      `json:"source_chart_digest"`
		Items             []map[string]any `json:"items"`
	}
	decode(t, raw, &noChart)
	if noChart.SourceChartDigest != "" {
		t.Fatalf("chartless source digest = %s, want empty", noChart.SourceChartDigest)
	}
}

func TestReusableOutputsCarrySourceChartIdentity(t *testing.T) {
	registry := testEngine()
	makeChart := func(method string) (map[string]any, string) {
		t.Helper()
		raw, err := registry.Call(context.Background(), method, mustJSON(t, map[string]any{
			"standard_time": "1990-05-20T12:49:00+08:00", "longitude": 116.4, "gender": "male",
		}))
		if err != nil {
			t.Fatal(err)
		}
		var chart map[string]any
		decode(t, raw, &chart)
		digest, _ := chart["chart_digest"].(string)
		if digest == "" {
			t.Fatalf("%s has no chart_digest", method)
		}
		return chart, digest
	}

	for method, downstream := range map[string]string{
		"bazi_chart": "bazi_fullchart", "ziwei_chart": "ziwei_fullchart",
	} {
		t.Run(method, func(t *testing.T) {
			chart, digest := makeChart(method)
			raw, err := registry.Call(context.Background(), downstream, mustJSON(t, map[string]any{"chart": chart}))
			if err != nil {
				t.Fatal(err)
			}
			var output map[string]any
			decode(t, raw, &output)
			if got, _ := output["source_chart_digest"].(string); got != digest {
				t.Fatalf("%s source_chart_digest = %s, want %s", downstream, got, digest)
			}
		})
	}
}

func TestSeriesSourceChartIdentityAndCountContract(t *testing.T) {
	registry := testEngine()
	chartRaw, err := registry.Call(context.Background(), "bazi_chart", mustJSON(t, BaziChartInput{
		StandardTime: "1990-05-20T12:49:00+08:00", Longitude: 116.4, Gender: "male",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var chart map[string]any
	decode(t, chartRaw, &chart)
	want := chart["chart_digest"].(string)

	input := map[string]any{"chart": chart, "standard_time": "2026-06-15T12:00:00+08:00"}
	raw, err := registry.Call(context.Background(), "bazi_liunian", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		SourceChartDigest ChartDigest      `json:"source_chart_digest"`
		Items             []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	if string(output.SourceChartDigest) != want || len(output.Items) != 1 {
		t.Fatalf("default series = digest %s, items %d", output.SourceChartDigest, len(output.Items))
	}

	input["count"] = 0
	if _, err := registry.Call(context.Background(), "bazi_liunian", mustJSON(t, input)); err == nil {
		t.Fatal("explicit series count zero was accepted")
	}
}
