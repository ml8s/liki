package bazi

import "liki-engine/internal/engine/ganzhi"

// Shensha category constants.
const (
	catJi        = "吉"
	catXiong     = "凶"
	catZhongXing = "中性"
)

// shenShaEntry describes a single shensha hit on a pillar.
type shenShaEntry struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type ganZhiPair struct {
	gan ganzhi.Gan
	zhi ganzhi.Zhi
}

// Package-level lookup maps, populated from data/shensha.json via data.go init().
var (
	taohuaBranchMap map[ganzhi.Zhi]ganzhi.Zhi
	yimaBranchMap   map[ganzhi.Zhi]ganzhi.Zhi
	huagaiBranchMap map[ganzhi.Zhi]ganzhi.Zhi
	yangRenLookup   map[ganzhi.Gan]ganzhi.Zhi
	jieshaBranch    map[ganzhi.Zhi]ganzhi.Zhi
	zaishaBranch    map[ganzhi.Zhi]ganzhi.Zhi
	hongluanLookup  map[ganzhi.Zhi]ganzhi.Zhi
	tianxiLookup    map[ganzhi.Zhi]ganzhi.Zhi
)

var tianYiLookup map[ganzhi.Gan][]ganzhi.Zhi

var (
	tiandeTargets   map[ganzhi.Zhi][]tianDeTarget
	yuedeStem       map[ganzhi.Zhi]ganzhi.Gan
	jiangxingLookup map[ganzhi.Zhi]ganzhi.Zhi
	jinyuLookup     map[ganzhi.Gan][]ganzhi.Zhi
	yueEnStems      map[ganzhi.Zhi][]ganzhi.Gan
	xueRenLookup    map[ganzhi.Gan]ganzhi.Zhi
	tianLuoDiWang   map[ganzhi.Zhi]string
	shiEDaBai       map[int]struct{}
)

// tianDeTarget 天德贵人的匹配目标：天德可为天干型（如正月见丁）或地支型（如二月见申）。
type tianDeTarget struct {
	IsZhi bool
	Gan   ganzhi.Gan
	Zhi   ganzhi.Zhi
}

// computeShenSha computes all shensha for the bazi chart, grouped by pillar.
func computeShenSha(bz ganzhi.Bazi) [4][]shenShaEntry {
	riYuan := bz.Ri.Gan
	monthBranch := bz.Yue.Zhi
	zhus := bz.Slice()
	var out [4][]shenShaEntry
	branches := [4]ganzhi.Zhi{zhus[0].Zhi, zhus[1].Zhi, zhus[2].Zhi, zhus[3].Zhi}
	// seasonIdx（月支三会组）供四废（四时废日）使用：寅卯辰→0 巳午未→1 申酉戌→2 亥子丑→3。
	seasonIdx := (int(monthBranch) - 1) / 3
	yearBranch := zhus[0].Zhi
	// yearSanHuiIdx（年支三会组）供孤辰寡宿使用：寅卯辰→0 巳午未→1 申酉戌→2 亥子丑→3。
	// 孤辰寡宿以年支为准（亥子丑人见寅为孤、见戌为寡），与月支三会组不同。
	yearSanHuiIdx := ((int(yearBranch) - 3 + 12) % 12) / 3

	addTianYi(&out, bz, riYuan, zhus[0].Gan)
	addWenChang(&out, bz, riYuan)
	addXueTang(&out, bz, riYuan)
	addLuShen(&out, bz, riYuan)
	addYangRen(&out, bz, riYuan)
	addTianDe(&out, bz, monthBranch)
	addYueDe(&out, bz, monthBranch)
	addTaoHua(&out, bz, branches)
	addYiMa(&out, bz, branches)
	addHuaGai(&out, bz, branches)
	addJiangXing(&out, bz, branches)
	addJieSha(&out, bz, branches)
	addZaiSha(&out, bz, branches)
	addGuChenGuaSu(&out, bz, yearSanHuiIdx)
	addHongLuanTianXi(&out, bz, yearBranch)
	addJinYu(&out, bz, riYuan)
	addCiGuan(&out, bz, riYuan)
	addYueEn(&out, bz, monthBranch)
	addTianShe(&out, bz, monthBranch)
	addTianLuoDiWang(&out, bz)
	addGouJiao(&out, bz, yearBranch)
	addYuanChen(&out, bz, yearBranch, zhus[0].Gan)
	addXueRen(&out, bz, riYuan)
	addSiFei(&out, bz, seasonIdx)
	addShiEDaBai(&out, bz)

	return out
}

