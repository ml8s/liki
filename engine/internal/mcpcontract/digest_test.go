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

func TestCanonicalDigestMatchesPythonReferenceVector(t *testing.T) {
	value := map[string]any{
		"b": json.Number("1"),
		"a": []any{"甲", map[string]any{"z": false, "y": nil}},
		"n": json.Number("12345678901234567890"),
		"s": "<a>&\"猴\"\u2028",
	}
	got, err := canonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(got)
	want := "bf15cbf7215513b63017a1a28b7f2a7590fffc4a23414a8a26e14fd119154d6d"
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

func TestXuankongSeriesValidatesReusableChart(t *testing.T) {
	registry := testEngine()
	chartRaw, err := registry.Call(context.Background(), "xuankong_chart", mustJSON(t, XuankongChartInput{
		ZhaiYunStartTime: "2026-07-31T12:00:00+08:00", ZuoShan: 2, XiangShan: 14,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var chart map[string]any
	decode(t, chartRaw, &chart)
	input := map[string]any{"chart": chart, "liu_nian_start_time": "2026-06-15T12:00:00+08:00"}
	raw, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		LiuNian []map[string]any `json:"liu_nian"`
	}
	decode(t, raw, &output)
	if len(output.LiuNian) != 1 || len(output.LiuNian[0]["house_overlay"].([]any)) == 0 {
		t.Fatalf("xuankong series output = %#v", output)
	}
	if _, exists := output.LiuNian[0]["source_chart_digest"]; exists {
		t.Fatal("derived series item carried source digest")
	}

	input["chart"].(map[string]any)["chart_digest"] = "0000000000000000000000000000000000000000000000000000000000000000"
	if _, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input)); err == nil {
		t.Fatal("tampered xuankong chart was accepted")
	}

	delete(input, "chart")
	if _, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, input)); err != nil {
		t.Fatalf("chartless xuankong series was rejected: %v", err)
	}
}

func TestNatalChartOutputsDoNotCarryDigests(t *testing.T) {
	registry := testEngine()
	for _, method := range []string{"bazi_chart", "ziwei_chart"} {
		t.Run(method, func(t *testing.T) {
			subject := map[string]any{
				"birth_time": "1990-05-20T12:49:00+08:00", "birth_longitude": 116.4, "gender": "male",
			}
			raw, err := registry.Call(context.Background(), method, mustJSON(t, map[string]any{"subject": subject}))
			if err != nil {
				t.Fatal(err)
			}
			var output map[string]any
			decode(t, raw, &output)
			for _, field := range []string{"source_chart_digest", "chart_digest"} {
				if _, exists := output[field]; exists {
					t.Fatalf("%s carried %s", method, field)
				}
			}
		})
	}
}

func TestSeriesSubjectOmitsSourceDigest(t *testing.T) {
	registry := testEngine()
	input := map[string]any{"subject": map[string]any{"birth_time": "1990-05-20T12:49:00+08:00", "birth_longitude": 116.4, "gender": "male"}, "liu_nian_start_time": "2026-06-15T12:00:00+08:00"}
	raw, err := registry.Call(context.Background(), "bazi_liunian", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		LiuNian []map[string]any `json:"liu_nian"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	if len(output.LiuNian) != 1 {
		t.Fatalf("omitted count items = %d, want 1", len(output.LiuNian))
	}
	if _, exists := output.LiuNian[0]["source_chart_digest"]; exists {
		t.Fatal("derived series carried source digest")
	}

	input["count"] = 0
	if _, err := registry.Call(context.Background(), "bazi_liunian", mustJSON(t, input)); err == nil {
		t.Fatal("explicit series count zero was accepted")
	}
}
