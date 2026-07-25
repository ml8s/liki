package qimen

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/stem_interactions.json
var stemInteractionsJSON []byte

//go:embed data/jushu.json
var jushuJSON []byte

//go:embed data/door_interactions.json
var doorInteractionsJSON []byte

//go:embed data/star_interactions.json
var starInteractionsJSON []byte

var (
	stemInteractionTable map[[2]ganzhi.Gan]stemEntry
	solarTermBureau      [24][4]int
	doorPalaceTable      map[[2]int]doorEntry
	starPalaceTable      map[[2]int]StarInteraction
)

func init() {
	if err := loadStemInteractions(); err != nil {
		log.Fatalf("qimen: load stem_interactions: %v", err)
	}
	if err := loadJushu(); err != nil {
		log.Fatalf("qimen: load jushu: %v", err)
	}
	if err := loadDoorInteractions(); err != nil {
		log.Fatalf("qimen: load door_interactions: %v", err)
	}
	if err := loadStarInteractions(); err != nil {
		log.Fatalf("qimen: load star_interactions: %v", err)
	}
}

func loadStemInteractions() error {
	var entries []struct {
		Earth      string `json:"earth"`
		Heaven     string `json:"heaven"`
		Name       string `json:"name"`
		Pattern    string `json:"pattern"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	if err := json.Unmarshal(stemInteractionsJSON, &entries); err != nil {
		return err
	}
	stemInteractionTable = make(map[[2]ganzhi.Gan]stemEntry, len(entries))
	for _, e := range entries {
		earth, err := ganzhi.ParseGan(e.Earth)
		if err != nil {
			return err
		}
		heaven, err := ganzhi.ParseGan(e.Heaven)
		if err != nil {
			return err
		}
		stemInteractionTable[[2]ganzhi.Gan{earth, heaven}] = stemEntry{
			Name:        e.Name,
			PatternName: e.Pattern,
			Meaning:     e.Meaning,
			Auspicious:  e.Auspicious,
		}
	}
	return nil
}

func loadJushu() error {
	var entries []struct {
		ShangYuan int  `json:"shang_yuan"`
		ZhongYuan int  `json:"zhong_yuan"`
		XiaYuan   int  `json:"xia_yuan"`
		YangDun   bool `json:"yang_dun"`
	}
	if err := json.Unmarshal(jushuJSON, &entries); err != nil {
		return err
	}
	for i, e := range entries {
		yd := 0
		if e.YangDun {
			yd = 1
		}
		solarTermBureau[i] = [4]int{e.ShangYuan, e.ZhongYuan, e.XiaYuan, yd}
	}
	return nil
}

func loadDoorInteractions() error {
	var entries []struct {
		Door    string `json:"door"`
		Palace  string `json:"palace"`
		Name    string `json:"name"`
		Meaning string `json:"meaning"`
	}
	if err := json.Unmarshal(doorInteractionsJSON, &entries); err != nil {
		return err
	}
	doorPalaceTable = make(map[[2]int]doorEntry, len(entries))
	for _, e := range entries {
		d, err := ParseDoorIndex(e.Door)
		if err != nil {
			return err
		}
		p, err := ParsePalaceIndex(e.Palace)
		if err != nil {
			return err
		}
		doorPalaceTable[[2]int{int(d), int(p) - 1}] = doorEntry{
			DoorName:   e.Door,
			PalaceName: e.Palace,
			Name:       e.Name,
			Meaning:    e.Meaning,
		}
	}
	return nil
}

func loadStarInteractions() error {
	var entries []struct {
		Star       string `json:"star"`
		Palace     string `json:"palace"`
		Name       string `json:"name"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	if err := json.Unmarshal(starInteractionsJSON, &entries); err != nil {
		return err
	}
	starPalaceTable = make(map[[2]int]StarInteraction, len(entries))
	for _, e := range entries {
		s, err := ParseStarIndex(e.Star)
		if err != nil {
			return err
		}
		p, err := ParsePalaceIndex(e.Palace)
		if err != nil {
			return err
		}
		starPalaceTable[[2]int{int(s), int(p) - 1}] = StarInteraction{
			Star:       e.Star,
			Palace:     e.Palace,
			Name:       e.Name,
			Meaning:    e.Meaning,
			Auspicious: e.Auspicious,
		}
	}
	return nil
}
