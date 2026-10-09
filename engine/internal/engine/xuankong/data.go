package xuankong

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"liki-engine/internal/engine/luoshu"
)

//go:embed data/xing_jiahui.json
var xingJiaHuiJSON []byte

var xingJiaHuiTable map[[2]int]xingJiaHui

var (
	loadOnce sync.Once
	loadErr  error
)

// Load parses the embedded xuankong tables. Idempotent; call from main/TestMain.
func Load() error {
	loadOnce.Do(func() {
		if err := luoshu.Load(); err != nil {
			loadErr = err
			return
		}
		if err := loadXingJiaHui(); err != nil {
			loadErr = fmt.Errorf("xuankong: load xing_jiahui: %w", err)
			return
		}
	})
	return loadErr
}

func loadXingJiaHui() error {
	var entries []struct {
		Shan    int    `json:"shan"`
		Xiang   int    `json:"xiang"`
		Name    string `json:"name"`
		Meaning string `json:"meaning"`
	}
	if err := json.Unmarshal(xingJiaHuiJSON, &entries); err != nil {
		return err
	}
	xingJiaHuiTable = make(map[[2]int]xingJiaHui, len(entries))
	for _, e := range entries {
		xingJiaHuiTable[[2]int{e.Shan, e.Xiang}] = xingJiaHui{
			ShanNum:  e.Shan,
			XiangNum: e.Xiang,
			Name:     e.Name,
			Meaning:  e.Meaning,
		}
	}
	return nil
}
