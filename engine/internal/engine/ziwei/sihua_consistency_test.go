package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 四化一致性：主星和辅星在宫位明细与汇总 chart.SiHua 中的标注须一致。
func TestSiHua_Consistency_StarLevel(t *testing.T) {
	// 辛年：化科=文曲、化忌=文昌（四化表）——宫位里的文曲/文昌必须带科/忌。
	st := tianwen.GregorianToSolar(
		time.Date(1991, 3, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8) // 立春后=辛年
	lt := tianwen.SolarToLunar(tianwen.GregorianTime(st.Time()))
	chart := ComputeChart(lt, ganzhi.Male)
	siHua := chart.SiHua

	// 汇总含文曲化科/文昌化忌
	ke, okKe := siHua[WenQu]
	ji, okJi := siHua[WenChang]
	if !okKe || ke != HuaKe {
		t.Fatalf("辛年汇总应含文曲化科, got %v", siHua)
	}
	if !okJi || ji != HuaJi {
		t.Fatalf("辛年汇总应含文昌化忌, got %v", siHua)
	}

	// 宫位星曜须与汇总一致：凡宫位出现文曲→带科；文昌→带忌
	for i, g := range chart.GongWei {
		for _, s := range g.Stars {
			want := ""
			if s.Star == WenQu {
				want = string(HuaKe)
			}
			if s.Star == WenChang {
				want = string(HuaJi)
			}
			if want == "" {
				continue
			}
			if s.SiHua != want {
				t.Errorf("宫%d %s 四化 = %q, want %q（与汇总一致）", i+1, s.Name, s.SiHua, want)
			}
		}
	}
}
