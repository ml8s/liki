package bazi

import "testing"

// TestBaziGolden6_QiyunPrecision verifies 起运精度 golden（bazi_golden6.json）：
// 起运年/月/日全部对齐 lunar（含节气边界的月/日级精度校准用例）。
func TestBaziGolden6_QiyunPrecision(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden6.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{QiYunMonthDay: true})
		})
	}
}
