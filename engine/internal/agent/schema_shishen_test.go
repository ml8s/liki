package agent

import (
	"encoding/json"
	"testing"
)

func TestSchema_ShiShenEnumMatchesTenGodClosure(t *testing.T) {
	reg := NewRPCRegistry()

	expected := []string{
		"比肩", "劫财", "食神", "伤官", "正财",
		"偏财", "七杀", "正官", "偏印", "正印",
	}
	names := []string{
		"bazi.liunian",
		"bazi.liuyue",
		"bazi.liuri",
		"bazi.liushi",
	}

	for _, name := range names {
		method, ok := reg.Method(name)
		if !ok {
			t.Fatalf("%s: missing registered method", name)
		}
		var result struct {
			Properties struct {
				Data struct {
					Properties struct {
						ShiShen struct {
							Enum []string `json:"enum"`
						} `json:"shi_shen"`
					} `json:"properties"`
				} `json:"data"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(method.Result, &result); err != nil {
			t.Fatalf("%s: unmarshal result schema: %v", name, err)
		}
		got := result.Properties.Data.Properties.ShiShen.Enum
		if len(got) != len(expected) {
			t.Fatalf("%s: shi_shen enum length = %d, want %d (%v)", name, len(got), len(expected), got)
		}
		expectedSet := map[string]bool{}
		for _, value := range expected {
			expectedSet[value] = true
		}
		gotSet := map[string]bool{}
		for _, value := range got {
			gotSet[value] = true
		}
		for _, value := range expected {
			if !gotSet[value] {
				t.Errorf("%s: shi_shen enum missing %q (%v)", method.Name, value, got)
			}
		}
		for value := range gotSet {
			if !expectedSet[value] {
				t.Errorf("%s: shi_shen enum has non-closure value %q (%v)", method.Name, value, got)
			}
		}
	}
}
