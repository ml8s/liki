package qimen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/gan_interaction.json
var ganInteractionJSON []byte

//go:embed data/jushu.json
var jushuJSON []byte

//go:embed data/men_interaction.json
var menInteractionJSON []byte

//go:embed data/catalog.json
var catalogJSON []byte

//go:embed data/plate.json
var plateJSON []byte

//go:embed data/patterns.json
var patternsJSON []byte

//go:embed data/xing_interaction.json
var xingInteractionJSON []byte

//go:embed data/yingqi.json
var yingqiJSON []byte

//go:embed data/zhirun.json
var zhirunJSON []byte

//go:embed data/maoshan.json
var maoshanJSON []byte

//go:embed data/jinhan.json
var jinhanJSON []byte

//go:embed data/quarter.json
var quarterJSON []byte

type wuxingRelationEntry struct {
	Name             string
	TraditionalLabel string
}

type yingqiRuleEntry struct {
	Mechanism   string
	Description string
}

type yingqiDateMatch struct {
	Match  string
	Reason string
}

type zhiRunStateEntry struct {
	State     string
	Name      string
	Condition string
}

type scopeEntry struct {
	Scope              string
	Name               string
	LeadPillar         string
	MethodSource       string
	DingjuMethods      []DingjuMethod
	DefaultDingju      DingjuMethod
	NoDingjuSchools    []School
	QuarterRules       []QuarterRule
	DefaultQuarterRule QuarterRule
	Features           []string
}

type schoolEntry struct {
	School               string
	Name                 string
	AllowedScopes        []Scope
	AllowedDingjuMethods []DingjuMethod
	AllowedQuarterRules  []QuarterRule
	SpiritMode           string
	Geometry             string
	StarMode             string
	DoorMode             string
	SpiritFlight         string
}

func (e schoolEntry) allowsScope(scope Scope) bool {
	for _, candidate := range e.AllowedScopes {
		if candidate == scope {
			return true
		}
	}
	return false
}

func (e schoolEntry) allowsDingju(method DingjuMethod) bool {
	for _, candidate := range e.AllowedDingjuMethods {
		if candidate == method {
			return true
		}
	}
	return false
}

func (e schoolEntry) allowsQuarterRule(rule QuarterRule) bool {
	for _, candidate := range e.AllowedQuarterRules {
		if candidate == rule {
			return true
		}
	}
	return false
}

func (e scopeEntry) allowsDingju(method DingjuMethod) bool {
	for _, candidate := range e.DingjuMethods {
		if candidate == method {
			return true
		}
	}
	return false
}

func (e scopeEntry) allowsQuarterRule(rule QuarterRule) bool {
	for _, candidate := range e.QuarterRules {
		if candidate == rule {
			return true
		}
	}
	return false
}

func (e scopeEntry) allowsNoDingjuSchool(school School) bool {
	for _, candidate := range e.NoDingjuSchools {
		if candidate == school {
			return true
		}
	}
	return false
}

func (e scopeEntry) hasFeature(feature string) bool {
	for _, candidate := range e.Features {
		if candidate == feature {
			return true
		}
	}
	return false
}

type monthYearGroupEntry struct {
	LeadingJu int
	Yuan      string
}

type quarterSanyuanEntry struct {
	Yuan   string
	YangJu int
	YinJu  int
}

type quarterVariant struct {
	Rule                QuarterRule
	Minutes             int
	Count               int
	LeadPillarMode      string
	YongJuJieQi         string
	BaseScope           Scope
	BaseSchool          School
	BaseDingjuMethods   map[DingjuMethod]bool
	DunSources          map[string]bool
	DefaultDunSource    string
	HourBoundaries      map[string]int
	DefaultHourBoundary string
	JuCycle             int
	YangStep            int
	YinStep             int
}

type quarterJSONEntry struct {
	Rule              string                `json:"rule"`
	Minutes           int                   `json:"quarter_minutes"`
	Count             int                   `json:"quarters_per_shichen"`
	LeadPillarMode    string                `json:"lead_pillar_mode"`
	YongJuJieQi       string                `json:"yong_ju_jieqi"`
	PillarRule        *quarterPillarJSON    `json:"pillar_rule"`
	DingjuRule        *quarterDingjuJSON    `json:"dingju_rule"`
	BaseScope         string                `json:"base_scope"`
	BaseSchool        string                `json:"base_school"`
	BaseDingjuMethods []string              `json:"base_dingju_methods"`
	DunSources        []quarterOptionJSON   `json:"dun_sources"`
	HourBoundaries    []quarterBoundaryJSON `json:"hour_boundaries"`
	JuShift           *quarterShiftJSON     `json:"ju_shift"`
}

type quarterOptionJSON struct {
	Source  string `json:"source"`
	Default bool   `json:"default"`
}

type quarterBoundaryJSON struct {
	Boundary       string `json:"boundary"`
	HourStartShift int    `json:"hour_start_shift"`
	Default        bool   `json:"default"`
}

type quarterPillarJSON struct {
	Source string                   `json:"source"`
	Starts []quarterPillarStartJSON `json:"starts"`
}

type quarterPillarStartJSON struct {
	HourGan string `json:"hour_gan"`
	Start   string `json:"start"`
}

type quarterDingjuJSON struct {
	YinYangSource string                    `json:"yin_yang_source"`
	YuanSource    string                    `json:"yuan_source"`
	FuTouUnits    int                       `json:"fu_tou_units"`
	Groups        []quarterSanyuanGroupJSON `json:"groups"`
}

type quarterSanyuanGroupJSON struct {
	Branches []string `json:"branches"`
	Yuan     string   `json:"yuan"`
	YangJu   int      `json:"yang_ju"`
	YinJu    int      `json:"yin_ju"`
}

type quarterShiftJSON struct {
	Cycle    int `json:"cycle"`
	YangStep int `json:"yang_step"`
	YinStep  int `json:"yin_step"`
}

var (
	ganInteractionTable map[[2]ganzhi.Gan]ganEntry
	solarTermJuTable    map[string][4]int
	solarTermLongitudes map[string]float64
	solarTermOrder      [24]string
	defaultScope        Scope
	defaultSchool       School
	scopeMethods        map[Scope]scopeEntry
	dingjuMethodNames   map[DingjuMethod]string
	schoolMethods       map[School]schoolEntry
	monthFirstBranch    ganzhi.Zhi
	monthGroups         map[ganzhi.Zhi]monthYearGroupEntry
	yearAnchor          int
	yearCycle           int
	yearJuByYuan        map[string]int
	qimenDayBoundary    string
	menGongTable        map[[2]int]doorEntry
	xingGongTable       map[[2]int]XingInteraction

	tianQinRule            TianQinRule
	tianQinStar            StarIndex
	tianQinFollows         StarIndex
	tianQinCarriesGan      bool
	outerRing              [8]GongIndex
	earthGanOrder          [9]ganzhi.Gan
	starOrder8             [8]StarIndex
	doorOrder              [8]DoorIndex
	mingfaDoorOrder        [9]DoorIndex
	spiritOrder            [8]SpiritIndex
	flySpiritOrder         [9]SpiritIndex
	flySpiritNames         [10]string
	mingfaSpiritOrder      [9]SpiritIndex
	mingfaSpiritNames      [10]string
	spiritYangNames        [10]string
	spiritYinNames         [10]string
	palaceStar             [9]StarIndex
	palaceDoor             [9]DoorIndex
	gongWuxingTable        [9]ganzhi.Wuxing
	starWuxingTable        [9]ganzhi.Wuxing
	doorWuxingTable        [9]ganzhi.Wuxing
	wuxingRelations        map[string]wuxingRelationEntry
	riShiRelations         map[string]string
	zhiPalaceTable         [12]GongIndex
	maXingTable            [12]ganzhi.Zhi
	liuJiaZhi              [6]ganzhi.Zhi
	liuJiaLiuYi            [6]ganzhi.Gan
	xunKongTable           [6][2]ganzhi.Zhi
	wuBuYuShiTable         map[[2]ganzhi.Gan]bool
	fuTouYuanTable         map[ganzhi.Zhi]int
	yuanNames              [3]string
	patternRules           []patternRule
	yingqiSummary          string
	yingqiRuleTable        map[string]yingqiRuleEntry
	yingqiHorizonDays      int
	yingqiDateMatches      map[string][]yingqiDateMatch
	zhiRunLeaderDays       int
	zhiRunThreshold        int
	zhiRunMaxIntervals     int
	zhiRunRepeatTerms      map[string]bool
	zhiRunStateNames       map[string]string
	zhiRunBasis            map[string]string
	maoShanShichenPerYuan  int
	maoShanMaxYuan         int
	centerLodgingPalace    GongIndex
	jinhanBasis            string
	jinhanMethodSource     string
	jinhanDunSource        string
	jinhanStarOrder        []string
	jinhanYangStarStarts   []GongIndex
	jinhanYinStarStarts    []GongIndex
	jinhanYangStarFlight   []GongIndex
	jinhanYinStarFlight    []GongIndex
	jinhanDoorStarts       []GongIndex
	jinhanDoorRing         []GongIndex
	jinhanDoorOrder        []DoorIndex
	jinhanDoorDays         int
	jinhanDaySpirits       map[ganzhi.Gan][12]string
	quarterRules           map[QuarterRule]quarterVariant
	quarterRuleNames       map[QuarterRule]string
	quarterFuTouUnits      int
	quarterStartIndex      map[ganzhi.Gan]int
	quarterYangBranch      map[ganzhi.Zhi]bool
	quarterSanyuanByBranch map[ganzhi.Zhi]quarterSanyuanEntry
)