// addTianYi 天乙贵人：按日干与年干双查（兼容两种流派），同一柱只标注一次（去重）。
func addTianYi(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan, nianGan ganzhi.Gan) {
	marked := map[int]bool{}
	mark := func(gan ganzhi.Gan) {
		targets, ok := tianYiLookup[gan]
		if !ok {
			return
		}
		zhus := bz.Slice()
		for pi, p := range zhus {
			for _, t := range targets {
				if p.Zhi == t && !marked[pi] {
					marked[pi] = true
					(*out)[pi] = append((*out)[pi], shenShaEntry{
						Name: "天乙贵人", Category: catJi, Description: "主贵人相助，逢凶化吉",
					})
				}
			}
		}
	}
	mark(riYuan)
	mark(nianGan)
}

var wenChangLookup map[ganzhi.Gan][]ganzhi.Zhi

func addWenChang(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	appendShenShaByStemLookup(out, bz, riYuan, wenChangLookup, "文昌", catJi, "主学业、文书、才华")
}

func addXueTang(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	addChangShengShenSha(out, bz, riYuan, 0, "学堂", catJi, "日主长生之位，主学业聪颖")
}

func addLuShen(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	addChangShengShenSha(out, bz, riYuan, 3, "禄神", catJi, "日主临官之位，主福禄安康")
}

func addCiGuan(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	addChangShengShenSha(out, bz, riYuan, 3, "词馆", catJi, "主文章、口才、文职")
}

func addChangShengShenSha(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan, stageIdx int, name, cat, desc string) {
	zhus := bz.Slice()
	stageRow := ganzhi.ChangShengTable[riYuan]
	if len(stageRow) != 12 {
		return
	}
	for pi, p := range zhus {
		if bn := p.Zhi; bn >= 1 && bn <= 12 && stageRow[stageIdx] == bn {
			(*out)[pi] = append((*out)[pi], shenShaEntry{Name: name, Category: cat, Description: desc})
		}
	}
}

func addYangRen(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	zhus := bz.Slice()
	for pi, p := range zhus {
		if yangRenLookup[riYuan] == p.Zhi {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "羊刃", Category: catXiong, Description: "日干帝旺/刃位，主刚强果断，但易冲动",
			})
		}
	}
}

func addTianDe(out *[4][]shenShaEntry, bz ganzhi.Bazi, monthBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	targets, ok := tiandeTargets[monthBranch]
	if !ok {
		return
	}
	for _, tgt := range targets {
		for pi, p := range zhus {
			hit := false
			if tgt.IsZhi {
				hit = p.Zhi == tgt.Zhi
			} else {
				hit = p.Gan == tgt.Gan
			}
			if hit {
				(*out)[pi] = append((*out)[pi], shenShaEntry{
					Name: "天德", Category: catJi, Description: "天德贵人，主福泽深厚，化险为夷",
				})
			}
		}
	}
}

func addYueDe(out *[4][]shenShaEntry, bz ganzhi.Bazi, monthBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	targetStem, ok := yuedeStem[monthBranch]
	if !ok {
		return
	}
	for pi, p := range zhus {
		if p.Gan == targetStem {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "月德", Category: catJi, Description: "月德贵人，主月令之德，人缘佳",
			})
		}
	}
}

// addTriadShenSha 三合类神煞（驿马/桃花/华盖/劫煞/灾煞/将星）：
// 以年支与日支为参考（兼容两种流派），同一柱同一神煞只标注一次（去重）。
func addTriadShenSha(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi, lookup map[ganzhi.Zhi]ganzhi.Zhi, name, cat, desc string) {
	zhus := bz.Slice()
	marked := map[int]bool{}
	for _, refIdx := range []int{0, 2} { // year & day branch
		if tb, ok := lookup[branches[refIdx]]; ok {
			for pi, p := range zhus {
				if p.Zhi == tb && !marked[pi] {
					marked[pi] = true
					(*out)[pi] = append((*out)[pi], shenShaEntry{Name: name, Category: cat, Description: desc})
				}
			}
		}
	}
}

func addTaoHua(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, taohuaBranchMap, "桃花", catZhongXing, "主异性缘佳，浪漫多情")
}

