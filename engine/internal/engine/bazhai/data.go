package bazhai

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"liki-engine/internal/engine/luoshu"
)

//go:embed data/bazhai.json
var bazhaiJSON []byte

var (
	eightMansionPatterns map[int]dirPattern
	westGroup            map[int]bool
)

var (
	loadOnce sync.Once
	loadErr  error
)

// Load parses the embedded bazhai table and builds the 命卦 table (which reads
// luoshu). Idempotent; call from main/TestMain.
func Load() error {
	loadOnce.Do(func() {
		if err := luoshu.Load(); err != nil {
			loadErr = err
			return
		}
		buildGuaTable()
		if err := loadBazhai(); err != nil {
			loadErr = fmt.Errorf("bazhai: load bazhai: %w", err)
			return
		}
	})
	return loadErr
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
		WestGroup []int `json:"west_group"`
	}
	if err := json.Unmarshal(bazhaiJSON, &data); err != nil {
		return err
	}

	eightMansionPatterns = make(map[int]dirPattern, 8)
	for name, p := range data.Patterns {
		num := luoshu.TrigramNumber(name)
		if num == 0 {
			return fmt.Errorf("bazhai: unknown gua name %q", name)
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

	return nil
}
