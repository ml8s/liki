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
			if tc.ChartFrom != "" {
				input["chart"] = loadChartFixture(t, tc.ChartFrom)
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

func loadChartFixture(t *testing.T, name string) any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "actual", name))
	if err != nil {
		t.Fatalf("read chart fixture %s: %v", name, err)
	}
	var doc struct {
		Input struct {
			Chart json.RawMessage `json:"chart"`
		} `json:"input"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode chart fixture %s: %v", name, err)
	}
	var chart any
	if err := json.Unmarshal(doc.Input.Chart, &chart); err != nil {
		t.Fatalf("decode chart %s: %v", name, err)
	}
	return chart
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
