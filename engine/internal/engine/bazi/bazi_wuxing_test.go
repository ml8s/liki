package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestWuxing_ElementThatGenerates(t *testing.T) {
	tests := []struct {
		name     string
		input    ganzhi.Wuxing
		expected ganzhi.Wuxing
	}{
		{"水生木", ganzhi.WxMu, ganzhi.WxShui},
		{"木生火", ganzhi.WxHuo, ganzhi.WxMu},
		{"火生土", ganzhi.WxTu, ganzhi.WxHuo},
		{"土生金", ganzhi.WxJin, ganzhi.WxTu},
		{"金生水", ganzhi.WxShui, ganzhi.WxJin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := elementThatGenerates(tt.input)
			if got != tt.expected {
				t.Errorf("elementThatGenerates(%s) = %s, want %s",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestWuxing_ElementThatControls(t *testing.T) {
	tests := []struct {
		name     string
		input    ganzhi.Wuxing
		expected ganzhi.Wuxing
	}{
		{"金克木", ganzhi.WxMu, ganzhi.WxJin},
		{"水克火", ganzhi.WxHuo, ganzhi.WxShui},
		{"木克土", ganzhi.WxTu, ganzhi.WxMu},
		{"火克金", ganzhi.WxJin, ganzhi.WxHuo},
		{"土克水", ganzhi.WxShui, ganzhi.WxTu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := elementThatControls(tt.input)
			if got != tt.expected {
				t.Errorf("elementThatControls(%s) = %s, want %s",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestWuxing_ElementThatDrains(t *testing.T) {
	tests := []struct {
		name     string
		input    ganzhi.Wuxing
		expected ganzhi.Wuxing
	}{
		{"木生火", ganzhi.WxMu, ganzhi.WxHuo},
		{"火生土", ganzhi.WxHuo, ganzhi.WxTu},
		{"土生金", ganzhi.WxTu, ganzhi.WxJin},
		{"金生水", ganzhi.WxJin, ganzhi.WxShui},
		{"水生木", ganzhi.WxShui, ganzhi.WxMu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := elementThatDrains(tt.input)
			if got != tt.expected {
				t.Errorf("elementThatDrains(%s) = %s, want %s",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestWuxing_ElementControlledBy(t *testing.T) {
	tests := []struct {
		name     string
		input    ganzhi.Wuxing
		expected ganzhi.Wuxing
	}{
		{"木克土", ganzhi.WxMu, ganzhi.WxTu},
		{"火克金", ganzhi.WxHuo, ganzhi.WxJin},
		{"土克水", ganzhi.WxTu, ganzhi.WxShui},
		{"金克木", ganzhi.WxJin, ganzhi.WxMu},
		{"水克火", ganzhi.WxShui, ganzhi.WxHuo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := elementControlledBy(tt.input)
			if got != tt.expected {
				t.Errorf("elementControlledBy(%s) = %s, want %s",
					tt.input, got, tt.expected)
			}
		})
	}
}

