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

var lookupTiaohou map[tiaohouKey]struct{ primary, secondary ganzhi.Gan }

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
}

func loadTiaohou() error {
	var entries []struct {
		RiYuan      string `json:"ri_yuan"`
		MonthBranch string `json:"month_branch"`
		Primary     string `json:"primary"`
		Secondary   string `json:"secondary"`
	}
	if err := json.Unmarshal(tiaohouJSON, &entries); err != nil {
		return err
	}
	lookupTiaohou = make(map[tiaohouKey]struct{ primary, secondary ganzhi.Gan }, len(entries))
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
		lookupTiaohou[tiaohouKey{int(dm), int(mb)}] = struct{ primary, secondary ganzhi.Gan }{pri, sec}
	}
	return nil
}

func loadShensha() error {
	var data struct {
		Triad        map[string]map[string]string   `json:"triad"`
		StemSingle   map[string]map[string]string   `json:"stem_single"`
		StemMulti    map[string]map[string][]string `json:"stem_multi"`
		BranchSingle map[string]map[string]string   `json:"branch_single"`
		MonthStems   struct {
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

	// --- triad maps (寅午戌 → individual branch → target) ---
	taohuaBranchMap = make(map[ganzhi.Zhi]ganzhi.Zhi)
	yimaBranchMap = make(map[ganzhi.Zhi]ganzhi.Zhi)
	huagaiBranchMap = make(map[ganzhi.Zhi]ganzhi.Zhi)
	jieshaBranch = make(map[ganzhi.Zhi]ganzhi.Zhi)
	zaishaBranch = make(map[ganzhi.Zhi]ganzhi.Zhi)
	jiangxingLookup = make(map[ganzhi.Zhi]ganzhi.Zhi)

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
		"taohua":    taohuaBranchMap,
		"yima":      yimaBranchMap,
		"huagai":    huagaiBranchMap,
		"jiesha":    jieshaBranch,
		"zaisha":    zaishaBranch,
		"jiangxing": jiangxingLookup,
	}
	for name, dst := range triadDsts {
		if err := loadTriad(dst, data.Triad[name]); err != nil {
			return fmt.Errorf("triad %s: %w", name, err)
		}
	}

	// --- stem → single branch ---
	loadStemSingle := func(src map[string]string) (map[ganzhi.Gan]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Gan]ganzhi.Zhi, len(src))
		for stemStr, branchStr := range src {
			stem, err := ganzhi.ParseGan(stemStr)
			if err != nil {
				return nil, fmt.Errorf("stem %q: %w", stemStr, err)
			}
			branch, err := ganzhi.ParseZhi(branchStr)
			if err != nil {
				return nil, fmt.Errorf("branch %q: %w", branchStr, err)
			}
			dst[stem] = branch
		}
		return dst, nil
	}

	var err error
	yangRenLookup, err = loadStemSingle(data.StemSingle["yang_ren"])
	if err != nil {
		return fmt.Errorf("yang_ren: %w", err)
	}
	xueRenLookup, err = loadStemSingle(data.StemSingle["xue_ren"])
	if err != nil {
		return fmt.Errorf("xue_ren: %w", err)
	}

	// --- stem → multi branch ---
	loadStemMulti := func(src map[string][]string) (map[ganzhi.Gan][]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Gan][]ganzhi.Zhi, len(src))
		for stemStr, branchStrs := range src {
			stem, err := ganzhi.ParseGan(stemStr)
			if err != nil {
				return nil, fmt.Errorf("stem %q: %w", stemStr, err)
			}
			branches := make([]ganzhi.Zhi, len(branchStrs))
			for i, bs := range branchStrs {
				b, err := ganzhi.ParseZhi(bs)
				if err != nil {
					return nil, fmt.Errorf("branch %q: %w", bs, err)
				}
				branches[i] = b
			}
			dst[stem] = branches
		}
		return dst, nil
	}

	tianYiLookup, err = loadStemMulti(data.StemMulti["tian_yi"])
	if err != nil {
		return fmt.Errorf("tian_yi: %w", err)
	}
	wenChangLookup, err = loadStemMulti(data.StemMulti["wen_chang"])
	if err != nil {
		return fmt.Errorf("wen_chang: %w", err)
	}
	jinyuLookup, err = loadStemMulti(data.StemMulti["jin_yu"])
	if err != nil {
		return fmt.Errorf("jin_yu: %w", err)
	}

	// --- branch → single branch ---
	loadBranchSingle := func(src map[string]string) (map[ganzhi.Zhi]ganzhi.Zhi, error) {
		dst := make(map[ganzhi.Zhi]ganzhi.Zhi, len(src))
		for branchStr, targetStr := range src {
			branch, err := ganzhi.ParseZhi(branchStr)
			if err != nil {
				return nil, fmt.Errorf("branch %q: %w", branchStr, err)
			}
			target, err := ganzhi.ParseZhi(targetStr)
			if err != nil {
				return nil, fmt.Errorf("target %q: %w", targetStr, err)
			}
			dst[branch] = target
		}
		return dst, nil
	}

	hongluanLookup, err = loadBranchSingle(data.BranchSingle["hong_luan"])
	if err != nil {
		return fmt.Errorf("hong_luan: %w", err)
	}
	tianxiLookup, err = loadBranchSingle(data.BranchSingle["tian_xi"])
	if err != nil {
		return fmt.Errorf("tian_xi: %w", err)
	}

	// --- month branch → stems (keys are branches, values are stems) ---
	loadBranchToStems := func(src map[string][]string) (map[ganzhi.Zhi][]ganzhi.Gan, error) {
		dst := make(map[ganzhi.Zhi][]ganzhi.Gan, len(src))
		for branchStr, stemStrs := range src {
			branch, err := ganzhi.ParseZhi(branchStr)
			if err != nil {
				return nil, fmt.Errorf("branch %q: %w", branchStr, err)
			}
			stems := make([]ganzhi.Gan, len(stemStrs))
			for i, ss := range stemStrs {
				s, err := ganzhi.ParseGan(ss)
				if err != nil {
					return nil, fmt.Errorf("stem %q: %w", ss, err)
				}
				stems[i] = s
			}
			dst[branch] = stems
		}
		return dst, nil
	}

	// --- month branch → 天德目标（天干型或地支型，如正月见丁、二月见申） ---
	loadBranchToTianDe := func(src map[string][]string) (map[ganzhi.Zhi][]tianDeTarget, error) {
		dst := make(map[ganzhi.Zhi][]tianDeTarget, len(src))
		for branchStr, targetStrs := range src {
			branch, err := ganzhi.ParseZhi(branchStr)
			if err != nil {
				return nil, fmt.Errorf("branch %q: %w", branchStr, err)
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
			dst[branch] = tgts
		}
		return dst, nil
	}

	tiandeTargets, err = loadBranchToTianDe(data.MonthStems.TianDe)
	if err != nil {
		return fmt.Errorf("tian_de: %w", err)
	}
	yueEnStems, err = loadBranchToStems(data.MonthStems.YueEn)
	if err != nil {
		return fmt.Errorf("yue_en: %w", err)
	}

	yuedeStem = make(map[ganzhi.Zhi]ganzhi.Gan, len(data.MonthStems.YueDe))
	for branchStr, stemStr := range data.MonthStems.YueDe {
		branch, err := ganzhi.ParseZhi(branchStr)
		if err != nil {
			return fmt.Errorf("yue_de branch %q: %w", branchStr, err)
		}
		stem, err := ganzhi.ParseGan(stemStr)
		if err != nil {
			return fmt.Errorf("yue_de stem %q: %w", stemStr, err)
		}
		yuedeStem[branch] = stem
	}

	// --- tian luo di wang ---
	tianLuoDiWang = make(map[ganzhi.Zhi]string, len(data.TianLuoDiWang))
	for branchStr, label := range data.TianLuoDiWang {
		branch, err := ganzhi.ParseZhi(branchStr)
		if err != nil {
			return fmt.Errorf("tian_luo_di_wang branch %q: %w", branchStr, err)
		}
		tianLuoDiWang[branch] = label
	}

	// --- shi e da bai ---
	shiEDaBai = make(map[int]struct{}, len(data.ShiEDaBai))
	for _, v := range data.ShiEDaBai {
		shiEDaBai[v] = struct{}{}
	}

	return nil
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
