package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func chartForTest() Chart {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st, ShiQiMen)
}

// TestParseYongShen 用神符号解析（门/星/神/干）
func TestParseYongShen(t *testing.T) {
	tests := []struct {
		in   string
		kind string // door/star/spirit/gan
	}{
		{"开门", "door"}, {"生门", "door"}, {"死门", "door"},
		{"天心", "star"}, {"天辅", "star"}, {"天芮", "star"},
		{"六合", "spirit"}, {"值符", "spirit"},
		{"戊", "gan"}, {"庚", "gan"}, {"乙", "gan"},
	}
	for _, tt := range tests {
		sym, err := ParseYongShen(tt.in)
		if err != nil {
			t.Errorf("ParseYongShen(%q): %v", tt.in, err)
			continue
		}
		switch tt.kind {
		case "door":
			if sym.Door == nil {
				t.Errorf("%q 应解析为门", tt.in)
			}
		case "star":
			if sym.Star == nil {
				t.Errorf("%q 应解析为星", tt.in)
			}
		case "spirit":
			if sym.Spirit == nil {
				t.Errorf("%q 应解析为神", tt.in)
			}
		case "gan":
			if sym.Stem == nil {
				t.Errorf("%q 应解析为干", tt.in)
			}
		}
	}
	// 未知报错
	if _, err := ParseYongShen("火星"); err == nil {
		t.Error("未知符号应报错")
	}
}

// TestComputeYongShen 求测人 + 符号组合聚合
func TestComputeYongShen(t *testing.T) {
	chart := chartForTest()
	// 求财：生门 + 戊
	sym1, _ := ParseYongShen("生门")
	sym2, _ := ParseYongShen("戊")
	ys := ComputeYongShen(chart, []YongShenSymbol{sym1, sym2})

	// 求测人定位由排盘固有字段提供（顶层）
	if chart.RiGanPalace == 0 {
		t.Error("日干落宫应为非零")
	}
	if chart.ShiGanPalace == 0 {
		t.Error("时干落宫应为非零")
	}
	// 符号组合
	if len(ys.Symbols) != 2 {
		t.Fatalf("应有 2 个符号结果，got %d", len(ys.Symbols))
	}
	// 生门落宫
	if ys.Symbols[0].Symbol != "生门" || ys.Symbols[0].Palace == 0 {
		t.Errorf("生门落宫应>0，got %+v", ys.Symbols[0])
	}
	// 戊干落宫
	if ys.Symbols[1].Symbol != "戊" || ys.Symbols[1].Palace == 0 {
		t.Errorf("戊干落宫应>0，got %+v", ys.Symbols[1])
	}
	// 落宫天盘干
	if ys.Symbols[0].TianGan == "" {
		t.Error("落宫天盘干不应为空")
	}
	// 日时生克（顶层）
	if chart.RiShiShengKe == "" {
		t.Error("日时生克不应为空")
	}
}

// TestRiGanJiaDun 日干为甲时按日支遁六仪
func TestRiGanJiaDun(t *testing.T) {
	// 2000-06-15 甲辰日 → 甲辰遁壬 → 壬落宫
	st := tianwen.GregorianToSolar(
		time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	ch := ComputeChart(st, ShiQiMen)
	if ch.Pan.RiGan != ganzhi.GanJia {
		t.Fatalf("预期甲日，got %s", ch.Pan.RiGan)
	}
	if ch.RiGanPalace == 0 {
		t.Fatal("甲辰日应遁壬落宫，而非 0")
	}
	want := findGanPalaceIdx(ch.Pan, ganzhi.GanRen)
	if want > 0 && ch.RiGanPalace != want {
		t.Errorf("甲辰遁壬落宫 = %d, want %d", ch.RiGanPalace, want)
	}
}

// TestComputeYongShen_NianGan 年命干聚合（需 birth_year）
func TestComputeYongShen_NianGan(t *testing.T) {
	chart := chartForTest()
	sym, _ := ParseYongShen("开门")
	// 1985 乙丑年，年干=乙
	ys := ComputeYongShenWithBirth(chart, []YongShenSymbol{sym}, 1985)
	if ys.NianGanPalace == nil {
		t.Fatal("有 birth_year 时年命干落宫不应为空")
	}
	// 无 birth_year → 年命干为空
	ys2 := ComputeYongShen(chart, []YongShenSymbol{sym})
	if ys2.NianGanPalace != nil {
		t.Error("无 birth_year 时年命干落宫应为空")
	}
}

// TestComputeYongShen_AJiaDun 甲年命遁六仪
func TestComputeYongShen_AJiaDun(t *testing.T) {
	chart := chartForTest()
	sym, _ := ParseYongShen("开门")
	// 1984 甲子年 → 甲遁戊 → 戊落宫
	ys := ComputeYongShenWithBirth(chart, []YongShenSymbol{sym}, 1984)
	if ys.NianGanPalace == nil {
		t.Fatal("1984甲子年命干甲应遁戊有落宫")
	}
	want := findGanPalaceIdx(chart.Pan, ganzhi.GanWu)
	if want > 0 && *ys.NianGanPalace != want {
		t.Errorf("甲遁戊落宫 = %d, want %d", *ys.NianGanPalace, want)
	}
}

// TestComputeYongShen_KongWangMaXing 符号落宫空亡/马星
func TestComputeYongShen_KongWangMaXing(t *testing.T) {
	chart := chartForTest()
	sym, _ := ParseYongShen("开门")
	ys := ComputeYongShen(chart, []YongShenSymbol{sym})
	if len(ys.Symbols) != 1 || ys.Symbols[0].Palace == 0 {
		t.Fatal("开门落宫应>0")
	}
	// 与盘一致
	for _, k := range chart.Pan.KongWang {
		if k == ys.Symbols[0].Palace && !ys.Symbols[0].KongWang {
			t.Errorf("落宫%d为空亡宫，kong_wang应为true", ys.Symbols[0].Palace)
		}
	}
	if ys.Symbols[0].Palace == chart.Pan.MaXing && !ys.Symbols[0].MaXing {
		t.Errorf("落宫%d为马星宫，ma_xing应为true", ys.Symbols[0].Palace)
	}
}
