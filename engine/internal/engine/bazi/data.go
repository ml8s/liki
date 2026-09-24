package bazi

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/tiaohou.json
var tiaohouJSON []byte

//go:embed data/shensha.json
var shenshaJSON []byte

//go:embed data/ride_rigui.json
var rideRiguiJSON []byte

//go:embed data/ten_god_strength_rules.json
var tenGodStrengthRulesJSON []byte

//go:embed data/yongji_rules.json
var yongjiRulesJSON []byte

//go:embed data/geju_rules.json
var gejuRulesJSON []byte

type tiaohouEntry struct {
	primary            ganzhi.Gan
	secondary          ganzhi.Gan
	secondaryCondition string
}

var lookupTiaohou map[tiaohouKey]tiaohouEntry

func init() {
	if err := loadTiaohou(); err != nil {
		log.Fatalf("bazi: load tiaohou: %v", err)
	}
	if err := loadShensha(); err != nil {
		log.Fatalf("bazi: load shensha: %v", err)
	}
	if err := loadRideRigui(); err != nil {
		log.Fatalf("bazi: load ride_rigui: %v", err)
	}
	if err := loadTenGodStrengthRules(); err != nil {
		log.Fatalf("bazi: load ten_god_strength_rules: %v", err)
	}
	if err := loadYongJiRules(); err != nil {
		log.Fatalf("bazi: load yongji_rules: %v", err)
	}
	if err := loadGeJuRules(); err != nil {
		log.Fatalf("bazi: load geju_rules: %v", err)
	}
}

// geJuPattern 格局规则（《子平真诠》）：格神十神 → 格名与顺逆用。
type geJuPattern struct {
	Name  string
	Usage string
}

var geJuRules map[string]geJuPattern

func loadGeJuRules() error {
	var raw struct {
		Patterns []struct {
			ShiShen string `json:"shishen"`
			Name    string `json:"name"`
			Usage   string `json:"usage"`
		} `json:"patterns"`
	}
	if err := json.Unmarshal(gejuRulesJSON, &raw); err != nil {
		return fmt.Errorf("unmarshal geju_rules.json: %w", err)
	}
	geJuRules = make(map[string]geJuPattern, len(raw.Patterns))
	for _, p := range raw.Patterns {
		geJuRules[p.ShiShen] = geJuPattern{Name: p.Name, Usage: p.Usage}
	}
	return nil
}

// yongJiRule 扶抑喜忌规则（《子平真诠》）：身强/身弱 → 用/喜/忌的五行关系。
type yongJiRule struct {
	Yong string // 克我者 / 生我者
	Xi   string // 生克我者 / 同我者
	Ji   string // 生我者 / 克我者
}

var yongJiRules map[string]yongJiRule

func loadYongJiRules() error {
	var raw []struct {
		Strength string `json:"strength"`
		Yong     string `json:"yong"`
		Xi       string `json:"xi"`
		Ji       string `json:"ji"`
		Basis    string `json:"basis"`
	}
	if err := json.Unmarshal(yongjiRulesJSON, &raw); err != nil {
		return fmt.Errorf("unmarshal yongji_rules.json: %w", err)
	}
	yongJiRules = make(map[string]yongJiRule, len(raw))
	for _, r := range raw {
		yongJiRules[r.Strength] = yongJiRule{Yong: r.Yong, Xi: r.Xi, Ji: r.Ji}
	}
	return nil
}

// tenGodStrengthRule 十神强度判定规则（《子平真诠》旺衰强弱）。
// 省略的字段为任意；season 取 element 的季节旺弱（strong/weak）。
type tenGodStrengthRule struct {
	Kind        string
	Season      *bool
	Transparent *bool
	Rooted      *bool
	Source      string
}

var tenGodStrengthRules []tenGodStrengthRule

