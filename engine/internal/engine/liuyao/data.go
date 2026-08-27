package liuyao

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/hexagrams.json
var hexagramsJSON []byte

var (
	guaTable   [64]guaMeta
	naGanTable [8][2]ganzhi.Gan // 纳甲天干：[内卦干, 外卦干]（京房：乾内甲外壬、坤内乙外癸，余宫内外同）
	naZhiTable [8][6]ganzhi.Zhi
)

func init() {
	if err := loadHexagrams(); err != nil {
		log.Fatalf("liuyao: load hexagrams: %v", err)
	}
}

func loadHexagrams() error {
	var data struct {
		Palaces   []string `json:"palaces"`
		Hexagrams []struct {
			Name   string `json:"name"`
			Palace string `json:"palace"`
			ShiPos int    `json:"shi_pos"`
		} `json:"hexagrams"`
		NaGan map[string]interface{} `json:"na_gan"`
		NaZhi map[string][]string    `json:"na_zhi"`
	}
	if err := json.Unmarshal(hexagramsJSON, &data); err != nil {
		return err
	}

	palaceIdx := make(map[string]int, 8)
	for i, name := range data.Palaces {
		palaceIdx[name] = i
	}

	for i, h := range data.Hexagrams {
		pi, ok := palaceIdx[h.Palace]
		if !ok {
			log.Fatalf("liuyao: unknown palace %q in hexagram %q", h.Palace, h.Name)
		}
		guaTable[i] = guaMeta{Name: h.Name, PalaceIdx: pi, ShiPos: h.ShiPos}
	}

	for palaceName, ganVal := range data.NaGan {
		pi, ok := palaceIdx[palaceName]
		if !ok {
			log.Fatalf("liuyao: unknown palace %q in na_gan", palaceName)
		}
		ganPair, err := parseGanPair(ganVal)
		if err != nil {
			return err
		}
		naGanTable[pi] = ganPair
	}

	for palaceName, zhiNames := range data.NaZhi {
		pi, ok := palaceIdx[palaceName]
		if !ok {
			log.Fatalf("liuyao: unknown palace %q in na_zhi", palaceName)
		}
		for j, zn := range zhiNames {
			z, err := ganzhi.ParseZhi(zn)
			if err != nil {
				return err
			}
			naZhiTable[pi][j] = z
		}
	}

	return nil
}

// parseGanPair parses a palace's 纳甲天干 pair (内卦干, 外卦干).
// JSON 存单干或双干（乾=甲壬、坤=乙癸，余宫单干内外相同）。
func parseGanPair(v interface{}) ([2]ganzhi.Gan, error) {
	var out [2]ganzhi.Gan
	switch t := v.(type) {
	case string:
		g, err := ganzhi.ParseGan(t)
		if err != nil {
			return out, err
		}
		out = [2]ganzhi.Gan{g, g}
	case []interface{}:
		for i, item := range t {
			name, ok := item.(string)
			if !ok || i >= 2 {
				return out, fmt.Errorf("invalid na_gan entry")
			}
			g, err := ganzhi.ParseGan(name)
			if err != nil {
				return out, err
			}
			out[i] = g
		}
	default:
		return out, fmt.Errorf("invalid na_gan entry")
	}
	return out, nil
}
