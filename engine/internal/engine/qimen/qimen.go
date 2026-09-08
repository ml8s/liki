package qimen

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// GongIndex is a 洛书九宫 index:
// 1=坎, 2=坤, 3=震, 4=巽, 5=中, 6=乾, 7=兑, 8=艮, 9=离.
type GongIndex int

const (
	GongKan GongIndex = 1 + iota
	GongKun
	GongZhen
	GongXun
	GongZhong
	GongQian
	GongDui
	GongGen
	GongLi
)

var gongNames = [10]string{"", "坎", "坤", "震", "巽", "中", "乾", "兑", "艮", "离"}

func (p GongIndex) String() string {
	if p >= 1 && p <= 9 {
		return gongNames[p]
	}
	return "?"
}

func (p GongIndex) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *GongIndex) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("gong must be a string, got %s", string(data))
	}
	for i, name := range gongNames {
		if i > 0 && name == s {
			*p = GongIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown gong: %q", s)
}

// StarIndex represents one of the 九星 (nine stars).
type StarIndex int

const (
	StarTianPeng StarIndex = 1 + iota
	StarTianRui
	StarTianChong
	StarTianFu
	StarTianQin
	StarTianXin
	StarTianZhu
	StarTianRen
	StarTianYing
)

var starNames = [10]string{"", "天蓬", "天芮", "天冲", "天辅", "天禽", "天心", "天柱", "天任", "天英"}

func (s StarIndex) String() string {
	if s >= 1 && s <= 9 {
		return starNames[s]
	}
	return "?"
}

func (s StarIndex) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *StarIndex) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("star must be a string (e.g. \"天蓬\"), got %s", string(data))
	}
	for i, n := range starNames {
		if i > 0 && n == name {
			*s = StarIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown star: %q", name)
}

// DoorIndex represents the eight doors and the Ming Fa center door.
type DoorIndex int

const (
	DoorXiu DoorIndex = 1 + iota
	DoorSheng
	DoorShang
	DoorDu
	DoorJing
	DoorSi
	DoorJingMen
	DoorKai
	DoorZhong
)

var doorNames = [10]string{"", "休", "生", "伤", "杜", "景", "死", "惊", "开", "中"}

func (d DoorIndex) String() string {
	if d >= 1 && d <= 9 {
		return doorNames[d] + "门"
	}
	return "?"
}

func (d DoorIndex) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *DoorIndex) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("door must be a string (e.g. \"休门\"), got %s", string(data))
	}
	for i, n := range doorNames {
		if i > 0 && n+"门" == name {
			*d = DoorIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown door: %q", name)
}

// SpiritIndex represents the rotate-plate 八神 and the fly-plate ninth spirit.
type SpiritIndex int

const (
	SpiritZhiFu SpiritIndex = 1 + iota
	SpiritTengShe
	SpiritTaiYin
	SpiritLiuHe
	SpiritGouChen // 阳遁=勾陈, 阴遁=白虎
	SpiritZhuQue  // 阳遁=朱雀, 阴遁=玄武
	SpiritJiuDi
	SpiritJiuTian
	SpiritTaiChang
)

// YangSpiritNames returns the spirit name for 阳遁.
func (s SpiritIndex) YangName() string {
	if s >= 1 && int(s) < len(spiritYangNames) {
		if spiritYangNames[s] != "" {
			return spiritYangNames[s]
		}
	}
	if s >= 1 && int(s) < len(flySpiritNames) && flySpiritNames[s] != "" {
		return flySpiritNames[s]
	}
	return "?"
}

// YinName returns the spirit name for 阴遁.
func (s SpiritIndex) YinName() string {
	if s >= 1 && int(s) < len(spiritYinNames) {
		if spiritYinNames[s] != "" {
			return spiritYinNames[s]
		}
	}
	if s >= 1 && int(s) < len(flySpiritNames) && flySpiritNames[s] != "" {
		return flySpiritNames[s]
	}
	return "?"
}

// UnmarshalJSON accepts a spirit name (值符/螣蛇/…).
func (s *SpiritIndex) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("spirit must be a string, got %s", string(data))
	}
	for i := 1; i < len(spiritYangNames); i++ {
		if SpiritIndex(i).YangName() == name || SpiritIndex(i).YinName() == name {
			*s = SpiritIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown spirit: %q", name)
}

// TianPanSymbol binds a heaven-plate star to the gan carried from its home palace.
type TianPanSymbol struct {
	Gan  ganzhi.Gan `json:"gan"`
	Star StarIndex  `json:"xing"`
}

type PalaceIdentity struct {
	Name      string `json:"name"`
	Luoshu    int    `json:"luoshu"`
	RingIndex *int   `json:"ring_index,omitempty"`
}

type BranchPalace struct {
	Branch ganzhi.Zhi `json:"branch"`
	Gong   GongIndex  `json:"gong"`
}

// Gong holds all layers of information for one 宫。
type Gong struct {
	Gong         PalaceIdentity
	DiPanGan     ganzhi.Gan
	AnGan        ganzhi.Gan
	HiddenPillar string `json:"an_gan_zhi,omitempty"`
	TianPan      []TianPanSymbol
	TianPanGan   *ganzhi.Gan
	Door         DoorIndex
	DoorSet      bool
	Spirit       SpiritIndex
	SpiritSet    bool
}

// pan is the complete 奇门遁甲排盘。
type pan struct {
	Jushu          int
	School         School
	YinDun         bool
	RiGan          ganzhi.Gan
	RiZhi          ganzhi.Zhi
	NianGan        ganzhi.Gan
	NianZhi        ganzhi.Zhi
	YueGan         ganzhi.Gan
	YueZhi         ganzhi.Zhi
	LeadGan        ganzhi.Gan
	LeadZhi        ganzhi.Zhi
	DutyStarPalace GongIndex
	DutyDoorPalace GongIndex
	DutyStar       StarIndex
	DutyDoor       DoorIndex
	GongWei        [9]Gong
	MaXing         BranchPalace
	HourGan        ganzhi.Gan
	HourZhi        ganzhi.Zhi
	KongWang       [2]BranchPalace
	WuBuYuShi      bool
}

// MarshalJSON projects the API fields and renders spirit names for the current
// 阴阳遁 and plate style.
func (p pan) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"jushu": p.Jushu, "yin_dun": p.YinDun,
		"ri_gan": p.RiGan, "ri_zhi": p.RiZhi,
		"nian_gan": p.NianGan, "nian_zhi": p.NianZhi,
		"yue_gan": p.YueGan, "yue_zhi": p.YueZhi,
		"zhi_fu_xing": p.DutyStar, "zhi_shi_men": p.DutyDoor,
		"ma_xing": p.MaXing, "shi_gan": p.HourGan, "shi_zhi": p.HourZhi,
		"kong_wang": p.KongWang, "wu_bu_yu_shi": p.WuBuYuShi,
	}
	palaces := make([]map[string]any, 9)
	for i, pl := range p.GongWei {
		pm := map[string]any{
			"gong": pl.Gong, "di_pan_gan": pl.DiPanGan, "an_gan": pl.AnGan,
			"men_present": pl.DoorSet, "shen_present": pl.SpiritSet,
		}
		if pl.HiddenPillar != "" {
			pm["an_gan_zhi"] = pl.HiddenPillar
		}
		if len(pl.TianPan) > 0 {
			pm["tian_pan"] = pl.TianPan
		}
		if pl.TianPanGan != nil {
			pm["tian_pan_gan"] = pl.TianPanGan
		}
		if pl.Door != 0 {
			pm["men"] = pl.Door
		}
		if pl.Spirit != 0 {
			pm["shen"] = spiritDisplayName(pl.Spirit, p.YinDun, p.School)
		}
		palaces[i] = pm
	}
	m["gong_wei"] = palaces
	return json.Marshal(m)
}

