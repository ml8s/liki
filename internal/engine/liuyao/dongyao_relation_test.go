package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// TestDongYaoRelations_Array 验证动爻关系是数组
func TestDongYaoRelations_Array(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	// 有动爻的卦（老阳/老阴）
	chart := ComputeChart(st, YongGuanGui, [6]int{9, 7, 7, 7, 7, 7})

	if chart.DongYaoRelations == nil {
		t.Fatal("DongYaoRelations should be an array (not nil)")
	}

	// 验证每个元素有 position + relation
	for _, r := range chart.DongYaoRelations {
		if r.Position < 1 || r.Position > 6 {
			t.Errorf("position %d out of range", r.Position)
		}
		if r.Relation == "" {
			t.Error("relation should not be empty")
		}
		// 验证 relation 是枚举
		valid := false
		for _, v := range []DongYaoRelationType{
			RelationShengYong, RelationKeYong, RelationBiHe, RelationChongYong,
			RelationShengYuan, RelationKeYuan, RelationShengJi, RelationKeJi,
		} {
			if r.Relation == v {
				valid = true
				break
			}
		}
		if !valid {
			t.Errorf("invalid relation type: %s", r.Relation)
		}
	}
}

// TestDongYaoRelations_None 验证静卦无动爻关系
func TestDongYaoRelations_None(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	// 纯少阳/少阴，无动爻
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	// 无动爻时，数组应为空
	if len(chart.DongYaoRelations) != 0 {
		t.Errorf("静卦 DongYaoRelations should be empty, got %d", len(chart.DongYaoRelations))
	}
}

// TestYongShen_AggregatedFields 验证用神聚合字段
func TestYongShen_AggregatedFields(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

	if chart.YongShen.Position == 0 {
		t.Fatal("用神应存在")
	}
	// 用神旺衰应已聚合
	if chart.YongShen.WangShuai == "" {
		t.Error("yong_shen.wang_shuai should be aggregated")
	}
	// 用神月破/旬空/入墓应存在
	if chart.YongShen.YuePo != chart.Lines[chart.YongShen.Position-1].YuePo {
		t.Error("yong_shen.yue_po should match line yue_po")
	}
	if chart.YongShen.XunKong != chart.Lines[chart.YongShen.Position-1].XunKong {
		t.Error("yong_shen.xun_kong should match line xun_kong")
	}
	if chart.YongShen.MuKu != chart.Lines[chart.YongShen.Position-1].MuKu {
		t.Error("yong_shen.mu_ku should match line mu_ku")
	}
}

// TestYongShen_InvalidPosition 用神不现时聚合字段为空
func TestYongShen_InvalidPosition(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	// 用神不现（卦中无此六亲）
	chart := ComputeChart(st, YongQiCai, [6]int{7, 7, 7, 7, 7, 7})

	if chart.YongShen.Position == 0 {
		// 用神不现时，聚合字段应为空，动爻关系应为空
		if chart.YongShen.WangShuai != "" {
			t.Error("用神不现时 wang_shuai 应为空")
		}
		if len(chart.DongYaoRelations) != 0 {
			t.Error("用神不现时 DongYaoRelations 应为空")
		}
	}
}

// TestDongYaoRelation_EnumTypes 验证 9 种枚举类型都存在
func TestDongYaoRelation_EnumTypes(t *testing.T) {
	types := []DongYaoRelationType{
		RelationShengYong, RelationKeYong, RelationBiHe, RelationChongYong,
		RelationShengYuan, RelationKeYuan, RelationShengJi, RelationKeJi, RelationNone,
	}
	expected := []string{
		"生用", "克用", "比和", "冲用",
		"生原神", "克原神", "生忌神", "克忌神", "无动爻",
	}
	for i, tt := range types {
		if string(tt) != expected[i] {
			t.Errorf("enum %d = %s, want %s", i, tt, expected[i])
		}
	}
}
