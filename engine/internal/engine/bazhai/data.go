package bazhai

import (
	_ "embed"
	"encoding/json"
	"log"
	"strconv"
)

//go:embed data/bazhai.json
var bazhaiJSON []byte

var (
	eightMansionPatterns map[int]dirPattern
	westGroup            map[int]bool
	palaceDirs           [10]string
)

func init() {
	if err := loadBazhai(); err != nil {
		log.Fatalf("bazhai: load bazhai: %v", err)
	}
}

func loadBazhai() error {
	var data struct {
		Patterns map[string]struct {
			ShengQi int `json:"sheng_qi"`
			TianYi  int `json:"tian_yi"`
			YanNian int `json:"yan_nian"`
			FuWei   int `json:"fu_wei"`
			HuoHai  int `json:"huo_hai"`
			WuGui   int `json:"wu_gui"`
			LiuSha  int `json:"liu_sha"`
			JueMing int `json:"jue_ming"`
		} `json:"eight_mansion_patterns"`
		WestGroup  []int             `json:"west_group"`
		PalaceDirs map[string]string `json:"palace_dirs"`
	}
	if err := json.Unmarshal(bazhaiJSON, &data); err != nil {
		return err
	}

	guaNameToNum := map[string]int{
		"坎": 1, "坤": 2, "震": 3, "巽": 4, "乾": 6, "兑": 7, "艮": 8, "离": 9,
	}

	eightMansionPatterns = make(map[int]dirPattern, 8)
	for name, p := range data.Patterns {
		num, ok := guaNameToNum[name]
		if !ok {
			log.Fatalf("bazhai: unknown gua name %q", name)
		}
		eightMansionPatterns[num] = dirPattern{
			shengQi: p.ShengQi, tianYi: p.TianYi, yanNian: p.YanNian, fuWei: p.FuWei,
			huoHai: p.HuoHai, wuGui: p.WuGui, liuSha: p.LiuSha, jueMing: p.JueMing,
		}
	}

	westGroup = make(map[int]bool, len(data.WestGroup))
	for _, v := range data.WestGroup {
		westGroup[v] = true
	}

	for k, v := range data.PalaceDirs {
		if k == "" {
			continue
		}
		idx, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if idx >= 0 && idx < 10 {
			palaceDirs[idx] = v
		}
	}

	return nil
}
