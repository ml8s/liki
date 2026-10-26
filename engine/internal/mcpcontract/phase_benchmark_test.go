package mcpcontract

import (
	"encoding/json"
	"testing"

	"liki-engine/internal/engine/bazi"
)

var benchmarkPhaseInputValid bool

func benchmarkDecodeInput[In any](b *testing.B, registry *Registry, name string, raw json.RawMessage) {
	b.Helper()
	tool, ok := registry.tools[name]
	if !ok {
		b.Fatalf("%s is not registered", name)
	}
	schemas, err := tool.schemas.get()
	if err != nil {
		b.Fatal(err)
	}
	if _, err := decodeInputForSchema[In](schemas, raw); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := decodeInputForSchema[In](schemas, raw); err != nil {
			b.Fatal(err)
		}
		benchmarkPhaseInputValid = true
	}
}

func benchmarkBaziFullChartPhases(b *testing.B) (json.RawMessage, BaziFullChartInput, bazi.Chart, bazi.FullChart, BaziFullChartOutput) {
	b.Helper()
	registry := testEngine()
	chartRaw := benchmarkBaziChart(b, registry, "1990-05-20T03:49:00Z", "male")
	arguments := mustBenchmarkJSON(b, map[string]any{"chart": chartRaw})
	var input BaziFullChartInput
	if err := json.Unmarshal(arguments, &input); err != nil {
		b.Fatal(err)
	}
	chart, err := input.Chart.domain()
	if err != nil {
		b.Fatal(err)
	}
	full := bazi.ComputeFullChart(chart)
	output, err := wireFromDomain[BaziFullChartOutput](full)
	if err != nil {
		b.Fatal(err)
	}
	return arguments, input, chart, full, output
}

func BenchmarkMCPBaziFullChartPhaseInputValidation(b *testing.B) {
	arguments, _, _, _, _ := benchmarkBaziFullChartPhases(b)
	benchmarkDecodeInput[BaziFullChartInput](b, testEngine(), "bazi_fullchart", arguments)
}

func BenchmarkMCPBaziFullChartPhaseTypedDecode(b *testing.B) {
	arguments, _, _, _, _ := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded BaziFullChartInput
		if err := json.Unmarshal(arguments, &decoded); err != nil {
			b.Fatal(err)
		}
		benchmarkPhaseInputValid = decoded.Chart.BirthYear != 0
	}
}

func BenchmarkMCPBaziFullChartPhaseDomainDecode(b *testing.B) {
	_, input, _, _, _ := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := input.Chart.domain(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziFullChartPhaseCompute(b *testing.B) {
	_, _, chart, _, _ := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bazi.ComputeFullChart(chart)
	}
}

func BenchmarkMCPBaziFullChartPhaseRawProjection(b *testing.B) {
	_, _, _, full, _ := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshalRawBaziFullChart(full); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziFullChartPhaseWireDTOConversion(b *testing.B) {
	_, _, _, full, _ := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := wireFromDomain[BaziFullChartOutput](full); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziFullChartPhaseOutputMarshal(b *testing.B) {
	_, _, _, _, output := benchmarkBaziFullChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(output); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziBondPhaseInputValidation(b *testing.B) {
	registry := testEngine()
	chartA := benchmarkBaziChart(b, registry, "1990-05-20T03:49:00Z", "male")
	chartB := benchmarkBaziChart(b, registry, "1992-08-10T02:20:00Z", "female")
	arguments := mustBenchmarkJSON(b, BaziBondInput{
		A: BaziBondInputA{Chart: mustDecodeBenchmarkChart(b, chartA)},
		B: BaziBondInputB{Chart: mustDecodeBenchmarkChart(b, chartB)},
	})
	benchmarkDecodeInput[BaziBondInput](b, registry, "bazi_bond", arguments)
}

func BenchmarkMCPBaziBondPhaseDomainDecode(b *testing.B) {
	registry := testEngine()
	chartA := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1990-05-20T03:49:00Z", "male",
	))
	chartB := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1992-08-10T02:20:00Z", "female",
	))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := chartA.domain(); err != nil {
			b.Fatal(err)
		}
		if _, err := chartB.domain(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziBondPhaseCompute(b *testing.B) {
	registry := testEngine()
	chartA := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1990-05-20T03:49:00Z", "male",
	))
	chartB := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1992-08-10T02:20:00Z", "female",
	))
	domainA, err := chartA.domain()
	if err != nil {
		b.Fatal(err)
	}
	domainB, err := chartB.domain()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bazi.ComputeBond(domainA, domainB)
	}
}

func BenchmarkMCPBaziBondPhaseOutputMarshal(b *testing.B) {
	registry := testEngine()
	chartA := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1990-05-20T03:49:00Z", "male",
	))
	chartB := mustDecodeBenchmarkChart(b, benchmarkBaziChart(
		b, registry, "1992-08-10T02:20:00Z", "female",
	))
	domainA, err := chartA.domain()
	if err != nil {
		b.Fatal(err)
	}
	domainB, err := chartB.domain()
	if err != nil {
		b.Fatal(err)
	}
	output := bazi.ComputeBond(domainA, domainB)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(output); err != nil {
			b.Fatal(err)
		}
	}
}

func mustDecodeBenchmarkChart(b *testing.B, raw json.RawMessage) BaziChart {
	b.Helper()
	var chart BaziChart
	if err := json.Unmarshal(raw, &chart); err != nil {
		b.Fatal(err)
	}
	return chart
}
