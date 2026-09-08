package qimen

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// YongShenSymbol 表示一个奇门事象用神符号（门/星/神/干）。
type YongShenSymbol struct {
	Door   *DoorIndex   `json:"door,omitempty"`   // 事象门
	Star   *StarIndex   `json:"star,omitempty"`   // 事象星
	Spirit *SpiritIndex `json:"spirit,omitempty"` // 事象神
	Gan    *ganzhi.Gan  `json:"stem,omitempty"`   // 事象干
}

// ParseYongShen 解析单个用神符号（门/星/神/干名）。
func ParseYongShen(s string) (YongShenSymbol, error) {
	if d, ok := parseDoor(s); ok {
		return YongShenSymbol{Door: &d}, nil
	}
	if st, ok := parseStar(s); ok {
		return YongShenSymbol{Star: &st}, nil
	}
	if sp, ok := parseSpirit(s); ok {
		return YongShenSymbol{Spirit: &sp}, nil
	}
	if g, ok := parseGan(s); ok {
		if g == ganzhi.GanJia {
			return YongShenSymbol{}, fmt.Errorf("甲遁不露盘，不能作为独立用神；请传入日干/时干/年命或对应六仪")
		}
		return YongShenSymbol{Gan: &g}, nil
	}
	return YongShenSymbol{}, fmt.Errorf("未知用神符号 %q，可选：门、星、神、干；甲遁不露盘", s)
}

func parseDoor(s string) (DoorIndex, bool) {
	// 门名格式："开门"/"生门"等（带门字）
	for d := DoorXiu; d <= DoorZhong; d++ {
		if d.String() == s {
			return d, true
		}
	}
	return 0, false
}

func parseStar(s string) (StarIndex, bool) {
	for st := StarTianPeng; st <= StarTianYing; st++ {
		if st.String() == s {
			return st, true
		}
	}
	return 0, false
}

func parseSpirit(s string) (SpiritIndex, bool) {
	for sp := SpiritZhiFu; sp <= SpiritTaiChang; sp++ {
		if sp.YangName() == s || sp.YinName() == s {
			return sp, true
		}
	}
	return 0, false
}

func parseGan(s string) (ganzhi.Gan, bool) {
	for g := ganzhi.GanJia; g <= ganzhi.GanGui; g++ {
		if g.String() == s {
			return g, true
		}
	}
	return 0, false
}

// SymbolResult 单个用神符号的落宫状态。
type SymbolResult struct {
	Symbol         string       `json:"symbol"`
	Palace         GongIndex    `json:"palace"`
	TianGan        []string     `json:"tian_gan"`
	KongWangBranch []ganzhi.Zhi `json:"kong_wang_branches"`
	MaXingBranch   *ganzhi.Zhi  `json:"ma_xing_branch,omitempty"`
}

// YongShenResult 奇门用神领域对象（用神符号组合落宫状态 + 年命干）。
// 求测人定位（日干/时干落宫、生克）由排盘固有字段提供，见 Chart 顶层字段。
type YongShenResult struct {
	NianGanPalace *GongIndex     `json:"nian_gan_gong,omitempty"` // 年命干落宫（需 birth_date；甲遁看六仪遁宫）
	Symbols       []SymbolResult `json:"symbols"`                 // 用神符号组合落宫状态
}

// jiaDunLiuYi 甲遁：六甲 → 六仪（甲子遁戊/甲戌遁己/甲申遁庚/甲午遁辛/甲辰遁壬/甲寅遁癸）。
// 传入遁甲所依的地支（年支/日支/时支），返回对应的六仪。
func jiaDunLiuYi(zhi ganzhi.Zhi) (ganzhi.Gan, bool) {
	for i, z := range liuJiaZhi {
		if z == zhi {
			return liuJiaLiuYi[i], true
		}
	}
	return 0, false
}

