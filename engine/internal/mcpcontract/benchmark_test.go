package mcpcontract

import (
	"context"
	"encoding/json"
	"testing"
)

var benchmarkOutputBytes int

func BenchmarkMCPNewEngine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewEngine()
	}
}

func BenchmarkMCPMaterializeAllDefinitions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		registry := NewEngine()
		for _, name := range registry.Names() {
			if _, err := registry.Definition(name); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkMCPMaterializeAllDefinitionsConcurrently(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := NewEngine().MaterializeSchemas(); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkCall(b *testing.B, registry *Registry, name string, arguments string) {
	b.Helper()
	raw := json.RawMessage(arguments)
	if _, err := registry.Call(context.Background(), name, raw); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output, err := registry.Call(context.Background(), name, raw)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkOutputBytes += len(output)
	}
}

func BenchmarkMCPBaziChart(b *testing.B) {
	benchmarkCall(b, testEngine(), "bazi_chart",
		`{"instant":"1990-05-20T03:49:00Z","timezone":"Asia/Shanghai","longitude":116.4,"gender":"male"}`)
}

func BenchmarkMCPZiweiChart(b *testing.B) {
	benchmarkCall(b, testEngine(), "ziwei_chart",
		`{"instant":"1984-02-16T04:00:00Z","timezone":"Asia/Shanghai","longitude":116.4,"gender":"male"}`)
}

func BenchmarkMCPQimenChart(b *testing.B) {
	benchmarkCall(b, testEngine(), "qimen_chart",
		`{"instant":"2000-06-15T04:00:00Z","timezone":"Asia/Shanghai","longitude":116.4,"method":{"scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}}`)
}

func BenchmarkMCPBaziFullChart(b *testing.B) {
	registry := testEngine()
	chart := benchmarkBaziChart(b, registry, "1990-05-20T03:49:00Z", "male")
	arguments := mustBenchmarkJSON(b, map[string]any{"chart": chart})
	benchmarkCall(b, registry, "bazi_fullchart", string(arguments))
}

func BenchmarkMCPBaziBond(b *testing.B) {
	registry := testEngine()
	chartA := benchmarkBaziChart(b, registry, "1990-05-20T03:49:00Z", "male")
	chartB := benchmarkBaziChart(b, registry, "1992-08-10T02:20:00Z", "female")
	arguments := mustBenchmarkJSON(b, map[string]any{
		"a": map[string]any{"chart": chartA},
		"b": map[string]any{"chart": chartB},
	})
	benchmarkCall(b, registry, "bazi_bond", string(arguments))
}

func BenchmarkMCPOutputValidationZiweiChart(b *testing.B) {
	registry := testEngine()
	arguments := json.RawMessage(`{"instant":"1984-02-16T04:00:00Z","timezone":"Asia/Shanghai","longitude":116.4,"gender":"male"}`)
	output, err := registry.Call(context.Background(), "ziwei_chart", arguments)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.ValidateOutput("ziwei_chart", output); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkBaziChart(b *testing.B, registry *Registry, instant, gender string) json.RawMessage {
	b.Helper()
	raw, err := registry.Call(context.Background(), "bazi_chart", json.RawMessage(
		`{"instant":"`+instant+`","timezone":"Asia/Shanghai","longitude":116.4,"gender":"`+gender+`"}`,
	))
	if err != nil {
		b.Fatal(err)
	}
	return raw
}

func mustBenchmarkJSON(b *testing.B, value any) json.RawMessage {
	b.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		b.Fatal(err)
	}
	return raw
}
