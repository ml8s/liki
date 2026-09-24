package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 格局格名（shiShenToPatternName）固化测试——入表前锁定当前命理输出。
// 《子平真诠》：月令格神定格局——官杀财印食伤各成一格。

func TestShiShenToPatternName(t *testing.T) {
	cases := []struct {
		shishen ganzhi.ShiShen
		want    string
	}{
		{ganzhi.ShiShenZhengGuan, "正官格"},
		{ganzhi.ShiShenQiSha, "七杀格"},
		{ganzhi.ShiShenZhengCai, "正财格"},
		{ganzhi.ShiShenPianCai, "偏财格"},
		{ganzhi.ShiShenZhengYin, "正印格"},
		{ganzhi.ShiShenPianYin, "偏印格"},
		{ganzhi.ShiShenShiShen, "食神格"},
		{ganzhi.ShiShenShangGuan, "伤官格"},
		{ganzhi.ShiShenBiJian, "杂格"},
		{ganzhi.ShiShenJieCai, "杂格"},
	}
	for _, tc := range cases {
		got := shiShenToPatternName(tc.shishen)
		if got != tc.want {
			t.Errorf("十神 %d → 格名 %s, want %s", tc.shishen, got, tc.want)
		}
	}
}