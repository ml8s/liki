package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fidelityMethods maps captured fixture suffixes to the wire tool they replay.
var fidelityMethods = []string{
	"liunian", "liuyue", "liuri", "liushi", "xiaoyun", "daxian", "bond", "chart",
}

// migratedWireTools are the tools this migration moved to the typed registry.
var migratedWireTools = []string{
	"bazi_bond", "bazi_liunian", "bazi_liuyue", "bazi_liuri",
	"bazi_liushi", "bazi_xiaoyun",
	"ziwei_chart", "ziwei_daxian", "ziwei_liunian", "ziwei_liuyue",
	"ziwei_liuri", "ziwei_liushi", "ziwei_bond",
	"qimen_chart",
}

// fixtureTool derives the tool name from a captured fixture file name such as
// "bazi_chart_1984_liunian" or "qimen_hour_zhuanpan_default".
func fixtureTool(name string) string {
	domain, _, ok := strings.Cut(name, "_")
	if !ok {
		return ""
	}
	if domain == "qimen" {
		return "qimen_chart"
	}
	rest := name[len(domain)+1:]
	tokens := strings.Split(rest, "_")
	for index := len(tokens) - 1; index >= 0; index-- {
		for _, method := range fidelityMethods {
			if tokens[index] == method {
				return domain + "_" + method
			}
		}
	}
	return ""
}

// normalizeLegacySeriesRoot is no longer needed: all captured series fixtures
// use named domain roots. Keep the call site stable while migration tests
// evolve, but make the no-adapter rule explicit.
func normalizeLegacySeriesRoot(tool string, want any) any {
	seriesRoots := map[string]string{
		"bazi_liunian": "liu_nian", "bazi_liuyue": "liu_yue",
		"bazi_liuri": "liu_ri", "bazi_liushi": "liu_shi",
		"bazi_xiaoyun": "xiao_yun", "ziwei_daxian": "da_xian",
		"ziwei_liunian": "liu_nian", "ziwei_liuyue": "liu_yue",
		"ziwei_liuri": "liu_ri", "ziwei_liushi": "liu_shi",
		"xuankong_liunian": "liu_nian",
	}
	root, isSeries := seriesRoots[tool]
	if !isSeries {
		return want
	}
	value, ok := want.(map[string]any)
	if !ok || value[root] == nil {
		panic("legacy series fixture must be regenerated: " + tool)
	}
	return want
}

