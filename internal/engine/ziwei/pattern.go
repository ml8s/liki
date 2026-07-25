package ziwei

type pattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Score       int    `json:"score"` // 0=下 1=中 2=上
}

func findPatterns(palaces [12]palace) []pattern {
	var p patterns
	ming := palaces[0]
	sf := sanFang(0)

	// --- A.1 命宫三维 ---
	if starWith(ming, ZiWei) {
		p.add("紫微朝垣", "紫微坐命，帝王气象", 2)
	}
	if starAtZhi(palaces, TaiYang, 7) { // 太阳居午宫（绝对支位午=7）
		p.add("日丽中天", "太阳居午，光明磊落", 2)
	}
	if starAtZhi(palaces, TaiYin, 12) { // 太阴居亥宫（绝对支位亥=12）
		p.add("月朗天门", "太阴居亥，清辉遍洒", 2)
	}
	if sunMoonBright(palaces) {
		p.add("日月并明", "太阳太阴双双庙旺，阴阳调和", 2)
	}
	if sunMoonDark(palaces) {
		p.add("日月反背", "太阳太阴双双落陷，光辉不显", 0)
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
	if starWith(ming, PoJun) && isMiao(ming.Zhi, PoJun) {
		p.add("雄宿乾元", "破军入庙，英雄独断", 2)
	}

	// --- A.2 命宫三方 ---
	if anyInSF(palaces, sf, TianFu) || anyInSF(palaces, sf, TianXiang) {
		p.add("府相朝垣", "天府或天相在三方拱照命宫", 1)
	}
	if anyInSF(palaces, sf, QiSha) && anyInSF(palaces, sf, PoJun) && anyInSF(palaces, sf, TanLang) {
		p.add("杀破狼", "七杀破军贪狼会聚命宫三方", 2)
	}
	if sfCount(palaces, sf, TianJi, TaiYin, TianTong, TianLiang) >= 3 {
		p.add("机月同梁", "天机太阴天同天梁汇聚命宫三方", 1)
	}
	if anyInSF(palaces, sf, WenChang) && anyInSF(palaces, sf, WenQu) {
		p.add("文星拱命", "文昌文曲在命宫三方", 1)
	}
	if anyInSF(palaces, sf, ZuoFu) && anyInSF(palaces, sf, YouBi) {
		p.add("辅弼拱主", "左辅右弼在命宫三方", 2)
	}
	// 夹: 命宫 index 0, 左右是 1(兄弟) 和 11(父母)
	if starAt(palaces[1], TianKui) && starAt(palaces[11], TianYue) {
		p.add("魁钺夹命", "天魁天钺夹命，贵人扶持", 2)
	}
	if starAt(palaces[1], ZuoFu) && starAt(palaces[11], YouBi) {
		p.add("左右夹命", "左辅右弼夹命，助力环绕", 2)
	}
	if sfSiHuaCount(palaces, sf, HuaLu) >= 2 {
		p.add("双禄朝垣", "两颗化禄在命宫三方，财禄丰厚", 2)
	}
	if starInSF(palaces, sf, LuCun) && starInSF(palaces, sf, TianMa) {
		p.add("禄马交驰", "禄存天马会聚命宫三方", 1)
	}
	if anyInSF(palaces, sf, TaiYang) && anyInSF(palaces, sf, TianLiang) && anyInSF(palaces, sf, WenChang) &&
		sfSiHuaCount(palaces, sf, HuaLu) >= 1 {
		p.add("阳梁昌禄", "太阳天梁文昌会照，化禄入命", 2)
	}

	// --- A.3 财帛官禄 ---
	if sfSiHuaCount(palaces, sanFang(4), HuaLu) >= 1 {
		p.add("财荫夹印", "财帛宫有化禄拱照", 1)
	}
	if starAt(palaces[8], TaiYang) && isMiao(palaces[8].Zhi, TaiYang) {
		p.add("金灿光辉", "太阳在官禄庙旺，功名显达", 1)
	}

	// --- A.4 单星 ---
	if starWith(ming, QingYang) && qingYangMiao(ming.Zhi) {
		p.add("擎羊入庙", "擎羊入庙，刚毅有威", 1)
	}

	return p.list
}

type patterns struct{ list []pattern }

func (p *patterns) add(name, desc string, score int) {
	p.list = append(p.list, pattern{name, desc, score})
}

// ------ helpers ------

func starAt(pa palace, star starIndex) bool { return starWith(pa, star) }
func starAtZhi(palaces [12]palace, star starIndex, zhi Zhi) bool {
	for _, p := range palaces {
		if p.Zhi == zhi && starWith(p, star) {
			return true
		}
	}
	return false
}
func starWith(pa palace, star starIndex) bool {
	for _, s := range pa.Stars {
		if s.Star == star {
			return true
		}
	}
	return false
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

func sunMoonBright(palaces [12]palace) bool {
	sunBright := false
	moonBright := false
	brightPalaces := []palaceIndex{0, 6, 8, 10} // 命迁移官禄福德
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

func sunMoonDark(palaces [12]palace) bool {
	sunDark := false
	moonDark := false
	for _, p := range palaces {
		for _, s := range p.Stars {
			if s.Star == TaiYang && miaoWang(TaiYang, p.Zhi) == Xian {
				sunDark = true
			}
			if s.Star == TaiYin && miaoWang(TaiYin, p.Zhi) == Xian {
				moonDark = true
			}
		}
	}
	return sunDark && moonDark
}

func sanFang(ming palaceIndex) [4]palaceIndex {
	return [4]palaceIndex{ming, (ming + 4) % 12, (ming + 8) % 12, (ming + 6) % 12}
}

func anyInSF(bz [12]palace, sf [4]palaceIndex, star starIndex) bool {
	for _, pi := range sf {
		if starAt(bz[pi], star) {
			return true
		}
	}
	return false
}

func starInSF(bz [12]palace, sf [4]palaceIndex, star starIndex) bool {
	return anyInSF(bz, sf, star)
}

func sfCount(bz [12]palace, sf [4]palaceIndex, stars ...starIndex) int {
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

func sfSiHuaCount(bz [12]palace, sf [4]palaceIndex, h siHuaType) int {
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
