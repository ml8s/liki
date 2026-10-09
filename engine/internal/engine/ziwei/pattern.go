package ziwei

type pattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func findPatterns(palaces [12]gong) []pattern {
	var p patterns
	ming := palaces[0]
	sf := sanFang(0)

	// --- A.1 命宫三维 ---
	if starWith(ming, ZiWei) &&
		sfDistinctCount(palaces, sf, ZuoFu, YouBi, TianKui, TianYue, WenChang, WenQu, LuCun) >= 2 {
		p.add("紫微朝垣", "紫微坐命百官朝拱，帝王气象", 2)
	}
	if ming.Zhi == 7 && starWith(ming, TaiYang) {
		p.add("日丽中天", "太阳居午，光明磊落", 2)
	}
	if ming.Zhi == 12 && starWith(ming, TaiYin) {
		p.add("月朗天门", "太阴居亥，清辉遍洒", 2)
	}
	if sunMoonBright(palaces) {
		p.add("日月并明", "太阳太阴双双庙旺，阴阳调和", 2)
	}
	if sunMoonDark(palaces) {
		p.add("日月反背", "太阳太阴按日戌月辰或日子月午反背，勿概以双落陷论", 0)
	}
	if starWith(ming, HuoXing) && starWith(ming, TanLang) && !isXian(ming.Zhi, TanLang) {
		p.add("火贪格", "火星贪狼同宫，爆发之格", 2)
	}
	if starWith(ming, LingXing) && starWith(ming, TanLang) && !isXian(ming.Zhi, TanLang) {
		p.add("铃贪格", "铃星贪狼同宫，暗发之格", 2)
	}
	if starWith(ming, JuMen) && starWith(ming, TaiYang) {
		p.add("巨日同宫", "巨门太阳同宫，照破暗昧", 1)
	}
	if starWith(ming, JuMen) && starWith(ming, TianJi) {
		p.add("巨机同宫", "巨门天机同宫，心思缜密", 1)
	}
	if starWith(ming, LianZhen) && starWith(ming, QiSha) {
		p.add("廉杀同宫", "廉贞七杀同宫，将星得地", 1)
	}
	if starWith(ming, LianZhen) && (ming.Zhi == 8 || ming.Zhi == 9) && noSixMalefic(ming) {
		p.add("雄宿朝元", "廉贞居未申无六煞，雄宿朝元", 2)
	}

	// --- A.2 命宫三方 ---
	if !starWith(ming, TianFu) && !starWith(ming, TianXiang) &&
		anyInSF(palaces, sf, TianFu) && anyInSF(palaces, sf, TianXiang) {
		p.add("府相朝垣", "天府或天相在三方拱照命宫", 1)
	}
	if anyInSF(palaces, sf, QiSha) && anyInSF(palaces, sf, PoJun) && anyInSF(palaces, sf, TanLang) {
		p.add("杀破狼", "七杀破军贪狼会聚命宫三方", 2)
	}
	if sfCount(palaces, sf, TianJi, TaiYin, TianTong, TianLiang) == 4 {
		p.add("机月同梁", "天机太阴天同天梁汇聚命宫三方", 1)
	}
	if anyInSF(palaces, sf, WenChang) && anyInSF(palaces, sf, WenQu) {
		p.add("文星拱命", "文昌文曲在命宫三方", 1)
	}
	if anyInSF(palaces, sf, ZuoFu) && anyInSF(palaces, sf, YouBi) {
		p.add("辅弼拱主", "左辅右弼在命宫三方", 2)
	}
	// 夹: 命宫 index 0, 左右是 1(兄弟) 和 11(父母)
	if (starAt(palaces[1], TianKui) && starAt(palaces[11], TianYue)) ||
		(starAt(palaces[1], TianYue) && starAt(palaces[11], TianKui)) {
		p.add("魁钺夹命", "天魁天钺夹命，贵人扶持", 2)
	}
	if (starAt(palaces[1], ZuoFu) && starAt(palaces[11], YouBi)) ||
		(starAt(palaces[1], YouBi) && starAt(palaces[11], ZuoFu)) {
		p.add("左右夹命", "左辅右弼夹命，助力环绕", 2)
	}
	if sfLuCount(palaces, sf) >= 2 {
		p.add("双禄朝垣", "禄存与化禄会命宫三方，财禄丰厚", 2)
	}
	if starInSF(palaces, sf, LuCun) && starInSF(palaces, sf, TianMa) {
		p.add("禄马交驰", "禄存天马会聚命宫三方", 1)
	}
	if anyInSF(palaces, sf, TaiYang) && anyInSF(palaces, sf, TianLiang) && anyInSF(palaces, sf, WenChang) &&
		(sfSiHuaCount(palaces, sf, HuaLu) >= 1 || anyInSF(palaces, sf, LuCun)) {
		p.add("阳梁昌禄", "太阳天梁文昌会照，化禄入命", 2)
	}

	// --- A.3 财荫夹印 ---
	if hasCaiYinJiaYin(palaces) {
		p.add("财荫夹印", "天相被禄财与天梁相夹，财荫护印", 1)
	}
	if majorCount(ming) == 1 && starWith(ming, TaiYang) && ming.Zhi == 7 && isMiao(ming.Zhi, TaiYang) {
		p.add("金灿光辉", "太阳独坐命宫午宫，光明磊落", 1)
	}

	// --- A.4 单星 ---
	if starWith(ming, QingYang) && qingYangMiao(ming.Zhi) {
		p.add("擎羊入庙", "擎羊入庙，刚毅有威", 1)
	}

	// --- A.5 巨火羊: 巨门+火星+擎羊同宫，是非官非 ---
	if starWith(ming, JuMen) && starWith(ming, HuoXing) && starWith(ming, QingYang) {
		p.add("巨火羊", "巨门火星擎羊同宫，是非官非", 0)
	}

	// --- A.6 刑忌夹印: 天相被化忌星+天刑夹制 ---
	if !starWith(palaces[0], TianXiang) {
		return p.list
	}
	prev, next := palaces[11], palaces[1]

	prevHasJi, nextHasJi := false, false
	for _, s := range prev.Stars {
		if s.SiHua == "忌" {
			prevHasJi = true
		}
	}
	for _, s := range next.Stars {
		if s.SiHua == "忌" {
			nextHasJi = true
		}
	}

	prevHasXing, nextHasXing := false, false
	for _, zy := range prev.ZaYao {
		if zy == "天刑" {
			prevHasXing = true
		}
	}
	for _, zy := range next.ZaYao {
		if zy == "天刑" {
			nextHasXing = true
		}
	}

	// 必须分居两侧才算"夹"
	huaJiFound := (prevHasJi && !nextHasJi) || (nextHasJi && !prevHasJi)
	tianXingFound := (prevHasXing && !nextHasXing) || (nextHasXing && !prevHasXing)
	_ = tianXingFound

	oppositeSides := (prevHasJi && nextHasXing) || (nextHasJi && prevHasXing)
	if huaJiFound && tianXingFound && oppositeSides {
		p.add("刑忌夹印", "命宫天相被化忌天刑夹制，受制受拖累", 0)
	}

	return p.list
}

