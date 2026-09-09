package liuyao

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestTombOfUsesElementSpecificBranch(t *testing.T) {
	tests := map[ganzhi.Wuxing]ganzhi.Zhi{
		ganzhi.WxMu:   ganzhi.ZhiWei,
		ganzhi.WxHuo:  ganzhi.ZhiXu,
		ganzhi.WxJin:  ganzhi.ZhiChou,
		ganzhi.WxShui: ganzhi.ZhiChen,
	}
	for element, want := range tests {
		if got := tombOf(element); got != want {
			t.Fatalf("tombOf(%v)=%v, want %v", element, got, want)
		}
	}
}

func TestLifeStageOfUsesElementAtBranch(t *testing.T) {
	tests := []struct {
		element ganzhi.Wuxing
		branch  ganzhi.Zhi
		want    string
	}{
		{ganzhi.WxMu, ganzhi.ZhiHai, "长生"},
		{ganzhi.WxMu, ganzhi.ZhiWei, "墓"},
		{ganzhi.WxHuo, ganzhi.ZhiYin, "长生"},
		{ganzhi.WxJin, ganzhi.ZhiSi, "长生"},
		{ganzhi.WxShui, ganzhi.ZhiShen, "长生"},
	}
	for _, tt := range tests {
		if got := lifeStageOf(tt.element, tt.branch); got != tt.want {
			t.Fatalf("lifeStageOf(%v,%v)=%q, want %q", tt.element, tt.branch, got, tt.want)
		}
	}
}
