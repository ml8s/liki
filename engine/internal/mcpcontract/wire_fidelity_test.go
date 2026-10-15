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
	"fullchart", "liunian", "liuyue", "liuri", "liushi", "xiaoyun", "daxian", "bond", "chart",
}

// migratedWireTools are the tools this migration moved to the typed registry.
var migratedWireTools = []string{
	"bazi.fullchart", "bazi.bond", "bazi.liunian", "bazi.liuyue", "bazi.liuri",
	"bazi.liushi", "bazi.xiaoyun",
	"ziwei.chart", "ziwei.fullchart", "ziwei.daxian", "ziwei.liunian", "ziwei.liuyue",
	"ziwei.liuri", "ziwei.liushi", "ziwei.bond",
	"qimen.chart",
}

// fixtureTool derives the tool name from a captured fixture file name such as
// "bazi_chart_1984_liunian" or "qimen_hour_zhuanpan_default".
func fixtureTool(name string) string {
	domain, _, ok := strings.Cut(name, "_")
	if !ok {
		return ""
	}
	if domain == "qimen" {
		return "qimen.chart"
	}
	rest := name[len(domain)+1:]
	tokens := strings.Split(rest, "_")
	for index := len(tokens) - 1; index >= 0; index-- {
		for _, method := range fidelityMethods {
			if tokens[index] == method {
				return domain + "." + method
			}
		}
	}
	return ""
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
	registry := NewEngine()
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
		if method == "bazi.chart" {
			// The lean chart carries clock-derived 大运 fields; it has its own
			// hand-written contract test.
			continue
		}
		if _, ok := registry.Definition(method); !ok {
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
		got, err := registry.Call(context.Background(), method, fixture.Input)
		if err != nil {
			t.Errorf("%s: %s: %v", file, method, err)
			continue
		}
		var gotDoc, wantDoc any
		if err := json.Unmarshal(got, &gotDoc); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(fixture.Output, &wantDoc); err != nil {
			t.Fatal(err)
		}
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
