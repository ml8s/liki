package mcpcontract

import (
	"fmt"
	"reflect"
	"strings"
)

// fixtureTool derives the tool name from a fixture file name such as
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
		switch tokens[index] {
		case "liunian", "liuyue", "liuri", "liushi", "xiaoyun", "bond", "chart":
			switch tokens[index] {
			case "liunian", "liuyue", "liuri", "liushi", "xiaoyun":
				return domain + "_liupan"
			default:
				return domain + "_" + tokens[index]
			}
		}
	}
	return ""
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
