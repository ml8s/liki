package bazi

import "testing"

// TestBaziGolden2_EdgeLayers verifies 边界 golden（bazi_golden2.json，16 例）：
// 立春/惊蛰前后、晚子时+立春、闰月前后、大月三十、跨年、夏令时、西半球、1900/2050、子时三刻。
// golden 由 lunar-typescript 生成；分钟级节气边界用例已调整为 30min 外（避开 lunar 节气近似误差）。
func TestBaziGolden2_EdgeLayers(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden2.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{MingGong: true})
		})
	}
}
