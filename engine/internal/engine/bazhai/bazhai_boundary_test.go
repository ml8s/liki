package bazhai

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestComputeChartAtStandardUsesLichunYear(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	// 1984 立春 is 1984-02-04. A 1984-01-20 birth belongs to the 1983 立春 year.
	before := time.Date(1984, 1, 20, 12, 0, 0, 0, cst)
	chart, err := ComputeChartAtStandard(ganzhi.Male, before)
	if err != nil {
		t.Fatal(err)
	}
	want := ComputeMingGua(ganzhi.Male, 1983)
	if chart.MingGua.Gua.Name != want.Gua.Name || chart.MingGua.Group != want.Group {
		t.Fatalf("命卦 = %s/%s, want 1983 %s/%s",
			chart.MingGua.Gua.Name, chart.MingGua.Group, want.Gua.Name, want.Group)
	}
	if chart.YearStars.Year != 1983 {
		t.Fatalf("year stars year = %d, want 1983", chart.YearStars.Year)
	}

	// On the 立春 date, the exact instant decides: 12:00 is before the term,
	// 18:00 is after it (2024 立春 is 2024-02-04 16:27 CST).
	beforeTerm := time.Date(2024, 2, 4, 12, 0, 0, 0, cst)
	chart, err = ComputeChartAtStandard(ganzhi.Female, beforeTerm)
	if err != nil {
		t.Fatal(err)
	}
	want = ComputeMingGua(ganzhi.Female, 2023)
	if chart.MingGua.Gua.Name != want.Gua.Name || chart.MingGua.Group != want.Group {
		t.Fatalf("立春 before term 命卦 = %s/%s, want 2023 %s/%s",
			chart.MingGua.Gua.Name, chart.MingGua.Group, want.Gua.Name, want.Group)
	}

	afterTerm := time.Date(2024, 2, 4, 18, 0, 0, 0, cst)
	chart, err = ComputeChartAtStandard(ganzhi.Female, afterTerm)
	if err != nil {
		t.Fatal(err)
	}
	want = ComputeMingGua(ganzhi.Female, 2024)
	if chart.MingGua.Gua.Name != want.Gua.Name || chart.MingGua.Group != want.Group {
		t.Fatalf("立春 after term 命卦 = %s/%s, want 2024 %s/%s",
			chart.MingGua.Gua.Name, chart.MingGua.Group, want.Gua.Name, want.Group)
	}
}
