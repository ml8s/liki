package ganzhi

import (
	"encoding/json"
	"testing"
)

// =============================================================================
// IsAnHe — 暗合
// =============================================================================

func TestIsAnHe_AllPairs(t *testing.T) {
	// 地支暗合: 寅丑, 卯申, 午亥, 子戌
	pairs := []BranchPair{
		{A: ZhiYin, B: ZhiChou}, // 寅丑
		{A: ZhiMao, B: ZhiShen}, // 卯申
		{A: ZhiWu, B: ZhiHai},   // 午亥
		{A: ZhiZi, B: ZhiXu},    // 子戌
	}
	for _, p := range pairs {
		a, b := p.A, p.B
		if !IsAnHe(a, b) {
			t.Errorf("IsAnHe(%s,%s)=false, want true", ZhiName(a), ZhiName(b))
		}
		if !IsAnHe(b, a) {
			t.Errorf("IsAnHe(%s,%s)=false, want true (reversed)", ZhiName(b), ZhiName(a))
		}
	}
}

func TestIsAnHe_NonPairs(t *testing.T) {
	nonPairs := []BranchPair{
		{A: ZhiZi, B: ZhiChou}, // 子丑合不是暗合
		{A: ZhiYin, B: ZhiMao}, // 寅卯会不是暗合
		{A: ZhiZi, B: ZhiWu},   // 子午冲不是暗合
	}
	for _, p := range nonPairs {
		if IsAnHe(p.A, p.B) {
			t.Errorf("IsAnHe(%s,%s)=true, want false", ZhiName(p.A), ZhiName(p.B))
		}
	}
}

func TestIsAnHe_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiYin, ZhiWu} {
		if IsAnHe(z, z) {
			t.Errorf("IsAnHe(%s,%s)=true, same branch should be false", ZhiName(z), ZhiName(z))
		}
	}
}

// =============================================================================
// IsPo — 相破
// =============================================================================

func TestIsPo_AllPairs(t *testing.T) {
	// 地支相破: 子酉, 寅亥, 辰丑, 午卯, 申巳, 戌未
	pairs := []BranchPair{
		{A: ZhiZi, B: ZhiYou},    // 子酉
		{A: ZhiYin, B: ZhiHai},   // 寅亥
		{A: ZhiChen, B: ZhiChou}, // 辰丑
		{A: ZhiWu, B: ZhiMao},    // 午卯
		{A: ZhiShen, B: ZhiSi},   // 申巳
		{A: ZhiXu, B: ZhiWei},    // 戌未
	}
	for _, p := range pairs {
		a, b := p.A, p.B
		if !IsPo(a, b) {
			t.Errorf("IsPo(%s,%s)=false, want true", ZhiName(a), ZhiName(b))
		}
		if !IsPo(b, a) {
			t.Errorf("IsPo(%s,%s)=false, want true (reversed)", ZhiName(b), ZhiName(a))
		}
	}
}

func TestIsPo_NonPairs(t *testing.T) {
	nonPairs := []BranchPair{
		{A: ZhiZi, B: ZhiChou}, // 子丑合不是破
		{A: ZhiYin, B: ZhiMao}, // 寅卯会不是破
		{A: ZhiZi, B: ZhiWu},   // 子午冲不是破
	}
	for _, p := range nonPairs {
		if IsPo(p.A, p.B) {
			t.Errorf("IsPo(%s,%s)=true, want false", ZhiName(p.A), ZhiName(p.B))
		}
	}
}

func TestIsPo_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiYin, ZhiWu} {
		if IsPo(z, z) {
			t.Errorf("IsPo(%s,%s)=true, same branch should be false", ZhiName(z), ZhiName(z))
		}
	}
}

// =============================================================================
// NayinLabel — 纳音
// =============================================================================

func TestNayinLabel_Known(t *testing.T) {
	tests := []struct {
		name string
		g    Gan
		z    Zhi
		want string
	}{
		{"甲子-海中金", GanJia, ZhiZi, "海中金"},
		{"乙丑-海中金", GanYi, ZhiChou, "海中金"},
		{"丙寅-炉中火", GanBing, ZhiYin, "炉中火"},
		{"丁卯-炉中火", GanDing, ZhiMao, "炉中火"},
		{"戊辰-大林木", GanWu, ZhiChen, "大林木"},
		{"己巳-大林木", GanJi, ZhiSi, "大林木"},
		{"庚午-路旁土", GanGeng, ZhiWu, "路旁土"},
		{"辛未-路旁土", GanXin, ZhiWei, "路旁土"},
		{"壬申-剑锋金", GanRen, ZhiShen, "剑锋金"},
		{"癸酉-剑锋金", GanGui, ZhiYou, "剑锋金"},
		{"甲戌-山头火", GanJia, ZhiXu, "山头火"},
		{"乙亥-山头火", GanYi, ZhiHai, "山头火"},
		{"壬戌-大海水", GanRen, ZhiXu, "大海水"},
		{"癸亥-大海水", GanGui, ZhiHai, "大海水"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NayinLabel(tt.g, tt.z)
			if got != tt.want {
				t.Errorf("NayinLabel(%s,%s)=%s, want %s",
					GanName(tt.g), ZhiName(tt.z), got, tt.want)
			}
		})
	}
}

func TestNayinLabel_All60HaveName(t *testing.T) {
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			if int(g)%2 != int(z)%2 {
				continue
			}
			label := NayinLabel(g, z)
			if label == "" || label == "未知" {
				t.Errorf("NayinLabel(%s,%s) empty or 未知", GanName(g), ZhiName(z))
			}
		}
	}
}

// =============================================================================
// CangGanForZhi — 藏干
// =============================================================================

func TestCangGanForZhi_All(t *testing.T) {
	tests := []struct {
		name     string
		z        Zhi
		mainGan  Gan
		midGan   Gan // 0 if nil
		minorGan Gan // 0 if nil
	}{
		{"子藏癸", ZhiZi, GanGui, 0, 0},
		{"丑藏己癸辛", ZhiChou, GanJi, GanGui, GanXin},
		{"寅藏甲丙戊", ZhiYin, GanJia, GanBing, GanWu},
		{"卯藏乙", ZhiMao, GanYi, 0, 0},
		{"辰藏戊乙癸", ZhiChen, GanWu, GanYi, GanGui},
		{"巳藏丙庚戊", ZhiSi, GanBing, GanGeng, GanWu},
		{"午藏丁己", ZhiWu, GanDing, GanJi, 0},
		{"未藏己乙丁", ZhiWei, GanJi, GanYi, GanDing},
		{"申藏庚壬戊", ZhiShen, GanGeng, GanRen, GanWu},
		{"酉藏辛", ZhiYou, GanXin, 0, 0},
		{"戌藏戊辛丁", ZhiXu, GanWu, GanXin, GanDing},
		{"亥藏壬甲", ZhiHai, GanRen, GanJia, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := CangGanForZhi(tt.z)
			if hs.Main == nil || *hs.Main != tt.mainGan {
				t.Errorf("%s main: got %v, want %d", ZhiName(tt.z), hs.Main, int(tt.mainGan))
			}
			if tt.midGan == 0 {
				if hs.Mid != nil {
					t.Errorf("%s mid: got %v, want nil", ZhiName(tt.z), hs.Mid)
				}
			} else {
				if hs.Mid == nil || *hs.Mid != tt.midGan {
					t.Errorf("%s mid: got %v, want %d", ZhiName(tt.z), hs.Mid, int(tt.midGan))
				}
			}
			if tt.minorGan == 0 {
				if hs.Minor != nil {
					t.Errorf("%s minor: got %v, want nil", ZhiName(tt.z), hs.Minor)
				}
			} else {
				if hs.Minor == nil || *hs.Minor != tt.minorGan {
					t.Errorf("%s minor: got %v, want %d", ZhiName(tt.z), hs.Minor, int(tt.minorGan))
				}
			}
		})
	}
}