// TestEngineContractWireFidelity replays every captured real handler input and
// output through the typed registry. The typed tool must return the same JSON
// document as the legacy handler did, which is the evidence that the wire DTOs
// did not drop, rename, or reshape any field.
func TestEngineContractWireFidelity(t *testing.T) {
	dir := filepath.Join("testdata", "actual")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	registry := testEngine()
	checked := map[string]int{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		file := entry.Name()
		method := fixtureTool(strings.TrimSuffix(file, ".json"))
		if method == "" {
			t.Errorf("%s: cannot derive tool name from fixture", file)
			continue
		}
		if method == "bazi_chart" {
			// The canonical chart has a dedicated hand-written contract test.
			continue
		}
		if _, err := registry.Definition(method); err != nil {
			t.Errorf("%s: %s is not registered", file, method)
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct {
			Input  json.RawMessage `json:"input"`
			Output json.RawMessage `json:"output"`
		}
		if err := json.Unmarshal(raw, &fixture); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		var oldInput map[string]any
		if err := json.Unmarshal(fixture.Input, &oldInput); err != nil {
			t.Fatalf("%s: decode input: %v", file, err)
		}
		inputMap := subjectFixtureInput(file, method, oldInput)
		if file == "bazi_chart_1984_liunian.json" {
			pretty, _ := json.MarshalIndent(inputMap, "", "  ")
			t.Logf(" transformed input: %s", pretty)
		}
		fixture.Input, err = json.Marshal(inputMap)
		if err != nil {
			t.Fatal(err)
		}
		got, err := registry.Call(context.Background(), method, fixture.Input)
		if err != nil {
			t.Errorf("%s: %s: %v", file, method, err)
			continue
		}
		if err := registry.ValidateOutput(method, got); err != nil {
			t.Errorf("%s: %s output contract: %v", file, method, err)
			continue
		}
		var gotDoc, wantDoc any
		if err := json.Unmarshal(got, &gotDoc); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(fixture.Output, &wantDoc); err != nil {
			t.Fatal(err)
		}
		wantDoc = normalizeLegacySeriesRoot(method, wantDoc)
		if !reflect.DeepEqual(gotDoc, wantDoc) {
			diffs := diffJSON("", gotDoc, wantDoc)
			if len(diffs) > 12 {
				diffs = append(diffs[:12], fmt.Sprintf("… %d more differences", len(diffs)-12))
			}
			t.Errorf("%s: %s wire output differs from the captured legacy output\n%s",
				file, method, strings.Join(diffs, "\n"))
		}
		checked[method]++
	}
	for _, method := range migratedWireTools {
		if checked[method] == 0 {
			t.Errorf("%s: no captured fixture replayed", method)
		}
	}
}

// subjectFixture returns the natal subject associated with a captured family.
func subjectFixture(file, method string) map[string]any {
	domain, _, _ := strings.Cut(method, "_")
	base := domain + "_chart.json"
	switch {
	case strings.Contains(file, "_female_") || strings.Contains(file, "_female."):
		base = domain + "_chart_female.json"
	case strings.Contains(file, "_1984_") || strings.Contains(file, "_1984."):
		base = domain + "_chart_1984.json"
	case strings.Contains(file, "_1990_") || strings.Contains(file, "_1990."):
		base = domain + "_chart_1990.json"
	}
	raw, err := os.ReadFile(filepath.Join("testdata", "actual", base))
	if err != nil {
		panic(fmt.Sprintf("load subject fixture %s: %v", base, err))
	}
	var fixture struct {
		Input struct {
			Subject NatalSubject `json:"subject"`
		} `json:"input"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		panic(fmt.Sprintf("decode subject fixture %s: %v", base, err))
	}
	return map[string]any{
		"birth_time":      fixture.Input.Subject.BirthTime,
		"birth_longitude": fixture.Input.Subject.BirthLongitude,
		"gender":          fixture.Input.Subject.Gender,
	}
}

// subjectFixtureInput migrates captured chart-transport fixtures to the
// subject contract while preserving their captured output for comparison.
func subjectFixtureInput(file, method string, old map[string]any) map[string]any {
	if !strings.HasPrefix(method, "bazi_") && !strings.HasPrefix(method, "ziwei_") {
		return old
	}
	subject := subjectFixture(file, method)
	switch method {
	case "bazi_chart", "ziwei_chart", "ziwei_daxian", "bazi_xiaoyun":
		target := map[string]any{"subject": subject}
		if method == "bazi_xiaoyun" {
			target["xiao_yun_count"] = 5
		}
		return target
	case "bazi_bond":
		a := subjectFixture("bazi_chart.json", "bazi_chart")
		b := subjectFixture("bazi_chart_female.json", "bazi_chart")
		return map[string]any{"a": map[string]any{"subject": a}, "b": map[string]any{"subject": b}}
	case "ziwei_bond":
		a := subjectFixture("ziwei_chart.json", "ziwei_chart")
		b := subjectFixture("ziwei_chart_female.json", "ziwei_chart")
		return map[string]any{"a": map[string]any{"subject": a}, "b": map[string]any{"subject": b}}
	default:
		target := map[string]any{"subject": subject}
		for key, value := range old {
			if key == "subject" {
				continue
			}
			switch key {
			case "start_time":
				target[methodSeriesFields(method)[0]] = value
				continue
			case "count":
				target[methodSeriesFields(method)[1]] = value
				continue
			case "longitude":
				if field := methodLongitudeField(method); field != "" {
					target[field] = value
					continue
				}
			}
			target[key] = value
		}
		return target
	}
}

// methodSeriesFields returns the method-owned input field names.
func methodSeriesFields(method string) [2]string {
	switch method {
	case "bazi_liunian", "ziwei_liunian":
		return [2]string{"liunian_start_time", "liunian_count"}
	case "bazi_liuyue", "ziwei_liuyue":
		return [2]string{"liuyue_start_time", "liuyue_count"}
	case "bazi_liuri", "ziwei_liuri":
		return [2]string{"liuri_start_time", "liuri_count"}
	case "bazi_liushi", "ziwei_liushi":
		return [2]string{"liushi_start_time", "liushi_count"}
	default:
		return [2]string{}
	}
}

func methodLongitudeField(method string) string {
	switch method {
	case "bazi_liuri":
		return "liuri_longitude"
	case "bazi_liushi", "ziwei_liushi":
		return "liushi_longitude"
	default:
		return ""
	}
}

// diffJSON lists structural differences between two decoded JSON documents.
func diffJSON(path string, got, want any) []string {
	var diffs []string
	switch wantValue := want.(type) {
	case map[string]any:
		gotValue, ok := got.(map[string]any)
		if !ok {
			return []string{fmt.Sprintf("%s: type %T, want object", pathOrRoot(path), got)}
		}
		for key, wantSub := range wantValue {
			gotSub, ok := gotValue[key]
			if !ok {
				diffs = append(diffs, fmt.Sprintf("%s.%s: missing in typed output", pathOrRoot(path), key))
				continue
			}
			diffs = append(diffs, diffJSON(path+"."+key, gotSub, wantSub)...)
		}
		for key := range gotValue {
			if _, ok := wantValue[key]; !ok {
				diffs = append(diffs, fmt.Sprintf("%s.%s: unexpected in typed output", pathOrRoot(path), key))
			}
		}
	case []any:
		gotValue, ok := got.([]any)
		if !ok {
			return []string{fmt.Sprintf("%s: type %T, want array", pathOrRoot(path), got)}
		}
		if len(gotValue) != len(wantValue) {
			diffs = append(diffs, fmt.Sprintf("%s: %d elements, want %d", pathOrRoot(path), len(gotValue), len(wantValue)))
		}
		for index := 0; index < len(gotValue) && index < len(wantValue); index++ {
			diffs = append(diffs, diffJSON(fmt.Sprintf("%s[%d]", path, index), gotValue[index], wantValue[index])...)
		}
	default:
		if !reflect.DeepEqual(got, want) {
			diffs = append(diffs, fmt.Sprintf("%s: %v, want %v", pathOrRoot(path), got, want))
		}
	}
	return diffs
}

func pathOrRoot(path string) string {
	if path == "" {
		return "(root)"
	}
	return strings.TrimPrefix(path, ".")
}