type patterns struct{ list []pattern }

func (p *patterns) add(name, desc string, _ int) {
	p.list = append(p.list, pattern{name, desc})
}

// ------ helpers ------

func starAt(pa gong, star starIndex) bool { return starWith(pa, star) }
func starWith(pa gong, star starIndex) bool {
	for _, s := range pa.Stars {
		if s.Star == star {
			return true
		}
	}
	return false
}

func majorCount(pa gong) int {
	count := 0
	for _, item := range pa.Stars {
		if mainStars[item.Star] {
			count++
		}
	}
	return count
}

func noSixMalefic(pa gong) bool {
	for _, item := range pa.Stars {
		switch item.Star {
		case QingYang, TuoLuo, HuoXing, LingXing, DiKong, DiJie:
			return false
		}
	}
	return true
}

func sfLuCount(bz [12]gong, sf [4]gongIndex) int {
	count := 0
	for _, pi := range sf {
		for _, s := range bz[pi].Stars {
			if s.Star == LuCun || s.SiHua == string(HuaLu) {
				count++
			}
		}
	}
	return count
}

func hasCaiYinJiaYin(palaces [12]gong) bool {
	if !starWith(palaces[0], TianXiang) {
		return false
	}
	prev, next := palaces[11], palaces[1]
	hasLuCai := func(pa gong) bool {
		for _, item := range pa.Stars {
			if item.Star == LuCun || item.SiHua == string(HuaLu) {
				return true
			}
		}
		return false
	}
	return (hasLuCai(prev) && starWith(next, TianLiang)) ||
		(starWith(prev, TianLiang) && hasLuCai(next))
}