func TestCangGanForZhi_Invalid(t *testing.T) {
	hs := CangGanForZhi(0)
	if hs.Main != nil {
		t.Error("CangGanForZhi(0) should have nil Main")
	}
}

func TestHiddenStems_Slice(t *testing.T) {
	hs := CangGanForZhi(ZhiChou)
	s := hs.Slice()
	if len(s) != 3 {
		t.Fatalf("Slice() len=%d, want 3", len(s))
	}
	if s[0] != hs.Main || s[1] != hs.Mid || s[2] != hs.Minor {
		t.Error("Slice() elements don't match struct fields")
	}
}

// =============================================================================
// RenYuanSiLingFenYeForZhi — 人元司令分野
// =============================================================================

func TestRenYuanSiLingFenYeForZhi_All(t *testing.T) {
	// Verify each month branch has phases totaling ~30 days
	for z := ZhiZi; z <= ZhiHai; z++ {
		phases := RenYuanSiLingFenYeForZhi(z)
		if len(phases) == 0 {
			t.Errorf("RenYuanSiLingFenYeForZhi(%s) empty", ZhiName(z))
			continue
		}
		totalDays := 0
		for _, p := range phases {
			if p.Gan < 1 || p.Gan > 10 {
				t.Errorf("%s phase gan=%d invalid", ZhiName(z), p.Gan)
			}
			if p.Days <= 0 {
				t.Errorf("%s phase days=%d <= 0", ZhiName(z), p.Days)
			}
			totalDays += p.Days
		}
		// Should be 30 or 31 days
		if totalDays < 29 || totalDays > 31 {
			t.Errorf("%s total days=%d, expected 29-31", ZhiName(z), totalDays)
		}
	}
}

func TestRenYuanSiLingFenYeForZhi_Known(t *testing.T) {
	// 寅月: 戊7→丙7→甲16 = 30天
	phases := RenYuanSiLingFenYeForZhi(ZhiYin)
	if len(phases) != 3 {
		t.Fatalf("寅月 phases len=%d, want 3", len(phases))
	}
	if phases[0].Gan != GanWu || phases[0].Days != 7 {
		t.Errorf("寅月[0]: gan=%s days=%d, want 戊 7", phases[0].GanName, phases[0].Days)
	}
	if phases[1].Gan != GanBing || phases[1].Days != 7 {
		t.Errorf("寅月[1]: gan=%s days=%d, want 丙 7", phases[1].GanName, phases[1].Days)
	}
	if phases[2].Gan != GanJia || phases[2].Days != 16 {
		t.Errorf("寅月[2]: gan=%s days=%d, want 甲 16", phases[2].GanName, phases[2].Days)
	}
}

func TestRenYuanSiLingFenYeForZhi_Invalid(t *testing.T) {
	phases := RenYuanSiLingFenYeForZhi(0)
	if phases != nil {
		t.Error("RenYuanSiLingFenYeForZhi(0) should be nil")
	}
}

// =============================================================================
// ShiShen — 十神
// =============================================================================

func TestShiShenType_AllRelations(t *testing.T) {
	dmElem := GanWuxing(GanJia)
	dmYY := GanYinYang(GanJia)

	tests := []struct {
		name     string
		other    Gan
		typeName string
	}{
		{"甲→甲-比肩", GanJia, "比肩"},
		{"甲→乙-劫财", GanYi, "劫财"},
		{"甲→丙-食神", GanBing, "食神"},
		{"甲→丁-伤官", GanDing, "伤官"},
		{"甲→戊-偏财", GanWu, "偏财"},
		{"甲→己-正财", GanJi, "正财"},
		{"甲→庚-七杀", GanGeng, "七杀"},
		{"甲→辛-正官", GanXin, "正官"},
		{"甲→壬-偏印", GanRen, "偏印"},
		{"甲→癸-正印", GanGui, "正印"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otherElem := GanWuxing(tt.other)
			otherYY := GanYinYang(tt.other)
			tgType := ShiShenType(dmElem, dmYY, otherElem, otherYY)
			gotName := ShiShenName(tgType)
			if gotName != tt.typeName {
				t.Errorf("甲→%s: got %s, want %s", GanName(tt.other), gotName, tt.typeName)
			}
		})
	}
}

func TestShiShenFromGan_All(t *testing.T) {
	tests := []struct {
		name  string
		other Gan
		want  ShiShen
	}{
		{"甲→甲-比肩", GanJia, ShiShenBiJian},
		{"甲→乙-劫财", GanYi, ShiShenJieCai},
		{"甲→丙-食神", GanBing, ShiShenShiShen},
		{"甲→丁-伤官", GanDing, ShiShenShangGuan},
		{"甲→戊-偏财", GanWu, ShiShenPianCai},
		{"甲→己-正财", GanJi, ShiShenZhengCai},
		{"甲→庚-七杀", GanGeng, ShiShenQiSha},
		{"甲→辛-正官", GanXin, ShiShenZhengGuan},
		{"甲→壬-偏印", GanRen, ShiShenPianYin},
		{"甲→癸-正印", GanGui, ShiShenZhengYin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShiShenFromGan(GanJia, tt.other)
			if got != tt.want {
				t.Errorf("ShiShenFromGan(甲,%s)=%s, want %s",
					GanName(tt.other), got, tt.want)
			}
		})
	}
}

func TestShiShenFromGan_DifferentDayMasters(t *testing.T) {
	// 丙(火阳)日主 → 庚(金阳)=偏财, 辛(金阴)=正财
	if got := ShiShenFromGan(GanBing, GanGeng); got != ShiShenPianCai {
		t.Errorf("ShiShenFromGan(丙,庚)=%s, want 偏财", got)
	}
	if got := ShiShenFromGan(GanBing, GanXin); got != ShiShenZhengCai {
		t.Errorf("ShiShenFromGan(丙,辛)=%s, want 正财", got)
	}
}

func TestShiShenName_Invalid(t *testing.T) {
	if got := ShiShenName(-1); got != "" {
		t.Errorf("ShiShenName(-1)=%s, want empty", got)
	}
	if got := ShiShenName(10); got != "" {
		t.Errorf("ShiShenName(10)=%s, want empty", got)
	}
}

// =============================================================================
// ParseWuxing — 五行 (中英文)
// =============================================================================