func loadTenGodStrengthRules() error {
	var raw []struct {
		Rule        int    `json:"rule"`
		Kind        string `json:"kind"`
		Season      string `json:"season,omitempty"`
		Transparent bool   `json:"transparent,omitempty"`
		Rooted      bool   `json:"rooted,omitempty"`
		Source      string `json:"source"`
	}
	if err := json.Unmarshal(tenGodStrengthRulesJSON, &raw); err != nil {
		return fmt.Errorf("unmarshal ten_god_strength_rules.json: %w", err)
	}
	tenGodStrengthRules = make([]tenGodStrengthRule, 0, len(raw))
	for _, r := range raw {
		rule := tenGodStrengthRule{Kind: r.Kind, Source: r.Source}
		switch r.Season {
		case "strong":
			rule.Season = boolPtr(true)
		case "weak":
			rule.Season = boolPtr(false)
		}
		rule.Transparent = boolPtr(r.Transparent)
		rule.Rooted = boolPtr(r.Rooted)
		tenGodStrengthRules = append(tenGodStrengthRules, rule)
	}
	return nil
}

func boolPtr(v bool) *bool { return &v }

func loadTiaohou() error {
	var entries []struct {
		RiYuan             string `json:"ri_yuan"`
		MonthBranch        string `json:"month_branch"`
		Primary            string `json:"primary"`
		Secondary          string `json:"secondary"`
		SecondaryCondition string `json:"secondary_condition,omitempty"`
	}
	if err := json.Unmarshal(tiaohouJSON, &entries); err != nil {
		return err
	}
	lookupTiaohou = make(map[tiaohouKey]tiaohouEntry, len(entries))
	for _, e := range entries {
		dm, err := ganzhi.ParseGan(e.RiYuan)
		if err != nil {
			return err
		}
		mb, err := ganzhi.ParseZhi(e.MonthBranch)
		if err != nil {
			return err
		}
		pri, err := ganzhi.ParseGan(e.Primary)
		if err != nil {
			return err
		}
		var sec ganzhi.Gan
		if e.Secondary != "" {
			sec, err = ganzhi.ParseGan(e.Secondary)
			if err != nil {
				return err
			}
		}
		lookupTiaohou[tiaohouKey{int(dm), int(mb)}] = tiaohouEntry{
			primary:            pri,
			secondary:          sec,
			secondaryCondition: e.SecondaryCondition,
		}
	}
	return nil
}

