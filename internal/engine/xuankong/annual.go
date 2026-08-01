package xuankong

import (
	"encoding/json"
	"fmt"
)

// Star9Index identifies one of the 紫白九星 (1-9).
type Star9Index int

const (
	Star9TanLang Star9Index = 1 + iota // 贪狼
	Star9JuMen                          // 巨门
	Star9LuCun                          // 禄存
	Star9WenQu                          // 文曲
	Star9LianZhen                       // 廉贞
	Star9WuQu                           // 武曲
	Star9PoJun                          // 破军
	Star9ZuoFu                          // 左辅
	Star9YouBi                          // 右弼
)

var star9Names = [10]string{"", "贪狼", "巨门", "禄存", "文曲", "廉贞", "武曲", "破军", "左辅", "右弼"}

func (s Star9Index) MarshalJSON() ([]byte, error) {
	if s < 1 || int(s) >= len(star9Names) {
		return json.Marshal("")
	}
	return json.Marshal(star9Names[s])
}

func (s *Star9Index) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return fmt.Errorf("star9 must be a string, got %s", string(data))
	}
	if name == "" {
		*s = 0
		return nil
	}
	for i, n := range star9Names {
		if i > 0 && n == name {
			*s = Star9Index(i)
			return nil
		}
	}
	return fmt.Errorf("unknown star9: %q", name)
}

// 流年飞星入中计算.
// 年飞星入中公式(下元): (年尾两位数 + 年尾两位数/4) mod 9
// 余数0则9入中.
func computeNianRuZhong(year int) Star9Index {
	tail := year % 100
	n := (tail + tail/4) % 9
	if n == 0 {
		return Star9Index(9)
	}
	return Star9Index(n)
}

// 九星顺序: 入中 → 乾 → 兑 → 艮 → 离 → 坎 → 坤 → 震 → 巽

// 八宫(无中宫): 西北=乾=1, 西=兑=2, 东北=艮=3, 南=离=4, 北=坎=5, 西南=坤=6, 东=震=7, 东南=巽=8
var palaceIndexFromFeiXu = [9]int{99, 6, 7, 8, 9, 1, 2, 3, 4} // 中(0)→无, 乾(1)→6, 兑(2)→7, 艮(3)→8, 离(4)→9, 坎(5)→1, 坤(6)→2, 震(7)→3, 巽(8)→4

type AnnualStar struct {
	Palace     int    `json:"gong"`     // 洛书宫位 1-9
	Star       int    `json:"xing"`       // 飞星号 1-9
	StarName   string `json:"xing_name"`
	Rating     string `json:"rating"` // 吉/凶/平
	RiZhong    bool           `json:"ru_zhong"`   // 是否入中
}

type AnnualBoard struct {
	Year       int            `json:"year"`
	RuZhong    Star9Index     `json:"ru_zhong"`   // 入中星
	Stars      []AnnualStar   `json:"xing_yao"`
}

// 九星名称
var starRatings = [10]string{"", "吉", "凶", "凶", "平", "大凶", "吉", "凶", "大吉", "吉"}

// ComputeAnnual computes the annual飞星 board.
func ComputeAnnual(year int) AnnualBoard {
	rz := computeNianRuZhong(year)
	var stars []AnnualStar
	for i := 0; i < 9; i++ {
		starNum := (int(rz) - 1 + i) % 9 + 1
		pi := palaceIndexFromFeiXu[i]
		if i == 0 {
			pi = 5 // 中宫
		}
		stars = append(stars, AnnualStar{
			Palace:   pi,
			Star:     starNum,
			StarName: star9Names[starNum],
			Rating:   starRatings[starNum],
			RiZhong:  i == 0,
		})
	}
	return AnnualBoard{
		Year:    year,
		RuZhong: rz,
		Stars:   stars,
	}
}
