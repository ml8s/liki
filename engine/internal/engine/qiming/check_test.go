package qiming

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEvaluateNames(t *testing.T) {
	results, err := EvaluateNames([]string{"林炎", "龍明", "病明"}, "火", []string{"木"}, []string{"水"})
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("results = %+v", results)
	}
	valid := results[0]
	if !valid.Valid || valid.GivenName != "林炎" || valid.Phonetic == nil || valid.Phonetic.Tones != "2-2" {
		t.Fatalf("valid result = %+v", valid)
	}
	if valid.Wuxing == nil ||
		!hasWuxingHit(valid.Wuxing.Yong, true) ||
		!hasWuxingHit(valid.Wuxing.Xi, true) ||
		!hasWuxingHit(valid.Wuxing.Ji, false) {
		t.Fatalf("wuxing = %+v", valid.Wuxing)
	}
	invalid := results[1]
	if invalid.Valid || len(invalid.Errors) == 0 || invalid.Errors[0].Code != "character_not_found" {
		t.Fatalf("invalid result = %+v", invalid)
	}
	negative := results[2]
	if negative.Valid || len(negative.Errors) == 0 || negative.Errors[0].Code != "negative_character_forbidden" {
		t.Fatalf("negative result = %+v", negative)
	}
}

func TestEvaluateNames_SerializesFalseYong(t *testing.T) {
	results, err := EvaluateNames([]string{"林炎"}, "土", nil, nil)
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	encoded, err := json.Marshal(results[0].Wuxing)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(encoded), `"yong":false`) {
		t.Fatalf("encoded wuxing = %s, want yong:false", encoded)
	}
	if strings.Contains(string(encoded), `"xi"`) || strings.Contains(string(encoded), `"ji"`) {
		t.Fatalf("encoded wuxing = %s, want only provided constraints", encoded)
	}
}

func TestEvaluateNames_OmitsInvalidNameFacts(t *testing.T) {
	results, err := EvaluateNames([]string{"龍明"}, "", nil, nil)
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	encoded, err := json.Marshal(results[0])
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, field := range []string{"characters", "phonetic", "wuxing"} {
		if strings.Contains(string(encoded), `"`+field+`"`) {
			t.Fatalf("encoded invalid evaluation = %s, contains %q", encoded, field)
		}
	}
}

func TestEvaluateNames_DeduplicatesRepeatedCharacterErrors(t *testing.T) {
	results, err := EvaluateNames([]string{"龍龍"}, "", nil, nil)
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	if len(results[0].Errors) != 1 {
		t.Fatalf("errors = %+v, want one deduplicated error", results[0].Errors)
	}
}

func TestEvaluateNames_OmitsWuxingWithoutConstraints(t *testing.T) {
	results, err := EvaluateNames([]string{"林炎"}, "", nil, nil)
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	if results[0].Wuxing != nil {
		t.Fatalf("wuxing = %+v, want nil", results[0].Wuxing)
	}
}

func TestEvaluateNames_RejectsEmptyBatch(t *testing.T) {
	if _, err := EvaluateNames(nil, "", nil, nil); err == nil {
		t.Fatal("EvaluateNames() error = nil, want error")
	}
}

func TestEvaluateNames_RejectsInvalidConstraints(t *testing.T) {
	tests := []struct {
		name     string
		yongShen string
		xiShen   []string
		jiShen   []string
	}{
		{"yongshen", "风", nil, nil},
		{"xishen", "", []string{"风"}, nil},
		{"jishen", "", nil, []string{"风"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := EvaluateNames([]string{"林炎"}, test.yongShen, test.xiShen, test.jiShen); err == nil {
				t.Fatal("EvaluateNames() error = nil, want error")
			}
		})
	}
}

func TestEvaluateNames_InvalidNameLength(t *testing.T) {
	results, err := EvaluateNames([]string{"林炎火"}, "", nil, nil)
	if err != nil {
		t.Fatalf("EvaluateNames() error = %v", err)
	}
	if results[0].Valid || len(results[0].Errors) != 1 || results[0].Errors[0].Code != "invalid_name_length" {
		t.Fatalf("result = %+v", results[0])
	}
}

func hasWuxingHit(value *bool, want bool) bool {
	return value != nil && *value == want
}