func loadShensha() error {
	var data struct {
		Triad       map[string]map[string]string   `json:"triad"`
		GanSingle   map[string]map[string]string   `json:"stem_single"`
		GanMulti    map[string]map[string][]string `json:"stem_multi"`
		ZhiSingle   map[string]map[string]string   `json:"branch_single"`
		MonthGroups map[string]map[string]struct {
			De  []string `json:"de"`
			Xiu []string `json:"xiu"`
		} `json:"month_groups"`
		DayPillars struct {
			YinChaYangCuo []string `json:"yin_cha_yang_cuo"`
		} `json:"day_pillars"`
		TongZi struct {
			Season map[string][]string `json:"season"`
			Nayin  map[string][]string `json:"nayin"`
		} `json:"tong_zi"`
		YueGan struct {
			TianDe map[string][]string `json:"tian_de"`
			YueDe  map[string]string   `json:"yue_de"`
			YueEn  map[string][]string `json:"yue_en"`
		} `json:"month_stems"`
		TianLuoDiWang map[string]string `json:"tian_luo_di_wang"`
		ShiEDaBai     []int             `json:"shi_e_da_bai"`
		Elements      struct {
			Yang map[string]string `json:"yang"`
			Yin  map[string]string `json:"yin"`
		} `json:"elements"`
	}
	if err := json.Unmarshal(shenshaJSON, &data); err != nil {
		return fmt.Errorf("unmarshal shensha.json: %w", err)
	}

	// --- triad maps (寅午戌 → individual zhi → target) ---
	taohuaZhiMap = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)
	yimaZhiMap = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)
	huagaiZhiMap = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)
	jieshaZhi = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)
	zaishaZhi = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)
	jiangxingLookup = make(map[ganzhi.Zhi]ganzhi.Zhi, 12)

	loadTriad := func(dst map[ganzhi.Zhi]ganzhi.Zhi, src map[string]string) error {
		for triadKey, targetStr := range src {
			target, err := ganzhi.ParseZhi(targetStr)
			if err != nil {
				return fmt.Errorf("triad target %q: %w", targetStr, err)
			}
			for _, r := range triadKey {
				zhi, err := ganzhi.ParseZhi(string(r))
				if err != nil {
					return fmt.Errorf("triad member %q in %q: %w", string(r), triadKey, err)
				}
				dst[zhi] = target
			}
		}
		return nil
	}

	triadDsts := map[string]map[ganzhi.Zhi]ganzhi.Zhi{
		"taohua":    taohuaZhiMap,
		"yima":      yimaZhiMap,
		"huagai":    huagaiZhiMap,
		"jiesha":    jieshaZhi,
		"zaisha":    zaishaZhi,
		"jiangxing": jiangxingLookup,
	}
	for name, dst := range triadDsts {
		if err := loadTriad(dst, data.Triad[name]); err != nil {
			return fmt.Errorf("triad %s: %w", name, err)
		}
	}

	// --- gan → single zhi ---
	loadGanSingle := func(src map[string]string) (map[ganzhi.Gan]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Gan]ganzhi.Zhi, len(src))
		for ganStr, zhiStr := range src {
			gan, err := ganzhi.ParseGan(ganStr)
			if err != nil {
				return nil, fmt.Errorf("gan %q: %w", ganStr, err)
			}
			zhi, err := ganzhi.ParseZhi(zhiStr)
			if err != nil {
				return nil, fmt.Errorf("zhi %q: %w", zhiStr, err)
			}
			dst[gan] = zhi
		}
		return dst, nil
	}

	var err error
	yangRenLookup, err = loadGanSingle(data.GanSingle["yang_ren"])
	if err != nil {
		return fmt.Errorf("yang_ren: %w", err)
	}
	xueRenLookup, err = loadGanSingle(data.GanSingle["xue_ren"])
	if err != nil {
		return fmt.Errorf("xue_ren: %w", err)
	}
	feiRenLookup, err = loadGanSingle(data.GanSingle["fei_ren"])
	if err != nil {
		return fmt.Errorf("fei_ren: %w", err)
	}

	// --- gan → multi zhi ---
	loadGanMulti := func(src map[string][]string) (map[ganzhi.Gan][]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Gan][]ganzhi.Zhi, len(src))
		for ganStr, zhiStrs := range src {
			gan, err := ganzhi.ParseGan(ganStr)
			if err != nil {
				return nil, fmt.Errorf("gan %q: %w", ganStr, err)
			}
			zhi := make([]ganzhi.Zhi, len(zhiStrs))
			for i, bs := range zhiStrs {
				b, err := ganzhi.ParseZhi(bs)
				if err != nil {
					return nil, fmt.Errorf("zhi %q: %w", bs, err)
				}
				zhi[i] = b
			}
			dst[gan] = zhi
		}
		return dst, nil
	}

	tianYiLookup, err = loadGanMulti(data.GanMulti["tian_yi"])
	if err != nil {
		return fmt.Errorf("tian_yi: %w", err)
	}
	wenChangLookup, err = loadGanMulti(data.GanMulti["wen_chang"])
	if err != nil {
		return fmt.Errorf("wen_chang: %w", err)
	}
	jinyuLookup, err = loadGanMulti(data.GanMulti["jin_yu"])
	if err != nil {
		return fmt.Errorf("jin_yu: %w", err)
	}
	taiJiLookup, err = loadGanMulti(data.GanMulti["tai_ji"])
	if err != nil {
		return fmt.Errorf("tai_ji: %w", err)
	}
	tianChuLookup, err = loadGanMulti(data.GanMulti["tian_chu"])
	if err != nil {
		return fmt.Errorf("tian_chu: %w", err)
	}
	fuXingLookup, err = loadGanMulti(data.GanMulti["fu_xing"])
	if err != nil {
		return fmt.Errorf("fu_xing: %w", err)
	}
	guoYinLookup, err = loadGanMulti(data.GanMulti["guo_yin"])
	if err != nil {
		return fmt.Errorf("guo_yin: %w", err)
	}
	// --- zhi → single zhi ---
	loadZhiSingle := func(src map[string]string) (map[ganzhi.Zhi]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Zhi]ganzhi.Zhi, len(src))
		for zhiStr, targetStr := range src {
			zhi, err := ganzhi.ParseZhi(zhiStr)
			if err != nil {
				return nil, fmt.Errorf("zhi %q: %w", zhiStr, err)
			}
			target, err := ganzhi.ParseZhi(targetStr)
			if err != nil {
				return nil, fmt.Errorf("target %q: %w", targetStr, err)
			}
			dst[zhi] = target
		}
		return dst, nil
	}

	wangShenZhi, err = loadZhiSingle(data.ZhiSingle["wang_shen"])
	if err != nil {
		return fmt.Errorf("wang_shen: %w", err)
	}

	hongluanLookup, err = loadZhiSingle(data.ZhiSingle["hong_luan"])
	if err != nil {
		return fmt.Errorf("hong_luan: %w", err)
	}
	tianxiLookup, err = loadZhiSingle(data.ZhiSingle["tian_xi"])
	if err != nil {
		return fmt.Errorf("tian_xi: %w", err)
	}

	// --- month zhi → gan (keys are zhi, values are gan) ---
	loadBranchToStems := func(src map[string][]string) (map[ganzhi.Zhi][]ganzhi.Gan, error) {
		dst := make(map[ganzhi.Zhi][]ganzhi.Gan, len(src))
		for zhiStr, ganStrs := range src {
			zhi, err := ganzhi.ParseZhi(zhiStr)
			if err != nil {
				return nil, fmt.Errorf("zhi %q: %w", zhiStr, err)
			}
			gan := make([]ganzhi.Gan, len(ganStrs))
			for i, ss := range ganStrs {
				s, err := ganzhi.ParseGan(ss)
				if err != nil {
					return nil, fmt.Errorf("gan %q: %w", ss, err)
				}
				gan[i] = s
			}
			dst[zhi] = gan
		}
		return dst, nil
	}

	// --- month zhi → 天德目标（天干型或地支型，如正月见丁、二月见申） ---
	loadBranchToTianDe := func(src map[string][]string) (map[ganzhi.Zhi][]tianDeTarget, error) {
		dst := make(map[ganzhi.Zhi][]tianDeTarget, len(src))
		for zhiStr, targetStrs := range src {
			zhi, err := ganzhi.ParseZhi(zhiStr)
			if err != nil {
				return nil, fmt.Errorf("zhi %q: %w", zhiStr, err)
			}
			var tgts []tianDeTarget
			for _, ts := range targetStrs {
				if gan, err := ganzhi.ParseGan(ts); err == nil {
					tgts = append(tgts, tianDeTarget{Gan: gan})
					continue
				}
				if zhi, err := ganzhi.ParseZhi(ts); err == nil {
					tgts = append(tgts, tianDeTarget{IsZhi: true, Zhi: zhi})
					continue
				}
				return nil, fmt.Errorf("tian_de target %q: neither gan nor zhi", ts)
			}
			dst[zhi] = tgts
		}
		return dst, nil
	}

	tiandeTargets, err = loadBranchToTianDe(data.YueGan.TianDe)
	if err != nil {
		return fmt.Errorf("tian_de: %w", err)
	}
	yueEnGan, err = loadBranchToStems(data.YueGan.YueEn)
	if err != nil {
		return fmt.Errorf("yue_en: %w", err)
	}

	yuedeGan = make(map[ganzhi.Zhi]ganzhi.Gan, len(data.YueGan.YueDe))
	for zhiStr, ganStr := range data.YueGan.YueDe {
		zhi, err := ganzhi.ParseZhi(zhiStr)
		if err != nil {
			return fmt.Errorf("yue_de zhi %q: %w", zhiStr, err)
		}
		gan, err := ganzhi.ParseGan(ganStr)
		if err != nil {
			return fmt.Errorf("yue_de gan %q: %w", ganStr, err)
		}
		yuedeGan[zhi] = gan
	}

	deXiuByMonth = make(map[ganzhi.Zhi]deXiuStems, 12)
	for triadKey, group := range data.MonthGroups["de_xiu"] {
		de, err := parseGans(group.De)
		if err != nil {
			return fmt.Errorf("de_xiu de %q: %w", triadKey, err)
		}
		xiu, err := parseGans(group.Xiu)
		if err != nil {
			return fmt.Errorf("de_xiu xiu %q: %w", triadKey, err)
		}
		for _, r := range triadKey {
			zhi, err := ganzhi.ParseZhi(string(r))
			if err != nil {
				return fmt.Errorf("de_xiu member %q in %q: %w", string(r), triadKey, err)
			}
			deXiuByMonth[zhi] = deXiuStems{De: de, Xiu: xiu}
		}
	}

	yinChaYangCuo = make(map[int]struct{}, len(data.DayPillars.YinChaYangCuo))
	for _, pair := range data.DayPillars.YinChaYangCuo {
		runes := []rune(pair)
		if len(runes) != 2 {
			return fmt.Errorf("yin_cha_yang_cuo: invalid pair %q", pair)
		}
		gan, err := ganzhi.ParseGan(string(runes[0]))
		if err != nil {
			return fmt.Errorf("yin_cha_yang_cuo gan %q: %w", pair, err)
		}
		zhi, err := ganzhi.ParseZhi(string(runes[1]))
		if err != nil {
			return fmt.Errorf("yin_cha_yang_cuo zhi %q: %w", pair, err)
		}
		yinChaYangCuo[ganzhi.SixtyCycleIndex(gan, zhi)] = struct{}{}
	}

	tongZiSeason = map[int][]ganzhi.Zhi{
		0: parseZhis(data.TongZi.Season["spring_autumn"]),
		1: parseZhis(data.TongZi.Season["summer_winter"]),
		2: parseZhis(data.TongZi.Season["spring_autumn"]),
		3: parseZhis(data.TongZi.Season["summer_winter"]),
	}
	tongZiNayin = make(map[ganzhi.Wuxing][]ganzhi.Zhi)
	for elements, branches := range data.TongZi.Nayin {
		zhis := parseZhis(branches)
		for _, r := range elements {
			element, err := ganzhi.ParseWuxing(string(r))
			if err != nil {
				return fmt.Errorf("tong_zi nayin element %q: %w", elements, err)
			}
			tongZiNayin[element] = zhis
		}
	}

	// --- tian luo di wang ---
	tianLuoDiWang = make(map[ganzhi.Zhi]string, len(data.TianLuoDiWang))
	for zhiStr, label := range data.TianLuoDiWang {
		zhi, err := ganzhi.ParseZhi(zhiStr)
		if err != nil {
			return fmt.Errorf("tian_luo_di_wang zhi %q: %w", zhiStr, err)
		}
		tianLuoDiWang[zhi] = label
	}

	// --- shi e da bai ---
	shiEDaBai = make(map[int]struct{}, len(data.ShiEDaBai))
	for _, v := range data.ShiEDaBai {
		shiEDaBai[v] = struct{}{}
	}

	return nil
}

