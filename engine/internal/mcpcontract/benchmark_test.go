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
		`{"subject":{"birth_time":"1990-05-20T12:49:00+08:00","birth_longitude":116.4,"gender":"male"}}`)
}

func BenchmarkMCPZiweiChart(b *testing.B) {
	benchmarkCall(b, testEngine(), "ziwei_chart",
		`{"subject":{"birth_time":"1984-02-16T12:00:00+08:00","birth_longitude":116.4,"gender":"male"}}`)
}

func BenchmarkMCPQimenChart(b *testing.B) {
	benchmarkCall(b, testEngine(), "qimen_chart",
		`{"standard_time":"2000-06-15T12:00:00+08:00","longitude":116.4,"method":{"scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}}`)
}

func BenchmarkMCPBaziBond(b *testing.B) {
	registry := testEngine()
	subjectA := benchmarkBaziSubject(b, "1990-05-20T12:49:00+08:00", "male")
	subjectB := benchmarkBaziSubject(b, "1992-08-10T10:20:00+08:00", "female")
	arguments := mustBenchmarkJSON(b, map[string]any{
		"a": map[string]any{"subject": subjectA},
		"b": map[string]any{"subject": subjectB},
	})
	benchmarkCall(b, registry, "bazi_bond", string(arguments))
}

func BenchmarkMCPOutputValidationZiweiChart(b *testing.B) {
	registry := testEngine()
	arguments := json.RawMessage(`{"subject":{"birth_time":"1984-02-16T12:00:00+08:00","birth_longitude":116.4,"gender":"male"}}`)
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

func benchmarkBaziSubject(b *testing.B, standardTime, gender string) NatalSubject {
	b.Helper()
	var subject NatalSubject
	if err := json.Unmarshal([]byte(
		`{"birth_time":"`+standardTime+`","birth_longitude":116.4,"gender":"`+gender+`"}`,
	), &subject); err != nil {
		b.Fatal(err)
	}
	return subject
}

func mustBenchmarkJSON(b *testing.B, value any) json.RawMessage {
	b.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		b.Fatal(err)
	}
	return raw
}
