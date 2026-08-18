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

// TestComputeYongShen_AJiaDun 甲年命遁六仪（1984-02-15 阳遁8局伏吟盘：戊天盘=地盘在宫8艮）
func TestComputeYongShen_AJiaDun(t *testing.T) {
	chart := chartForTest()
	sym, _ := ParseYongShen("开门")
	// 1984 甲子年 → 甲遁戊 → 戊落艮8（盘面独立锚定）
	ys := ComputeYongShenWithBirth(chart, []YongShenSymbol{sym}, 1984)
	if ys.NianGanPalace == nil {
		t.Fatal("1984甲子年命干甲应遁戊有落宫")
	}
	if *ys.NianGanPalace != GongGen {
		t.Errorf("1984甲子年命甲遁戊落宫 = %s(%d), want 艮(8)", *ys.NianGanPalace, *ys.NianGanPalace)
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

// TestYongShenSymbolAnchors 用神符号落宫数据驱动锚点（门/星/神/干四类）。
// 期望值由盘面九宫数据独立确定（用神落宫以天盘为核心判断依据，甲遁看六仪遁宫），
// 非用 findGanPalaceIdx 自证，避免实现错误被测试掩盖。
//
// 2000-06-15 午时（阳遁9局，非伏吟盘，能验证天盘/地盘取值）盘面：
//   宫1坎=天任/死/九天/天盘乙   宫2坤=天英/惊/值符/天盘戊
//   宫3震=天蓬/开/螣蛇/天盘己   宫4巽=天芮/休/太阴/天盘庚
//   宫5中=天冲/天盘辛           宫6乾=天辅/生/六合/天盘壬
//   宫7兑=天禽/伤/勾陈/天盘癸   宫8艮=天心/杜/朱雀/天盘丁
//   宫9离=天柱/景/九地/天盘丙
func TestYongShenSymbolAnchors(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ShiQiMen)
	cases := []struct {
		symbol string
		want   string // 期望落宫（乾/坤/…）
	}{
		// 门（2000-06-15）
		{"生门", "乾"}, // 生门在宫6乾
		{"开门", "震"}, // 开门在宫3震
		{"休门", "巽"}, // 休门在宫4巽
		{"死门", "坎"}, // 死门在宫1坎
		// 星
		{"天辅", "兑"}, // 天辅在宫7兑
		{"天芮", "巽"}, // 天芮在宫4巽
		{"天心", "艮"}, // 天心在宫8艮
		// 神
		{"六合", "乾"}, // 六合神在宫6乾
		{"值符", "坤"}, // 值符在宫2坤
		// 干（用神落宫看天盘，验证天盘优先）
		{"戊", "坤"}, // 戊天盘在宫2坤（地盘在宫9离）
		{"庚", "巽"}, // 庚天盘在宫4巽（地盘在宫2坤）
		{"乙", "坎"}, // 乙天盘在宫1坎（地盘在宫8艮）
	}
	for _, c := range cases {
		sym, err := ParseYongShen(c.symbol)
		if err != nil {
			t.Errorf("ParseYongShen(%q): %v", c.symbol, err)
			continue
		}
		ys := ComputeYongShen(chart, []YongShenSymbol{sym})
		if len(ys.Symbols) != 1 {
			t.Errorf("%s: 应有 1 个符号结果", c.symbol)
			continue
		}
		got := ys.Symbols[0].Palace.String()
		if got != c.want {
			t.Errorf("%s 落宫 = %s(%d), want %s（盘面独立锚定）", c.symbol, got, ys.Symbols[0].Palace, c.want)
		}
	}
}

// TestYongShenSymbolAnchor_JiaDun 用神干为甲时的锚点：甲辰日→甲辰遁壬（天盘壬在宫6乾）。
func TestYongShenSymbolAnchor_JiaDun(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	ch := ComputeChart(st, ShiQiMen)
	if ch.Pan.RiGan != ganzhi.GanJia {
		t.Fatalf("预期甲日，got %s", ch.Pan.RiGan)
	}
	// 甲辰遁壬：壬天盘在宫7兑（地盘壬在宫4巽）。用神落宫以天盘为核心。
	if ch.RiGanPalace != GongDui {
		t.Errorf("甲辰日甲遁壬落宫 = %s(%d), want 兑(7)", ch.RiGanPalace, ch.RiGanPalace)
	}
}

// TestJiaDunAllSix 六甲遁六仪全映射（领域规则：甲子遁戊/甲戌遁己/甲申遁庚/甲午遁辛/甲辰遁壬/甲寅遁癸）。
func TestJiaDunAllSix(t *testing.T) {
	cases := []struct {
		zhi  ganzhi.Zhi
		want ganzhi.Gan
	}{
		{ganzhi.ZhiZi, ganzhi.GanWu},   // 甲子→戊
		{ganzhi.ZhiXu, ganzhi.GanJi},   // 甲戌→己
		{ganzhi.ZhiShen, ganzhi.GanGeng}, // 甲申→庚
		{ganzhi.ZhiWu, ganzhi.GanXin},  // 甲午→辛
		{ganzhi.ZhiChen, ganzhi.GanRen}, // 甲辰→壬
		{ganzhi.ZhiYin, ganzhi.GanGui}, // 甲寅→癸
	}
	for _, c := range cases {
		got, ok := jiaDunLiuYi(c.zhi)
		if !ok {
			t.Errorf("jiaDunLiuYi(%s) 应命中", c.zhi)
			continue
		}
		if got != c.want {
			t.Errorf("jiaDunLiuYi(%s) = %s, want %s", c.zhi, got, c.want)
		}
	}
	// 非六甲支不遁
	if _, ok := jiaDunLiuYi(ganzhi.ZhiChou); ok {
		t.Error("非六甲支(丑)不应遁甲")
	}
}

// TestParseYongShen_SpiritYinYang 八神阴遁名（白虎/玄武）与阳遁名等价解析。
func TestParseYongShen_SpiritYinYang(t *testing.T) {
	cases := []struct {
		yang, yin string
	}{
		{"勾陈", "白虎"}, // 阳遁勾陈=阴遁白虎
		{"朱雀", "玄武"}, // 阳遁朱雀=阴遁玄武
	}
	for _, c := range cases {
		yangSym, err := ParseYongShen(c.yang)
		if err != nil {
			t.Errorf("ParseYongShen(%q): %v", c.yang, err)
			continue
		}
		yinSym, err := ParseYongShen(c.yin)
		if err != nil {
			t.Errorf("ParseYongShen(%q): %v", c.yin, err)
			continue
		}
		if yangSym.Spirit == nil || yinSym.Spirit == nil || *yangSym.Spirit != *yinSym.Spirit {
			t.Errorf("%q(%v) 与 %q(%v) 应为同一八神", c.yang, yangSym.Spirit, c.yin, yinSym.Spirit)
		}
	}
}

// TestYongShenSymbolStateAnchors 符号落宫完整状态锚点（Palace/TianGan/KongWang/MaXing）。
// 2000-06-15 午时（阳遁9局）：时柱庚午（甲子旬）空戌亥→乾；午属寅午戌三合，马在申→坤。
// 独立盘面锚定，非自证。
func TestYongShenSymbolStateAnchors(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ShiQiMen)
	cases := []struct {
		symbol    string
		palace    string
		tianGan   string
		kongWang  bool
		maXing    bool
	}{
		{"生门", "乾", "辛", true, false},  // 生门落乾6，天盘辛，乾空亡
		{"戊", "坤", "戊", false, true},    // 戊天盘落坤2，坤为马星
		{"庚", "巽", "庚", false, false},   // 庚天盘落巽4
		{"开门", "震", "己", false, false}, // 开门落震3，天盘己
		{"六合", "乾", "辛", true, false},  // 六合神落乾6（同宫）
	}
	for _, c := range cases {
		sym, err := ParseYongShen(c.symbol)
		if err != nil {
			t.Errorf("ParseYongShen(%q): %v", c.symbol, err)
			continue
		}
		ys := ComputeYongShen(chart, []YongShenSymbol{sym})
		if len(ys.Symbols) != 1 {
			t.Errorf("%s: 应有 1 个符号结果", c.symbol)
			continue
		}
		r := ys.Symbols[0]
		if r.Palace.String() != c.palace || r.TianGan != c.tianGan || r.KongWang != c.kongWang || r.MaXing != c.maXing {
			t.Errorf("%s 状态 = {宫%s 天盘%s 空亡%v 马星%v}, want {宫%s 天盘%s 空亡%v 马星%v}",
				c.symbol, r.Palace, r.TianGan, r.KongWang, r.MaXing,
				c.palace, c.tianGan, c.kongWang, c.maXing)
		}
	}
}

