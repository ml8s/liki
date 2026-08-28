package ziwei

import (
	"testing"
)

func makeGong(idx gongIndex, stars []starInfo, zaYao []string) gong {
	return gong{
		Index: idx,
		Name:  gongLabels[idx],
		Stars: stars,
		ZaYao: zaYao,
	}
}

func makeStar(star starIndex, name string, siHua string) starInfo {
	return starInfo{Star: star, Name: name, IsMajor: star < 14, SiHua: siHua}
}

func hasPattern(patterns []pattern, name string) bool {
	for _, p := range patterns {
		if p.Name == name {
			return true
		}
	}
	return false
}

func TestJuoHuoYang(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[0].Stars = []starInfo{
		makeStar(JuMen, "巨门", ""),
		makeStar(HuoXing, "火星", ""),
		makeStar(QingYang, "擎羊", ""),
	}
	result := findPatterns(palaces)
	if !hasPattern(result, "巨火羊") {
		t.Errorf("expected 巨火羊 pattern, got: %v", result)
	}
}

func TestJuoHuoYangNegative(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[0].Stars = []starInfo{
		makeStar(JuMen, "巨门", ""),
		makeStar(HuoXing, "火星", ""),
	}
	result := findPatterns(palaces)
	if hasPattern(result, "巨火羊") {
		t.Errorf("should NOT have 巨火羊 without 擎羊")
	}
}

func TestXingJiJiaYin(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[3].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	palaces[2].Stars = []starInfo{makeStar(TianLiang, "天梁", "忌")}
	palaces[4].ZaYao = []string{"天刑"}
	result := findPatterns(palaces)
	if !hasPattern(result, "刑忌夹印") {
		t.Errorf("expected 刑忌夹印 pattern, got: %v", result)
	}
}

func TestXingJiJiaYinReverse(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[3].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	palaces[2].ZaYao = []string{"天刑"}
	palaces[4].Stars = []starInfo{makeStar(TianLiang, "天梁", "忌")}
	result := findPatterns(palaces)
	if !hasPattern(result, "刑忌夹印") {
		t.Errorf("expected 刑忌夹印 (reverse), got: %v", result)
	}
}

func TestXingJiJiaYinNegative(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[3].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	palaces[2].Stars = []starInfo{makeStar(TianLiang, "天梁", "")}
	palaces[4].ZaYao = []string{"天刑"}
	result := findPatterns(palaces)
	if hasPattern(result, "刑忌夹印") {
		t.Errorf("should NOT have 刑忌夹印 without 化忌")
	}
}

func TestXingJiJiaYinNoTianXing(t *testing.T) {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = makeGong(gongIndex(i), nil, nil)
	}
	palaces[3].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	palaces[2].Stars = []starInfo{makeStar(TianLiang, "天梁", "忌")}
	result := findPatterns(palaces)
	if hasPattern(result, "刑忌夹印") {
		t.Errorf("should NOT have 刑忌夹印 without 天刑")
	}
}
