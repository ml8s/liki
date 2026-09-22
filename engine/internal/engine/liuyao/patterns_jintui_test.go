package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestComputeJinTui(t *testing.T) {
	tests := []struct {
		name    string
		ben     ganzhi.Zhi
		bian    ganzhi.Zhi
		subType string
	}{
		{name: "亥化子进神", ben: ganzhi.ZhiHai, bian: ganzhi.ZhiZi, subType: "进神"},
		{name: "寅化卯进神", ben: ganzhi.ZhiYin, bian: ganzhi.ZhiMao, subType: "进神"},
		{name: "巳化午进神", ben: ganzhi.ZhiSi, bian: ganzhi.ZhiWu, subType: "进神"},
		{name: "申化酉进神", ben: ganzhi.ZhiShen, bian: ganzhi.ZhiYou, subType: "进神"},
		{name: "子化亥退神", ben: ganzhi.ZhiZi, bian: ganzhi.ZhiHai, subType: "退神"},
		{name: "卯化寅退神", ben: ganzhi.ZhiMao, bian: ganzhi.ZhiYin, subType: "退神"},
		{name: "午化巳退神", ben: ganzhi.ZhiWu, bian: ganzhi.ZhiSi, subType: "退神"},
		{name: "酉化申退神", ben: ganzhi.ZhiYou, bian: ganzhi.ZhiShen, subType: "退神"},
		{name: "子化辰不是进退", ben: ganzhi.ZhiZi, bian: ganzhi.ZhiChen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := &Chart{
				Lines: [6]Line{
					{Position: 1, Type: LaoYang, Zhi: tt.ben, LiuQin: QinQiCai},
				},
				BianLines: [6]Line{
					{Position: 1, Type: ShaoYang, Zhi: tt.bian},
				},
			}
			patterns := computeJinTui(chart, YongQiCai)
			if tt.subType == "" {
				if len(patterns) != 0 {
					t.Fatalf("patterns = %v, want none", patterns)
				}
				return
			}
			if len(patterns) != 1 {
				t.Fatalf("patterns count = %d, want 1", len(patterns))
			}
			got := patterns[0]
			if got.Type != PatternJinTui || got.SubType != tt.subType || !got.IsTrue {
				t.Fatalf("pattern = %+v, want %s with is_true=true", got, tt.subType)
			}
		})
	}
}

func TestTrueVacantAndBrokenSubtypes(t *testing.T) {
	chart := &Chart{
		XunKong: [2]ganzhi.Zhi{ganzhi.ZhiZi},
		Lines: [6]Line{
			{Position: 1, Type: ShaoYin, Zhi: ganzhi.ZhiZi, LiuQin: QinQiCai, YuePo: true},
		},
		WangShuai: [6]ganzhi.WangShuai{ganzhi.WSQiu},
	}

	vacant := computeXunKong(chart, YongQiCai)
	if len(vacant) != 1 || vacant[0].SubType != "真空" || !vacant[0].IsTrue {
		t.Fatalf("vacant = %+v, want 真空 with is_true=true", vacant)
	}

	broken := computeYuePo(chart, YongQiCai)
	if len(broken) != 1 || broken[0].SubType != "真破" || !broken[0].IsTrue {
		t.Fatalf("broken = %+v, want 真破 with is_true=true", broken)
	}
}