// TestYongShenCombination_Marriage 多符号组合（婚姻：六合神 + 庚 + 乙）。
func TestYongShenCombination_Marriage(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ShiQiMen)
	names := []string{"六合", "庚", "乙"}
	syms := make([]YongShenSymbol, 0, len(names))
	for _, n := range names {
		sym, _ := ParseYongShen(n)
		syms = append(syms, sym)
	}
	ys := ComputeYongShen(chart, syms)
	if len(ys.Symbols) != 3 {
		t.Fatalf("应有 3 个符号结果，got %d", len(ys.Symbols))
	}
	// 六合→乾6，庚→巽4，乙→坎1（天盘优先，独立盘面锚定）
	want := []struct{ symbol, palace string }{
		{"六合", "乾"},
		{"庚", "巽"},
		{"乙", "坎"},
	}
	for i, w := range want {
		if ys.Symbols[i].Symbol != w.symbol || ys.Symbols[i].Palace.String() != w.palace {
			t.Errorf("组合[%d] = %s@%s, want %s@%s",
				i, ys.Symbols[i].Symbol, ys.Symbols[i].Palace, w.symbol, w.palace)
		}
	}
}

// TestComputeYongShen_EmptyDuplicate 空符号组合与重复符号边界。
func TestComputeYongShen_EmptyDuplicate(t *testing.T) {
	chart := chartForTest()
	// 空符号组合 → 无符号结果（不 panic）
	ys := ComputeYongShen(chart, nil)
	if len(ys.Symbols) != 0 {
		t.Errorf("空符号组合应无结果，got %d", len(ys.Symbols))
	}
	// 重复符号 → 各符号独立成项
	sym, _ := ParseYongShen("开门")
	ys = ComputeYongShen(chart, []YongShenSymbol{sym, sym})
	if len(ys.Symbols) != 2 {
		t.Errorf("重复符号应 2 项，got %d", len(ys.Symbols))
	}
	if ys.Symbols[0].Palace != ys.Symbols[1].Palace {
		t.Error("重复同符号落宫应一致")
	}
}

// TestYongShenNianGanAnchor 年命干具体落宫锚定（非仅非空断言）。
func TestYongShenNianGanAnchor(t *testing.T) {
	chart := chartForTest() // 1984-02-15 阳遁8局伏吟盘
	sym, _ := ParseYongShen("开门")
	// 1985 乙丑年 → 年干乙 → 天盘乙落宫7兑（盘面独立锚定）
	ys := ComputeYongShenWithBirth(chart, []YongShenSymbol{sym}, 1985)
	if ys.NianGanPalace == nil {
		t.Fatal("有 birth_year 时年命干落宫不应为空")
	}
	if *ys.NianGanPalace != GongDui {
		t.Errorf("1985乙丑年命乙落宫 = %s(%d), want 兑(7)", *ys.NianGanPalace, *ys.NianGanPalace)
	}
}