func addYiMa(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, yimaBranchMap, "驿马", catZhongXing, "主动荡、奔波、迁移")
}

func addHuaGai(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, huagaiBranchMap, "华盖", catZhongXing, "主孤独清高，聪明好学，有艺术天赋")
}

func addJiangXing(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, jiangxingLookup, "将星", catJi, "主领导才能，有权威")
}

func addJieSha(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, jieshaBranch, "劫煞", catXiong, "主破财、意外、是非")
}

func addZaiSha(out *[4][]shenShaEntry, bz ganzhi.Bazi, branches [4]ganzhi.Zhi) {
	addTriadShenSha(out, bz, branches, zaishaBranch, "灾煞", catXiong, "主灾祸、疾病、横事")
}

func addGuChenGuaSu(out *[4][]shenShaEntry, bz ganzhi.Bazi, seasonIdx int) {
	zhus := bz.Slice()
	guchenBranches := [4]ganzhi.Zhi{6, 9, 12, 3}
	guasuBranches := [4]ganzhi.Zhi{2, 5, 8, 11}
	for pi, p := range zhus {
		if p.Zhi == guchenBranches[seasonIdx] {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "孤辰", Category: catXiong, Description: "主性格孤僻，晚婚或婚姻不顺",
			})
		}
		if p.Zhi == guasuBranches[seasonIdx] {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "寡宿", Category: catXiong, Description: "主孤独寂寞，夫妻缘薄",
			})
		}
	}
}

func addHongLuanTianXi(out *[4][]shenShaEntry, bz ganzhi.Bazi, yearBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	if target, ok := hongluanLookup[yearBranch]; ok {
		for pi, p := range zhus {
			if p.Zhi == target {
				(*out)[pi] = append((*out)[pi], shenShaEntry{
					Name: "红鸾", Category: catJi, Description: "主婚喜、恋爱、添丁",
				})
			}
		}
	}
	if target, ok := tianxiLookup[yearBranch]; ok {
		for pi, p := range zhus {
			if p.Zhi == target {
				(*out)[pi] = append((*out)[pi], shenShaEntry{
					Name: "天喜", Category: catJi, Description: "主喜庆之事，婚恋吉兆",
				})
			}
		}
	}
}

func addJinYu(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	appendShenShaByStemLookup(out, bz, riYuan, jinyuLookup, "金舆", catJi, "主财运、车辆、出行顺利")
}

func addYueEn(out *[4][]shenShaEntry, bz ganzhi.Bazi, monthBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	targets, ok := yueEnStems[monthBranch]
	if !ok {
		return
	}
	for _, ts := range targets {
		for pi, p := range zhus {
			if p.Gan == ts {
				(*out)[pi] = append((*out)[pi], shenShaEntry{
					Name: "月恩", Category: catJi, Description: "月令之恩，主福佑加持",
				})
			}
		}
	}
}

func addTianShe(out *[4][]shenShaEntry, bz ganzhi.Bazi, monthBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	season := (int(monthBranch) - 1) / 3
	tianSheChecks := [4]ganZhiPair{{5, 3}, {1, 7}, {5, 9}, {1, 1}} // 戊寅, 甲午, 戊申, 甲子
	if season >= 0 && season < 4 {
		pair := tianSheChecks[season]
		if zhus[2].Gan == pair.gan && zhus[2].Zhi == pair.zhi {
			(*out)[2] = append((*out)[2], shenShaEntry{
				Name: "天赦", Category: catJi, Description: "天赦日出生，主逢凶化吉，宽恕赦免",
			})
		}
	}
}

func addTianLuoDiWang(out *[4][]shenShaEntry, bz ganzhi.Bazi) {
	zhus := bz.Slice()
	for pi, p := range zhus {
		if label, ok := tianLuoDiWang[p.Zhi]; ok {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: label, Category: catXiong, Description: "主运势阻滞，有志难伸",
			})
		}
	}
}

func addGouJiao(out *[4][]shenShaEntry, bz ganzhi.Bazi, yearBranch ganzhi.Zhi) {
	zhus := bz.Slice()
	gouShen := ganzhi.Zhi((int(yearBranch)+2)%12 + 1)
	jiaoShen := ganzhi.Zhi((int(yearBranch)+4)%12 + 1)
	for pi, p := range zhus {
		if p.Zhi == gouShen {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "勾神", Category: catXiong, Description: "主纠缠牵连，是非官讼",
			})
		}
		if p.Zhi == jiaoShen {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "绞神", Category: catXiong, Description: "主受困被缚，身不由己",
			})
		}
	}
}