func isMiao(z Zhi, star starIndex) bool { return miaoWang(star, z) <= Wang }
func isXian(z Zhi, star starIndex) bool { return miaoWang(star, z) == Xian }

func qingYangMiao(z Zhi) bool {
	switch z {
	case 5, 11, 2, 8: // 辰戌丑未=5,11,2,8
		return true
	}
	return false
}

func sunMoonBright(palaces [12]gong) bool {
	sunBright := false
	moonBright := false
	brightPalaces := []gongIndex{0, 4, 6, 8} // 命财官迁（三方四正）
	for _, bp := range brightPalaces {
		for _, s := range palaces[bp].Stars {
			if s.Star == TaiYang && miaoWang(TaiYang, palaces[bp].Zhi) <= Wang {
				sunBright = true
			}
			if s.Star == TaiYin && miaoWang(TaiYin, palaces[bp].Zhi) <= Wang {
				moonBright = true
			}
		}
	}
	return sunBright && moonBright
}

func sunMoonDark(palaces [12]gong) bool {
	// 《紫微斗数全书·卷三》：“日月最嫌反背……若反背日戌月辰，子月午。”
	// 这里只取原文明确结构，不把命宫三方四正任意双落陷泛化为反背。
	var sunZhis, moonZhis []Zhi
	for _, cp := range []gongIndex{0, 4, 6, 8} { // 命财官迁（三方四正）
		for _, s := range palaces[cp].Stars {
			if s.Star == TaiYang {
				sunZhis = append(sunZhis, palaces[cp].Zhi)
			}
			if s.Star == TaiYin {
				moonZhis = append(moonZhis, palaces[cp].Zhi)
			}
		}
	}
	for _, sun := range sunZhis {
		for _, moon := range moonZhis {
			if (sun == zhiXu && moon == zhiChen) || (sun == zhiZi && moon == zhiWu) {
				return true
			}
		}
	}
	return false
}

func sanFang(ming gongIndex) [4]gongIndex {
	return [4]gongIndex{ming, (ming + 4) % 12, (ming + 8) % 12, (ming + 6) % 12}
}

func anyInSF(bz [12]gong, sf [4]gongIndex, star starIndex) bool {
	for _, pi := range sf {
		if starAt(bz[pi], star) {
			return true
		}
	}
	return false
}

func starInSF(bz [12]gong, sf [4]gongIndex, star starIndex) bool {
	return anyInSF(bz, sf, star)
}

func sfCount(bz [12]gong, sf [4]gongIndex, stars ...starIndex) int {
	count := 0
	for _, pi := range sf {
		for _, s := range stars {
			if starAt(bz[pi], s) {
				count++
			}
		}
	}
	return count
}

func sfDistinctCount(bz [12]gong, sf [4]gongIndex, stars ...starIndex) int {
	count := 0
	for _, star := range stars {
		if anyInSF(bz, sf, star) {
			count++
		}
	}
	return count
}

func sfSiHuaCount(bz [12]gong, sf [4]gongIndex, h siHuaType) int {
	count := 0
	for _, pi := range sf {
		for _, s := range bz[pi].Stars {
			if s.SiHua == string(h) {
				count++
			}
		}
	}
	return count
}
