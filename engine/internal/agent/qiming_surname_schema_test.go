package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandler_QimingSurnameContract(t *testing.T) {
	reg := NewRPCRegistry()
	result, err := reg.Execute(context.TODO(), "qiming.surname", json.RawMessage(`{"source_surname":"Lee","max_candidates":2}`))
	if err != nil {
		t.Fatalf("qiming.surname: %v", err)
	}
	if getStr(result, "_product") != "qiming_surname" {
		t.Fatalf("_product = %q, want qiming_surname", getStr(result, "_product"))
	}
	var envelope struct {
		Data struct {
			Strategy   string `json:"strategy"`
			Candidates []struct {
				Surname    string `json:"surname"`
				MatchLevel string `json:"match_level"`
			} `json:"candidates"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Strategy != "phonetic" || len(envelope.Data.Candidates) == 0 || envelope.Data.Candidates[0].Surname != "李" {
		t.Fatalf("result = %+v", envelope.Data)
	}
	var rawEnvelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(result, &rawEnvelope); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"boundary"} {
		if _, ok := rawEnvelope.Data[key]; ok {
			t.Fatalf("result data has redundant %q", key)
		}
	}
	first := rawEnvelope.Data["candidates"].([]any)[0].(map[string]any)
	for _, key := range []string{"plain_pinyin"} {
		if _, ok := first[key]; ok {
			t.Fatalf("first candidate has redundant %q", key)
		}
	}
}

func TestHandler_QimingSurname_RejectsNonLatinSource(t *testing.T) {
	reg := NewRPCRegistry()
	if _, err := reg.Execute(context.TODO(), "qiming.surname", json.RawMessage(`{"source_surname":"王小明"}`)); err == nil {
		t.Fatal("expected non-Latin source surname to fail")
	}
}