func TestParseWuxing_Chinese(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Wuxing
	}{
		{"木", "木", WxMu},
		{"火", "火", WxHuo},
		{"土", "土", WxTu},
		{"金", "金", WxJin},
		{"水", "水", WxShui},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWuxing(tt.input)
			if err != nil {
				t.Errorf("ParseWuxing(%s) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseWuxing(%s)=%d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWuxing_English(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Wuxing
	}{
		{"wood", "wood", WxMu},
		{"fire", "fire", WxHuo},
		{"earth", "earth", WxTu},
		{"metal", "metal", WxJin},
		{"water", "water", WxShui},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWuxing(tt.input)
			if err != nil {
				t.Errorf("ParseWuxing(%s) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseWuxing(%s)=%d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWuxing_Invalid(t *testing.T) {
	for _, s := range []string{"", "X", "unknown", "土土"} {
		got, err := ParseWuxing(s)
		if err == nil {
			t.Errorf("ParseWuxing(%s)=%d, want error", s, got)
		}
	}
}

// =============================================================================
// SixtyToZhu — 六十甲子序号转柱
// =============================================================================

func TestSixtyToZhu_Known(t *testing.T) {
	tests := []struct {
		name string
		idx  int
		gan  Gan
		zhi  Zhi
	}{
		{"甲子", 0, GanJia, ZhiZi},
		{"乙丑", 1, GanYi, ZhiChou},
		{"甲戌", 10, GanJia, ZhiXu},
		{"壬辰", 28, GanRen, ZhiChen},
		{"癸巳", 29, GanGui, ZhiSi},
		{"甲寅", 50, GanJia, ZhiYin},
		{"癸亥", 59, GanGui, ZhiHai},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := SixtyToZhu(tt.idx)
			if z.Gan != tt.gan || z.Zhi != tt.zhi {
				t.Errorf("SixtyToZhu(%d)=%s%s, want %s%s",
					tt.idx, GanName(z.Gan), ZhiName(z.Zhi),
					GanName(tt.gan), ZhiName(tt.zhi))
			}
		})
	}
}

func TestSixtyToZhu_Roundtrip(t *testing.T) {
	// SixtyToZhu → SixtyCycleIndex should be identity
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			if int(g)%2 != int(z)%2 {
				continue
			}
			idx := SixtyCycleIndex(g, z)
			zhu := SixtyToZhu(idx)
			if zhu.Gan != g || zhu.Zhi != z {
				t.Errorf("roundtrip: %s%s(idx=%d) → SixtyToZhu → %s%s",
					GanName(g), ZhiName(z), idx, GanName(zhu.Gan), ZhiName(zhu.Zhi))
			}
		}
	}
}

func TestSixtyToZhu_All60(t *testing.T) {
	seen := make(map[int]bool)
	for i := 0; i < 60; i++ {
		z := SixtyToZhu(i)
		idx := SixtyCycleIndex(z.Gan, z.Zhi)
		if seen[idx] {
			t.Errorf("SixtyToZhu(%d): duplicate index %d", i, idx)
		}
		seen[idx] = true
	}
	if len(seen) != 60 {
		t.Errorf("SixtyToZhu produced %d unique pairs, want 60", len(seen))
	}
}

// =============================================================================
// MarshalJSON — Gan/Zhi/Wuxing
// =============================================================================

func TestGan_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(GanJia)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"甲"` {
		t.Errorf("GanJia.MarshalJSON=%s, want \"甲\"", string(b))
	}
}

func TestZhi_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(ZhiZi)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"子"` {
		t.Errorf("ZhiZi.MarshalJSON=%s, want \"子\"", string(b))
	}
}

func TestWuxing_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(WxMu)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(b) != `"木"` {
		t.Errorf("WxMu.MarshalJSON=%s, want \"木\"", string(b))
	}
}

// =============================================================================
// String — Gan/Zhi
// =============================================================================

func TestGan_String(t *testing.T) {
	if GanJia.String() != "甲" {
		t.Errorf("GanJia.String()=%s, want 甲", GanJia.String())
	}
}

func TestZhi_String(t *testing.T) {
	if ZhiZi.String() != "子" {
		t.Errorf("ZhiZi.String()=%s, want 子", ZhiZi.String())
	}
}

func TestWuxing_String_Invalid(t *testing.T) {
	w := Wuxing(99)
	if w.String() != "未知" {
		t.Errorf("Wuxing(99).String()=%s, want 未知", w.String())
	}
}

// =============================================================================
// Wuxing UnmarshalJSON — 错误路径
// =============================================================================

func TestWuxing_UnmarshalJSON_Invalid(t *testing.T) {
	var w Wuxing
	if err := json.Unmarshal([]byte(`"X"`), &w); err == nil {
		t.Error("expected error for invalid wuxing name")
	}
}

func TestGan_UnmarshalJSON_Invalid(t *testing.T) {
	var g Gan
	if err := json.Unmarshal([]byte(`"X"`), &g); err == nil {
		t.Error("expected error for invalid gan name")
	}
}

func TestZhi_UnmarshalJSON_Invalid(t *testing.T) {
	var z Zhi
	if err := json.Unmarshal([]byte(`"X"`), &z); err == nil {
		t.Error("expected error for invalid zhi name")
	}
}

// =============================================================================
// LoadHeHua — 合化结果验证
// =============================================================================

func TestZhiHe_Result(t *testing.T) {
	for _, zh := range ZhiHes {
		if zh.Result < 1 || zh.Result > 5 {
			t.Errorf("ZhiHe(%s,%s): invalid result %d", ZhiName(zh.A), ZhiName(zh.B), zh.Result)
		}
	}
}

func TestGanHe_Result(t *testing.T) {
	for _, gh := range GanHes {
		if gh.Result < 1 || gh.Result > 5 {
			t.Errorf("GanHe(%s,%s): invalid result %d", GanName(gh.A), GanName(gh.B), gh.Result)
		}
	}
}

func TestTripleHe_Consistency(t *testing.T) {
	// 三合局每局3个地支
	for _, th := range TripleHeList {
		if len(th.Branches) != 3 {
			t.Errorf("TripleHe element=%d: got %d branches, want 3", th.Element, len(th.Branches))
		}
		if th.Element < 1 || th.Element > 5 {
			t.Errorf("TripleHe: invalid element %d", th.Element)
		}
	}
}

func TestTripleHui_Consistency(t *testing.T) {
	// 三会局每局3个连续地支
	for _, th := range TripleHuiList {
		if len(th.Branches) != 3 {
			t.Errorf("TripleHui element=%d: got %d branches, want 3", th.Element, len(th.Branches))
		}
	}
}

// =============================================================================
// NaYinTable — 纳音验证
// =============================================================================

func TestNayinTable_All60(t *testing.T) {
	if len(NayinTable) != 60 {
		t.Errorf("NayinTable len=%d, want 60", len(NayinTable))
	}
	for i := 0; i < 60; i++ {
		if name, ok := NayinTable[i]; !ok || name == "" {
			t.Errorf("NayinTable[%d] missing", i)
		}
	}
}

// =============================================================================
// ChangShengTable — 十二长生
// =============================================================================

func TestChangShengTable_AllStems(t *testing.T) {
	// Each stem should map to exactly 12 stages (branch positions)
	for g := GanJia; g <= GanGui; g++ {
		stages, ok := ChangShengTable[g]
		if !ok {
			t.Errorf("ChangShengTable[%s] missing", GanName(g))
			continue
		}
		if len(stages) != 12 {
			t.Errorf("%s: got %d stages, want 12", GanName(g), len(stages))
		}
	}
}

func TestStageNamesZH(t *testing.T) {
	if len(StageNamesZH) != 12 {
		t.Errorf("StageNamesZH len=%d, want 12", len(StageNamesZH))
	}
	expected := [12]string{
		"长生", "沐浴", "冠带", "临官", "帝旺",
		"衰", "病", "死", "墓", "绝", "胎", "养",
	}
	for i, want := range expected {
		if StageNamesZH[i] != want {
			t.Errorf("StageNamesZH[%d]=%s, want %s", i, StageNamesZH[i], want)
		}
	}
}

// =============================================================================
// ChongPairs — 六冲对
// =============================================================================

