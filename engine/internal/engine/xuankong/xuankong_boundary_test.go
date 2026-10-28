package xuankong

import (
	"testing"
	"time"
)

func TestComputeChartAtInstantUsesLichunYear(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)

	// 2024 立春 is 2024-02-04. Before it, 2023 is still 八运.
	before := time.Date(2024, 1, 15, 12, 0, 0, 0, cst)
	chart, err := ComputeChartAtInstant(before, 2, 14)
	if err != nil {
		t.Fatal(err)
	}
	if chart.Yun.Year != 2023 || chart.Yun.YunNumber != 8 {
		t.Fatalf("before 立春 = year %d yun %d, want 2023 / 八运", chart.Yun.Year, chart.Yun.YunNumber)
	}

	after := time.Date(2024, 2, 5, 12, 0, 0, 0, cst)
	chart, err = ComputeChartAtInstant(after, 2, 14)
	if err != nil {
		t.Fatal(err)
	}
	if chart.Yun.Year != 2024 || chart.Yun.YunNumber != 9 {
		t.Fatalf("after 立春 = year %d yun %d, want 2024 / 九运", chart.Yun.Year, chart.Yun.YunNumber)
	}
}

func TestComputeLiuNianAtInstantUsesLichunYear(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	before := time.Date(2024, 1, 15, 12, 0, 0, 0, cst)
	result, err := ComputeLiuNianAtInstant(before, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Year != 2023 {
		t.Fatalf("before 立春 year = %d, want 2023", result.Year)
	}
}