// duty holds the value符 star and value使 door.
type duty struct {
	Star   StarIndex
	Door   DoorIndex
	Palace GongIndex
}

// juShu holds the result of dingju determination.
type juShu struct {
	Method      Method
	Number      int
	YinDun      bool
	Yuan        string // 上元/中元/下元
	JieQi       string
	YongJuJieQi string
	ZhiRunState string
	Quarter     *quarterInfo
}

type TianQinRule struct {
	Star                  string `json:"star"`
	HomeGong              string `json:"home_gong"`
	LodgingGong           string `json:"lodging_gong"`
	Follows               string `json:"follows"`
	CarriesCenterEarthGan bool   `json:"carries_center_earth_gan"`
	LodgingRule           string `json:"lodging_rule"`
}

type ChartMethod struct {
	Scope                string        `json:"scope"`
	ScopeName            string        `json:"scope_name"`
	School               string        `json:"school"`
	SchoolName           string        `json:"school_name"`
	DingjuMethod         string        `json:"dingju_method,omitempty"`
	DingjuMethodName     string        `json:"dingju_method_name,omitempty"`
	QuarterRule          string        `json:"quarter_rule,omitempty"`
	QuarterRuleName      string        `json:"quarter_rule_name,omitempty"`
	BaseDingjuMethod     string        `json:"base_dingju_method,omitempty"`
	BaseDingjuMethodName string        `json:"base_dingju_method_name,omitempty"`
	MethodSource         string        `json:"method_source"`
	DayBoundary          string        `json:"day_boundary"`
	SpiritMode           string        `json:"spirit_mode"`
	Geometry             string        `json:"geometry"`
	StarMode             string        `json:"star_mode"`
	DoorMode             string        `json:"door_mode"`
	SpiritFlight         string        `json:"spirit_flight"`
	TianQin              *TianQinRule  `json:"tian_qin,omitempty"`
	JieQi                string        `json:"jie_qi"`
	YongJuJieQi          string        `json:"yong_ju_jie_qi"`
	Yuan                 string        `json:"yuan"`
	ZhiRunState          string        `json:"zhi_run_state,omitempty"`
	LeadPillar           PillarInfo    `json:"lead_pillar"`
	LeadXunShou          PillarInfo    `json:"lead_xun_shou"`
	LeadXunShouGong      GongIndex     `json:"lead_xun_shou_gong"`
	DayFuTou             PillarInfo    `json:"day_fu_tou"`
	Quarter              *ChartQuarter `json:"quarter,omitempty"`
}