func TestChongPairs_Count(t *testing.T) {
	if len(ChongPairs) != 6 {
		t.Errorf("ChongPairs len=%d, want 6", len(ChongPairs))
	}
}

func TestHaiPairs_Count(t *testing.T) {
	if len(HaiPairs) != 6 {
		t.Errorf("HaiPairs len=%d, want 6", len(HaiPairs))
	}
}

func TestZhiHes_Count(t *testing.T) {
	if len(ZhiHes) != 6 {
		t.Errorf("ZhiHes len=%d, want 6", len(ZhiHes))
	}
}

func TestGanHes_Count(t *testing.T) {
	if len(GanHes) != 5 {
		t.Errorf("GanHes len=%d, want 5", len(GanHes))
	}
}

// =============================================================================
// XingGroups — 刑
// =============================================================================

func TestXingGroups_NotEmpty(t *testing.T) {
	if len(XingGroups) == 0 {
		t.Error("XingGroups is empty")
	}
	for _, x := range XingGroups {
		if x.Type == "" {
			t.Error("XingGroup has empty type")
		}
		if len(x.Branches) == 0 {
			t.Error("XingGroup has empty branches")
		}
	}
}

// ── IsGanHe ──

func TestIsGanHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Gan
	}{
		{"甲己合", GanJia, GanJi},
		{"乙庚合", GanYi, GanGeng},
		{"丙辛合", GanBing, GanXin},
		{"丁壬合", GanDing, GanRen},
		{"戊癸合", GanWu, GanGui},
	}
	for _, tc := range tests {
		if !IsGanHe(tc.a, tc.b) {
			t.Errorf("IsGanHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsGanHe(tc.b, tc.a) {
			t.Errorf("IsGanHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsGanHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Gan
	}{
		{"甲乙", GanJia, GanYi},
		{"甲丙", GanJia, GanBing},
		{"己庚", GanJi, GanGeng},
		{"乙丙", GanYi, GanBing},
	}
	for _, tc := range tests {
		if IsGanHe(tc.a, tc.b) {
			t.Errorf("IsGanHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsGanHe_SameStem(t *testing.T) {
	for _, g := range []Gan{GanJia, GanYi, GanBing, GanDing, GanWu} {
		if IsGanHe(g, g) {
			t.Errorf("IsGanHe(%d,%d)=true, same stem should be false", g, g)
		}
	}
}

// ── IsZhiHe ──

func TestIsZhiHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合", ZhiZi, ZhiChou},
		{"寅亥合", ZhiYin, ZhiHai},
		{"卯戌合", ZhiMao, ZhiXu},
		{"辰酉合", ZhiChen, ZhiYou},
		{"巳申合", ZhiSi, ZhiShen},
		{"午未合", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if !IsZhiHe(tc.a, tc.b) {
			t.Errorf("IsZhiHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsZhiHe(tc.b, tc.a) {
			t.Errorf("IsZhiHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsZhiHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子寅", ZhiZi, ZhiYin},
		{"丑卯", ZhiChou, ZhiMao},
		{"子午冲不是合", ZhiZi, ZhiWu},
	}
	for _, tc := range tests {
		if IsZhiHe(tc.a, tc.b) {
			t.Errorf("IsZhiHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsTripleHe ──

func TestIsTripleHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"申子（水局）", ZhiShen, ZhiZi},
		{"子辰（水局）", ZhiZi, ZhiChen},
		{"亥卯（木局）", ZhiHai, ZhiMao},
		{"卯未（木局）", ZhiMao, ZhiWei},
		{"寅午（火局）", ZhiYin, ZhiWu},
		{"午戌（火局）", ZhiWu, ZhiXu},
		{"巳酉（金局）", ZhiSi, ZhiYou},
		{"酉丑（金局）", ZhiYou, ZhiChou},
	}
	for _, tc := range tests {
		if !IsTripleHe(tc.a, tc.b) {
			t.Errorf("IsTripleHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsTripleHe(tc.b, tc.a) {
			t.Errorf("IsTripleHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsTripleHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"申寅（水火不同局）", ZhiShen, ZhiYin},
		{"子午（水土不同局）", ZhiZi, ZhiWu},
		{"亥寅（木木但不同三合）", ZhiHai, ZhiYin},
	}
	for _, tc := range tests {
		if IsTripleHe(tc.a, tc.b) {
			t.Errorf("IsTripleHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsTripleHui ──

func TestIsTripleHui_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"寅卯（东方木）", ZhiYin, ZhiMao},
		{"卯辰（东方木）", ZhiMao, ZhiChen},
		{"巳午（南方火）", ZhiSi, ZhiWu},
		{"午未（南方火）", ZhiWu, ZhiWei},
		{"申酉（西方金）", ZhiShen, ZhiYou},
		{"酉戌（西方金）", ZhiYou, ZhiXu},
		{"亥子（北方水）", ZhiHai, ZhiZi},
		{"子丑（北方水）", ZhiZi, ZhiChou},
	}
	for _, tc := range tests {
		if !IsTripleHui(tc.a, tc.b) {
			t.Errorf("IsTripleHui(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsTripleHui(tc.b, tc.a) {
			t.Errorf("IsTripleHui(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsTripleHui_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"寅巳（不同方）", ZhiYin, ZhiSi},
		{"子午（不同方）", ZhiZi, ZhiWu},
		{"申寅（不同方）", ZhiShen, ZhiYin},
	}
	for _, tc := range tests {
		if IsTripleHui(tc.a, tc.b) {
			t.Errorf("IsTripleHui(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsLiuChong ──

func TestIsLiuChong_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子午冲", ZhiZi, ZhiWu},
		{"丑未冲", ZhiChou, ZhiWei},
		{"寅申冲", ZhiYin, ZhiShen},
		{"卯酉冲", ZhiMao, ZhiYou},
		{"辰戌冲", ZhiChen, ZhiXu},
		{"巳亥冲", ZhiSi, ZhiHai},
	}
	for _, tc := range tests {
		if !IsLiuChong(tc.a, tc.b) {
			t.Errorf("IsLiuChong(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsLiuChong(tc.b, tc.a) {
			t.Errorf("IsLiuChong(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsLiuChong_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合不是冲", ZhiZi, ZhiChou},
		{"寅亥合不是冲", ZhiYin, ZhiHai},
		{"午未合不是冲", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if IsLiuChong(tc.a, tc.b) {
			t.Errorf("IsLiuChong(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsLiuChong_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiYin, ZhiWu} {
		if IsLiuChong(z, z) {
			t.Errorf("IsLiuChong(%d,%d)=true, same branch should be false", z, z)
		}
	}
}

// ── IsXing ──

func TestIsXing_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		// 无礼之刑
		{"子卯无礼之刑", ZhiZi, ZhiMao},
		// 无恩之刑
		{"寅巳无恩之刑", ZhiYin, ZhiSi},
		{"巳申无恩之刑", ZhiSi, ZhiShen},
		{"寅申无恩之刑", ZhiYin, ZhiShen},
		// 恃势之刑
		{"丑未恃势之刑", ZhiChou, ZhiWei},
		{"未戌恃势之刑", ZhiWei, ZhiXu},
		{"丑戌恃势之刑", ZhiChou, ZhiXu},
	}
	for _, tc := range tests {
		if !IsXing(tc.a, tc.b) {
			t.Errorf("IsXing(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsXing(tc.b, tc.a) {
			t.Errorf("IsXing(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsXing_SelfXing(t *testing.T) {
	// 自刑：辰午酉亥
	for _, z := range []Zhi{ZhiChen, ZhiWu, ZhiYou, ZhiHai} {
		if !IsXing(z, z) {
			t.Errorf("IsXing(%d,%d)=false, self-xing should be true", z, z)
		}
	}
	// Non-self-xing branches should not be self-xing
	for _, z := range []Zhi{ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiSi, ZhiWei, ZhiShen, ZhiXu} {
		if IsXing(z, z) {
			t.Errorf("IsXing(%d,%d)=true, should not be self-xing", z, z)
		}
	}
}

func TestIsXing_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑无刑", ZhiZi, ZhiChou},
		{"寅卯无刑", ZhiYin, ZhiMao},
		{"子丑合不是刑", ZhiZi, ZhiChou},
		{"午未合不是刑", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if IsXing(tc.a, tc.b) {
			t.Errorf("IsXing(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsHai ──

func TestIsHai_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子未害", ZhiZi, ZhiWei},
		{"丑午害", ZhiChou, ZhiWu},
		{"寅巳害", ZhiYin, ZhiSi},
		{"卯辰害", ZhiMao, ZhiChen},
		{"申亥害", ZhiShen, ZhiHai},
		{"酉戌害", ZhiYou, ZhiXu},
	}
	for _, tc := range tests {
		if !IsHai(tc.a, tc.b) {
			t.Errorf("IsHai(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsHai(tc.b, tc.a) {
			t.Errorf("IsHai(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsHai_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合不是害", ZhiZi, ZhiChou},
		{"子午冲不是害", ZhiZi, ZhiWu},
		{"寅亥合不是害", ZhiYin, ZhiHai},
	}
	for _, tc := range tests {
		if IsHai(tc.a, tc.b) {
			t.Errorf("IsHai(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsHai_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiWu, ZhiYin} {
		if IsHai(z, z) {
			t.Errorf("IsHai(%d,%d)=true, same branch should be false", z, z)
		}
	}
}

// ── inBranchList ──

func TestInBranchList(t *testing.T) {
	branches := []Zhi{1, 3, 5}
	tests := []struct {
		name string
		z    Zhi
		want bool
	}{
		{"子在内", ZhiZi, true},
		{"寅在内", ZhiYin, true},
		{"辰在内", ZhiChen, true},
		{"丑不在内", ZhiChou, false},
		{"卯不在内", ZhiMao, false},
	}
	for _, tc := range tests {
		got := inBranchList(branches, tc.z)
		if got != tc.want {
			t.Errorf("inBranchList(%v,%d)=%v, want %v (%s)", branches, tc.z, got, tc.want, tc.name)
		}
	}
}

func TestInBranchList_Empty(t *testing.T) {
	if inBranchList(nil, ZhiZi) {
		t.Error("inBranchList(nil, 子)=true, want false")
	}
	if inBranchList([]Zhi{}, ZhiZi) {
		t.Error("inBranchList([], 子)=true, want false")
	}
}

// -- ZhiSeasonLabel --

func TestZhiSeasonLabel_All(t *testing.T) {
	tests := []struct {
		name string
		z    Zhi
		want string
	}{
		{"寅春", ZhiYin, "春"}, {"卯春", ZhiMao, "春"}, {"辰春", ZhiChen, "春"},
		{"巳夏", ZhiSi, "夏"}, {"午夏", ZhiWu, "夏"}, {"未夏", ZhiWei, "夏"},
		{"申秋", ZhiShen, "秋"}, {"酉秋", ZhiYou, "秋"}, {"戌秋", ZhiXu, "秋"},
		{"亥冬", ZhiHai, "冬"}, {"子冬", ZhiZi, "冬"}, {"丑冬", ZhiChou, "冬"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ZhiSeasonLabel(tc.z)
			if got != tc.want {
				t.Errorf("ZhiSeasonLabel(%d)=%s, want %s", tc.z, got, tc.want)
			}
		})
	}
}

func TestZhiSeasonLabel_Invalid(t *testing.T) {
	if got := ZhiSeasonLabel(0); got != "未知" {
		t.Errorf("ZhiSeasonLabel(0)=%s, want 未知", got)
	}
	if got := ZhiSeasonLabel(13); got != "未知" {
		t.Errorf("ZhiSeasonLabel(13)=%s, want 未知", got)
	}
}

// -- zhiLunarMonth complete --

func TestZhiLunarMonth_All(t *testing.T) {
	tests := []struct {
		name string
		z    Zhi
		want string
	}{
		{"子-十一月", ZhiZi, "十一月"},
		{"丑-十二月", ZhiChou, "十二月"},
		{"寅-正月", ZhiYin, "正月"},
		{"卯-二月", ZhiMao, "二月"},
		{"辰-三月", ZhiChen, "三月"},
		{"巳-四月", ZhiSi, "四月"},
		{"午-五月", ZhiWu, "五月"},
		{"未-六月", ZhiWei, "六月"},
		{"申-七月", ZhiShen, "七月"},
		{"酉-八月", ZhiYou, "八月"},
		{"戌-九月", ZhiXu, "九月"},
		{"亥-十月", ZhiHai, "十月"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := zhiLunarMonth(tc.z)
			if !ok {
				t.Errorf("zhiLunarMonth(%d) not ok", tc.z)
			}
			if got != tc.want {
				t.Errorf("zhiLunarMonth(%d)=%s, want %s", tc.z, got, tc.want)
			}
		})
	}
}

func TestZhiLunarMonth_Invalid(t *testing.T) {
	for _, z := range []Zhi{0, 13} {
		if _, ok := zhiLunarMonth(z); ok {
			t.Errorf("zhiLunarMonth(%d) should not be ok", z)
		}
	}
}

// -- zhiLunarMonthLabel --

func TestZhiLunarMonthLabel_Valid(t *testing.T) {
	if got := zhiLunarMonthLabel(ZhiYin); got != "正月" {
		t.Errorf("zhiLunarMonthLabel(寅)=%s, want 正月", got)
	}
}

func TestZhiLunarMonthLabel_Invalid(t *testing.T) {
	if got := zhiLunarMonthLabel(0); got != "未知" {
		t.Errorf("zhiLunarMonthLabel(0)=%s, want 未知", got)
	}
}

// -- zhiHourRange complete --

func TestZhiHourRange_All(t *testing.T) {
	tests := []struct {
		name string
		z    Zhi
		want string
	}{
		{"子时", ZhiZi, "23:00-01:00"},
		{"丑时", ZhiChou, "01:00-03:00"},
		{"寅时", ZhiYin, "03:00-05:00"},
		{"卯时", ZhiMao, "05:00-07:00"},
		{"辰时", ZhiChen, "07:00-09:00"},
		{"巳时", ZhiSi, "09:00-11:00"},
		{"午时", ZhiWu, "11:00-13:00"},
		{"未时", ZhiWei, "13:00-15:00"},
		{"申时", ZhiShen, "15:00-17:00"},
		{"酉时", ZhiYou, "17:00-19:00"},
		{"戌时", ZhiXu, "19:00-21:00"},
		{"亥时", ZhiHai, "21:00-23:00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := zhiHourRange(tc.z)
			if !ok {
				t.Errorf("zhiHourRange(%d) not ok", tc.z)
			}
			if got != tc.want {
				t.Errorf("zhiHourRange(%d)=%s, want %s", tc.z, got, tc.want)
			}
		})
	}
}

func TestZhiHourRange_Invalid(t *testing.T) {
	for _, z := range []Zhi{0, 13} {
		if _, ok := zhiHourRange(z); ok {
			t.Errorf("zhiHourRange(%d) should not be ok", z)
		}
	}
}

// -- ZhiHourRangeLabel --

func TestZhiHourRangeLabel_Valid(t *testing.T) {
	if got := ZhiHourRangeLabel(ZhiWu); got != "11:00-13:00" {
		t.Errorf("ZhiHourRangeLabel(午)=%s, want 11:00-13:00", got)
	}
}

func TestZhiHourRangeLabel_Invalid(t *testing.T) {
	if got := ZhiHourRangeLabel(0); got != "未知" {
		t.Errorf("ZhiHourRangeLabel(0)=%s, want 未知", got)
	}
}

// -- GanName / ZhiName —
func TestGanName_Invalid(t *testing.T) {
	if got := GanName(0); got != "" {
		t.Errorf("GanName(0)=%s, want empty", got)
	}
}
func TestZhiName_Invalid(t *testing.T) {
	if got := ZhiName(0); got != "" {
		t.Errorf("ZhiName(0)=%s, want empty", got)
	}
}

func TestSixtyCycleIndex_Known(t *testing.T) {
	tests := []struct {
		name string
		gan  Gan
		zhi  Zhi
		want int
	}{
		{"甲子", GanJia, ZhiZi, 0},
		{"乙丑", GanYi, ZhiChou, 1},
		{"癸亥", GanGui, ZhiHai, 59},
		{"甲戌", GanJia, ZhiXu, 10},
		{"丙寅", GanBing, ZhiYin, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SixtyCycleIndex(tc.gan, tc.zhi)
			if got != tc.want {
				t.Errorf("SixtyCycleIndex(%s,%s)=%d, want %d", tc.gan, tc.zhi, got, tc.want)
			}
		})
	}
}

func TestSixtyCycleIndex_Range(t *testing.T) {
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			idx := SixtyCycleIndex(g, z)
			if idx < 0 || idx > 59 {
				t.Errorf("SixtyCycleIndex(%s,%s)=%d out of [0,59]", g, z, idx)
			}
		}
	}
}

func TestSixtyCycleIndex_Bijection(t *testing.T) {
	// The 60 JiaZi cycle only includes pairs where gan and zhi have the same
	// yin-yang parity (both odd or both even). All 60 valid pairs must produce
	// distinct indices with no collisions.
	seen := make(map[int]bool)
	count := 0
	for g := GanJia; g <= GanGui; g++ {
		for z := ZhiZi; z <= ZhiHai; z++ {
			// Only valid 60-cycle pairs: same parity (both yang or both yin).
			if int(g)%2 != int(z)%2 {
				continue
			}
			count++
			idx := SixtyCycleIndex(g, z)
			if seen[idx] {
				t.Errorf("SixtyCycleIndex collision at index %d (gan=%s, zhi=%s)", idx, g, z)
			}
			seen[idx] = true
		}
	}
	if len(seen) != 60 {
		t.Errorf("expected 60 unique indices, got %d", len(seen))
	}
	if count != 60 {
		t.Errorf("expected 60 valid pairs, counted %d", count)
	}
}

func TestSixtyCycleIndex_Consecutive(t *testing.T) {
	// In the 60-cycle, consecutive pairs advance both gan and zhi by 1.
	// Starting from 甲子(0), next is 乙丑(1), then 丙寅(2), etc.
	ganVals := []Gan{GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui,
		GanJia, GanYi, GanBing, GanDing, GanWu, GanJi, GanGeng, GanXin, GanRen, GanGui}
	zhiVals := []Zhi{ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai,
		ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiChen, ZhiSi, ZhiWu, ZhiWei, ZhiShen, ZhiYou, ZhiXu, ZhiHai}

	for i := 0; i < 59; i++ {
		idx1 := SixtyCycleIndex(ganVals[i], zhiVals[i])
		idx2 := SixtyCycleIndex(ganVals[i+1], zhiVals[i+1])
		if idx2 != (idx1+1)%60 {
			t.Errorf("%s%s=%d → %s%s=%d, want +1 mod 60",
				ganVals[i], zhiVals[i], idx1, ganVals[i+1], zhiVals[i+1], idx2)
		}
	}
}

// -- sheng / ke --

func TestSheng_Known(t *testing.T) {
	tests := []struct {
		name string
		from, to Wuxing
	}{
		{"木生火", WxMu, WxHuo},
		{"火生土", WxHuo, WxTu},
		{"土生金", WxTu, WxJin},
		{"金生水", WxJin, WxShui},
		{"水生木", WxShui, WxMu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !Sheng(tt.from, tt.to) {
				t.Errorf("Sheng(%s,%s)=false, want true", tt.from, tt.to)
			}
		})
	}
}

func TestSheng_NonPairs(t *testing.T) {
	tests := []struct {
		name string
		from, to Wuxing
	}{
		{"木土非生", WxMu, WxTu},
		{"木金非生", WxMu, WxJin},
		{"木水非生", WxMu, WxShui},
		{"木木非生", WxMu, WxMu},
		{"火木非生", WxHuo, WxMu},
		{"火水非生", WxHuo, WxShui},
		{"火火非生", WxHuo, WxHuo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Sheng(tt.from, tt.to) {
				t.Errorf("Sheng(%s,%s)=true, want false", tt.from, tt.to)
			}
		})
	}
}

func TestKe_Known(t *testing.T) {
	tests := []struct {
		name string
		from, to Wuxing
	}{
		{"木克土", WxMu, WxTu},
		{"土克水", WxTu, WxShui},
		{"水克火", WxShui, WxHuo},
		{"火克金", WxHuo, WxJin},
		{"金克木", WxJin, WxMu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !Ke(tt.from, tt.to) {
				t.Errorf("Ke(%s,%s)=false, want true", tt.from, tt.to)
			}
		})
	}
}

func TestKe_NonPairs(t *testing.T) {
	tests := []struct {
		name string
		from, to Wuxing
	}{
		{"木火生非克", WxMu, WxHuo},
		{"火土生非克", WxHuo, WxTu},
		{"土金生非克", WxTu, WxJin},
		{"金水生非克", WxJin, WxShui},
		{"水木生非克", WxShui, WxMu},
		{"木木同非克", WxMu, WxMu},
		{"火火同非克", WxHuo, WxHuo},
		{"木水逆生非克", WxMu, WxShui},
		{"火木逆生非克", WxHuo, WxMu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Ke(tt.from, tt.to) {
				t.Errorf("Ke(%s,%s)=true, want false", tt.from, tt.to)
			}
		})
	}
}

// -- ganWuxing / zhiWuxing --

func TestGanWuxing_All(t *testing.T) {
	want := map[Gan]Wuxing{
		GanJia: WxMu, GanYi: WxMu,
		GanBing: WxHuo, GanDing: WxHuo,
		GanWu: WxTu, GanJi: WxTu,
		GanGeng: WxJin, GanXin: WxJin,
		GanRen: WxShui, GanGui: WxShui,
	}
	for g, w := range want {
		if got := GanWuxing(g); got != w {
			t.Errorf("GanWuxing(%s)=%s, want %s", g, got, w)
		}
	}
}

func TestZhiWuxing_All(t *testing.T) {
	want := map[Zhi]Wuxing{
		ZhiYin: WxMu, ZhiMao: WxMu,
		ZhiSi: WxHuo, ZhiWu: WxHuo,
		ZhiChen: WxTu, ZhiXu: WxTu, ZhiChou: WxTu, ZhiWei: WxTu,
		ZhiShen: WxJin, ZhiYou: WxJin,
		ZhiHai: WxShui, ZhiZi: WxShui,
	}
	for z, w := range want {
		if got := ZhiWuxing(z); got != w {
			t.Errorf("ZhiWuxing(%s)=%s, want %s", z, got, w)
		}
	}
}

// -- ganYinYang --

func TestGanYinYang_All(t *testing.T) {
	yang := []Gan{GanJia, GanBing, GanWu, GanGeng, GanRen}
	yin := []Gan{GanYi, GanDing, GanJi, GanXin, GanGui}
	for _, g := range yang {
		if got := GanYinYang(g); got != Yang {
			t.Errorf("GanYinYang(%s)=Yin, want Yang", g)
		}
	}
	for _, g := range yin {
		if got := GanYinYang(g); got != Yin {
			t.Errorf("GanYinYang(%s)=Yang, want Yin", g)
		}
	}
}

// -- Wuxing.String / ParseWuxing round-trip --

func TestWuxing_String(t *testing.T) {
	want := map[Wuxing]string{WxMu: "木", WxHuo: "火", WxTu: "土", WxJin: "金", WxShui: "水"}
	for w, s := range want {
		if got := w.String(); got != s {
			t.Errorf("Wuxing(%d).String()=%s, want %s", w, got, s)
		}
	}
}

func TestParseWuxing_Roundtrip(t *testing.T) {
	names := []string{"木", "火", "土", "金", "水"}
	for _, s := range names {
		w, err := ParseWuxing(s)
		if err != nil {
			t.Errorf("ParseWuxing(%s) error: %v", s, err)
		}
		if w.String() != s {
			t.Errorf("roundtrip: %s → %d → %s", s, w, w.String())
		}
	}
	if _, err := ParseWuxing("X"); err == nil {
		t.Error("ParseWuxing(X) should error")
	}
}

// -- GanName / ZhiName --

func TestGanName_All(t *testing.T) {
	names := []string{"", "甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	for i := 1; i <= 10; i++ {
		if got := GanName(Gan(i)); got != names[i] {
			t.Errorf("GanName(%d)=%s, want %s", i, got, names[i])
		}
	}
}

func TestZhiName_All(t *testing.T) {
	names := []string{"", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	for i := 1; i <= 12; i++ {
		if got := ZhiName(Zhi(i)); got != names[i] {
			t.Errorf("ZhiName(%d)=%s, want %s", i, got, names[i])
		}
	}
}

// -- zodiac / zhiSeason / zhiLunarMonth / zhiHourRange --

func TestZodiac_All(t *testing.T) {
	want := []string{"", "鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	for i := 1; i <= 12; i++ {
		s, ok := zodiac(Zhi(i))
		if !ok {
			t.Errorf("zodiac(%d) not found", i)
		}
		if s != want[i] {
			t.Errorf("zodiac(%d)=%s, want %s", i, s, want[i])
		}
	}
}

func TestZodiac_Invalid(t *testing.T) {
	if _, ok := zodiac(0); ok {
		t.Error("zodiac(0) should not be ok")
	}
	if _, ok := zodiac(13); ok {
		t.Error("zodiac(13) should not be ok")
	}
	if got := ZodiacLabel(0); got != "未知" {
		t.Errorf("ZodiacLabel(0)=%s, want 未知", got)
	}
}

func TestZhiSeason_Known(t *testing.T) {
	spring := []Zhi{ZhiYin, ZhiMao, ZhiChen}
	summer := []Zhi{ZhiSi, ZhiWu, ZhiWei}
	autumn := []Zhi{ZhiShen, ZhiYou, ZhiXu}
	winter := []Zhi{ZhiHai, ZhiZi, ZhiChou}
	for _, z := range spring {
		s, ok := zhiSeason(z)
		if !ok || s != "春" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (春,true)", z, s, ok)
		}
	}
	for _, z := range summer {
		s, ok := zhiSeason(z)
		if !ok || s != "夏" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (夏,true)", z, s, ok)
		}
	}
	for _, z := range autumn {
		s, ok := zhiSeason(z)
		if !ok || s != "秋" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (秋,true)", z, s, ok)
		}
	}
	for _, z := range winter {
		s, ok := zhiSeason(z)
		if !ok || s != "冬" {
			t.Errorf("zhiSeason(%s)=(%s,%v), want (冬,true)", z, s, ok)
		}
	}
}

func TestZhiLunarMonth_Known(t *testing.T) {
	// 正月 = 寅(3)
	s, ok := zhiLunarMonth(ZhiYin)
	if !ok || s != "正月" {
		t.Errorf("zhiLunarMonth(寅)=(%s,%v), want (正月,true)", s, ok)
	}
	// 十一月 = 子(1)
	s, ok = zhiLunarMonth(ZhiZi)
	if !ok || s != "十一月" {
		t.Errorf("zhiLunarMonth(子)=(%s,%v), want (十一月,true)", s, ok)
	}
}

func TestZhiHourRange_Known(t *testing.T) {
	// 子 = 23:00-01:00
	s, ok := zhiHourRange(ZhiZi)
	if !ok || s != "23:00-01:00" {
		t.Errorf("zhiHourRange(子)=(%s,%v), want (23:00-01:00,true)", s, ok)
	}
	// 午 = 11:00-13:00
	s, ok = zhiHourRange(ZhiWu)
	if !ok || s != "11:00-13:00" {
		t.Errorf("zhiHourRange(午)=(%s,%v), want (11:00-13:00,true)", s, ok)
	}
}

// -- JSON serialization --

func TestZhu_MarshalJSON(t *testing.T) {
	z := Zhu{Gan: GanJia, Zhi: ZhiZi}
	b, err := json.Marshal(z)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if m["gan"] != "甲" || m["zhi"] != "子" {
		t.Errorf("marshaled Zhu = %v, want {甲 子}", m)
	}
}

func TestGan_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Gan
	}{
		{"甲-字符串", `"甲"`, GanJia},
		{"丙-字符串", `"丙"`, GanBing},
		{"癸-字符串", `"癸"`, GanGui},
		{"甲-数字", `1`, Gan(1)},
		{"癸-数字", `10`, Gan(10)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g Gan
			if err := json.Unmarshal([]byte(tc.input), &g); err != nil {
				t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
			}
			if g != tc.want {
				t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, g, tc.want)
			}
		})
	}
}

func TestZhi_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Zhi
	}{
		{"子-字符串", `"子"`, ZhiZi},
		{"午-字符串", `"午"`, ZhiWu},
		{"亥-字符串", `"亥"`, ZhiHai},
		{"子-数字", `1`, Zhi(1)},
		{"亥-数字", `12`, Zhi(12)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var z Zhi
			if err := json.Unmarshal([]byte(tc.input), &z); err != nil {
				t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
			}
			if z != tc.want {
				t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, z, tc.want)
			}
		})
	}
}

func TestWuxing_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Wuxing
	}{
		{"木", `"木"`, WxMu},
		{"火", `"火"`, WxHuo},
		{"土", `"土"`, WxTu},
		{"金", `"金"`, WxJin},
		{"水", `"水"`, WxShui},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w Wuxing
			if err := json.Unmarshal([]byte(tc.input), &w); err != nil {
				t.Errorf("UnmarshalJSON(%s): %v", tc.input, err)
			}
			if w != tc.want {
				t.Errorf("UnmarshalJSON(%s)=%d, want %d", tc.input, w, tc.want)
			}
		})
	}
}

// -- Bazi.Validate --

func TestBazi_Validate_Valid(t *testing.T) {
	bz := Bazi{
		Nian: Zhu{Gan: GanJia, Zhi: ZhiZi},
		Yue:  Zhu{Gan: GanYi, Zhi: ZhiChou},
		Ri:   Zhu{Gan: GanBing, Zhi: ZhiYin},
		Shi:  Zhu{Gan: GanDing, Zhi: ZhiMao},
	}
	if err := bz.Validate(); err != nil {
		t.Errorf("valid Bazi should not error: %v", err)
	}
}

func TestBazi_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		bz   Bazi
	}{
		{"gan=0", Bazi{Nian: Zhu{Gan: 0, Zhi: ZhiZi}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"gan=11", Bazi{Nian: Zhu{Gan: 11, Zhi: ZhiZi}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"zhi=0", Bazi{Nian: Zhu{Gan: GanJia, Zhi: 0}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
		{"zhi=13", Bazi{Nian: Zhu{Gan: GanJia, Zhi: 13}, Yue: Zhu{Gan: GanYi, Zhi: ZhiChou}, Ri: Zhu{Gan: GanBing, Zhi: ZhiYin}, Shi: Zhu{Gan: GanDing, Zhi: ZhiMao}}},
	}
	for _, tc := range tests {
		if err := tc.bz.Validate(); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

func TestBazi_Slice(t *testing.T) {
	bz := Bazi{
		Nian: Zhu{Gan: GanJia, Zhi: ZhiZi},
		Yue:  Zhu{Gan: GanYi, Zhi: ZhiChou},
		Ri:   Zhu{Gan: GanBing, Zhi: ZhiYin},
		Shi:  Zhu{Gan: GanDing, Zhi: ZhiMao},
	}
	s := bz.Slice()
	if len(s) != 4 {
		t.Fatalf("Slice() len=%d, want 4", len(s))
	}
	names := []string{"年", "月", "日", "时"}
	for i, name := range names {
		if s[i].Gan == 0 || s[i].Zhi == 0 {
			t.Errorf("%s pillar is empty", name)
		}
	}
}

// ── WangShuai ──

func TestParseWangShuai(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    WangShuai
		wantErr bool
	}{
		{"旺", "旺", WSWang, false},
		{"相", "相", WSXiang, false},
		{"休", "休", WSXiu, false},
		{"囚", "囚", WSQiu, false},
		{"死", "死", WSSi, false},
		{"空字符串", "", 0, true},
		{"无效值", "invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWangShuai(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWangShuai(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("ParseWangShuai(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestWangShuai_String(t *testing.T) {
	tests := []struct {
		ws   WangShuai
		want string
	}{
		{WSWang, "旺"},
		{WSXiang, "相"},
		{WSXiu, "休"},
		{WSQiu, "囚"},
		{WSSi, "死"},
		{WangShuai(-1), ""},
		{WangShuai(99), ""},
	}
	for _, tt := range tests {
		if got := tt.ws.String(); got != tt.want {
			t.Errorf("WangShuai(%d).String() = %q, want %q", tt.ws, got, tt.want)
		}
	}
}

func TestWangShuaiOf(t *testing.T) {
	// 寅月(正月)=木 → 当令者旺(木), 我生者相(木生火), 生我者休(水生木), 克我者囚(金克木), 我克者死(木克土)
	t.Run("寅月", func(t *testing.T) {
		if got := WangShuaiOf(WxMu, ZhiYin); got != WSWang {
			t.Errorf("木在寅月 = %v, want 旺", got)
		}
		if got := WangShuaiOf(WxHuo, ZhiYin); got != WSXiang {
			t.Errorf("火在寅月 = %v, want 相", got)
		}
		if got := WangShuaiOf(WxShui, ZhiYin); got != WSXiu {
			t.Errorf("水在寅月 = %v, want 休", got)
		}
		if got := WangShuaiOf(WxJin, ZhiYin); got != WSQiu {
			t.Errorf("金在寅月 = %v, want 囚", got)
		}
		if got := WangShuaiOf(WxTu, ZhiYin); got != WSSi {
			t.Errorf("土在寅月 = %v, want 死", got)
		}
	})

	// 巳月(四月)=火 → 当令者旺(火), 我生者相(火生土), 生我者休(木生火), 克我者囚(水克火), 我克者死(火克金)
	t.Run("巳月", func(t *testing.T) {
		if got := WangShuaiOf(WxHuo, ZhiSi); got != WSWang {
			t.Errorf("火在巳月 = %v, want 旺", got)
		}
		if got := WangShuaiOf(WxTu, ZhiSi); got != WSXiang {
			t.Errorf("土在巳月 = %v, want 相", got)
		}
		if got := WangShuaiOf(WxMu, ZhiSi); got != WSXiu {
			t.Errorf("木在巳月 = %v, want 休", got)
		}
		if got := WangShuaiOf(WxShui, ZhiSi); got != WSQiu {
			t.Errorf("水在巳月 = %v, want 囚", got)
		}
		if got := WangShuaiOf(WxJin, ZhiSi); got != WSSi {
			t.Errorf("金在巳月 = %v, want 死", got)
		}
	})

	// 申月(七月)=金
	t.Run("申月", func(t *testing.T) {
		if got := WangShuaiOf(WxJin, ZhiShen); got != WSWang {
			t.Errorf("金在申月 = %v, want 旺", got)
		}
		if got := WangShuaiOf(WxShui, ZhiShen); got != WSXiang {
			t.Errorf("水在申月 = %v, want 相", got)
		}
		if got := WangShuaiOf(WxTu, ZhiShen); got != WSXiu {
			t.Errorf("土在申月 = %v, want 休", got)
		}
		if got := WangShuaiOf(WxHuo, ZhiShen); got != WSQiu {
			t.Errorf("火在申月 = %v, want 囚", got)
		}
		if got := WangShuaiOf(WxMu, ZhiShen); got != WSSi {
			t.Errorf("木在申月 = %v, want 死", got)
		}
	})

	// 亥月(十月)=水
	t.Run("亥月", func(t *testing.T) {
		if got := WangShuaiOf(WxShui, ZhiHai); got != WSWang {
			t.Errorf("水在亥月 = %v, want 旺", got)
		}
		if got := WangShuaiOf(WxMu, ZhiHai); got != WSXiang {
			t.Errorf("木在亥月 = %v, want 相", got)
		}
		if got := WangShuaiOf(WxJin, ZhiHai); got != WSXiu {
			t.Errorf("金在亥月 = %v, want 休", got)
		}
		if got := WangShuaiOf(WxTu, ZhiHai); got != WSQiu {
			t.Errorf("土在亥月 = %v, want 囚", got)
		}
		if got := WangShuaiOf(WxHuo, ZhiHai); got != WSSi {
			t.Errorf("火在亥月 = %v, want 死", got)
		}
	})

	// 丑月(十二月)=土
	t.Run("丑月", func(t *testing.T) {
		if got := WangShuaiOf(WxTu, ZhiChou); got != WSWang {
			t.Errorf("土在丑月 = %v, want 旺", got)
		}
	})

	// Edge: invalid month branch
	t.Run("无效月支", func(t *testing.T) {
		if got := WangShuaiOf(WxMu, 0); got != -1 {
			t.Errorf("WangShuaiOf(木, 0) = %d, want -1", got)
		}
		if got := WangShuaiOf(WxMu, 13); got != -1 {
			t.Errorf("WangShuaiOf(木, 13) = %d, want -1", got)
		}
	})
}
