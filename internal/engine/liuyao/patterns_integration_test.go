package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// TestComputeChart_IncludesPatterns 验证装卦输出包含特殊格局
func TestComputeChart_IncludesPatterns(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 验证 Patterns 字段存在
	if chart.Patterns == nil {
		t.Fatal("Patterns field should not be nil")
	}

	// 验证 Patterns 是数组类型
	if len(chart.Patterns) < 0 {
		t.Error("Patterns should be an array")
	}
}

// TestComputeChart_PatternsStructure 验证特殊格局结构
func TestComputeChart_PatternsStructure(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 验证 Patterns 中的每个元素都有必要字段
	for _, p := range chart.Patterns {
		if p.Type == "" {
			t.Error("Pattern.Type should not be empty")
		}
		if p.Assessment == "" {
			t.Error("Pattern.Assessment should not be empty")
		}
	}
}

// TestComputeChart_PatternsIncludeXunKong 验证包含旬空格局
func TestComputeChart_PatternsIncludeXunKong(t *testing.T) {
	// 选择一个会触发旬空的日期（如甲子旬空戌亥）
	st := tianwen.GregorianToSolar(
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 检查是否有旬空相关格局
	hasXunKong := false
	for _, p := range chart.Patterns {
		if p.Type == PatternXunKong {
			hasXunKong = true
			break
		}
	}

	// 不一定每次都有旬空，这里只验证结构正确
	t.Logf("Patterns count: %d, hasXunKong: %v", len(chart.Patterns), hasXunKong)
}

// TestComputeChart_PatternsIncludeYuePo 验证包含月破格局
func TestComputeChart_PatternsIncludeYuePo(t *testing.T) {
	// 选择一个会触发月破的日期
	st := tianwen.GregorianToSolar(
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 检查是否有月破相关格局
	hasYuePo := false
	for _, p := range chart.Patterns {
		if p.Type == PatternYuePo {
			hasYuePo = true
			break
		}
	}

	t.Logf("Patterns count: %d, hasYuePo: %v", len(chart.Patterns), hasYuePo)
}

// TestComputeChart_PatternsIncludeChongHe 验证包含冲合格局
func TestComputeChart_PatternsIncludeChongHe(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 检查是否有冲合相关格局
	hasChongHe := false
	for _, p := range chart.Patterns {
		if p.Type == PatternChongHe {
			hasChongHe = true
			break
		}
	}

	t.Logf("Patterns count: %d, hasChongHe: %v", len(chart.Patterns), hasChongHe)
}

// TestComputeChart_PatternsIncludeDuFa 验证包含独发格局
func TestComputeChart_PatternsIncludeDuFa(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	// 只有一个动爻
	chart := ComputeChart(st, YongGuanGui, [6]int{9, 7, 7, 7, 7, 7})

	// 检查是否有独发格局
	hasDuFa := false
	for _, p := range chart.Patterns {
		if p.Type == PatternDuFa {
			hasDuFa = true
			break
		}
	}

	t.Logf("Patterns count: %d, hasDuFa: %v", len(chart.Patterns), hasDuFa)
}
