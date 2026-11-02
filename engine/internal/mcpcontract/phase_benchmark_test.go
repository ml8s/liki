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

func benchmarkBaziChartPhases(b *testing.B) (json.RawMessage, BaziChartInput, bazi.Chart, bazi.Chart, BaziChartOutput) {
	b.Helper()
	subject := testNatalSubject(b, standardInputFromInstant(b, "1990-05-20T03:49:00Z", "Asia/Shanghai"), "male")
	arguments := mustBenchmarkJSON(b, map[string]any{"subject": subject})
	var input BaziChartInput
	if err := json.Unmarshal(arguments, &input); err != nil {
		b.Fatal(err)
	}
	chart, err := input.Subject.baziChart()
	if err != nil {
		b.Fatal(err)
	}
	output, err := wireFromDomain[BaziChartOutput](chart)
	if err != nil {
		b.Fatal(err)
	}
	return arguments, input, chart, chart, output
}

func BenchmarkMCPBaziChartPhaseInputValidation(b *testing.B) {
	arguments, _, _, _, _ := benchmarkBaziChartPhases(b)
	benchmarkDecodeInput[BaziChartInput](b, testEngine(), "bazi_chart", arguments)
}

func BenchmarkMCPBaziChartPhaseTypedDecode(b *testing.B) {
	arguments, _, _, _, _ := benchmarkBaziChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded BaziChartInput
		if err := json.Unmarshal(arguments, &decoded); err != nil {
			b.Fatal(err)
		}
		benchmarkPhaseInputValid = decoded.Subject.BirthTime != ""
	}
}

func BenchmarkMCPBaziChartPhaseDomainDecode(b *testing.B) {
	_, input, _, _, _ := benchmarkBaziChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := input.Subject.baziChart(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziChartPhaseCompute(b *testing.B) {
	_, _, chart, _, _ := benchmarkBaziChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bazi.ComputeChartFromCore(bazi.NatalCore{
			Pillars:   chart.ToBazi().Slice(),
			NaYin:     chart.NaYinArray(),
			DaYun:     chart.DaYun,
			Gender:    chart.Gender,
			BirthYear: chart.BirthYear,
		})
	}
}

func BenchmarkMCPBaziChartPhaseRawProjection(b *testing.B) {
	_, _, _, full, _ := benchmarkBaziChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshalRawBaziChart(full); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziChartPhaseWireDTOConversion(b *testing.B) {
	_, _, _, full, _ := benchmarkBaziChartPhases(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := wireFromDomain[BaziChartOutput](full); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziChartPhaseOutputMarshal(b *testing.B) {
	_, _, _, _, output := benchmarkBaziChartPhases(b)
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
	arguments := mustBenchmarkJSON(b, BaziBondInput{
		A: BaziBondInputA{Subject: testNatalSubject(b, standardInputFromInstant(b, "1990-05-20T03:49:00Z", "Asia/Shanghai"), "male")},
		B: BaziBondInputB{Subject: testNatalSubject(b, standardInputFromInstant(b, "1992-08-10T02:20:00Z", "Asia/Shanghai"), "female")},
	})
	benchmarkDecodeInput[BaziBondInput](b, registry, "bazi_bond", arguments)
}

func BenchmarkMCPBaziBondPhaseDomainDecode(b *testing.B) {
	a := testNatalSubject(b, standardInputFromInstant(b, "1990-05-20T03:49:00Z", "Asia/Shanghai"), "male")
	bb := testNatalSubject(b, standardInputFromInstant(b, "1992-08-10T02:20:00Z", "Asia/Shanghai"), "female")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := a.baziNatalCore()
		if err != nil {
			b.Fatal(err)
		}
		_, err = bb.baziNatalCore()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPBaziBondPhaseCompute(b *testing.B) {
	a := testNatalSubject(b, standardInputFromInstant(b, "1990-05-20T03:49:00Z", "Asia/Shanghai"), "male")
	bb := testNatalSubject(b, standardInputFromInstant(b, "1992-08-10T02:20:00Z", "Asia/Shanghai"), "female")
	domainA, err := a.baziNatalCore()
	if err != nil {
		b.Fatal(err)
	}
	domainB, err := bb.baziNatalCore()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bazi.ComputeBond(domainA, domainB)
	}
}
