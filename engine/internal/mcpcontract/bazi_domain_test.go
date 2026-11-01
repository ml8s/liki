package mcpcontract

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"liki-engine/internal/engine/bazi"
)

func TestBaziChartSubjectProducesStableChart(t *testing.T) {
	registry := testEngine()
	for _, subject := range []NatalSubject{
		{BirthTime: "1990-05-20T11:49:00+08:00", BirthLongitude: 116.4, Gender: "male"},
		{BirthTime: "1984-02-16T08:00:00+08:00", BirthLongitude: 116.4, Gender: "female"},
		{BirthTime: "2000-09-20T08:30:00+08:00", BirthLongitude: 118.5, Gender: "female"},
	} {
		raw, err := registry.Call(context.Background(), "bazi_chart", mustJSON(t, BaziChartInput{Subject: subject}))
		if err != nil {
			t.Fatalf("bazi.chart %s: %v", subject.BirthTime, err)
		}
		var first BaziChartOutput
		decode(t, raw, &first)
		raw, err = registry.Call(context.Background(), "bazi_chart", mustJSON(t, BaziChartInput{Subject: subject}))
		if err != nil {
			t.Fatal(err)
		}
		var second BaziChartOutput
		decode(t, raw, &second)
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("subject %s produced unstable chart", subject.BirthTime)
		}
	}
}

func TestBaziChartSubjectRejectsInvalidLongitude(t *testing.T) {
	registry := testEngine()
	subject := NatalSubject{BirthTime: "1990-05-20T12:49:00+08:00", BirthLongitude: 500, Gender: "male"}
	if _, err := registry.Call(context.Background(), "bazi_chart", mustJSON(t, BaziChartInput{Subject: subject})); err == nil {
		t.Fatal("invalid subject longitude was accepted")
	}
}

func TestBaziChartRawProjectionMatchesWireDTO(t *testing.T) {
	registry := testEngine()
	subject := NatalSubject{BirthTime: "1990-05-20T12:49:00+08:00", BirthLongitude: 116.4, Gender: "male"}
	chart, err := subject.baziChart()
	if err != nil {
		t.Fatal(err)
	}
	full := bazi.ComputeFullChart(chart)
	for name, secondary := range map[string]*bazi.TiaoHouAvailability{
		"present": full.TiaoHou.Secondary,
		"absent":  nil,
	} {
		t.Run(name, func(t *testing.T) {
			value := full
			value.TiaoHou.Secondary = secondary
			raw, err := marshalRawBaziCompleteChart(value, chart.ZiShiRule)
			if err != nil {
				t.Fatal(err)
			}
			if err := registry.ValidateOutput("bazi_chart", raw); err != nil {
				t.Fatalf("raw output contract: %v", err)
			}
			want, err := wireFromDomain[BaziChartOutput](value)
			if err != nil {
				t.Fatal(err)
			}
			want.ZiShiRule = chart.ZiShiRule
			wantRaw, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var gotDocument, wantDocument any
			if err := json.Unmarshal(raw, &gotDocument); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wantRaw, &wantDocument); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotDocument, wantDocument) {
				t.Fatalf("raw chart projection differs from typed wire DTO: %v", diffJSON("", gotDocument, wantDocument))
			}
		})
	}
}
