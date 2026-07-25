package xuankong

import (
	_ "embed"
	"encoding/json"
	"log"
)

//go:embed data/xing_jiahui.json
var xingJiaHuiJSON []byte

var xingJiaHuiTable map[[2]int]xingJiaHui

func init() {
	if err := loadXingJiaHui(); err != nil {
		log.Fatalf("xuankong: load xing_jiahui: %v", err)
	}
}

func loadXingJiaHui() error {
	var entries []struct {
		Shan       int    `json:"shan"`
		Xiang      int    `json:"xiang"`
		Name       string `json:"name"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	if err := json.Unmarshal(xingJiaHuiJSON, &entries); err != nil {
		return err
	}
	xingJiaHuiTable = make(map[[2]int]xingJiaHui, len(entries))
	for _, e := range entries {
		xingJiaHuiTable[[2]int{e.Shan, e.Xiang}] = xingJiaHui{
			ShanNum:    e.Shan,
			XiangNum:   e.Xiang,
			Name:       e.Name,
			Meaning:    e.Meaning,
			Auspicious: e.Auspicious,
		}
	}
	return nil
}
