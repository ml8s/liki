// Package luoshu holds the shared 洛书九宫（后天八卦）知识表：
// 宫号(1-9) ↔ 卦名 ↔ 方位 ↔ 五行，及卦名→宫号。
//
// 数据文件 data/luoshu.json 是单一事实源（engine 知识表统一以 embedded data 维护）。
// 它是风水（八宅 bazhai、玄空 xuankong）与奇门（qimen）共用的术数基础，
// 不归任何单一术数域。飞星（紫白）与 24 山等风水专有表不在此包。
package luoshu

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/luoshu.json
var luoshuJSON []byte

// Palace is one of the nine Luoshu palaces. Field names/tags are the shared
// wire contract consumed by fengshui, bazhai, xuankong and qimen.
type Palace struct {
	Number    int           `json:"number"`
	Name      string        `json:"name"`
	Direction string        `json:"direction"`
	Element   ganzhi.Wuxing `json:"wuxing"`
	YinYang   string        `json:"yin_yang,omitempty"`
}

// PalaceTable holds all nine palaces indexed by palace number (1-9).
var PalaceTable [10]Palace

func init() {
	var rows []struct {
		Number    int    `json:"number"`
		Name      string `json:"name"`
		Direction string `json:"direction"`
		Wuxing    string `json:"wuxing"`
		YinYang   string `json:"yin_yang"`
	}
	if err := json.Unmarshal(luoshuJSON, &rows); err != nil {
		log.Fatalf("luoshu: parse data/luoshu.json: %v", err)
	}
	for _, row := range rows {
		if row.Number < 1 || row.Number > 9 {
			log.Fatalf("luoshu: palace number out of range: %d", row.Number)
		}
		if PalaceTable[row.Number].Number != 0 {
			log.Fatalf("luoshu: duplicate palace number: %d", row.Number)
		}
		element, err := ganzhi.ParseWuxing(row.Wuxing)
		if err != nil {
			log.Fatalf("luoshu: palace %d: %v", row.Number, err)
		}
		PalaceTable[row.Number] = Palace{
			Number:    row.Number,
			Name:      row.Name,
			Direction: row.Direction,
			Element:   element,
			YinYang:   row.YinYang,
		}
	}
	for n := 1; n <= 9; n++ {
		if PalaceTable[n].Name == "" {
			log.Fatalf("luoshu: missing palace %d", n)
		}
	}
}

// ByNumber returns the palace for a given number (1-9), or the zero value.
func ByNumber(n int) Palace {
	if n >= 1 && n <= 9 {
		return PalaceTable[n]
	}
	return Palace{}
}

// TrigramNumber returns the palace number for a trigram name, or 0 if unknown.
func TrigramNumber(name string) int {
	for _, palace := range PalaceTable {
		if palace.Name == name {
			return palace.Number
		}
	}
	return 0
}