func init() {
	loaders := []func() error{
		loadGanInteractions, loadJushu, loadMenInteractions,
		loadXingInteractions, loadCatalog, loadQuarter, loadPlate, loadPatterns,
		loadYingQi, loadZhiRun, loadMaoShan, loadJinhan,
	}
	for _, load := range loaders {
		if err := load(); err != nil {
			log.Fatalf("qimen: %v", err)
		}
	}
}

type interactionTableDocument[T any] struct {
	Provenance string `json:"provenance"`
	Entries    []T    `json:"entries"`
}

func decodeInteractionTable[T any](data []byte) (interactionTableDocument[T], error) {
	var document interactionTableDocument[T]
	if err := json.Unmarshal(data, &document); err != nil {
		return document, err
	}
	if document.Provenance != "curated" || len(document.Entries) == 0 {
		return document, fmt.Errorf("interaction table provenance/entries = %q/%d, want curated/non-empty", document.Provenance, len(document.Entries))
	}
	return document, nil
}

func loadGanInteractions() error {
	type entry struct {
		Earth      string `json:"di_pan_gan"`
		Heaven     string `json:"tian_pan_gan"`
		Pattern    string `json:"pattern"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	document, err := decodeInteractionTable[entry](ganInteractionJSON)
	if err != nil {
		return err
	}
	entries := document.Entries
	ganInteractionTable = make(map[[2]ganzhi.Gan]ganEntry, len(entries))
	for _, entry := range entries {
		earth, err := ganzhi.ParseGan(entry.Earth)
		if err != nil {
			return err
		}
		heaven, err := ganzhi.ParseGan(entry.Heaven)
		if err != nil {
			return err
		}
		ganInteractionTable[[2]ganzhi.Gan{earth, heaven}] = ganEntry{
			PatternName: entry.Pattern,
			Meaning:     entry.Meaning,
			Auspicious:  entry.Auspicious,
		}
	}
	return nil
}

func loadJushu() error {
	var entries []struct {
		JieQi     string  `json:"jie_qi"`
		Longitude float64 `json:"longitude"`
		ShangYuan int     `json:"shang_yuan"`
		ZhongYuan int     `json:"zhong_yuan"`
		XiaYuan   int     `json:"xia_yuan"`
		YangDun   bool    `json:"yang_dun"`
	}
	if err := json.Unmarshal(jushuJSON, &entries); err != nil {
		return err
	}
	if len(entries) != 24 {
		return fmt.Errorf("load jushu: want 24 terms, got %d", len(entries))
	}
	solarTermJuTable = make(map[string][4]int, len(entries))
	solarTermLongitudes = make(map[string]float64, len(entries))
	for _, entry := range entries {
		if _, exists := solarTermJuTable[entry.JieQi]; exists {
			return fmt.Errorf("load jushu: duplicate term %q", entry.JieQi)
		}
		if entry.Longitude < 0 || entry.Longitude >= 360 {
			return fmt.Errorf("load jushu: invalid longitude for %q", entry.JieQi)
		}
		index := solarTermCanonicalIndex(entry.Longitude)
		if index < 0 || solarTermOrder[index] != "" {
			return fmt.Errorf("load jushu: invalid or duplicate canonical position for %q", entry.JieQi)
		}
		for name, longitude := range solarTermLongitudes {
			if longitude == entry.Longitude {
				return fmt.Errorf("load jushu: %q and %q share longitude %v", name, entry.JieQi, longitude)
			}
		}
		if entry.ShangYuan < 1 || entry.ShangYuan > 9 || entry.ZhongYuan < 1 ||
			entry.ZhongYuan > 9 || entry.XiaYuan < 1 || entry.XiaYuan > 9 {
			return fmt.Errorf("load jushu: invalid dingju for %q", entry.JieQi)
		}
		yang := 0
		if entry.YangDun {
			yang = 1
		}
		solarTermLongitudes[entry.JieQi] = entry.Longitude
		solarTermJuTable[entry.JieQi] = [4]int{
			entry.ShangYuan, entry.ZhongYuan, entry.XiaYuan, yang,
		}
		solarTermOrder[index] = entry.JieQi
	}
	for index, name := range solarTermOrder {
		if name == "" {
			return fmt.Errorf("load jushu: canonical term %d is missing", index)
		}
	}
	return nil
}

func loadMenInteractions() error {
	type entry struct {
		Door    string `json:"door"`
		Gong    string `json:"gong"`
		Name    string `json:"name"`
		Meaning string `json:"meaning"`
	}
	document, err := decodeInteractionTable[entry](menInteractionJSON)
	if err != nil {
		return err
	}
	entries := document.Entries
	menGongTable = make(map[[2]int]doorEntry, len(entries))
	for _, entry := range entries {
		door, err := ParseDoorIndex(entry.Door)
		if err != nil {
			return err
		}
		palace, err := ParsePalaceIndex(entry.Gong)
		if err != nil {
			return err
		}
		menGongTable[[2]int{int(door), int(palace) - 1}] = doorEntry{
			Name: entry.Name, Meaning: entry.Meaning,
		}
	}
	return nil
}

func loadXingInteractions() error {
	type entry struct {
		Star       string `json:"xing"`
		Gong       string `json:"gong"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	document, err := decodeInteractionTable[entry](xingInteractionJSON)
	if err != nil {
		return err
	}
	entries := document.Entries
	xingGongTable = make(map[[2]int]XingInteraction, len(entries))
	for _, e := range entries {
		s, err := ParseStarIndex(e.Star)
		if err != nil {
			return err
		}
		p, err := ParsePalaceIndex(e.Gong)
		if err != nil {
			return err
		}
		xingGongTable[[2]int{int(s), int(p) - 1}] = XingInteraction{
			Star: e.Star, Gong: e.Gong,
			Meaning: e.Meaning, Auspicious: e.Auspicious,
		}
	}
	return nil
}

func loadCatalog() error {
	var table struct {
		DayBoundary string `json:"day_boundary"`
		Default     struct {
			Scope        string `json:"scope"`
			School       string `json:"school"`
			DingjuMethod string `json:"dingju_method"`
		} `json:"default"`
		Scopes []struct {
			Scope              string   `json:"scope"`
			Name               string   `json:"name"`
			LeadPillar         string   `json:"lead_pillar"`
			MethodSource       string   `json:"method_source"`
			DingjuMethods      []string `json:"dingju_methods"`
			DefaultDingju      string   `json:"default_dingju_method"`
			NoDingjuSchools    []string `json:"no_dingju_schools"`
			QuarterRules       []string `json:"quarter_rules"`
			DefaultQuarterRule string   `json:"default_quarter_rule"`
			Features           []string `json:"features"`
		} `json:"scopes"`
		Schools []struct {
			School               string   `json:"school"`
			Name                 string   `json:"name"`
			AllowedScopes        []string `json:"allowed_scopes"`
			AllowedDingjuMethods []string `json:"allowed_dingju_methods"`
			AllowedQuarterRules  []string `json:"allowed_quarter_rules"`
			SpiritMode           string   `json:"spirit_mode"`
			Geometry             string   `json:"geometry"`
			StarMode             string   `json:"star_mode"`
			DoorMode             string   `json:"door_mode"`
			SpiritFlight         string   `json:"spirit_flight"`
		} `json:"schools"`
		DingjuMethods []struct {
			DingjuMethod string `json:"dingju_method"`
			Name         string `json:"name"`
		} `json:"dingju_methods"`
		QuarterRules []struct {
			QuarterRule string `json:"quarter_rule"`
			Name        string `json:"name"`
		} `json:"quarter_rules"`
		MonthDingju struct {
			FirstMonthBranch string `json:"first_month_branch"`
			Direction        string `json:"direction"`
			YinDun           bool   `json:"yin_dun"`
			Groups           []struct {
				Branches  []string `json:"branches"`
				LeadingJu int      `json:"leading_ju"`
				Yuan      string   `json:"yuan"`
			} `json:"groups"`
		} `json:"month_dingju"`
		YearDingju struct {
			AnchorYear int  `json:"anchor_year"`
			CycleYears int  `json:"cycle_years"`
			YinDun     bool `json:"yin_dun"`
			YuanJu     []struct {
				Yuan string `json:"yuan"`
				Ju   int    `json:"ju"`
			} `json:"yuan_ju"`
		} `json:"year_dingju"`
	}
	if err := json.Unmarshal(catalogJSON, &table); err != nil {
		return err
	}
	if table.DayBoundary != "late_zi_rolls_day_pillar" {
		return fmt.Errorf("load catalog: unsupported day boundary")
	}
	qimenDayBoundary = table.DayBoundary
	parsedDefaultScope, err := ParseScope(table.Default.Scope)
	if err != nil {
		return fmt.Errorf("load catalog: default scope: %w", err)
	}
	parsedDefaultSchool, err := ParseSchool(table.Default.School)
	if err != nil {
		return fmt.Errorf("load catalog: default school: %w", err)
	}
	defaultScope, defaultSchool = parsedDefaultScope, parsedDefaultSchool
	defaultDingju, err := ParseDingjuMethod(table.Default.DingjuMethod)
	if err != nil {
		return fmt.Errorf("load catalog: default dingju method: %w", err)
	}
	if len(table.Scopes) != 5 || len(table.Schools) != 4 || len(table.DingjuMethods) != 4 ||
		len(table.QuarterRules) != 2 || len(table.MonthDingju.Groups) != 3 || len(table.YearDingju.YuanJu) != 3 {
		return fmt.Errorf("load catalog: invalid dimensions")
	}

	dingjuMethodNames = make(map[DingjuMethod]string, len(table.DingjuMethods))
	for _, entry := range table.DingjuMethods {
		method, err := ParseDingjuMethod(entry.DingjuMethod)
		if err != nil {
			return fmt.Errorf("load catalog: %w", err)
		}
		if _, exists := dingjuMethodNames[method]; exists || entry.Name == "" {
			return fmt.Errorf("load catalog: invalid dingju method %q", entry.DingjuMethod)
		}
		dingjuMethodNames[method] = entry.Name
	}
	for _, method := range []DingjuMethod{DingjuChaiBu, DingjuZhiRun, DingjuMaoShan, DingjuNone} {
		if dingjuMethodNames[method] == "" {
			return fmt.Errorf("load catalog: missing dingju method %q", method)
		}
	}
	quarterRuleNames = make(map[QuarterRule]string, len(table.QuarterRules))
	for _, entry := range table.QuarterRules {
		rule, err := ParseQuarterRule(entry.QuarterRule)
		if err != nil {
			return fmt.Errorf("load catalog: %w", err)
		}
		if _, exists := quarterRuleNames[rule]; exists || entry.Name == "" {
			return fmt.Errorf("load catalog: invalid quarter rule %q", entry.QuarterRule)
		}
		quarterRuleNames[rule] = entry.Name
	}
	for _, rule := range []QuarterRule{QuarterTenMinuteSanYuan, QuarterTwelveMinuteTenDivision} {
		if quarterRuleNames[rule] == "" {
			return fmt.Errorf("load catalog: missing quarter rule %q", rule)
		}
	}

	scopeMethods = make(map[Scope]scopeEntry, len(table.Scopes))
	for _, entry := range table.Scopes {
		scope, err := ParseScope(entry.Scope)
		if err != nil {
			return fmt.Errorf("load catalog: %w", err)
		}
		if _, exists := scopeMethods[scope]; exists {
			return fmt.Errorf("load catalog: duplicate scope %q", entry.Scope)
		}
		if entry.Name == "" || entry.LeadPillar != scope.String() {
			return fmt.Errorf("load catalog: invalid scope %q", entry.Scope)
		}
		switch entry.MethodSource {
		case "solar_term", "quarter_rule", "month_cycle", "year_cycle":
		default:
			return fmt.Errorf("load catalog: scope %q has invalid method source", entry.Scope)
		}
		loaded := scopeEntry{
			Scope: entry.Scope, Name: entry.Name, LeadPillar: entry.LeadPillar,
			MethodSource: entry.MethodSource,
		}
		for _, name := range entry.DingjuMethods {
			method, err := ParseDingjuMethod(name)
			if err != nil {
				return fmt.Errorf("load catalog: scope %q: %w", entry.Scope, err)
			}
			if dingjuMethodNames[method] == "" || loaded.allowsDingju(method) {
				return fmt.Errorf("load catalog: scope %q has invalid dingju method %q", entry.Scope, name)
			}
			loaded.DingjuMethods = append(loaded.DingjuMethods, method)
		}
		if entry.DefaultDingju != "" {
			method, err := ParseDingjuMethod(entry.DefaultDingju)
			if err != nil || !loaded.allowsDingju(method) {
				return fmt.Errorf("load catalog: scope %q has invalid default dingju method %q", entry.Scope, entry.DefaultDingju)
			}
			loaded.DefaultDingju = method
		} else if len(entry.DingjuMethods) > 0 {
			return fmt.Errorf("load catalog: scope %q lacks a default dingju method", entry.Scope)
		}
		for _, name := range entry.QuarterRules {
			rule, err := ParseQuarterRule(name)
			if err != nil {
				return fmt.Errorf("load catalog: scope %q: %w", entry.Scope, err)
			}
			if quarterRuleNames[rule] == "" || loaded.allowsQuarterRule(rule) {
				return fmt.Errorf("load catalog: scope %q has invalid quarter rule %q", entry.Scope, name)
			}
			loaded.QuarterRules = append(loaded.QuarterRules, rule)
		}
		if entry.DefaultQuarterRule != "" {
			rule, err := ParseQuarterRule(entry.DefaultQuarterRule)
			if err != nil || !loaded.allowsQuarterRule(rule) {
				return fmt.Errorf("load catalog: scope %q has invalid default quarter rule %q", entry.Scope, entry.DefaultQuarterRule)
			}
			loaded.DefaultQuarterRule = rule
		} else if len(entry.QuarterRules) > 0 {
			return fmt.Errorf("load catalog: scope %q lacks a default quarter rule", entry.Scope)
		}
		for _, feature := range entry.Features {
			if feature != "wu_bu_yu_shi" || loaded.hasFeature(feature) {
				return fmt.Errorf("load catalog: scope %q has invalid feature %q", entry.Scope, feature)
			}
			loaded.Features = append(loaded.Features, feature)
		}
		for _, name := range entry.NoDingjuSchools {
			school, err := ParseSchool(name)
			if err != nil || loaded.allowsNoDingjuSchool(school) {
				return fmt.Errorf("load catalog: scope %q has invalid no-dingju school %q", entry.Scope, name)
			}
			loaded.NoDingjuSchools = append(loaded.NoDingjuSchools, school)
		}
		scopeMethods[scope] = loaded
	}
	for _, scope := range []Scope{ScopeHour, ScopeDay, ScopeQuarter, ScopeMonth, ScopeYear} {
		if _, ok := scopeMethods[scope]; !ok {
			return fmt.Errorf("load catalog: missing scope %q", scope)
		}
	}

	schoolMethods = make(map[School]schoolEntry, len(table.Schools))
	for _, entry := range table.Schools {
		school, err := ParseSchool(entry.School)
		if err != nil {
			return fmt.Errorf("load catalog: %w", err)
		}
		if _, exists := schoolMethods[school]; exists || entry.Name == "" || entry.SpiritMode == "" ||
			entry.Geometry == "" || entry.StarMode == "" || entry.DoorMode == "" || entry.SpiritFlight == "" {
			return fmt.Errorf("load catalog: invalid school %q", entry.School)
		}
		loadedSchool := schoolEntry{
			School: entry.School, Name: entry.Name, SpiritMode: entry.SpiritMode,
			Geometry: entry.Geometry, StarMode: entry.StarMode,
			DoorMode: entry.DoorMode, SpiritFlight: entry.SpiritFlight,
		}
		for _, name := range entry.AllowedScopes {
			allowedScope, err := ParseScope(name)
			if err != nil || loadedSchool.allowsScope(allowedScope) {
				return fmt.Errorf("load catalog: school %q has invalid scope %q", entry.School, name)
			}
			loadedSchool.AllowedScopes = append(loadedSchool.AllowedScopes, allowedScope)
		}
		for _, name := range entry.AllowedDingjuMethods {
			method, err := ParseDingjuMethod(name)
			if err != nil || loadedSchool.allowsDingju(method) {
				return fmt.Errorf("load catalog: school %q has invalid dingju method %q", entry.School, name)
			}
			loadedSchool.AllowedDingjuMethods = append(loadedSchool.AllowedDingjuMethods, method)
		}
		for _, name := range entry.AllowedQuarterRules {
			rule, err := ParseQuarterRule(name)
			if err != nil || loadedSchool.allowsQuarterRule(rule) {
				return fmt.Errorf("load catalog: school %q has invalid quarter rule %q", entry.School, name)
			}
			loadedSchool.AllowedQuarterRules = append(loadedSchool.AllowedQuarterRules, rule)
		}
		if len(loadedSchool.AllowedScopes) == 0 || len(loadedSchool.AllowedDingjuMethods) == 0 {
			return fmt.Errorf("load catalog: school %q lacks allowed scope/dingju method", entry.School)
		}
		schoolMethods[school] = loadedSchool
	}
	for _, school := range []School{SchoolZhuanPan, SchoolLuoShuFeiPan, SchoolMingFaFeiPan, SchoolJinhanYuJing} {
		if _, ok := schoolMethods[school]; !ok {
			return fmt.Errorf("load catalog: missing school %q", school)
		}
	}
	for scope, entry := range scopeMethods {
		for _, school := range entry.NoDingjuSchools {
			schoolEntry, ok := schoolMethods[school]
			if !ok || !schoolEntry.allowsScope(scope) || !schoolEntry.allowsDingju(DingjuNone) {
				return fmt.Errorf("load catalog: scope %q has invalid no-dingju school %q", scope, school)
			}
		}
	}

	defaultMethod := Method{Scope: defaultScope, School: defaultSchool, Dingju: defaultDingju}
	if _, err := ParseMethod(defaultMethod.Scope.String(), defaultMethod.School.String(), defaultMethod.Dingju.String(), ""); err != nil {
		return fmt.Errorf("load catalog: invalid default method: %w", err)
	}
	if table.MonthDingju.Direction != "retreat" || !table.MonthDingju.YinDun || !table.YearDingju.YinDun {
		return fmt.Errorf("load catalog: unsupported month/year dingju rule")
	}
	firstBranch, err := ganzhi.ParseZhi(table.MonthDingju.FirstMonthBranch)
	if err != nil {
		return fmt.Errorf("load catalog: month first branch: %w", err)
	}
	monthFirstBranch = firstBranch
	monthGroups = make(map[ganzhi.Zhi]monthYearGroupEntry)
	for _, group := range table.MonthDingju.Groups {
		if group.LeadingJu < 1 || group.LeadingJu > 9 || group.Yuan == "" {
			return fmt.Errorf("load catalog: invalid month group")
		}
		for _, name := range group.Branches {
			branch, err := ganzhi.ParseZhi(name)
			if err != nil {
				return fmt.Errorf("load catalog: month branch: %w", err)
			}
			if _, exists := monthGroups[branch]; exists {
				return fmt.Errorf("load catalog: duplicate month branch %q", name)
			}
			monthGroups[branch] = monthYearGroupEntry{LeadingJu: group.LeadingJu, Yuan: group.Yuan}
		}
	}
	if len(monthGroups) != 12 {
		return fmt.Errorf("load catalog: month groups must cover twelve branches")
	}

	yearAnchor, yearCycle = table.YearDingju.AnchorYear, table.YearDingju.CycleYears
	yearJuByYuan = make(map[string]int, len(table.YearDingju.YuanJu))
	for _, entry := range table.YearDingju.YuanJu {
		if entry.Ju < 1 || entry.Ju > 9 || entry.Yuan == "" {
			return fmt.Errorf("load catalog: invalid year dingju entry")
		}
		if _, exists := yearJuByYuan[entry.Yuan]; exists {
			return fmt.Errorf("load catalog: duplicate year yuan %q", entry.Yuan)
		}
		yearJuByYuan[entry.Yuan] = entry.Ju
	}
	for _, yuan := range []string{"上元", "中元", "下元"} {
		if _, ok := yearJuByYuan[yuan]; !ok {
			return fmt.Errorf("load catalog: missing year yuan %q", yuan)
		}
	}
	if yearAnchor <= 0 || yearCycle != 60 {
		return fmt.Errorf("load catalog: invalid year cycle")
	}
	return nil
}

func loadQuarter() error {
	var table struct {
		HourBoundary string             `json:"hour_boundary"`
		YangBranches []string           `json:"yang_branches"`
		YinBranches  []string           `json:"yin_branches"`
		Methods      []quarterJSONEntry `json:"methods"`
	}
	if err := json.Unmarshal(quarterJSON, &table); err != nil {
		return fmt.Errorf("load quarter: %w", err)
	}
	if table.HourBoundary != qimenDayBoundary || len(table.Methods) != 2 ||
		len(table.YangBranches) != 6 || len(table.YinBranches) != 6 {
		return fmt.Errorf("load quarter: unsupported rule")
	}

	quarterYangBranch = make(map[ganzhi.Zhi]bool, len(table.YangBranches))
	for _, name := range table.YangBranches {
		branch, err := ganzhi.ParseZhi(name)
		if err != nil {
			return fmt.Errorf("load quarter: yang branch: %w", err)
		}
		if quarterYangBranch[branch] {
			return fmt.Errorf("load quarter: duplicate yang branch %q", name)
		}
		quarterYangBranch[branch] = true
	}
	for _, name := range table.YinBranches {
		branch, err := ganzhi.ParseZhi(name)
		if err != nil {
			return fmt.Errorf("load quarter: yin branch: %w", err)
		}
		if quarterYangBranch[branch] {
			return fmt.Errorf("load quarter: branch %q has duplicate yin-yang class", name)
		}
		quarterYangBranch[branch] = false
	}
	if len(quarterYangBranch) != 12 {
		return fmt.Errorf("load quarter: yin-yang branches must cover twelve branches")
	}

	quarterRules = make(map[QuarterRule]quarterVariant, len(table.Methods))
	for _, entry := range table.Methods {
		rule, err := ParseQuarterRule(entry.Rule)
		if err != nil {
			return fmt.Errorf("load quarter: %w", err)
		}
		if _, exists := quarterRules[rule]; exists {
			return fmt.Errorf("load quarter: duplicate rule %q", entry.Rule)
		}
		if entry.Minutes < 1 || entry.Count < 1 || entry.Minutes*entry.Count != 120 ||
			(entry.LeadPillarMode != "quarter" && entry.LeadPillarMode != "hour") ||
			entry.YongJuJieQi == "" {
			return fmt.Errorf("load quarter: invalid rule %q", entry.Rule)
		}
		variant := quarterVariant{
			Rule: rule, Minutes: entry.Minutes, Count: entry.Count,
			LeadPillarMode: entry.LeadPillarMode, YongJuJieQi: entry.YongJuJieQi,
		}
		switch rule {
		case QuarterTenMinuteSanYuan:
			if err := loadTenMinuteQuarter(entry); err != nil {
				return err
			}
		case QuarterTwelveMinuteTenDivision:
			loaded, err := loadTwelveMinuteQuarter(entry, variant)
			if err != nil {
				return err
			}
			variant = loaded
		default:
			return fmt.Errorf("load quarter: unsupported rule %q", rule)
		}
		quarterRules[rule] = variant
	}
	for _, rule := range []QuarterRule{QuarterTenMinuteSanYuan, QuarterTwelveMinuteTenDivision} {
		if _, ok := quarterRules[rule]; !ok {
			return fmt.Errorf("load quarter: missing rule %q", rule)
		}
	}
	return nil
}

func loadTenMinuteQuarter(entry quarterJSONEntry) error {
	if entry.PillarRule == nil || entry.DingjuRule == nil ||
		entry.Minutes != 10 || entry.Count != 12 || entry.LeadPillarMode != "quarter" ||
		entry.PillarRule.Source != "hour_stem_five_rat_escape" ||
		entry.DingjuRule.YinYangSource != "hour_branch" ||
		entry.DingjuRule.YuanSource != "hour_five_unit_fu_tou" ||
		entry.DingjuRule.FuTouUnits != 5 ||
		len(entry.PillarRule.Starts) != 10 || len(entry.DingjuRule.Groups) != 3 {
		return fmt.Errorf("load quarter: invalid ten-minute rule")
	}
	quarterFuTouUnits = entry.DingjuRule.FuTouUnits
	quarterStartIndex = make(map[ganzhi.Gan]int, len(entry.PillarRule.Starts))
	for _, start := range entry.PillarRule.Starts {
		hourGan, err := ganzhi.ParseGan(start.HourGan)
		if err != nil {
			return fmt.Errorf("load quarter: hour gan: %w", err)
		}
		if len([]rune(start.Start)) != 2 {
			return fmt.Errorf("load quarter: invalid start pillar %q", start.Start)
		}
		startGan, err := ganzhi.ParseGan(string([]rune(start.Start)[0]))
		startZhi, err2 := ganzhi.ParseZhi(string([]rune(start.Start)[1]))
		if err != nil || err2 != nil || startZhi != ganzhi.ZhiZi {
			return fmt.Errorf("load quarter: invalid start pillar %q", start.Start)
		}
		if _, exists := quarterStartIndex[hourGan]; exists {
			return fmt.Errorf("load quarter: duplicate hour gan %q", start.HourGan)
		}
		quarterStartIndex[hourGan] = ganzhi.SixtyCycleIndex(startGan, startZhi)
	}

	quarterSanyuanByBranch = make(map[ganzhi.Zhi]quarterSanyuanEntry, 12)
	for _, group := range entry.DingjuRule.Groups {
		if len(group.Branches) != 4 || group.Yuan == "" ||
			group.YangJu < 1 || group.YangJu > 9 || group.YinJu < 1 || group.YinJu > 9 {
			return fmt.Errorf("load quarter: invalid dingju group %+v", group)
		}
		for _, name := range group.Branches {
			branch, err := ganzhi.ParseZhi(name)
			if err != nil {
				return fmt.Errorf("load quarter: dingju branch: %w", err)
			}
			if _, exists := quarterSanyuanByBranch[branch]; exists {
				return fmt.Errorf("load quarter: duplicate dingju branch %q", name)
			}
			quarterSanyuanByBranch[branch] = quarterSanyuanEntry{
				Yuan: group.Yuan, YangJu: group.YangJu, YinJu: group.YinJu,
			}
		}
	}
	if len(quarterSanyuanByBranch) != 12 {
		return fmt.Errorf("load quarter: dingju groups must cover twelve branches")
	}
	return nil
}

func loadTwelveMinuteQuarter(entry quarterJSONEntry, variant quarterVariant) (quarterVariant, error) {
	if len(entry.BaseDingjuMethods) != 3 || entry.JuShift == nil ||
		entry.Minutes != 12 || entry.Count != 10 || entry.LeadPillarMode != "hour" ||
		entry.BaseScope != ScopeHour.String() || entry.BaseSchool != SchoolZhuanPan.String() ||
		len(entry.DunSources) != 2 || len(entry.HourBoundaries) != 2 ||
		entry.JuShift.Cycle != 9 || entry.JuShift.YangStep != 1 || entry.JuShift.YinStep != -1 {
		return variant, fmt.Errorf("load quarter: invalid twelve-minute rule")
	}
	baseScope, err := ParseScope(entry.BaseScope)
	if err != nil {
		return variant, fmt.Errorf("load quarter: twelve-minute base scope: %w", err)
	}
	baseSchool, err := ParseSchool(entry.BaseSchool)
	if err != nil {
		return variant, fmt.Errorf("load quarter: twelve-minute base school: %w", err)
	}
	variant.BaseScope, variant.BaseSchool = baseScope, baseSchool
	variant.BaseDingjuMethods = make(map[DingjuMethod]bool, len(entry.BaseDingjuMethods))
	for _, name := range entry.BaseDingjuMethods {
		method, err := ParseDingjuMethod(name)
		if err != nil || variant.BaseDingjuMethods[method] {
			return variant, fmt.Errorf("load quarter: invalid base dingju method %q", name)
		}
		if !schoolMethods[baseSchool].allowsDingju(method) {
			return variant, fmt.Errorf("load quarter: base dingju method %q is unsupported", name)
		}
		variant.BaseDingjuMethods[method] = true
	}
	for _, method := range []DingjuMethod{DingjuChaiBu, DingjuZhiRun, DingjuMaoShan} {
		if !variant.BaseDingjuMethods[method] {
			return variant, fmt.Errorf("load quarter: twelve-minute rule lacks base dingju method %q", method)
		}
	}
	variant.DunSources = make(map[string]bool, len(entry.DunSources))
	defaultDunCount := 0
	for _, option := range entry.DunSources {
		if option.Source != "hour_branch" && option.Source != "solar_term" {
			return variant, fmt.Errorf("load quarter: invalid dun source %q", option.Source)
		}
		if variant.DunSources[option.Source] {
			return variant, fmt.Errorf("load quarter: duplicate dun source %q", option.Source)
		}
		if option.Default {
			variant.DefaultDunSource = option.Source
			defaultDunCount++
		}
		variant.DunSources[option.Source] = true
	}
	if defaultDunCount != 1 || !variant.DunSources["hour_branch"] || !variant.DunSources["solar_term"] {
		return variant, fmt.Errorf("load quarter: invalid dun source defaults")
	}
	variant.HourBoundaries = make(map[string]int, len(entry.HourBoundaries))
	defaultBoundaryCount := 0
	for _, option := range entry.HourBoundaries {
		if option.Boundary != "late_zi" && option.Boundary != "zi_zheng" {
			return variant, fmt.Errorf("load quarter: invalid hour boundary %q", option.Boundary)
		}
		if _, exists := variant.HourBoundaries[option.Boundary]; exists {
			return variant, fmt.Errorf("load quarter: duplicate hour boundary %q", option.Boundary)
		}
		if option.HourStartShift != 0 && option.HourStartShift != 1 {
			return variant, fmt.Errorf("load quarter: invalid hour boundary shift %d", option.HourStartShift)
		}
		if option.Default {
			variant.DefaultHourBoundary = option.Boundary
			defaultBoundaryCount++
		}
		variant.HourBoundaries[option.Boundary] = option.HourStartShift
	}
	if defaultBoundaryCount != 1 || variant.HourBoundaries["late_zi"] != 1 || variant.HourBoundaries["zi_zheng"] != 0 {
		return variant, fmt.Errorf("load quarter: invalid hour boundary defaults")
	}
	variant.JuCycle, variant.YangStep, variant.YinStep =
		entry.JuShift.Cycle, entry.JuShift.YangStep, entry.JuShift.YinStep
	return variant, nil
}

func loadPlate() error {
	var table struct {
		TianQin struct {
			Star                  string `json:"star"`
			HomeGong              string `json:"home_gong"`
			LodgingGong           string `json:"lodging_gong"`
			Follows               string `json:"follows"`
			CarriesCenterEarthGan bool   `json:"carries_center_earth_gan"`
			LodgingRule           string `json:"lodging_rule"`
		} `json:"tian_qin"`
		OuterRing     []string `json:"outer_ring"`
		EarthGanOrder []string `json:"earth_gan_order"`
		StarOrder     []string `json:"star_order"`
		DoorOrder     []string `json:"door_order"`
		Spirits       []struct {
			Slot int    `json:"slot"`
			Yang string `json:"yang"`
			Yin  string `json:"yin"`
		} `json:"spirits"`
		FlySpiritOrder    []string `json:"fly_spirit_order"`
		MingFaDoorOrder   []string `json:"mingfa_door_order"`
		MingFaSpiritOrder []string `json:"mingfa_spirit_order"`
		GongWuxing        []struct {
			Gong   string `json:"gong"`
			Wuxing string `json:"wuxing"`
		} `json:"gong_wuxing"`
		StarWuxing []struct {
			Star   string `json:"star"`
			Wuxing string `json:"wuxing"`
		} `json:"star_wuxing"`
		DoorWuxing []struct {
			Door   string `json:"door"`
			Wuxing string `json:"wuxing"`
		} `json:"door_wuxing"`
		WuxingRelations []struct {
			Relation         string `json:"relation"`
			Name             string `json:"name"`
			TraditionalLabel string `json:"traditional_label"`
		} `json:"wuxing_relations"`
		RiShiRelations []struct {
			Relation string `json:"relation"`
			Name     string `json:"name"`
		} `json:"ri_shi_relations"`
		ZhiGong map[string]string `json:"zhi_gong"`
		MaXing  []struct {
			Branches []string `json:"branches"`
			Ma       string   `json:"ma"`
		} `json:"ma_xing"`
		LiuJiaLiuYi []struct {
			Xun   string `json:"xun"`
			Zhi   string `json:"zhi"`
			LiuYi string `json:"liu_yi"`
		} `json:"liu_jia_liu_yi"`
		XunKong []struct {
			Xun      string   `json:"xun"`
			Branches []string `json:"branches"`
		} `json:"xun_kong"`
		WuBuYuShi []string `json:"wu_bu_yu_shi"`
		FuTouYuan []struct {
			Branches []string `json:"branches"`
			Yuan     string   `json:"yuan"`
		} `json:"fu_tou_yuan"`
	}
	if err := json.Unmarshal(plateJSON, &table); err != nil {
		return err
	}
	if len(table.OuterRing) != 8 || len(table.EarthGanOrder) != 9 ||
		len(table.StarOrder) != 8 || len(table.DoorOrder) != 8 ||
		len(table.Spirits) != 8 || len(table.GongWuxing) != 9 ||
		len(table.StarWuxing) != 9 || len(table.DoorWuxing) != 9 ||
		len(table.WuxingRelations) != 5 || len(table.RiShiRelations) != 5 ||
		len(table.ZhiGong) != 12 || len(table.LiuJiaLiuYi) != 6 ||
		len(table.XunKong) != 6 || len(table.WuBuYuShi) != 10 ||
		len(table.MingFaDoorOrder) != 9 || len(table.MingFaSpiritOrder) != 9 ||
		len(table.FuTouYuan) != 3 {
		return fmt.Errorf("load plate: invalid table dimensions")
	}

	for i, name := range table.OuterRing {
		palace, err := ParsePalaceIndex(name)
		if err != nil {
			return err
		}
		outerRing[i] = palace
	}
	for i, name := range table.EarthGanOrder {
		gan, err := ganzhi.ParseGan(name)
		if err != nil {
			return err
		}
		earthGanOrder[i] = gan
	}
	for i, name := range table.StarOrder {
		star, err := ParseStarIndex(name)
		if err != nil {
			return err
		}
		starOrder8[i] = star
	}
	for i, name := range table.DoorOrder {
		door, err := ParseDoorIndex(name)
		if err != nil {
			return err
		}
		doorOrder[i] = door
	}
	for i, name := range table.MingFaDoorOrder {
		door, err := ParseDoorIndex(name)
		if err != nil {
			return fmt.Errorf("load plate: Ming Fa door %q: %w", name, err)
		}
		mingfaDoorOrder[i] = door
	}
	for i, palace := range outerRing {
		palaceStar[int(palace)-1] = starOrder8[i]
		palaceDoor[int(palace)-1] = doorOrder[i]
	}
	for i, entry := range table.Spirits {
		if entry.Slot < 1 || entry.Slot > 8 || entry.Yang == "" || entry.Yin == "" {
			return fmt.Errorf("load plate: invalid spirit entry %d", i)
		}
		spiritOrder[i] = SpiritIndex(entry.Slot)
		spiritYangNames[entry.Slot] = entry.Yang
		spiritYinNames[entry.Slot] = entry.Yin
	}
	if len(table.FlySpiritOrder) != 9 {
		return fmt.Errorf("load plate: fly spirit order must contain nine members")
	}
	for i, name := range table.FlySpiritOrder {
		var spirit SpiritIndex
		if name == "太常" {
			spirit = SpiritTaiChang
		} else {
			parsed, err := ParseSpiritIndex(name)
			if err != nil {
				return fmt.Errorf("load plate: fly spirit %q: %w", name, err)
			}
			spirit = parsed
		}
		flySpiritOrder[i] = spirit
		flySpiritNames[int(spirit)] = name
	}
	for i, name := range table.MingFaSpiritOrder {
		var spirit SpiritIndex
		parsed, err := ParseSpiritIndex(name)
		if err != nil {
			return fmt.Errorf("load plate: Ming Fa spirit %q: %w", name, err)
		}
		spirit = parsed
		mingfaSpiritOrder[i] = spirit
		mingfaSpiritNames[int(spirit)] = name
	}
	for _, entry := range table.GongWuxing {
		palace, err := ParsePalaceIndex(entry.Gong)
		if err != nil {
			return err
		}
		if gongWuxingTable[int(palace)-1] != 0 {
			return fmt.Errorf("load plate: duplicate gong wuxing %q", entry.Gong)
		}
		element, err := ganzhi.ParseWuxing(entry.Wuxing)
		if err != nil {
			return err
		}
		gongWuxingTable[int(palace)-1] = element
	}
	for _, entry := range table.StarWuxing {
		star, err := ParseStarIndex(entry.Star)
		if err != nil {
			return err
		}
		if starWuxingTable[int(star)-1] != 0 {
			return fmt.Errorf("load plate: duplicate star wuxing %q", entry.Star)
		}
		element, err := ganzhi.ParseWuxing(entry.Wuxing)
		if err != nil {
			return err
		}
		starWuxingTable[int(star)-1] = element
	}
	for _, entry := range table.DoorWuxing {
		door, err := ParseDoorIndex(entry.Door)
		if err != nil {
			return err
		}
		if doorWuxingTable[int(door)-1] != 0 {
			return fmt.Errorf("load plate: duplicate door wuxing %q", entry.Door)
		}
		element, err := ganzhi.ParseWuxing(entry.Wuxing)
		if err != nil {
			return err
		}
		doorWuxingTable[int(door)-1] = element
	}
	wuxingRelations = make(map[string]wuxingRelationEntry, len(table.WuxingRelations))
	for _, entry := range table.WuxingRelations {
		if _, exists := wuxingRelations[entry.Relation]; exists {
			return fmt.Errorf("load plate: duplicate wuxing relation %q", entry.Relation)
		}
		wuxingRelations[entry.Relation] = wuxingRelationEntry{
			Name: entry.Name, TraditionalLabel: entry.TraditionalLabel,
		}
	}
	riShiRelations = make(map[string]string, len(table.RiShiRelations))
	for _, entry := range table.RiShiRelations {
		if _, exists := riShiRelations[entry.Relation]; exists {
			return fmt.Errorf("load plate: duplicate ri-shi relation %q", entry.Relation)
		}
		riShiRelations[entry.Relation] = entry.Name
	}
	for branch, palaceName := range table.ZhiGong {
		zhi, err := ganzhi.ParseZhi(branch)
		if err != nil {
			return err
		}
		palace, err := ParsePalaceIndex(palaceName)
		if err != nil {
			return err
		}
		zhiPalaceTable[int(zhi)-1] = palace
	}
	for _, entry := range table.MaXing {
		ma, err := ganzhi.ParseZhi(entry.Ma)
		if err != nil {
			return err
		}
		for _, branch := range entry.Branches {
			zhi, err := ganzhi.ParseZhi(branch)
			if err != nil {
				return err
			}
			if maXingTable[int(zhi)-1] != 0 {
				return fmt.Errorf("load plate: duplicate ma-xing branch %q", branch)
			}
			maXingTable[int(zhi)-1] = ma
		}
	}
	for i, entry := range table.LiuJiaLiuYi {
		zhi, err := ganzhi.ParseZhi(entry.Zhi)
		if err != nil {
			return err
		}
		gan, err := ganzhi.ParseGan(entry.LiuYi)
		if err != nil {
			return err
		}
		liuJiaZhi[i] = zhi
		liuJiaLiuYi[i] = gan
	}
	for _, entry := range table.XunKong {
		if len(entry.Branches) != 2 {
			return fmt.Errorf("load plate: invalid xun_kong %q", entry.Xun)
		}
		xunGan, err := ganzhi.ParseGan(string([]rune(entry.Xun)[:1]))
		if err != nil {
			return err
		}
		xunZhi, err := ganzhi.ParseZhi(string([]rune(entry.Xun)[1:]))
		if err != nil {
			return err
		}
		xunIdx := -1
		for j, zhi := range liuJiaZhi {
			if zhi == xunZhi && liuJiaLiuYi[j] != 0 && ganzhi.SixtyCycleIndex(xunGan, xunZhi) == j*10 {
				xunIdx = j
				break
			}
		}
		if xunIdx < 0 {
			return fmt.Errorf("load plate: unknown xun_kong %q", entry.Xun)
		}
		for j, branch := range entry.Branches {
			zhi, err := ganzhi.ParseZhi(branch)
			if err != nil {
				return err
			}
			xunKongTable[xunIdx][j] = zhi
		}
	}
	wuBuYuShiTable = make(map[[2]ganzhi.Gan]bool, len(table.WuBuYuShi))
	for _, pair := range table.WuBuYuShi {
		if len([]rune(pair)) != 2 {
			return fmt.Errorf("load plate: invalid wu_bu_yu_shi %q", pair)
		}
		day, err := ganzhi.ParseGan(string([]rune(pair)[0]))
		if err != nil {
			return err
		}
		hour, err := ganzhi.ParseGan(string([]rune(pair)[1]))
		if err != nil {
			return err
		}
		wuBuYuShiTable[[2]ganzhi.Gan{day, hour}] = true
	}
	fuTouYuanTable = make(map[ganzhi.Zhi]int, 12)
	for _, entry := range table.FuTouYuan {
		var yuan int
		switch entry.Yuan {
		case "上元":
			yuan = 0
		case "中元":
			yuan = 1
		case "下元":
			yuan = 2
		default:
			return fmt.Errorf("load plate: invalid yuan %q", entry.Yuan)
		}
		yuanNames[yuan] = entry.Yuan
		for _, branch := range entry.Branches {
			zhi, err := ganzhi.ParseZhi(branch)
			if err != nil {
				return err
			}
			if _, exists := fuTouYuanTable[zhi]; exists {
				return fmt.Errorf("load plate: duplicate fu-tou branch %q", branch)
			}
			fuTouYuanTable[zhi] = yuan
		}
	}
	lodgingPalace, err := ParsePalaceIndex(table.TianQin.LodgingGong)
	if err != nil {
		return fmt.Errorf("load plate: tian_qin lodging_gong: %w", err)
	}
	if lodgingPalace == GongZhong {
		return fmt.Errorf("load plate: tian_qin lodging_gong cannot be 中")
	}
	homePalace, err := ParsePalaceIndex(table.TianQin.HomeGong)
	if err != nil {
		return fmt.Errorf("load plate: tian_qin home_gong: %w", err)
	}
	follows, err := ParseStarIndex(table.TianQin.Follows)
	if err != nil {
		return fmt.Errorf("load plate: tian_qin follows: %w", err)
	}
	star, err := ParseStarIndex(table.TianQin.Star)
	if err != nil {
		return fmt.Errorf("load plate: tian_qin star: %w", err)
	}
	if homePalace != GongZhong || star != StarTianQin || follows != StarTianRui ||
		!table.TianQin.CarriesCenterEarthGan {
		return fmt.Errorf("load plate: unsupported tian_qin rule")
	}
	palaceStar[int(homePalace)-1] = star
	centerLodgingPalace = lodgingPalace
	tianQinRule = TianQinRule{
		Star: table.TianQin.Star, HomeGong: table.TianQin.HomeGong,
		LodgingGong:           table.TianQin.LodgingGong,
		Follows:               table.TianQin.Follows,
		CarriesCenterEarthGan: table.TianQin.CarriesCenterEarthGan,
		LodgingRule:           table.TianQin.LodgingRule,
	}
	tianQinStar = star
	tianQinFollows = follows
	tianQinCarriesGan = table.TianQin.CarriesCenterEarthGan
	if err := validatePlateTables(); err != nil {
		return err
	}
	return nil
}

func validatePlateTables() error {
	seen := func(values ...int) bool {
		unique := make(map[int]bool, len(values))
		for _, value := range values {
			if value == 0 || unique[value] {
				return false
			}
			unique[value] = true
		}
		return true
	}
	if !seen(intValues(outerRing)...) || !seen(intValues(starOrder8)...) ||
		!seen(intValues(doorOrder)...) || !seen(intValues(spiritOrder)...) {
		return fmt.Errorf("load plate: ring/order tables must contain unique non-empty members")
	}
	if !seen(intValues9(earthGanOrder)...) {
		return fmt.Errorf("load plate: earth gan order must contain all stems")
	}
	flySeen := make(map[SpiritIndex]bool, len(flySpiritOrder))
	for _, spirit := range flySpiritOrder {
		if spirit == 0 || flySeen[spirit] {
			return fmt.Errorf("load plate: fly spirit order must contain unique non-empty members")
		}
		flySeen[spirit] = true
	}
	if !flySeen[SpiritTaiChang] {
		return fmt.Errorf("load plate: fly spirit order must contain 太常")
	}
	mingFaDoorSeen := make(map[DoorIndex]bool, len(mingfaDoorOrder))
	for _, door := range mingfaDoorOrder {
		if door == 0 || mingFaDoorSeen[door] {
			return fmt.Errorf("load plate: Ming Fa door order must contain unique non-empty members")
		}
		mingFaDoorSeen[door] = true
	}
	if !mingFaDoorSeen[DoorZhong] {
		return fmt.Errorf("load plate: Ming Fa door order must contain 中门")
	}
	mingFaSpiritSeen := make(map[SpiritIndex]bool, len(mingfaSpiritOrder))
	for _, spirit := range mingfaSpiritOrder {
		if spirit == 0 || mingFaSpiritSeen[spirit] {
			return fmt.Errorf("load plate: Ming Fa spirit order must contain unique non-empty members")
		}
		mingFaSpiritSeen[spirit] = true
	}
	for i, palace := range outerRing {
		if palace == GongZhong {
			return fmt.Errorf("load plate: outer ring contains 中 at %d", i)
		}
	}
	for palace, element := range gongWuxingTable {
		if element == 0 {
			return fmt.Errorf("load plate: gong %d has no five-element", palace+1)
		}
	}
	for star, element := range starWuxingTable {
		if element == 0 {
			return fmt.Errorf("load plate: star %d has no five-element", star+1)
		}
	}
	for door, element := range doorWuxingTable {
		if element == 0 {
			return fmt.Errorf("load plate: door %d has no five-element", door+1)
		}
	}
	for slot, name := range spiritYangNames {
		if slot > 0 && slot < 9 && name == "" {
			return fmt.Errorf("load plate: yang spirit %d is empty", slot)
		}
	}
	for slot, name := range spiritYinNames {
		if slot > 0 && slot < 9 && name == "" {
			return fmt.Errorf("load plate: yin spirit %d is empty", slot)
		}
	}
	for branch, palace := range zhiPalaceTable {
		if palace == 0 || palace == GongZhong {
			return fmt.Errorf("load plate: branch %d has invalid palace", branch+1)
		}
	}
	for branch, horse := range maXingTable {
		if horse == 0 {
			return fmt.Errorf("load plate: branch %d has no horse", branch+1)
		}
	}
	for i, branch := range liuJiaZhi {
		if branch == 0 || liuJiaLiuYi[i] == 0 || xunKongTable[i][0] == 0 || xunKongTable[i][1] == 0 {
			return fmt.Errorf("load plate: liu-jia entry %d is incomplete", i)
		}
	}
	if len(wuBuYuShiTable) != 10 {
		return fmt.Errorf("load plate: wu_bu_yu_shi must contain ten pairs")
	}
	for branch, yuan := range fuTouYuanTable {
		if yuan < 0 || yuan > 2 || yuanNames[yuan] == "" {
			return fmt.Errorf("load plate: branch %d has invalid fu-tou yuan", branch+1)
		}
	}
	for palace, star := range palaceStar {
		if star == 0 {
			return fmt.Errorf("load plate: palace %d has no home star", palace+1)
		}
	}
	for palace, door := range palaceDoor {
		hasDoor := door != 0
		wantDoor := GongIndex(palace+1) != GongZhong
		if hasDoor != wantDoor {
			return fmt.Errorf("load plate: palace %d home door is invalid", palace+1)
		}
	}
	for _, relation := range []string{
		"same", "gong_generates_xing", "xing_generates_gong",
		"xing_controls_gong", "gong_controls_xing",
	} {
		if _, ok := wuxingRelations[relation]; !ok {
			return fmt.Errorf("load plate: missing wuxing relation %q", relation)
		}
	}
	for _, relation := range []string{
		"same", "ri_generates_shi", "shi_generates_ri",
		"ri_controls_shi", "shi_controls_ri",
	} {
		if _, ok := riShiRelations[relation]; !ok {
			return fmt.Errorf("load plate: missing ri-shi relation %q", relation)
		}
	}
	return nil
}

func intValues[T ~int](values [8]T) []int {
	result := make([]int, len(values))
	for i, value := range values {
		result[i] = int(value)
	}
	return result
}

func intValues9[T ~int](values [9]T) []int {
	result := make([]int, len(values))
	for i, value := range values {
		result[i] = int(value)
	}
	return result
}

func loadPatterns() error {
	rules, err := decodePatternTable(patternsJSON)
	if err != nil {
		return err
	}
	patternRules = rules
	return nil
}

func decodePatternTable(raw []byte) ([]patternRule, error) {
	var rules []struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		Basis            string `json:"basis"`
		Auspicious       bool   `json:"auspicious"`
		GanInteraction   string `json:"gan_interaction"`
		RequiresDutyDoor bool   `json:"requires_duty_door"`
		DutyStarPosition string `json:"duty_star_position"`
		Conditions       []struct {
			HeavenGan []string `json:"heaven_gan"`
			EarthGan  []string `json:"earth_gan"`
			Doors     []string `json:"doors"`
			Spirits   []string `json:"spirits"`
			Palaces   []string `json:"palaces"`
		} `json:"conditions"`
	}
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}
	rulesOut := make([]patternRule, 0, len(rules))
	seen := map[string]bool{}
	for _, rule := range rules {
		if rule.Name == "" || rule.Description == "" || rule.Basis == "" || seen[rule.Name] {
			return nil, fmt.Errorf("load patterns: invalid or duplicate rule %q", rule.Name)
		}
		seen[rule.Name] = true
		if rule.DutyStarPosition != "" {
			if rule.DutyStarPosition != "home" && rule.DutyStarPosition != "opposite" {
				return nil, fmt.Errorf("load patterns: %q has invalid duty star position %q", rule.Name, rule.DutyStarPosition)
			}
			if len(rule.Conditions) > 0 || rule.RequiresDutyDoor || rule.GanInteraction != "" {
				return nil, fmt.Errorf("load patterns: %q duty star position must not define conditions", rule.Name)
			}
		}
		if rule.DutyStarPosition == "" && rule.GanInteraction == "" && len(rule.Conditions) == 0 {
			return nil, fmt.Errorf("load patterns: %q requires conditions", rule.Name)
		}
		loadedRule := patternRule{
			Name: rule.Name, Description: rule.Description, Basis: rule.Basis,
			Auspicious: rule.Auspicious, RequiresDutyDoor: rule.RequiresDutyDoor,
			DutyStarPosition: rule.DutyStarPosition,
		}
		if rule.GanInteraction != "" {
			if rule.Auspicious || len(rule.Conditions) > 0 {
				return nil, fmt.Errorf("load patterns: %q must only reference its gan interaction", rule.Name)
			}
			key, entry, err := uniqueGanInteraction(rule.GanInteraction)
			if err != nil {
				return nil, fmt.Errorf("load patterns: %w", err)
			}
			loadedRule.Auspicious = entry.Auspicious
			loadedRule.Conditions = []patternCondition{{
				HeavenGan: []ganzhi.Gan{key[1]},
				EarthGan:  []ganzhi.Gan{key[0]},
			}}
		}
		for _, condition := range rule.Conditions {
			var loadedCondition patternCondition
			if len(condition.HeavenGan)+len(condition.EarthGan)+
				len(condition.Doors)+len(condition.Spirits)+len(condition.Palaces) == 0 {
				return nil, fmt.Errorf("load patterns: %q has an empty condition", rule.Name)
			}
			for _, name := range condition.HeavenGan {
				gan, err := ganzhi.ParseGan(name)
				if err != nil {
					return nil, err
				}
				if slices.Contains(loadedCondition.HeavenGan, gan) {
					return nil, fmt.Errorf("load patterns: %q has duplicate heaven_gan %q", rule.Name, name)
				}
				loadedCondition.HeavenGan = append(loadedCondition.HeavenGan, gan)
			}
			for _, name := range condition.EarthGan {
				gan, err := ganzhi.ParseGan(name)
				if err != nil {
					return nil, err
				}
				if slices.Contains(loadedCondition.EarthGan, gan) {
					return nil, fmt.Errorf("load patterns: %q has duplicate earth_gan %q", rule.Name, name)
				}
				loadedCondition.EarthGan = append(loadedCondition.EarthGan, gan)
			}
			for _, name := range condition.Doors {
				door, err := ParseDoorIndex(name)
				if err != nil {
					return nil, err
				}
				if slices.Contains(loadedCondition.Doors, door) {
					return nil, fmt.Errorf("load patterns: %q has duplicate door %q", rule.Name, name)
				}
				loadedCondition.Doors = append(loadedCondition.Doors, door)
			}
			for _, name := range condition.Spirits {
				spirit, err := ParseSpiritIndex(name)
				if err != nil {
					return nil, err
				}
				if slices.Contains(loadedCondition.Spirits, spirit) {
					return nil, fmt.Errorf("load patterns: %q has duplicate spirit %q", rule.Name, name)
				}
				loadedCondition.Spirits = append(loadedCondition.Spirits, spirit)
			}
			for _, name := range condition.Palaces {
				palace, err := ParsePalaceIndex(name)
				if err != nil {
					return nil, err
				}
				if palace == GongZhong {
					return nil, fmt.Errorf("load patterns: %q places a condition in the center palace", rule.Name)
				}
				if slices.Contains(loadedCondition.Palaces, palace) {
					return nil, fmt.Errorf("load patterns: %q has duplicate palace %q", rule.Name, name)
				}
				loadedCondition.Palaces = append(loadedCondition.Palaces, palace)
			}
			loadedRule.Conditions = append(loadedRule.Conditions, loadedCondition)
		}
		rulesOut = append(rulesOut, loadedRule)
	}
	return rulesOut, nil
}

