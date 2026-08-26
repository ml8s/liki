package qiming

import (
	"strings"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// =============================================================================
// 五格笔画计算
// =============================================================================

func TestComputeWuGeFromStrokes_LiMingHui(t *testing.T) {
	// 李(7) + 明(8) + 辉(15)
	wg := computeWuGeFromStrokes(singleStrokes(7), 8, 15)

	// 天格=8, 人格=15, 地格=23, 总格=30, 外格=30-15+1=16
	if wg.TianGe.Stroke != 8 {
		t.Errorf("天格 = %d, want 8", wg.TianGe.Stroke)
	}
	if wg.RenGe.Stroke != 15 {
		t.Errorf("人格 = %d, want 15", wg.RenGe.Stroke)
	}
	if wg.DiGe.Stroke != 23 {
		t.Errorf("地格 = %d, want 23", wg.DiGe.Stroke)
	}
	if wg.ZongGe.Stroke != 30 {
		t.Errorf("总格 = %d, want 30", wg.ZongGe.Stroke)
	}
	if wg.WaiGe.Stroke != 16 {
		t.Errorf("外格 = %d, want 16", wg.WaiGe.Stroke)
	}
}

func TestComputeWuGeFromStrokes_EdgeCases(t *testing.T) {
	// Large strokes: all >81 should wrap
	wg := computeWuGeFromStrokes(singleStrokes(50), 40, 30)
	// 天格=51, 人格=90→9, 地格=70, 总格=120→39, 外格=39-9+1=31
	if wg.RenGe.Stroke != 9 {
		t.Errorf("人格(90) = %d, want 9 (wrap)", wg.RenGe.Stroke)
	}
	if wg.ZongGe.Stroke != 39 {
		t.Errorf("总格(120) = %d, want 39 (wrap)", wg.ZongGe.Stroke)
	}
}

// =============================================================================
// 五格五行 — 尾数决定
// =============================================================================

func TestStrokeResult_KnownValues(t *testing.T) {
	// Per sanCaiNums, verify known stroke→element→fortune for common values
	tests := []struct {
		stroke  int
		element string
	}{
		{1, "木"},  // 1=木
		{5, "土"},  // 5=土
		{8, "金"},  // 8=金
		{13, "火"}, // 13=火
		{21, "木"}, // 21=木
		{24, "火"}, // 24=火
		{31, "木"}, // 31=木
		{37, "金"}, // 37=金
		{45, "土"}, // 45=土
	}

	for _, tt := range tests {
		r := strokeResult(tt.stroke)
		if r.Element != tt.element {
			t.Errorf("stroke %d: element=%s, want %s", tt.stroke, r.Element, tt.element)
		}
	}
}

// =============================================================================
// 三才配置
// =============================================================================

func TestComputeSanCai_KnownConfigs(t *testing.T) {
	tests := []struct {
		config string // e.g. "木木木"
	}{
		{"木木木"}, {"木木火"}, {"木木土"}, {"木木金"}, {"木木水"},
		{"木火木"}, {"木火火"}, {"木火土"}, {"木火金"}, {"木火水"},
		{"木土木"}, {"木土火"}, {"木土土"}, {"木土金"}, {"木土水"},
		{"木金木"}, {"木金火"}, {"木金土"}, {"木金金"}, {"木金水"},
		{"木水木"}, {"木水火"}, {"木水土"}, {"木水金"}, {"木水水"},
		{"火木木"}, {"火木火"}, {"火木土"},
	}

	for _, tt := range tests {
		runes := []rune(tt.config)
		sc := computeSanCai(string(runes[0:1]), string(runes[1:2]), string(runes[2:3]))
		if sc.Configuration != tt.config {
			t.Errorf("config=%s, got %s", tt.config, sc.Configuration)
		}
		if sc.Fortune == "" {
			t.Errorf("config=%s: empty fortune", tt.config)
		}
	}
}

func TestComputeSanCai_UnknownConfig(t *testing.T) {
	// 使用不存在的三才组合 → 默认半吉
	sc := computeSanCai("水", "水", "水")
	if sc.Configuration != "水水水" {
		t.Errorf("config = %s, want 水水水", sc.Configuration)
	}
	if sc.Fortune != "半吉" && sc.Fortune == "" {
		t.Errorf("unknown config fortune = %s, want non-empty", sc.Fortune)
	}
}

// =============================================================================
// 吉凶判断
// =============================================================================

func TestIsAuspicious(t *testing.T) {
	if !isAuspicious("吉") {
		t.Error("吉 should be auspicious")
	}
	if !isAuspicious("大吉") {
		t.Error("大吉 should be auspicious")
	}
	if isAuspicious("凶") {
		t.Error("凶 should NOT be auspicious")
	}
	if isAuspicious("半吉") {
		t.Error("半吉 should NOT be auspicious")
	}
	if isAuspicious("") {
		t.Error("empty should NOT be auspicious")
	}
}

// =============================================================================
// 音韵分析 — analyzePhonetic
// =============================================================================

func TestAnalyzePhonetic(t *testing.T) {
	// Empty chars
	phon := analyzePhonetic(nil)
	if phon.Tones != "" {
		t.Errorf("empty tones = %q, want empty", phon.Tones)
	}

	// Single char
	phon2 := analyzePhonetic([]Character{{Char: "明", Tone: 2}})
	if phon2.Tones != "2" {
		t.Errorf("single tones = %q, want 2", phon2.Tones)
	}

	// Two chars
	phon3 := analyzePhonetic([]Character{
		{Char: "明", Tone: 2},
		{Char: "亮", Tone: 4},
	})
	if phon3.Tones != "2-4" {
		t.Errorf("two-char tones = %q, want 2-4", phon3.Tones)
	}
}

// =============================================================================
// 汉字笔画查询
// =============================================================================

func TestLookupKangxiStroke_KnownChars(t *testing.T) {
	// Verify commonly used surname characters are in the database
	tests := []struct {
		char   string
		expect bool // expect >0 strokes (in DB)
	}{
		{"王", true},
		{"李", true},
		{"张", true},
		{"明", true},
		{"文", true},
		{"xyz", false}, // non-existent
		{"𠀀", false},   // very rare (probably not in DB)
	}

	for _, tt := range tests {
		strokes := lookupKangxiStroke(tt.char)
		if tt.expect && strokes == 0 {
			t.Errorf("lookupKangxiStroke(%q) = 0, want >0 (expected in DB)", tt.char)
		}
		if !tt.expect && strokes != 0 {
			t.Errorf("lookupKangxiStroke(%q) = %d, want 0 (not in DB)", tt.char, strokes)
		}
	}
}

// =============================================================================
// SurnameStroke
// =============================================================================

func TestSurnameStroke_KnownSurnames(t *testing.T) {
	tests := []struct {
		surname string
		stroke  int
	}{
		{"王", 4},
		{"李", 7},
		{"张", 7},
	}
	for _, tt := range tests {
		got, err := SurnameStroke(tt.surname)
		if err != nil {
			t.Errorf("SurnameStroke(%q): %v", tt.surname, err)
			continue
		}
		if got != tt.stroke {
			t.Errorf("SurnameStroke(%q) = %d, want %d", tt.surname, got, tt.stroke)
		}
	}
}

func TestSurnameStroke_NotFound(t *testing.T) {
	_, err := SurnameStroke("𠀀")
	if err == nil {
		t.Error("expected error for unknown surname")
	}
}

// =============================================================================
// 五行字符分组 — getCharsByElement
// =============================================================================

func TestGetCharsByElement_Structure(t *testing.T) {
	for _, elem := range []string{"木", "火", "土", "金", "水"} {
		wx := wuxingFromChinese(elem)
		chars := getCharsByElement(wx)
		if len(chars) == 0 {
			t.Errorf("getCharsByElement(%s): empty result", elem)
		}
		// Each char should be sorted within its stroke group
		for stroke, group := range chars {
			if stroke < 1 || stroke > 50 {
				t.Errorf("unexpected stroke %d in group", stroke)
			}
			for i := 1; i < len(group); i++ {
				if group[i-1].Char >= group[i].Char {
					t.Errorf("group %d not sorted at index %d: %q >= %q",
						stroke, i, group[i-1].Char, group[i].Char)
				}
			}
		}
	}
}

// =============================================================================
// 部首→五行推断
// =============================================================================

func TestInferElementFromRadical_DirectMatch(t *testing.T) {
	tests := []struct {
		radical string
		want    string
	}{
		{"木", "木"}, {"火", "火"}, {"土", "土"}, {"金", "金"}, {"水", "水"},
		{"艹", "木"}, {"氵", "水"}, {"忄", "火"}, {"钅", "金"}, {"山", "土"},
		{"石", "土"}, {"日", "火"}, {"心", "火"}, {"戈", "金"}, {"玉", "土"},
	}

	for _, tt := range tests {
		elem, ok := inferElementFromRadical(tt.radical)
		if !ok {
			t.Errorf("radical %q: not found", tt.radical)
			continue
		}
		if elem.String() != tt.want {
			t.Errorf("radical %q: element=%s, want %s", tt.radical, elem.String(), tt.want)
		}
	}
}

func TestInferElementFromRadical_UnknownRadical(t *testing.T) {
	// Unknown radicals are not resolved — no fallback scanning.
	_, ok := inferElementFromRadical("unknown")
	if ok {
		t.Error("unknown radical should not resolve to an element")
	}
}

// =============================================================================
// ComposeNames
// =============================================================================

// =============================================================================
// elementYAMLToChinese — 默认路径
// =============================================================================

func TestElementYAMLToChinese_Default(t *testing.T) {
	// Unknown element values returned as-is.
	if got := elementYAMLToChinese("unknown"); got != "unknown" {
		t.Errorf("elementYAMLToChinese(unknown) = %s, want unknown", got)
	}
}

// =============================================================================
// fortuneYAMLToChinese — 默认路径
// =============================================================================

func TestFortuneYAMLToChinese_Default(t *testing.T) {
	// Unknown fortune values returned as-is.
	if got := fortuneYAMLToChinese("unknown"); got != "unknown" {
		t.Errorf("fortuneYAMLToChinese(unknown) = %s, want unknown", got)
	}
}

// =============================================================================
// strokeResult — stroke >81 包装验证
// =============================================================================

func TestStrokeResult_WrapAbove81(t *testing.T) {
	// stroke >81 wraps: (stroke-1)%81+1. All 1-81 are in sanCaiNums.
	// 1000 → (1000-1)%81+1 = 999%81+1 = 27+1 = 28
	// 28 = metal(金), xiong(凶)
	r := strokeResult(1000)
	if r.Stroke != 28 {
		t.Errorf("stroke 1000 wraps to stroke=%d, want 28", r.Stroke)
	}
	if r.Element != "金" {
		t.Errorf("stroke 1000 wrap element = %s, want 金", r.Element)
	}
	if r.Fortune != "凶" {
		t.Errorf("stroke 1000 wrap fortune = %s, want 凶", r.Fortune)
	}
}

// =============================================================================
// EvaluateNames — 批量两字名评估（仅名字部分，不含姓）
// =============================================================================

func TestEvaluateNames_BatchTwoCharNames(t *testing.T) {
	// 佳桐、若薇 — given names only
	results, err := EvaluateNames("沙", []string{"佳桐", "若薇"}, "", nil, nil, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateNames: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First name: 沙佳桐
	r0 := results[0]
	if r0.Surname != "沙" {
		t.Errorf("result[0] surname = %q, want 沙", r0.Surname)
	}
	if r0.GivenName != "佳桐" {
		t.Errorf("result[0] givenName = %q, want 佳桐", r0.GivenName)
	}
	if len(r0.Characters) != 2 {
		t.Errorf("result[0] chars count = %d, want 2", len(r0.Characters))
	}
	if r0.Name != "沙佳桐" {
		t.Errorf("result[0] Name = %q, want 沙佳桐", r0.Name)
	}

	// Second name: 沙若薇
	r1 := results[1]
	if r1.Surname != "沙" {
		t.Errorf("result[1] surname = %q, want 沙", r1.Surname)
	}
	if r1.GivenName != "若薇" {
		t.Errorf("result[1] givenName = %q, want 若薇", r1.GivenName)
	}
	if len(r1.Characters) != 2 {
		t.Errorf("result[1] chars count = %d, want 2", len(r1.Characters))
	}
	if r1.Name != "沙若薇" {
		t.Errorf("result[1] Name = %q, want 沙若薇", r1.Name)
	}
}

func TestEvaluateNames_SingleFullName(t *testing.T) {
	// Single full name: 王明辉
	results, err := EvaluateNames("王", []string{"明辉"}, "", nil, nil, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateNames: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].GivenName != "明辉" {
		t.Errorf("givenName = %q, want 明辉", results[0].GivenName)
	}
	if len(results[0].Characters) != 2 {
		t.Errorf("chars count = %d, want 2", len(results[0].Characters))
	}
	if results[0].Name != "王明辉" {
		t.Errorf("Name = %q, want 王明辉", results[0].Name)
	}
}

func TestEvaluateNames_WithWuxing(t *testing.T) {
	results, err := EvaluateNames("王", []string{"明辉"}, "火", []string{"木"}, nil, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateNames: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Wuxing == nil {
		t.Fatal("Wuxing should be set when yong/xi/ji provided")
	}
	if results[0].Wuxing.Yong != results[0].WuxingMatch {
		t.Errorf("Wuxing.Yong (%v) should match WuxingMatch (%v)", results[0].Wuxing.Yong, results[0].WuxingMatch)
	}
	if results[0].Wuxing.Ji {
		t.Error("Wuxing.Ji should be false when jiShen is nil")
	}
}

// =============================================================================
// ComposeNames — 实际字符数据
// =============================================================================

// =============================================================================
// ComposeNames — xiChars 含字符时 yong+xi 路径
// =============================================================================

// =============================================================================
// ComposeNames — pairs 约束
// =============================================================================

// =============================================================================
// BUG-9 regression: radical element corrections
// =============================================================================

func TestRadicalToElement_SilkRadical(t *testing.T) {
	// 纟(糸部) → 金 per Kangxi dictionary.
	elem, ok := radicalToElement["纟"]
	if !ok {
		t.Fatal("纟 radical not found in radicalToElement")
	}
	if elem.String() != "金" {
		t.Errorf("纟 element = %s, want 金", elem.String())
	}
}

func TestRadicalToElement_MeatRadical(t *testing.T) {
	// 肉/⺼(肉部) → 土 per Kangxi dictionary.
	for _, r := range []string{"肉", "⺼"} {
		elem, ok := radicalToElement[r]
		if !ok {
			t.Errorf("%q radical not found in radicalToElement", r)
			continue
		}
		if elem.String() != "土" {
			t.Errorf("%q element = %s, want 土", r, elem.String())
		}
	}
}

func TestRadicalToElement_MoonRadical(t *testing.T) {
	// 月(月部) → 水 per Kangxi dictionary.
	elem, ok := radicalToElement["月"]
	if !ok {
		t.Fatal("月 radical not found in radicalToElement")
	}
	if elem.String() != "水" {
		t.Errorf("月 element = %s, want 水", elem.String())
	}
}

// =============================================================================
// BUG-8 regression: inferElementFromRadical no char-scanning fallback
// =============================================================================

func TestInferElementFromRadical_NoCharFallback(t *testing.T) {
	// Even if the char contains a known radical component, the fallback
	// should NOT scan the char string — only direct radical match.
	tests := []string{"明", "沐", "灶", "针"}
	for _, char := range tests {
		_, ok := inferElementFromRadical(char)
		if ok {
			t.Errorf("inferElementFromRadical(%q) should not resolve via char scan", char)
		}
	}
}

// =============================================================================
// BUG-7 regression: negative chars filtered in GetChars
// =============================================================================

func TestGetChars_NegativeCharFiltered(t *testing.T) {
	// Verify that characters listed in negative_chars.txt are filtered by GetChars.
	// "死" and "亡" are in negative_chars.txt — they must not appear in any wuxing pool.
	negativeChars["死"] = true
	negativeChars["亡"] = true
	t.Cleanup(func() {
		delete(negativeChars, "死")
		delete(negativeChars, "亡")
	})

	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		chars, err := GetChars(wx)
		if err != nil {
			t.Fatalf("GetChars(%q): %v", wx, err)
		}
		for stroke, group := range chars {
			for _, c := range group {
				if negativeChars[c.Char] {
					t.Errorf("GetChars(%q): stroke=%d contains negative char %q", wx, stroke, c.Char)
				}
			}
		}
	}
}

// =============================================================================
func TestComputeWuGeFromStrokes(t *testing.T) {
	// 王(4) + 1st名(9) + 2nd名(16) → standard example
	wg := computeWuGeFromStrokes(singleStrokes(4), 9, 16)

	if wg.TianGe.Stroke != 5 {
		t.Errorf("天格 stroke = %d, want 5", wg.TianGe.Stroke)
	}
	if wg.RenGe.Stroke != 13 {
		t.Errorf("人格 stroke = %d, want 13", wg.RenGe.Stroke)
	}
	if wg.DiGe.Stroke != 25 {
		t.Errorf("地格 stroke = %d, want 25", wg.DiGe.Stroke)
	}
	if wg.ZongGe.Stroke != 29 {
		t.Errorf("总格 stroke = %d, want 29", wg.ZongGe.Stroke)
	}
	// 外格 = 总格 - 人格 + 1 = 29 - 13 + 1 = 17
	if wg.WaiGe.Stroke != 17 {
		t.Errorf("外格 stroke = %d, want 17", wg.WaiGe.Stroke)
	}
}

// TestComputeWuGeFromStrokes_Minimum strokes verifies boundary case.
func TestComputeWuGeFromStrokes_MinStrokes(t *testing.T) {
	wg := computeWuGeFromStrokes(singleStrokes(1), 1, 1)

	// 天格=2, 人格=2, 地格=2, 总格=3, 外格=总格-人格+1=3-2+1=2
	if wg.TianGe.Stroke != 2 {
		t.Errorf("天格 = %d, want 2", wg.TianGe.Stroke)
	}
	if wg.WaiGe.Stroke != 2 {
		t.Errorf("外格 = %d, want 2", wg.WaiGe.Stroke)
	}
}

// TestWuGeElements verifies five-element attribution for known stroke counts.
func TestWuGeElements(t *testing.T) {
	// According to 81-number五格, the last digit determines element:
	// 1,2=木 3,4=火 5,6=土 7,8=金 9,0=水
	tests := []struct {
		stroke     int
		wantWuxing string
	}{
		{1, "木"}, {2, "木"},
		{3, "火"}, {4, "火"},
		{5, "土"}, {6, "土"},
		{7, "金"}, {8, "金"},
		{9, "水"}, {10, "水"},
		{11, "木"}, {21, "木"},
		{13, "火"}, {24, "火"},
	}

	for _, tt := range tests {
		got := strokeResult(tt.stroke)
		if got.Element != tt.wantWuxing {
			t.Errorf("strokeResult(%d).Element = %s, want %s",
				tt.stroke, got.Element, tt.wantWuxing)
		}
	}
}

// TestStrokeOverflow verifies stroke count wrapping above 81 and edge values.
func TestStrokeOverflow(t *testing.T) {
	r0 := strokeResult(0)
	if r0.Stroke != 1 {
		t.Errorf("stroke 0 should become 1, got %d", r0.Stroke)
	}

	r81 := strokeResult(81)
	if r81.Stroke != 81 {
		t.Errorf("stroke 81 should stay 81, got %d", r81.Stroke)
	}

	r1 := strokeResult(1)
	r82 := strokeResult(82)
	if r1.Element != r82.Element {
		t.Errorf("stroke 82 should wrap to 1: element %s != %s", r82.Element, r1.Element)
	}
	if r82.Stroke != 1 {
		t.Errorf("stroke 82 should wrap to stroke 1, got %d", r82.Stroke)
	}

	r162 := strokeResult(162)
	if r162.Stroke != 81 {
		t.Errorf("stroke 162 → %d, want 81", r162.Stroke)
	}
}

// TestSanCai verifies三才 configuration.
func TestSanCai(t *testing.T) {
	// 木木木 → 大吉 (all wood, mutual support)
	sc := computeSanCai("木", "木", "木")
	if sc.Configuration != "木木木" {
		t.Errorf("configuration = %s, want 木木木", sc.Configuration)
	}
	if sc.Fortune == "" {
		t.Error("fortune should not be empty")
	}

	// 金金金 → should exist in config
	sc2 := computeSanCai("金", "金", "金")
	if sc2.Configuration != "金金金" {
		t.Errorf("configuration = %s, want 金金金", sc2.Configuration)
	}

	// Unknown combo → default
	sc3 := computeSanCai("木", "火", "金")
	if sc3.Configuration != "木火金" {
		t.Errorf("configuration = %s, want 木火金", sc3.Configuration)
	}
}

// =============================================================================
// Golden: 首页 36 个示例名 — 验证引擎能力覆盖
// =============================================================================

var exampleNames = []string{
	"林观澜", "赵知微", "徐望舒", "王砚清", "李鹿鸣", "刘予安",
	"黄文茵", "吴佩弦", "张知行", "陈思诚", "杨明哲", "孙思远",
	"马归真", "朱修远", "周如玉", "郑含章", "谢清风", "唐致远",
	"于若水", "邓景行", "钱浩然", "薛养正", "卢思齐", "戴知远",
	"邵明德", "雷敬之", "方敏行", "袁守拙", "乔清和", "秦云舒",
	"任心怡", "苏子衿", "罗静言", "夏砚耕", "顾逢春", "汤书白",
}

func TestExampleNames_CharsInDatabase(t *testing.T) {
	for _, full := range exampleNames {
		rs := []rune(full)
		surname := string(rs[0])
		g1 := string(rs[1])
		g2 := string(rs[2])

		// 姓氏必须在字典中
		if lookupKangxiStroke(surname) == 0 {
			t.Errorf("%s: surname %q not in database", full, surname)
			continue
		}
		// 两个名字字必须在字典中
		ce1, ok1 := charByRune[rs[1]]
		if !ok1 {
			t.Errorf("%s: char %q not in charByRune", full, g1)
			continue
		}
		ce2, ok2 := charByRune[rs[2]]
		if !ok2 {
			t.Errorf("%s: char %q not in charByRune", full, g2)
			continue
		}
		// 五格必须可计算
		ss := singleStrokes(lookupKangxiStroke(surname))
		wg := computeWuGeFromStrokes(ss, ce1.Stroke, ce2.Stroke)
		if wg.TianGe.Stroke == 0 || wg.RenGe.Stroke == 0 || wg.DiGe.Stroke == 0 {
			t.Errorf("%s: wuge calculation failed: %+v", full, wg)
		}
	}
}

// TestWuxingFromChinese verifies element mapping.
func TestWuxingFromChinese(t *testing.T) {
	tests := []struct {
		ch   string
		want Wuxing
	}{
		{"木", ganzhi.WxMu},
		{"火", ganzhi.WxHuo},
		{"土", ganzhi.WxTu},
		{"金", ganzhi.WxJin},
		{"水", ganzhi.WxShui},
		{"x", ganzhi.Wuxing(0)},
	}

	for _, tt := range tests {
		got := wuxingFromChinese(tt.ch)
		if got != tt.want {
			t.Errorf("wuxingFromChinese(%q) = %d, want %d", tt.ch, got, tt.want)
		}
	}
}