// resolveJiaDunGan 若天干为甲，按所依地支（年/日/时支）遁入六仪。
func resolveJiaDunGan(gan ganzhi.Gan, zhi ganzhi.Zhi) ganzhi.Gan {
	if gan == ganzhi.GanJia {
		if liuYi, ok := jiaDunLiuYi(zhi); ok {
			return liuYi
		}
	}
	return gan
}

// resolveNianGan 出生年份 → 年命干落宫（甲年命遁六仪）。
func resolveNianGan(chart Chart, birthDate BirthDate) *GongIndex {
	nian := tianwen.NianZhu(tianwen.GregorianTime(birthDate.Time))
	if nian.Gan == 0 {
		return nil
	}
	gan := resolveJiaDunGan(nian.Gan, nian.Zhi)
	palace := findGanPalaceIdx(chart.Pan, gan)
	if palace > 0 {
		return &palace
	}
	return nil
}

// symbolName 符号名称。
func symbolName(s YongShenSymbol, chart Chart) string {
	switch {
	case s.Door != nil:
		return s.Door.String()
	case s.Star != nil:
		return s.Star.String()
	case s.Spirit != nil:
		return spiritDisplayName(*s.Spirit, chart.Pan.YinDun, chart.Pan.School)
	case s.Gan != nil:
		return s.Gan.String()
	}
	return ""
}

// computeYongShenSymbols 聚合奇门用神（求测人 + 用神符号组合落宫）。
func computeYongShenSymbols(chart Chart, syms []YongShenSymbol) YongShenResult {
	return computeYongShen(chart, syms, BirthDate{}, false)
}

// computeYongShenWithBirth 聚合奇门用神，并按精确立春计算年命干落宫。
func computeYongShenWithBirth(chart Chart, syms []YongShenSymbol, birthDate BirthDate) YongShenResult {
	return computeYongShen(chart, syms, birthDate, true)
}

func computeYongShen(chart Chart, syms []YongShenSymbol, birthDate BirthDate, hasBirth bool) YongShenResult {
	ys := YongShenResult{Symbols: []SymbolResult{}}

	// 用神符号组合：逐个定位落宫取状态
	for _, s := range syms {
		sr := SymbolResult{
			Symbol: symbolName(s, chart), TianGan: []string{}, KongWangBranch: []ganzhi.Zhi{},
		}
		switch {
		case s.Door != nil:
			sr.Palace = findDoorPalaceIdx(chart.Pan, *s.Door)
		case s.Star != nil:
			sr.Palace = findStarPalace(chart.Pan, *s.Star)
		case s.Spirit != nil:
			sr.Palace = findSpiritPalaceIdx(chart.Pan, *s.Spirit)
		case s.Gan != nil:
			sr.Palace = findGanPalaceIdx(chart.Pan, *s.Gan)
		}
		if sr.Palace > 0 {
			for _, void := range chart.Pan.KongWang {
				if void.Gong == sr.Palace {
					sr.KongWangBranch = append(sr.KongWangBranch, void.Branch)
				}
			}
			if chart.Pan.MaXing.Gong == sr.Palace {
				branch := chart.Pan.MaXing.Branch
				sr.MaXingBranch = &branch
			}
			if int(sr.Palace) >= 1 && int(sr.Palace) <= 9 {
				for _, item := range chart.Pan.GongWei[sr.Palace-1].TianPan {
					sr.TianGan = append(sr.TianGan, item.Gan.String())
				}
				if pointer := chart.Pan.GongWei[sr.Palace-1].TianPanGan; pointer != nil {
					sr.TianGan = append(sr.TianGan, pointer.String())
				}
			}
		}
		ys.Symbols = append(ys.Symbols, sr)
	}

	// 年命干落宫（需出生年份；甲年命遁六仪）
	if hasBirth && birthDate.Has {
		ys.NianGanPalace = resolveNianGan(chart, birthDate)
	}

	return ys
}

func findSpiritPalaceIdx(p pan, s SpiritIndex) GongIndex {
	for i, pg := range p.GongWei {
		if pg.Spirit == s {
			return GongIndex(i + 1)
		}
	}
	return 0
}