func parseGans(values []string) ([]ganzhi.Gan, error) {
	out := make([]ganzhi.Gan, 0, len(values))
	for _, value := range values {
		gan, err := ganzhi.ParseGan(value)
		if err != nil {
			return nil, err
		}
		out = append(out, gan)
	}
	return out, nil
}

func parseZhis(values []string) []ganzhi.Zhi {
	out := make([]ganzhi.Zhi, 0, len(values))
	for _, value := range values {
		zhi, err := ganzhi.ParseZhi(value)
		if err != nil {
			panic(fmt.Sprintf("bazi: parse zhi %q: %v", value, err))
		}
		out = append(out, zhi)
	}
	return out
}

func loadRideRigui() error {
	var data struct {
		RiDe  [][]string `json:"ri_de"`
		RiGui [][]string `json:"ri_gui"`
	}
	if err := json.Unmarshal(rideRiguiJSON, &data); err != nil {
		return fmt.Errorf("unmarshal ride_rigui.json: %w", err)
	}

	riDeSet = make(map[[2]int]bool, len(data.RiDe))
	for _, pair := range data.RiDe {
		if len(pair) != 2 {
			continue
		}
		gan, err := ganzhi.ParseGan(pair[0])
		if err != nil {
			return fmt.Errorf("ri_de gan %q: %w", pair[0], err)
		}
		zhi, err := ganzhi.ParseZhi(pair[1])
		if err != nil {
			return fmt.Errorf("ri_de zhi %q: %w", pair[1], err)
		}
		riDeSet[[2]int{int(gan), int(zhi)}] = true
	}

	riGuiSet = make(map[[2]int]bool, len(data.RiGui))
	for _, pair := range data.RiGui {
		if len(pair) != 2 {
			continue
		}
		gan, err := ganzhi.ParseGan(pair[0])
		if err != nil {
			return fmt.Errorf("ri_gui gan %q: %w", pair[0], err)
		}
		zhi, err := ganzhi.ParseZhi(pair[1])
		if err != nil {
			return fmt.Errorf("ri_gui zhi %q: %w", pair[1], err)
		}
		riGuiSet[[2]int{int(gan), int(zhi)}] = true
	}

	return nil
}
