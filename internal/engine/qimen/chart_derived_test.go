package qimen

import (
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// qimen.chart 确定性派生字段 — 命理锚点测试。
// 用与 golden 相同的输入（2026-06-28 12:00 CST 时家奇门），断言派生字段的已知值。

func goldenChart(t *testing.T) Chart {
	t.Helper()
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st, "时家")
}

func TestChartDerived_RiShiGanPalace(t *testing.T) {
	c := goldenChart(t)
	if c.RiGanPalace != GongKan {
		t.Errorf("ri_gan_gong = %d, want 坎(1)", c.RiGanPalace)
	}
	if c.ShiGanPalace != GongKun {
		t.Errorf("shi_gan_gong = %d, want 坤(2)", c.ShiGanPalace)
	}
}

func TestChartDerived_ShengKe(t *testing.T) {
	c := goldenChart(t)
	if c.RiShiShengKe == "" {
		t.Fatal("ri_shi_sheng_ke empty")
	}
	// 生克文案必须与宫五行一致：坤土 克 坎水 → "时干(2宫)克日干(1宫)"
	want := "时干(2宫)克日干(1宫)"
	if c.RiShiShengKe != want {
		t.Errorf("ri_shi_sheng_ke = %q, want %q", c.RiShiShengKe, want)
	}
	// 自洽校验：宫五行关系与文案方向一致（坤土克坎水）
	sp := palaceWuxing(c.ShiGanPalace)
	rp := palaceWuxing(c.RiGanPalace)
	if !ganzhi.Ke(sp, rp) {
		t.Errorf("宫五行不匹配：坤(%s) 应克 坎(%s)", sp, rp)
	}
	if !strings.Contains(c.RiShiShengKe, "克") {
		t.Errorf("ri_shi_sheng_ke 应含生克方向")
	}
}

func TestChartDerived_KongWangMaXingAffected(t *testing.T) {
	c := goldenChart(t)
	if !c.KongWangAffected {
		t.Error("kong_wang_affected = false, want true（golden 锚点）")
	}
	if !c.MaXingAffected {
		t.Error("ma_xing_affected = false, want true（golden 锚点）")
	}
	// 自洽校验：affected 必须与 pan 的空亡/马星数据一致
	ri, shi := c.RiGanPalace, c.ShiGanPalace
	kw := false
	for _, k := range c.Pan.KongWang {
		if k == ri || k == shi {
			kw = true
			break
		}
	}
	if kw != c.KongWangAffected {
		t.Errorf("kong_wang_affected(%v) 与 pan.KongWang 推导(%v) 不一致", c.KongWangAffected, kw)
	}
	ma := c.Pan.MaXing == ri || c.Pan.MaXing == shi
	if ma != c.MaXingAffected {
		t.Errorf("ma_xing_affected(%v) 与 pan.MaXing 推导(%v) 不一致", c.MaXingAffected, ma)
	}
}

func TestChartDerived_AllAffectedFalse(t *testing.T) {
	// 另一个时刻：空亡/马星未必同时命中日时干宫，验证标记会随盘变化（非恒真）
	st := tianwen.GregorianToSolar(
		time.Date(2026, 1, 1, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	c := ComputeChart(st, "时家")
	ri, shi := c.RiGanPalace, c.ShiGanPalace
	kw := false
	for _, k := range c.Pan.KongWang {
		if k == ri || k == shi {
			kw = true
			break
		}
	}
	if kw != c.KongWangAffected {
		t.Errorf("kong_wang_affected(%v) 与推导(%v) 不一致（盘=%d/%d）", c.KongWangAffected, kw, ri, shi)
	}
	ma := c.Pan.MaXing == ri || c.Pan.MaXing == shi
	if ma != c.MaXingAffected {
		t.Errorf("ma_xing_affected(%v) 与推导(%v) 不一致", c.MaXingAffected, ma)
	}
}
