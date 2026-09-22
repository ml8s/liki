package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBaziChart_ExternalOracle_TrueSolarChain(t *testing.T) {
	r := NewRPCRegistry()

	timeResult, err := r.Execute(context.Background(), "tianwen.time", json.RawMessage(`{
		"time":"1981-08-26T00:20:00+08:00",
		"longitude":132.044137
	}`))
	if err != nil {
		t.Fatalf("tianwen.time: %v", err)
	}
	var timeEnvelope struct {
		Data struct {
			Solar string `json:"solar"`
			Lunar struct {
				Shichen string `json:"shichen"`
			} `json:"lunar"`
		} `json:"data"`
	}
	if err := json.Unmarshal(timeResult, &timeEnvelope); err != nil {
		t.Fatal(err)
	}
	if got, want := timeEnvelope.Data.Solar, "1981-08-26T01:06:00+08:00"; got != want {
		t.Fatalf("solar time = %s, want %s", got, want)
	}
	if got, want := timeEnvelope.Data.Lunar.Shichen, "丑"; got != want {
		t.Fatalf("shichen = %s, want %s", got, want)
	}

	params, err := json.Marshal(map[string]any{
		"solar_time": timeEnvelope.Data.Solar,
		"gender":     "male",
	})
	if err != nil {
		t.Fatal(err)
	}
	chartResult, err := r.Execute(context.Background(), "bazi.chart", params)
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	var chartEnvelope struct {
		Data struct {
			Nian struct {
				Gan string `json:"gan"`
				Zhi string `json:"zhi"`
			} `json:"nian"`
			Yue struct {
				Gan string `json:"gan"`
				Zhi string `json:"zhi"`
			} `json:"yue"`
			Ri struct {
				Gan string `json:"gan"`
				Zhi string `json:"zhi"`
			} `json:"ri"`
			Shi struct {
				Gan string `json:"gan"`
				Zhi string `json:"zhi"`
			} `json:"shi"`
		} `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &chartEnvelope); err != nil {
		t.Fatal(err)
	}
	got := []string{
		chartEnvelope.Data.Nian.Gan + chartEnvelope.Data.Nian.Zhi,
		chartEnvelope.Data.Yue.Gan + chartEnvelope.Data.Yue.Zhi,
		chartEnvelope.Data.Ri.Gan + chartEnvelope.Data.Ri.Zhi,
		chartEnvelope.Data.Shi.Gan + chartEnvelope.Data.Shi.Zhi,
	}
	want := []string{"辛酉", "丙申", "丙子", "己丑"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pillar[%d] = %s, want %s; all=%v", i, got[i], want[i], got)
		}
	}
}
