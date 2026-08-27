package qimen

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/gan_interaction.json
var ganInteractionJSON []byte

//go:embed data/jushu.json
var jushuJSON []byte

//go:embed data/men_interaction.json
var menInteractionJSON []byte

//go:embed data/xing_interaction.json
var xingInteractionJSON []byte

var (
	ganInteractionTable map[[2]ganzhi.Gan]ganEntry
	solarTermBureau     [24][4]int
	menGongTable        map[[2]int]doorEntry
	xingGongTable       map[[2]int]XingInteraction
)

func init() {
	if err := loadGanInteractions(); err != nil {
		log.Fatalf("qimen: load gan_interaction: %v", err)
	}
	if err := loadJushu(); err != nil {
		log.Fatalf("qimen: load jushu: %v", err)
	}
	if err := loadMenInteractions(); err != nil {
		log.Fatalf("qimen: load men_interaction: %v", err)
	}
	if err := loadXingInteractions(); err != nil {
		log.Fatalf("qimen: load xing_interaction: %v", err)
	}
}

func loadGanInteractions() error {
	var entries []struct {
		Earth      string `json:"di_pan_gan"`
		Heaven     string `json:"tian_pan_gan"`
		Name       string `json:"name"`
		Pattern    string `json:"pattern"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	if err := json.Unmarshal(ganInteractionJSON, &entries); err != nil {
		return err
	}
	ganInteractionTable = make(map[[2]ganzhi.Gan]ganEntry, len(entries))
	for _, e := range entries {
		earth, err := ganzhi.ParseGan(e.Earth)
		if err != nil {
			return err
		}
		heaven, err := ganzhi.ParseGan(e.Heaven)
		if err != nil {
			return err
		}
		ganInteractionTable[[2]ganzhi.Gan{earth, heaven}] = ganEntry{
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

func loadMenInteractions() error {
	var entries []struct {
		Door    string `json:"door"`
		Gong    string `json:"gong"`
		Name    string `json:"name"`
		Meaning string `json:"meaning"`
	}
	if err := json.Unmarshal(menInteractionJSON, &entries); err != nil {
		return err
	}
	menGongTable = make(map[[2]int]doorEntry, len(entries))
	for _, e := range entries {
		d, err := ParseDoorIndex(e.Door)
		if err != nil {
			return err
		}
		p, err := ParsePalaceIndex(e.Gong)
		if err != nil {
			return err
		}
		menGongTable[[2]int{int(d), int(p) - 1}] = doorEntry{
			DoorName: e.Door,
			GongName: e.Gong,
			Name:     e.Name,
			Meaning:  e.Meaning,
		}
	}
	return nil
}

func loadXingInteractions() error {
	var entries []struct {
		Star       string `json:"xing"`
		Gong       string `json:"gong"`
		Name       string `json:"name"`
		Meaning    string `json:"meaning"`
		Auspicious bool   `json:"auspicious"`
	}
	if err := json.Unmarshal(xingInteractionJSON, &entries); err != nil {
		return err
	}
	xingGongTable = make(map[[2]int]XingInteraction, len(entries))
	for _, e := range entries {
		s, err := ParseStarIndex(e.Star)
		if err != nil {
			return err
		}
		p, err := ParsePalaceIndex(e.Gong)
		if err != nil {
			return err
		}
		xingGongTable[[2]int{int(s), int(p) - 1}] = XingInteraction{
			Star:       e.Star,
			Gong:       e.Gong,
			Name:       e.Name,
			Meaning:    e.Meaning,
			Auspicious: e.Auspicious,
		}
	}
	return nil
}