type PillarInfo struct {
	Name  string `json:"name"`
	Gan   string `json:"gan"`
	Zhi   string `json:"zhi"`
	LiuYi string `json:"liu_yi,omitempty"`
}

// GanInteraction represents a 十干克应 between earth and heaven gan.
type GanInteraction struct {
	Gong       GongIndex  `json:"gong"`
	DiPanGan   ganzhi.Gan `json:"di_pan_gan"`
	TianPanGan ganzhi.Gan `json:"tian_pan_gan"`
	Name       string     `json:"name"`
	Meaning    string     `json:"meaning"`
	Auspicious bool       `json:"auspicious"`
}

// MenInteraction represents an 八门克应 for a door in a specific gong.
type MenInteraction struct {
	Door    DoorIndex `json:"door,omitempty"`
	Gong    GongIndex `json:"gong,omitempty"`
	Name    string    `json:"name"`
	Meaning string    `json:"meaning"`
}

// Pattern represents a detected 格局 in the pan.
type Pattern struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Basis       string      `json:"basis,omitempty"`
	Auspicious  bool        `json:"auspicious"`
	GongWei     []GongIndex `json:"gong_wei,omitempty"`
}

type patternRule struct {
	Name             string
	Description      string
	Basis            string
	Auspicious       bool
	RequiresDutyDoor bool
	DutyStarPosition string
	Conditions       []patternCondition
}

type patternCondition struct {
	HeavenGan []ganzhi.Gan
	EarthGan  []ganzhi.Gan
	Doors     []DoorIndex
	Spirits   []SpiritIndex
	Palaces   []GongIndex
}
