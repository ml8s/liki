package liuyao

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/hexagrams.json
var hexagramsJSON []byte

var (
	guaTable   [64]guaMeta
	naGanTable [8]ganzhi.Gan
	naZhiTable [8][6]ganzhi.Zhi
)

func init() {
	if err := loadHexagrams(); err != nil {
		log.Fatalf("liuyao: load hexagrams: %v", err)
	}
}

func loadHexagrams() error {
	var data struct {
		Palaces    []string `json:"palaces"`
		Hexagrams  []struct {
			Name    string `json:"name"`
			Palace  string `json:"palace"`
			ShiPos  int    `json:"shi_pos"`
		} `json:"hexagrams"`
		NaGan map[string]string     `json:"na_gan"`
		NaZhi map[string][]string   `json:"na_zhi"`
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

	for palaceName, stemName := range data.NaGan {
		pi, ok := palaceIdx[palaceName]
		if !ok {
			log.Fatalf("liuyao: unknown palace %q in na_gan", palaceName)
		}
		stem, err := ganzhi.ParseGan(stemName)
		if err != nil {
			return err
		}
		naGanTable[pi] = stem
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
