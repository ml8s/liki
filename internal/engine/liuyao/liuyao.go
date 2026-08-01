package liuyao

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// YaoType classifies a single coin-toss result.
type YaoType int

const (
	LaoYin   YaoType = 6 // 老阴 ⚋→⚊ (6, 三个2 = 三背)
	ShaoYang YaoType = 7 // 少阳 ⚊ (7, 两背一面)
	ShaoYin  YaoType = 8 // 少阴 ⚋ (8, 两面一背)
	LaoYang  YaoType = 9 // 老阳 ⚊→⚋ (9, 三个3 = 三面)
)

func (y YaoType) IsYang() bool     { return y == ShaoYang || y == LaoYang }
func (y YaoType) IsChanging() bool { return y == LaoYang || y == LaoYin }

// LiuQin is the 六亲 classification.
type LiuQin int

const (
	QinFumu    LiuQin = iota // 父母
	QinXiongDi               // 兄弟
	QinGuanGui               // 官鬼
	QinQiCai                 // 妻财
	QinZiSun                 // 子孙
)

var liuQinNames = [5]string{"父母", "兄弟", "官鬼", "妻财", "子孙"}

func (q LiuQin) MarshalJSON() ([]byte, error) { return json.Marshal(q.String()) }

func (q *LiuQin) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("liu_qin: expected string, got %s", string(data))
	}
	for i, n := range liuQinNames {
		if n == s {
			*q = LiuQin(i)
			return nil
		}
	}
	return fmt.Errorf("liu_qin: unknown %q", s)
}

func (q LiuQin) String() string {
	if q >= 0 && q <= 4 {
		return liuQinNames[q]
	}
	return "?"
}

// LiuShou is the 六兽.
type LiuShou int

const (
	ShouQingLong LiuShou = iota
	ShouZhuQue
	ShouGouChen
	ShouTengShe
	ShouBaiHu
	ShouXuanWu
)

var liuShouNames = [6]string{"青龙", "朱雀", "勾陈", "螣蛇", "白虎", "玄武"}

func (l LiuShou) MarshalJSON() ([]byte, error) { return json.Marshal(l.String()) }

func (l *LiuShou) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("LiuShou must be a string, got %s", string(data))
	}
	for i, n := range liuShouNames {
		if n == name {
			*l = LiuShou(i)
			return nil
		}
	}
	return fmt.Errorf("unknown liushou: %q", name)
}

func (l LiuShou) String() string {
	if l >= 0 && l <= 5 {
		return liuShouNames[l]
	}
	return "?"
}

// Line is a single爻 in a hexagram.
type Line struct {
	Position int           `json:"position"` // 1-6, bottom to top
	Type     YaoType       `json:"type"`
	Gan      ganzhi.Gan    `json:"gan"`
	Zhi      ganzhi.Zhi    `json:"zhi"`
	Wuxing   ganzhi.Wuxing `json:"wuxing"`
	LiuQin  LiuQin `json:"liu_qin"`
	ShiYing  string `json:"shi_ying"` // "世"/"应"/"""
	LiuShou LiuShou `json:"liu_shou"`
}

// guaIndex identifies one of the 64 hexagrams.
type guaIndex int // 0-63, upper trigram 0-7, lower trigram 0-7

func (g guaIndex) MarshalJSON() ([]byte, error) {
	if int(g) < 0 || int(g) >= 64 {
		return json.Marshal("")
	}
	return json.Marshal(zhouyiTable[g].Name)
}

func (g *guaIndex) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("guaIndex must be a string, got %s", string(data))
	}
	for i, gc := range zhouyiTable {
		if gc.Name == s {
			*g = guaIndex(i)
			return nil
		}
	}
	return fmt.Errorf("unknown gua: %q", s)
}

// guaMeta holds static data for a hexagram.
type guaMeta struct {
	Name      string `json:"name"`       // 卦名
	PalaceIdx int    `json:"palace_idx"` // 0-7, which palace
	ShiPos    int    `json:"shi_pos"`    // 1-6,世爻 position
}

// Chart is the complete 六爻排盘 with all analysis layers.
type Chart struct {
	Name          string         `json:"name"`
	BenGua        guaIndex       `json:"ben_gua"`
	BianGua       guaIndex       `json:"bian_gua,omitempty"` // 0 if no change
	Palace        string         `json:"gong"`
	PalaceWuxing  ganzhi.Wuxing  `json:"palace_wuxing"`
	Lines         [6]Line        `json:"lines"`
	BianLines     [6]Line        `json:"bian_lines,omitempty"`
	DayGan        ganzhi.Gan     `json:"ri_chen_gan"`
	DayZhi        ganzhi.Zhi     `json:"ri_chen_zhi"`
	MonthZhi      ganzhi.Zhi     `json:"yue_jian_zhi"`
	MonthGan      ganzhi.Gan     `json:"yue_jian_gan"`
	DongYao   []int          `json:"dong_yao"` // 动爻位置 1-6
	// Analysis layers set by ComputeChart.
	YongShen     YongShenResult `json:"yong_shen"`
	GuaCi	GuaCi	`json:"gua_ci,omitempty"`
	WangShuai    [6]ganzhi.WangShuai   `json:"wang_shuai"`
	DayRelations [6]DayRelation `json:"ri_chen_relations"`
	YingQi    YingQi         `json:"ying_qi"`
}

// palaceNames.
var palaceNames = [8]string{"乾", "兑", "离", "震", "巽", "坎", "艮", "坤"}

// palaceWuxing maps palace index → five element (for六亲).
var palaceWuxing = [8]ganzhi.Wuxing{
	ganzhi.WxJin,  // 乾=金
	ganzhi.WxJin,  // 兑=金
	ganzhi.WxHuo,  // 离=火
	ganzhi.WxMu,   // 震=木
	ganzhi.WxMu,   // 巽=木
	ganzhi.WxShui, // 坎=水
	ganzhi.WxTu,   // 艮=土
	ganzhi.WxTu,   // 坤=土
}
