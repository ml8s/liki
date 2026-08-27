package qimen

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// GongIndex is a 洛书九宫 index. 1-9 map to:
//
//	生-4 立-9 杜-2
//		伤-3 中-5 景-7
//			休-8 开-1 惊-6
//
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

// DoorIndex represents one of the 八门 (eight doors).
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
)

var doorNames = [9]string{"", "休", "生", "伤", "杜", "景", "死", "惊", "开"}

func (d DoorIndex) String() string {
	if d >= 1 && d <= 8 {
		return doorNames[d]
	}
	return "?"
}

func (d DoorIndex) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String() + "门")
}

func (d *DoorIndex) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("door must be a string (e.g. \"休门\"), got %s", string(data))
	}
	// strip trailing "门" (3 bytes) if present
	base := name
	if len(name) >= 3 && name[len(name)-3:] == "门" {
		base = name[:len(name)-3]
	}
	for i, n := range doorNames {
		if i > 0 && n == base {
			*d = DoorIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown door: %q", name)
}

// SpiritIndex represents one of the 八神 (eight spirits).
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
)

// YangSpiritNames returns the spirit name for 阳遁.
func (s SpiritIndex) YangName() string {
	names := [9]string{"", "值符", "螣蛇", "太阴", "六合", "勾陈", "朱雀", "九地", "九天"}
	if s >= 1 && s <= 8 {
		return names[s]
	}
	return "?"
}

// YinSpiritNames returns the spirit name for 阴遁.
func (s SpiritIndex) YinName() string {
	names := [9]string{"", "值符", "螣蛇", "太阴", "六合", "白虎", "玄武", "九地", "九天"}
	if s >= 1 && s <= 8 {
		return names[s]
	}
	return "?"
}

// UnmarshalJSON accepts either a spirit name (值符/螣蛇/…) or an integer.
func (s *SpiritIndex) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*s = SpiritIndex(n)
		return nil
	}
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("spirit must be a string or int, got %s", string(data))
	}
	for i := 1; i <= 8; i++ {
		if SpiritIndex(i).YangName() == name || SpiritIndex(i).YinName() == name {
			*s = SpiritIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown spirit: %q", name)
}

// Gong holds all layers of information for one 宫。
type Gong struct {
	DiPanGan   ganzhi.Gan  `json:"di_pan_gan"`
	TianPanGan ganzhi.Gan  `json:"tian_pan_gan,omitempty"`
	Star       StarIndex   `json:"xing,omitempty"`
	Door       DoorIndex   `json:"men,omitempty"`
	Spirit     SpiritIndex `json:"shen"`
	CangGan    ganzhi.Gan  `json:"an_gan,omitempty"`
}

// pan is the complete 奇门遁甲排盘。
type pan struct {
	Jushu     int          `json:"jushu"`
	YinDun    bool         `json:"yin_dun"`
	RiGan     ganzhi.Gan   `json:"ri_gan"`
	RiZhi     ganzhi.Zhi   `json:"ri_zhi"`
	DutyStar  StarIndex    `json:"zhi_fu_xing"`
	DutyDoor  DoorIndex    `json:"zhi_shi_men"`
	GongWei   [9]Gong      `json:"gong_wei"`
	MaXing    GongIndex    `json:"ma_xing"`
	DriveGan  ganzhi.Gan   `json:"shi_gan"`
	DriveZhi  ganzhi.Zhi   `json:"shi_zhi"`
	KongWang  [2]GongIndex `json:"kong_wang"`
	WuBuYuShi bool         `json:"wu_bu_yu_shi"`
}

// MarshalJSON outputs spirit names per 阴/阳遁, keeping the rest as-is.
func (p pan) MarshalJSON() ([]byte, error) {
	// 用 map 保留默认序列化，只覆盖 palaces 里的 spirit 为名称
	m := map[string]any{
		"jushu": p.Jushu, "yin_dun": p.YinDun,
		"ri_gan": p.RiGan, "ri_zhi": p.RiZhi,
		"zhi_fu_xing": p.DutyStar, "zhi_shi_men": p.DutyDoor,
		"ma_xing": p.MaXing, "shi_gan": p.DriveGan, "shi_zhi": p.DriveZhi,
		"kong_wang": p.KongWang, "wu_bu_yu_shi": p.WuBuYuShi,
	}
	palaces := make([]map[string]any, 9)
	for i, pl := range p.GongWei {
		pm := map[string]any{
			"di_pan_gan": pl.DiPanGan, "tian_pan_gan": pl.TianPanGan,
		}
		if pl.Star != 0 {
			pm["xing"] = pl.Star
		}
		if pl.Door != 0 {
			pm["men"] = pl.Door
		}
		if pl.Spirit != 0 {
			if p.YinDun {
				pm["shen"] = pl.Spirit.YinName()
			} else {
				pm["shen"] = pl.Spirit.YangName()
			}
		}
		if pl.CangGan != 0 {
			pm["an_gan"] = pl.CangGan
		}
		palaces[i] = pm
	}
	m["gong_wei"] = palaces
	return json.Marshal(m)
}

// duty holds the value符 star and value使 door.
type duty struct {
	Star StarIndex
	Door DoorIndex
}

// juShu holds the result of bureau determination.
type juShu struct {
	Number int
	YinDun bool
	Yuan   string // 上元/中元/下元
}

// GanInteraction represents a 十干克应 between earth and heaven gan.
type GanInteraction struct {
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
	Auspicious  bool        `json:"auspicious"`
	GongWei     []GongIndex `json:"gong_wei,omitempty"`
}