func uniqueGanInteraction(name string) ([2]ganzhi.Gan, ganEntry, error) {
	var found [2]ganzhi.Gan
	var entry ganEntry
	count := 0
	for key, candidate := range ganInteractionTable {
		if candidate.PatternName != name {
			continue
		}
		found, entry, count = key, candidate, count+1
	}
	if count != 1 {
		return found, entry, fmt.Errorf("gan interaction %q has %d definitions, want 1", name, count)
	}
	return found, entry, nil
}

func loadYingQi() error {
	var table struct {
		Summary     string `json:"summary"`
		HorizonDays int    `json:"horizon_days"`
		DateMatches map[string][]struct {
			Match  string `json:"match"`
			Reason string `json:"reason"`
		} `json:"date_matches"`
		Rules []struct {
			Type        string `json:"type"`
			Mechanism   string `json:"mechanism"`
			Description string `json:"description"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(yingqiJSON, &table); err != nil {
		return err
	}
	if table.Summary == "" || len(table.Rules) != 4 ||
		table.HorizonDays <= 0 || table.HorizonDays > 60 || len(table.DateMatches) != 4 {
		return fmt.Errorf("load yingqi: invalid table dimensions")
	}
	yingqiSummary = table.Summary
	yingqiHorizonDays = table.HorizonDays
	yingqiRuleTable = make(map[string]yingqiRuleEntry, len(table.Rules))
	yingqiDateMatches = make(map[string][]yingqiDateMatch, len(table.DateMatches))
	for _, entry := range table.Rules {
		if _, exists := yingqiRuleTable[entry.Type]; exists {
			return fmt.Errorf("load yingqi: duplicate type %q", entry.Type)
		}
		yingqiRuleTable[entry.Type] = yingqiRuleEntry{
			Mechanism: entry.Mechanism, Description: entry.Description,
		}
		matches := make([]yingqiDateMatch, 0, len(table.DateMatches[entry.Type]))
		for _, match := range table.DateMatches[entry.Type] {
			if match.Match != "same" && match.Match != "opposite" || match.Reason == "" {
				return fmt.Errorf("load yingqi: invalid date match for %q", entry.Type)
			}
			matches = append(matches, yingqiDateMatch{Match: match.Match, Reason: match.Reason})
		}
		yingqiDateMatches[entry.Type] = matches
	}
	for _, ruleType := range []string{"ma_xing", "kong_wang", "duty_star", "duty_door"} {
		rule, ok := yingqiRuleTable[ruleType]
		if !ok || rule.Mechanism == "" || rule.Description == "" {
			return fmt.Errorf("load yingqi: missing rule %q", ruleType)
		}
	}
	return nil
}

func loadZhiRun() error {
	var table struct {
		Method             string             `json:"method"`
		LeaderIntervalDays int                `json:"leader_interval_days"`
		AdoptThresholdDays int                `json:"adopt_threshold_days"`
		MaxTermIntervals   int                `json:"max_term_intervals"`
		RepeatableTerms    []string           `json:"repeatable_terms"`
		States             []zhiRunStateEntry `json:"states"`
		Basis              struct {
			LeaderIntervalDays string `json:"leader_interval_days"`
			AdoptThresholdDays string `json:"adopt_threshold_days"`
			MaxTermIntervals   string `json:"max_term_intervals"`
			RepeatableTerms    string `json:"repeatable_terms"`
		} `json:"basis"`
	}
	if err := json.Unmarshal(zhirunJSON, &table); err != nil {
		return err
	}
	if table.Method != "zhirun" || dingjuMethodNames[DingjuZhiRun] == "" ||
		table.LeaderIntervalDays != 15 || table.AdoptThresholdDays < 0 ||
		table.AdoptThresholdDays >= table.LeaderIntervalDays ||
		table.MaxTermIntervals != 12 || len(table.RepeatableTerms) != 2 ||
		len(table.States) != 4 {
		return fmt.Errorf("load zhirun: invalid rule dimensions")
	}
	zhiRunLeaderDays = table.LeaderIntervalDays
	zhiRunThreshold = table.AdoptThresholdDays
	zhiRunMaxIntervals = table.MaxTermIntervals
	zhiRunRepeatTerms = make(map[string]bool, len(table.RepeatableTerms))
	for _, name := range table.RepeatableTerms {
		if _, ok := solarTermJuTable[name]; !ok {
			return fmt.Errorf("load zhirun: unknown repeatable term %q", name)
		}
		if zhiRunRepeatTerms[name] {
			return fmt.Errorf("load zhirun: duplicate repeatable term %q", name)
		}
		zhiRunRepeatTerms[name] = true
	}
	zhiRunStateNames = make(map[string]string, len(table.States))
	for _, state := range table.States {
		if state.State == "" || state.Name == "" || state.Condition == "" {
			return fmt.Errorf("load zhirun: incomplete state %+v", state)
		}
		if _, exists := zhiRunStateNames[state.State]; exists {
			return fmt.Errorf("load zhirun: duplicate state %q", state.State)
		}
		zhiRunStateNames[state.State] = state.Name
	}
	for _, state := range []string{"exact", "ahead", "behind", "repeat"} {
		if zhiRunStateNames[state] == "" {
			return fmt.Errorf("load zhirun: missing state %q", state)
		}
	}
	zhiRunBasis = map[string]string{
		"leader_interval_days": table.Basis.LeaderIntervalDays,
		"adopt_threshold_days": table.Basis.AdoptThresholdDays,
		"max_term_intervals":   table.Basis.MaxTermIntervals,
		"repeatable_terms":     table.Basis.RepeatableTerms,
	}
	for name, basis := range zhiRunBasis {
		if basis == "" {
			return fmt.Errorf("load zhirun: missing basis for %q", name)
		}
	}
	return nil
}

func loadMaoShan() error {
	var table struct {
		Method         string `json:"method"`
		ShichenPerYuan int    `json:"shichen_per_yuan"`
		MaxYuan        int    `json:"max_yuan"`
		Basis          string `json:"basis"`
	}
	if err := json.Unmarshal(maoshanJSON, &table); err != nil {
		return err
	}
	if table.Method != "maoshan" || dingjuMethodNames[DingjuMaoShan] == "" ||
		table.ShichenPerYuan != 60 || table.MaxYuan != 3 || table.Basis == "" {
		return fmt.Errorf("load maoshan: invalid rule")
	}
	maoShanShichenPerYuan, maoShanMaxYuan = table.ShichenPerYuan, table.MaxYuan
	return nil
}

func loadJinhan() error {
	var table struct {
		Method            string   `json:"method"`
		MethodSource      string   `json:"method_source"`
		DunSource         string   `json:"dun_source"`
		Basis             string   `json:"basis"`
		StarOrder         []string `json:"star_order"`
		StarAnchorPalaces struct {
			Yang []string `json:"yang"`
			Yin  []string `json:"yin"`
		} `json:"star_anchor_palaces"`
		StarFlightPalaces struct {
			Yang []string `json:"yang"`
			Yin  []string `json:"yin"`
		} `json:"star_flight_palaces"`
		DoorStartPalaces []string                     `json:"door_start_palaces"`
		DoorRingPalaces  []string                     `json:"door_ring_palaces"`
		DoorOrder        []string                     `json:"door_order"`
		DoorDays         int                          `json:"door_days_per_group"`
		DaySpirits       map[string]map[string]string `json:"day_spirits"`
	}
	if err := json.Unmarshal(jinhanJSON, &table); err != nil {
		return err
	}
	if table.Method != SchoolJinhanYuJing.String() || dingjuMethodNames[DingjuNone] == "" ||
		table.MethodSource == "" || table.DunSource == "" || table.Basis == "" || table.DoorDays != 3 {
		return fmt.Errorf("load jinhan: invalid rule")
	}
	if len(table.StarOrder) != 9 || hasDuplicate(table.StarOrder) {
		return fmt.Errorf("load jinhan: invalid star order")
	}
	yangStarts, err := parseJinhanPalaces(table.StarAnchorPalaces.Yang, 9)
	if err != nil {
		return fmt.Errorf("load jinhan: yang star anchors: %w", err)
	}
	yinStarts, err := parseJinhanPalaces(table.StarAnchorPalaces.Yin, 9)
	if err != nil {
		return fmt.Errorf("load jinhan: yin star anchors: %w", err)
	}
	yangFlight, err := parseJinhanPalaces(table.StarFlightPalaces.Yang, 9)
	if err != nil {
		return fmt.Errorf("load jinhan: yang star flight: %w", err)
	}
	yinFlight, err := parseJinhanPalaces(table.StarFlightPalaces.Yin, 9)
	if err != nil {
		return fmt.Errorf("load jinhan: yin star flight: %w", err)
	}
	doorStarts, err := parseJinhanPalaces(table.DoorStartPalaces, 8)
	if err != nil {
		return fmt.Errorf("load jinhan: door starts: %w", err)
	}
	doorRing, err := parseJinhanPalaces(table.DoorRingPalaces, 8)
	if err != nil {
		return fmt.Errorf("load jinhan: door ring: %w", err)
	}
	if len(table.DoorOrder) != 8 || hasDuplicate(table.DoorOrder) {
		return fmt.Errorf("load jinhan: invalid door order")
	}
	doors := make([]DoorIndex, 0, len(table.DoorOrder))
	for _, name := range table.DoorOrder {
		door, err := ParseDoorIndex(name)
		if err != nil || door == DoorZhong {
			return fmt.Errorf("load jinhan: invalid door %q", name)
		}
		doors = append(doors, door)
	}
	if len(table.DaySpirits) != 10 {
		return fmt.Errorf("load jinhan: day spirits must cover ten stems")
	}
	spiritsByGan := make(map[ganzhi.Gan][12]string, len(table.DaySpirits))
	for ganName, branches := range table.DaySpirits {
		gan, err := ganzhi.ParseGan(ganName)
		if err != nil {
			return fmt.Errorf("load jinhan: day spirit stem: %w", err)
		}
		if len(branches) != 12 {
			return fmt.Errorf("load jinhan: day spirits for %s must cover twelve branches", ganName)
		}
		var row [12]string
		for branchName, spirit := range branches {
			branch, err := ganzhi.ParseZhi(branchName)
			if err != nil || spirit == "" {
				return fmt.Errorf("load jinhan: invalid day spirit %s/%s", ganName, branchName)
			}
			row[int(branch)-1] = spirit
		}
		for _, spirit := range row {
			if spirit == "" {
				return fmt.Errorf("load jinhan: incomplete day spirits for %s", ganName)
			}
		}
		spiritsByGan[gan] = row
	}

	jinhanBasis, jinhanMethodSource, jinhanDunSource = table.Basis, table.MethodSource, table.DunSource
	jinhanStarOrder = append([]string(nil), table.StarOrder...)
	jinhanYangStarStarts, jinhanYinStarStarts = yangStarts, yinStarts
	jinhanYangStarFlight, jinhanYinStarFlight = yangFlight, yinFlight
	jinhanDoorStarts, jinhanDoorRing = doorStarts, doorRing
	jinhanDoorOrder, jinhanDoorDays = doors, table.DoorDays
	jinhanDaySpirits = spiritsByGan
	return nil
}

func parseJinhanPalaces(names []string, want int) ([]GongIndex, error) {
	if len(names) != want {
		return nil, fmt.Errorf("length = %d, want %d", len(names), want)
	}
	result := make([]GongIndex, 0, len(names))
	seen := make(map[GongIndex]bool, len(names))
	for _, name := range names {
		palace, err := ParsePalaceIndex(name)
		if err != nil {
			return nil, err
		}
		if seen[palace] || (want == 8 && palace == GongZhong) {
			return nil, fmt.Errorf("invalid palace %q", name)
		}
		seen[palace] = true
		result = append(result, palace)
	}
	return result, nil
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}
