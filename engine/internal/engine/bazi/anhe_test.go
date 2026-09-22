package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"testing"
)

func TestAnHe(t *testing.T) {
	// 暗合对: 寅丑, 卯申, 午亥, 子戌
	tests := []struct {
		name  string
		zhi   [4]ganzhi.Zhi
		wantA string
		wantB string
	}{
		{"子戌暗合", [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiXu, ganzhi.ZhiChen}, "子", "戌"},
		{"寅丑暗合", [4]ganzhi.Zhi{ganzhi.ZhiYin, ganzhi.ZhiChou, ganzhi.ZhiZi, ganzhi.ZhiWu}, "寅", "丑"},
		{"卯申暗合", [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiShen, ganzhi.ZhiMao, ganzhi.ZhiWu}, "卯", "申"},
		{"午亥暗合", [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiWu, ganzhi.ZhiHai}, "午", "亥"},
		{"无暗合", [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiChen, ganzhi.ZhiSi}, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.zhi[0]},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi[1]},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: tt.zhi[2]},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: tt.zhi[3]},
			}
			hehui := ComputeHeHui(canonicalChart(bz, ganzhi.Male, 2000))

			if tt.wantA == "" {
				if len(hehui.AnHe) > 0 {
					t.Errorf("expected no an_he, got %d", len(hehui.AnHe))
				}
				return
			}
			found := false
			for _, ah := range hehui.AnHe {
				if (ah.ZhiA == tt.wantA && ah.ZhiB == tt.wantB) || (ah.ZhiA == tt.wantB && ah.ZhiB == tt.wantA) {
					found = true
				}
			}
			if !found {
				t.Errorf("暗合 %s%s not found in an_he (%d entries)", tt.wantA, tt.wantB, len(hehui.AnHe))
			}
		})
	}
}