func addYuanChen(out *[4][]shenShaEntry, bz ganzhi.Bazi, yearBranch ganzhi.Zhi, nianGan ganzhi.Gan) {
	zhus := bz.Slice()
	ycBranch := yuanChenBranch(yearBranch, nianGan)
	for pi, p := range zhus {
		if p.Zhi == ycBranch {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "元辰", Category: catXiong, Description: "主波折反复，好事多磨",
			})
		}
	}
}

func addXueRen(out *[4][]shenShaEntry, bz ganzhi.Bazi, riYuan ganzhi.Gan) {
	zhus := bz.Slice()
	for pi, p := range zhus {
		if xueRenLookup[riYuan] == p.Zhi {
			(*out)[pi] = append((*out)[pi], shenShaEntry{
				Name: "血刃", Category: catXiong, Description: "主意外血光，手术外伤",
			})
		}
	}
}

func addSiFei(out *[4][]shenShaEntry, bz ganzhi.Bazi, seasonIdx int) {
	zhus := bz.Slice()
	siFeiZhus := [4][]ganZhiPair{
		{{7, 9}, {8, 10}}, {{9, 1}, {10, 12}}, {{1, 3}, {2, 4}}, {{3, 7}, {4, 6}},
	}
	if seasonIdx < 0 || seasonIdx >= 4 {
		return
	}
	for pi, p := range zhus {
		for _, pair := range siFeiZhus[seasonIdx] {
			if p.Gan == pair.gan && p.Zhi == pair.zhi {
				(*out)[pi] = append((*out)[pi], shenShaEntry{
					Name: "四废", Category: catXiong, Description: "四季废日，主事业阻滞，有志难伸",
				})
			}
		}
	}
}

func addShiEDaBai(out *[4][]shenShaEntry, bz ganzhi.Bazi) {
	zhus := bz.Slice()
	if _, ok := shiEDaBai[ganzhi.SixtyCycleIndex(zhus[2].Gan, zhus[2].Zhi)]; ok {
		(*out)[2] = append((*out)[2], shenShaEntry{
			Name: "十恶大败", Category: catXiong, Description: "日柱十恶大败日，主财库不聚，须谨慎理财",
		})
	}
}

func appendShenShaByStemLookup(out *[4][]shenShaEntry, bz ganzhi.Bazi, s ganzhi.Gan, lookup map[ganzhi.Gan][]ganzhi.Zhi, name, cat, desc string) {
	zhus := bz.Slice()
	targets, ok := lookup[s]
	if !ok {
		return
	}
	for pi, p := range zhus {
		for _, t := range targets {
			if p.Zhi == t {
				(*out)[pi] = append((*out)[pi], shenShaEntry{Name: name, Category: cat, Description: desc})
			}
		}
	}
}

// computeKongWang returns pillar indices whose branches fall in the void (空亡)
// of the day pillar's 旬.
func computeKongWang(bz ganzhi.Bazi) []int {
	sbIdx := ganzhi.SixtyCycleIndex(bz.Ri.Gan, bz.Ri.Zhi)
	xunIdx := sbIdx / 10

	voidPairs := [6][2]ganzhi.Zhi{
		{11, 12}, {9, 10}, {7, 8}, {5, 6}, {3, 4}, {1, 2},
	}
	v1, v2 := voidPairs[xunIdx][0], voidPairs[xunIdx][1]

	var hits []int
	for pi, p := range bz.Slice() {
		b := p.Zhi
		if b == v1 || b == v2 {
			hits = append(hits, pi)
		}
	}
	return hits
}

