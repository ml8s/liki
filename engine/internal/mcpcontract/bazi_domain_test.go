package mcpcontract

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"liki-engine/internal/engine/bazi"
)

func TestBaziChartDomainMapperMatchesJSONProjection(t *testing.T) {
	registry := testEngine()
	for _, tc := range []struct {
		civilTime string
		gender    string
	}{
		{civilTime: "1990-05-20T11:49:00+08:00", gender: "male"},
		{civilTime: "1984-02-16T08:00:00+08:00", gender: "female"},
		{civilTime: "2000-09-20T08:30:00+08:00", gender: "female"},
	} {
		raw, err := registry.Call(context.Background(), "bazi_chart", json.RawMessage(
			`{"standard_time":"`+tc.civilTime+`","longitude":116.4,"gender":"`+tc.gender+`"}`,
		))
		if err != nil {
			t.Fatalf("bazi.chart: %v", err)
		}
		var wire BaziChart
		if err := json.Unmarshal(raw, &wire); err != nil {
			t.Fatal(err)
		}

		got, err := wire.domain()
		if err != nil {
			t.Fatalf("typed domain decode: %v", err)
		}
		projection, err := json.Marshal(wire)
		if err != nil {
			t.Fatal(err)
		}
		var want bazi.Chart
		if err := json.Unmarshal(projection, &want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("typed domain decode differs from JSON projection:\ngot: %#v\nwant:%#v", got, want)
		}
	}
}

func TestBaziChartDomainMapperRejectsInvalidSymbols(t *testing.T) {
	registry := testEngine()
	raw, err := registry.Call(context.Background(), "bazi_chart", json.RawMessage(
		`{"standard_time":"1990-05-20T12:49:00+08:00","longitude":116.4,"gender":"male"}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	var wire BaziChart
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	wire.Ri.Gan = "bad"
	if _, err := wire.domain(); err == nil {
		t.Fatal("invalid ri gan was accepted")
	}
}

func TestBaziFullChartRawProjectionMatchesWireDTO(t *testing.T) {
	registry := testEngine()
	chartRaw, err := registry.Call(context.Background(), "bazi_chart", json.RawMessage(
		`{"standard_time":"1990-05-20T12:49:00+08:00","longitude":116.4,"gender":"male"}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	var wire BaziChart
	if err := json.Unmarshal(chartRaw, &wire); err != nil {
		t.Fatal(err)
	}
	chart, err := wire.domain()
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
			raw, err := marshalRawBaziFullChart(value)
			if err != nil {
				t.Fatal(err)
			}
			if err := registry.ValidateOutput("bazi_fullchart", raw); err != nil {
				t.Fatalf("raw output contract: %v", err)
			}
			want, err := wireFromDomain[BaziFullChartOutput](value)
			if err != nil {
				t.Fatal(err)
			}
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
				t.Fatal("raw fullchart projection differs from typed wire DTO")
			}
		})
	}
}
