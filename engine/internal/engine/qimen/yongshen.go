package qimen

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// YongShenSymbol 用神符号（奇门事象用神：门/星/神/干，35 种封闭）。
// LLM 读 skill 规则（domains/qimen/yongshen.md）按问事确定符号组合。
type YongShenSymbol struct {
	Door   *DoorIndex   `json:"door,omitempty"`   // 事象门
	Star   *StarIndex   `json:"star,omitempty"`   // 事象星
	Spirit *SpiritIndex `json:"spirit,omitempty"` // 事象神
	Stem   *ganzhi.Gan  `json:"stem,omitempty"`   // 事象干
	Raw    string       `json:"-"`                // 用户原始输入（门/星/神/干名），保留展示原样
}

// ParseYongShen 解析单个用神符号（门/星/神/干名）。
func ParseYongShen(s string) (YongShenSymbol, error) {
	if d, ok := parseDoor(s); ok {
		return YongShenSymbol{Door: &d, Raw: s}, nil
	}
	if st, ok := parseStar(s); ok {
		return YongShenSymbol{Star: &st, Raw: s}, nil
	}
	if sp, ok := parseSpirit(s); ok {
		return YongShenSymbol{Spirit: &sp, Raw: s}, nil
	}
	if g, ok := parseGan(s); ok {
		return YongShenSymbol{Stem: &g, Raw: s}, nil
	}
	return YongShenSymbol{}, fmt.Errorf("未知用神符号 %q，可选：门(休门/生门/伤门/杜门/景门/死门/惊门/开门)、星(天蓬/天芮/天冲/天辅/天禽/天心/天柱/天任/天英)、神(值符/螣蛇/太阴/六合/勾陈/朱雀/九地/九天，阴遁白虎/玄武)、干(甲/乙/丙/丁/戊/己/庚/辛/壬/癸)", s)
}

func parseDoor(s string) (DoorIndex, bool) {
	// 门名格式："开门"/"生门"等（带门字）
	for _, d := range []DoorIndex{DoorXiu, DoorSheng, DoorShang, DoorDu, DoorJing, DoorSi, DoorJingMen, DoorKai} {
		if d.String()+"门" == s {
			return d, true
		}
	}
	return 0, false
}

func parseStar(s string) (StarIndex, bool) {
	for _, st := range []StarIndex{StarTianPeng, StarTianRui, StarTianChong, StarTianFu, StarTianQin, StarTianXin, StarTianZhu, StarTianRen, StarTianYing} {
		if st.String() == s {
			return st, true
		}
	}
	return 0, false
}

func parseSpirit(s string) (SpiritIndex, bool) {
	for _, sp := range []SpiritIndex{SpiritZhiFu, SpiritTengShe, SpiritTaiYin, SpiritLiuHe, SpiritGouChen, SpiritZhuQue, SpiritJiuDi, SpiritJiuTian} {
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
	Symbol   string    `json:"symbol"`    // 符号名（开门/天辅/六合/戊）
	Palace   GongIndex `json:"palace"`    // 符号落宫
	TianGan  string    `json:"tian_gan"`  // 落宫天盘干（十干克应）
	KongWang bool      `json:"kong_wang"` // 落宫是否空亡
	MaXing   bool      `json:"ma_xing"`   // 落宫是否马星
}

// YongShenResult 奇门用神领域对象（用神符号组合落宫状态 + 年命干）。
// 求测人定位（日干/时干落宫、生克）由排盘固有字段提供，见 Chart 顶层字段。
type YongShenResult struct {
	NianGanPalace *GongIndex     `json:"nian_gan_gong,omitempty"` // 年命干落宫（需 birth_year；甲遁看六仪遁宫）
	Symbols       []SymbolResult `json:"symbols"`                 // 用神符号组合落宫状态
}

// liuJiaLiuYi 六甲 → 六仪（甲遁）。顺序：甲子/甲戌/甲申/甲午/甲辰/甲寅 → 戊/己/庚/辛/壬/癸。
var liuJiaZhi = [6]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiXu, ganzhi.ZhiShen, ganzhi.ZhiWu, ganzhi.ZhiChen, ganzhi.ZhiYin}

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
func resolveNianGan(chart Chart, birthYear int) *GongIndex {
	nian := tianwen.NianZhu(tianwen.GregorianTime(time.Date(birthYear, 2, 15, 0, 0, 0, 0, time.UTC)))
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
func symbolName(s YongShenSymbol) string {
	// 优先返回用户原始输入（保留"白虎/玄武"等阴遁名原样）。
	if s.Raw != "" {
		return s.Raw
	}
	switch {
	case s.Door != nil:
		return s.Door.String() + "门"
	case s.Star != nil:
		return s.Star.String()
	case s.Spirit != nil:
		return s.Spirit.YangName()
	case s.Stem != nil:
		return s.Stem.String()
	}
	return ""
}

// ComputeYongShen 聚合奇门用神（求测人 + 用神符号组合落宫）。
func ComputeYongShen(chart Chart, syms []YongShenSymbol) YongShenResult {
	return computeYongShen(chart, syms, 0, false)
}

// ComputeYongShenWithBirth 聚合奇门用神，并计算年命干落宫（需出生年份）。
func ComputeYongShenWithBirth(chart Chart, syms []YongShenSymbol, birthYear int) YongShenResult {
	return computeYongShen(chart, syms, birthYear, true)
}

func computeYongShen(chart Chart, syms []YongShenSymbol, birthYear int, hasBirth bool) YongShenResult {
	ys := YongShenResult{}

	// 用神符号组合：逐个定位落宫取状态
	for _, s := range syms {
		sr := SymbolResult{Symbol: symbolName(s)}
		switch {
		case s.Door != nil:
			sr.Palace = findDoorPalaceIdx(chart.Pan, *s.Door)
		case s.Star != nil:
			sr.Palace = findStarPalaceIdx(chart.Pan, *s.Star)
		case s.Spirit != nil:
			sr.Palace = findSpiritPalaceIdx(chart.Pan, *s.Spirit)
		case s.Stem != nil:
			// 用神干为甲时甲遁于六仪（按日支遁），与日干处理一致；否则直接找落宫。
			gan := resolveJiaDunGan(*s.Stem, chart.Pan.RiZhi)
			sr.Palace = findGanPalaceIdx(chart.Pan, gan)
		}
		if sr.Palace > 0 {
			for _, k := range chart.Pan.KongWang {
				if k == sr.Palace {
					sr.KongWang = true
					break
				}
			}
			sr.MaXing = sr.Palace == chart.Pan.MaXing
			if int(sr.Palace) >= 1 && int(sr.Palace) <= 9 {
				sr.TianGan = chart.Pan.GongWei[sr.Palace-1].HeavenStem.String()
			}
		}
		ys.Symbols = append(ys.Symbols, sr)
	}

	// 年命干落宫（需出生年份；甲年命遁六仪）
	if hasBirth && birthYear > 0 {
		ys.NianGanPalace = resolveNianGan(chart, birthYear)
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
