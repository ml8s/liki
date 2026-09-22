package liuyao

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
)

// ── 周易卦爻辞 (卦名、卦辞、爻辞) ──

// GuaCi holds the 周易 text for one hexagram.
type GuaCi struct {
	Name  string    `json:"name"`
	GuaCi string    `json:"gua_ci"`
	YaoCi [6]string `json:"yao_ci"` // from 初(0) to 上(5)
}

//go:embed data/zhouyi.json
var zhouyiJSON []byte

var zhouyiTable []GuaCi

func init() {
	if err := json.Unmarshal(zhouyiJSON, &zhouyiTable); err != nil {
		log.Fatalf("liuyao: load zhouyi: %v", err)
	}
	if len(zhouyiTable) != 64 {
		log.Fatalf("liuyao: zhouyi table has %d entries, want 64", len(zhouyiTable))
	}
}

// GetGuaCi returns the hexagram text for a given gua index (0-63).
func GetGuaCi(idx int) (GuaCi, error) {
	if idx < 0 || idx >= 64 {
		return GuaCi{}, fmt.Errorf("gua index %d out of range [0,63]", idx)
	}
	return zhouyiTable[idx], nil
}
