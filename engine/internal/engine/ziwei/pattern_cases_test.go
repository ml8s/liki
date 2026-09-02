package ziwei

import "testing"

func blankPalaces() [12]gong {
	var palaces [12]gong
	for i := range palaces {
		palaces[i] = gong{Index: gongIndex(i), Name: gongLabels[i]}
	}
	return palaces
}

func firstBrightZhi(star starIndex) Zhi {
	for zhi := Zhi(1); zhi <= 12; zhi++ {
		if miaoWang(star, zhi) <= Wang {
			return zhi
		}
	}
	panic("no bright palace for star")
}

func TestFuXiangChaoyuanRequiresBothStars(t *testing.T) {
	palaces := blankPalaces()
	palaces[4].Stars = []starInfo{makeStar(TianFu, "天府", "")}
	palaces[8].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	if !hasPattern(findPatterns(palaces), "府相朝垣") {
		t.Error("府相朝垣 should match when 天府 and 天相 are present")
	}
}

func TestFuXiangChaoyuanOnlyFuShouldNotMatch(t *testing.T) {
	palaces := blankPalaces()
	palaces[4].Stars = []starInfo{makeStar(TianFu, "天府", "")}
	if hasPattern(findPatterns(palaces), "府相朝垣") {
		t.Error("府相朝垣 should require both 天府 and 天相")
	}
}

func TestFuXiangChaoyuanOnlyXiangShouldNotMatch(t *testing.T) {
	palaces := blankPalaces()
	palaces[4].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	if hasPattern(findPatterns(palaces), "府相朝垣") {
		t.Error("府相朝垣 should require both 天府 and 天相")
	}
}

func TestSunMoonBrightIncludesWealthAndTravelPalaces(t *testing.T) {
	palaces := blankPalaces()
	palaces[4].Zhi = firstBrightZhi(TaiYang)
	palaces[4].Stars = []starInfo{makeStar(TaiYang, "太阳", "")}
	palaces[6].Zhi = firstBrightZhi(TaiYin)
	palaces[6].Stars = []starInfo{makeStar(TaiYin, "太阴", "")}
	if !sunMoonBright(palaces) {
		t.Error("日月并明 should include bright sun and moon in wealth and travel palaces")
	}
}

func TestXingJiJiaYinSameSideShouldNotMatch(t *testing.T) {
	palaces := blankPalaces()
	palaces[3].Stars = []starInfo{makeStar(TianXiang, "天相", "")}
	palaces[2].Stars = []starInfo{makeStar(TianLiang, "天梁", "忌")}
	palaces[2].ZaYao = []string{"天刑"}
	if hasPattern(findPatterns(palaces), "刑忌夹印") {
		t.Error("刑忌夹印 should require 化忌 and 天刑 on opposite sides")
	}
}

func TestJinCanGuangHuiRequiresSunInWuPalace(t *testing.T) {
	palaces := blankPalaces()
	palaces[0].Zhi = Zhi(7)
	palaces[0].Stars = []starInfo{makeStar(TaiYang, "太阳", "")}
	if !hasPattern(findPatterns(palaces), "金灿光辉") {
		t.Error("金灿光辉 should match a sun alone in the ming palace at 午")
	}

	palaces[0].Zhi = Zhi(1)
	if hasPattern(findPatterns(palaces), "金灿光辉") {
		t.Error("金灿光辉 should require the ming palace at 午")
	}
}

func TestZiWeiChaoyuanRequiresAuxiliaryStars(t *testing.T) {
	palaces := blankPalaces()
	palaces[0].Stars = []starInfo{makeStar(ZiWei, "紫微", "")}
	if hasPattern(findPatterns(palaces), "紫微朝垣") {
		t.Error("紫微朝垣 should require auxiliary stars")
	}

	palaces[4].Stars = []starInfo{makeStar(ZuoFu, "左辅", "")}
	palaces[8].Stars = []starInfo{makeStar(YouBi, "右弼", "")}
	if !hasPattern(findPatterns(palaces), "紫微朝垣") {
		t.Error("紫微朝垣 should match with 左辅 and 右弼")
	}
}

func TestKuiYueJiaMatchesReverseDirection(t *testing.T) {
	palaces := blankPalaces()
	palaces[1].Stars = []starInfo{makeStar(TianYue, "天钺", "")}
	palaces[11].Stars = []starInfo{makeStar(TianKui, "天魁", "")}
	if !hasPattern(findPatterns(palaces), "魁钺夹命") {
		t.Error("魁钺夹命 should match both directions")
	}
}

func TestZuoYouJiaMatchesReverseDirection(t *testing.T) {
	palaces := blankPalaces()
	palaces[1].Stars = []starInfo{makeStar(YouBi, "右弼", "")}
	palaces[11].Stars = []starInfo{makeStar(ZuoFu, "左辅", "")}
	if !hasPattern(findPatterns(palaces), "左右夹命") {
		t.Error("左右夹命 should match both directions")
	}
}

func TestYangLiangChangLuAcceptsLuCun(t *testing.T) {
	palaces := blankPalaces()
	palaces[4].Stars = []starInfo{
		makeStar(TaiYang, "太阳", ""),
		makeStar(TianLiang, "天梁", ""),
		makeStar(WenChang, "文昌", ""),
		makeStar(LuCun, "禄存", ""),
	}
	if !hasPattern(findPatterns(palaces), "阳梁昌禄") {
		t.Error("阳梁昌禄 should accept 禄存")
	}
}