// computeDynamicShenSha computes shensha triggered by an external branch against the bazi chart.
func computeDynamicShenSha(b ganzhi.Zhi, yearBranch, riBranch ganzhi.Zhi, riYuan ganzhi.Gan) []shenShaEntry {
	var result []shenShaEntry
	seen := map[string]bool{}
	add := func(name, cat, desc string) {
		if !seen[name] {
			seen[name] = true
			result = append(result, shenShaEntry{Name: name, Category: cat, Description: desc})
		}
	}
	// 三合局系神煞（桃花/驿马/华盖/劫煞/灾煞）——年支+日支双查（《三命通会》年支/日支桃花驿马）
	for _, rb := range []ganzhi.Zhi{yearBranch, riBranch} {
		if tb, ok := taohuaBranchMap[rb]; ok && tb == b {
			add("桃花", catZhongXing, "流运桃花，异性缘佳")
		}
		if tb, ok := yimaBranchMap[rb]; ok && tb == b {
			add("驿马", catZhongXing, "流运驿马，动象奔波")
		}
		if tb, ok := huagaiBranchMap[rb]; ok && tb == b {
			add("华盖", catZhongXing, "流运华盖，宜静思")
		}
		if js, ok := jieshaBranch[rb]; ok && js == b {
			add("劫煞", catXiong, "流运劫煞，防破财是非")
		}
		if zs, ok := zaishaBranch[rb]; ok && zs == b {
			add("灾煞", catXiong, "流运灾煞，防意外灾祸")
		}
	}
	// 红鸾/天喜——仅年支查（红鸾属年支体系，日支不取）
	if hl, ok := hongluanLookup[yearBranch]; ok && hl == b {
		add("红鸾", catJi, "流运红鸾，主婚喜添丁")
	}
	if tx, ok := tianxiLookup[yearBranch]; ok && tx == b {
		add("天喜", catJi, "流运天喜，喜庆之事")
	}
	// 天乙贵人/羊刃——按日干
	if targets, ok := tianYiLookup[riYuan]; ok {
		for _, t := range targets {
			if t == b {
				add("天乙贵人", catJi, "流运天乙贵人，有贵人相助")
				break
			}
		}
	}
	if yr, ok := yangRenLookup[riYuan]; ok && yr == b {
		add("羊刃", catXiong, "流运羊刃，防冲动冲突")
	}

	if result == nil {
		return []shenShaEntry{}
	}
	return result
}

// computeAnnualShenSha 值年神煞（《协纪辨方书》——按太岁/流年支查表，命局四柱逢煞支即应）。
// 病符=太岁后1辰、丧门=后2辰、吊客=前2辰、大耗（岁破）=对冲。白虎查表有版本争议，不做。
func computeAnnualShenSha(b ganzhi.Zhi, bz ganzhi.Bazi) []shenShaEntry {
	annual := []struct {
		name string
		zhi  ganzhi.Zhi
		desc string
	}{
		{"病符", zhiOffset(b, -1), "流运病符临命，主病灾"},
		{"丧门", zhiOffset(b, -2), "流运丧门临命，主孝服/丧事"},
		{"吊客", zhiOffset(b, +2), "流运吊客临命，主吊丧/孝服"},
		{"大耗", zhiOffset(b, +6), "流运大耗临命，主破财大耗"},
	}
	var result []shenShaEntry
	seen := map[string]bool{}
	for _, a := range annual {
		for _, p := range bz.Slice() {
			if p.Zhi == a.zhi && !seen[a.name] {
				seen[a.name] = true
				result = append(result, shenShaEntry{Name: a.name, Category: catXiong, Description: a.desc})
				break
			}
		}
	}
	if result == nil {
		return []shenShaEntry{}
	}
	return result
}

// zhiOffset 地支循环偏移（z 后移 n 位，n 可为负）。
func zhiOffset(z ganzhi.Zhi, n int) ganzhi.Zhi {
	return ganzhi.Zhi((int(z)-1+n+120)%12 + 1)
}

func yuanChenBranch(yearBranch ganzhi.Zhi, nianGan ganzhi.Gan) ganzhi.Zhi {
	for _, p := range ganzhi.ChongPairs {
		if p.A == yearBranch {
			return yuanChenOffset(p.B, nianGan)
		}
		if p.B == yearBranch {
			return yuanChenOffset(p.A, nianGan)
		}
	}
	return 0
}

// yuanChenOffset applies the yin/yang offset to the clash branch.
// 阳年: +1, 阴年: -1.
func yuanChenOffset(clashBranch ganzhi.Zhi, nianGan ganzhi.Gan) ganzhi.Zhi {
	isYang := int(nianGan)%2 == 1
	if isYang {
		return clashBranch%12 + 1
	}
	return (clashBranch-2+12)%12 + 1
}
