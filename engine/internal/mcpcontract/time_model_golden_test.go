package mcpcontract

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTimeModelGolden(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "golden", "time_model.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Version int `json:"version"`
		Cases   []struct {
			Name      string         `json:"name"`
			Tool      string         `json:"tool"`
			Input     map[string]any `json:"input"`
			ChartFrom string         `json:"chart_from"`
			Expect    map[string]any `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != 1 || len(doc.Cases) == 0 {
		t.Fatalf("invalid time-model golden: version=%d cases=%d", doc.Version, len(doc.Cases))
	}

	registry := testEngine()
	for _, tc := range doc.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			input := make(map[string]any, len(tc.Input)+1)
			for key, value := range tc.Input {
				input[key] = value
			}
			if tc.Tool == "bazi_chart" || tc.Tool == "ziwei_chart" {
				input["subject"] = map[string]any{
					"standard_time": tc.Input["standard_time"],
					"longitude":     tc.Input["longitude"],
					"gender":        tc.Input["gender"],
				}
				delete(input, "standard_time")
				delete(input, "longitude")
				delete(input, "gender")
			} else if tc.ChartFrom != "" {
				input["subject"] = loadSubjectFixture(t, tc.ChartFrom, tc.Tool)
			}
			gotRaw, err := registry.Call(context.Background(), tc.Tool, mustJSON(t, input))
			if err != nil {
				t.Fatalf("%s: %v", tc.Tool, err)
			}
			var got any
			if err := json.Unmarshal(gotRaw, &got); err != nil {
				t.Fatal(err)
			}
			for path, want := range tc.Expect {
				gotValue, ok := lookupJSONPath(got, path)
				if !ok {
					t.Fatalf("%s: missing %s", tc.Tool, path)
				}
				// JSON numbers decode to float64; compare normalized values.
				if !reflect.DeepEqual(gotValue, want) {
					t.Fatalf("%s.%s = %#v, want %#v", tc.Tool, path, gotValue, want)
				}
			}
		})
	}
}

func loadSubjectFixture(t *testing.T, name, tool string) map[string]any {
	t.Helper()
	subject := map[string]any{
		"standard_time": "1990-05-20T12:49:00+08:00",
		"longitude":     116.4,
		"gender":        "male",
	}
	domain, _, _ := strings.Cut(tool, "_")
	switch {
	case strings.Contains(name, "_female_"):
		subject["gender"] = "female"
		if domain == "bazi" {
			subject["standard_time"] = "2000-09-20T08:30:00+08:00"
		} else {
			subject["standard_time"] = "2000-09-20T08:00:00+08:00"
		}
	case strings.Contains(name, "_1984_"):
		if domain == "bazi" {
			subject["standard_time"] = "1984-02-15T08:00:00+08:00"
		} else {
			subject["standard_time"] = "1984-02-16T12:00:00+08:00"
		}
	}
	return subject
}

func lookupJSONPath(root any, path string) (any, bool) {
	current := root
	for _, segment := range strings.Split(path, ".") {
		switch node := current.(type) {
		case map[string]any:
			value, ok := node[segment]
			if !ok {
				return nil, false
			}
			current = value
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil, false
			}
			current = node[index]
		default:
			return nil, false
		}
	}
	return current, true
}
