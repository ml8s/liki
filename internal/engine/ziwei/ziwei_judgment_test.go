package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestZiWeiJudgment_KnownPatterns(t *testing.T) {
	tests := []struct {
		name      string
		time      string
		gender    ganzhi.Gender
		wantPatterns int  // minimum patterns found
		wantRating string
	}{
		{
			// 1990-06-15 10:00 北京, 男
			// 已知: 紫微在命宫(紫微朝垣)者比例较高
			// 具体格局取决于星曜分布
			name: "1990年北京男",
			time: "1990-06-15T10:00:00+08:00",
			gender: ganzhi.Male,
			wantPatterns: 0, // just verify no crash
			wantRating: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.FixedZone("CST", 8*3600)
			bt, err := time.ParseInLocation("2006-01-02T15:04:05Z07:00", tt.time, loc)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, tt.gender)
			result := ComputeJudgment(chart)
			if tt.wantPatterns > 0 && len(result.Patterns) < tt.wantPatterns {
				t.Errorf("patterns=%d, want >=%d", len(result.Patterns), tt.wantPatterns)
			}
			if tt.wantRating != "" && result.Rating != tt.wantRating {
				t.Errorf("rating=%q, want %q", result.Rating, tt.wantRating)
			}
			t.Logf("patterns=%d, rating=%q, sihua=%v", len(result.Patterns), result.Rating, result.SiHua)
			for _, p := range result.Patterns {
				t.Logf("  %s(%d): %s", p.Name, p.Score, p.Description)
			}
		})
	}
}
