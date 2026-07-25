package ziwei

import (
	_ "embed"
	"encoding/json"
	"log"

	"liki-engine/internal/engine/ganzhi"
)

//go:embed data/tables.json
var tablesJSON []byte

//go:embed data/miao_wang.json
var miaoWangJSON []byte

var (
	siHuaTable    map[Gan][4]starIndex
	ziweiStartPos map[juShu]int
	luCunTable    [10]int
	tianKuiTable  [10]int
	tianMaTable   [12]int
)

func init() {
	if err := loadTables(); err != nil {
		log.Fatalf("ziwei: load tables: %v", err)
	}
}

func loadTables() error {
	var data struct {
		SiHua       map[string]map[string]string `json:"si_hua"`
		ZiweiStart  map[string]int               `json:"ziwei_start"`
		LuCun       map[string]int               `json:"lu_cun"`
		TianKui     map[string]int               `json:"tian_kui"`
		TianMa      map[string]int               `json:"tian_ma"`
	}
	if err := json.Unmarshal(tablesJSON, &data); err != nil {
		return err
	}

	// siHuaTable
	siHuaTable = make(map[Gan][4]starIndex, 10)
	for stemName, h := range data.SiHua {
		stem, err := ganzhi.ParseGan(stemName)
		if err != nil {
			return err
		}
		stars := [4]starIndex{
			nameToStar(h["hua_lu"]), nameToStar(h["hua_quan"]),
			nameToStar(h["hua_ke"]), nameToStar(h["hua_ji"]),
		}
		siHuaTable[stem] = stars
	}

	// ziweiStartPos
	juNameToJuShu := map[string]juShu{
		"水二局": JuWater, "木三局": JuWood, "金四局": JuMetal,
		"土五局": JuEarth, "火六局": JuFire,
	}
	ziweiStartPos = make(map[juShu]int, 5)
	for name, pos := range data.ZiweiStart {
		js, ok := juNameToJuShu[name]
		if !ok {
			log.Fatalf("ziwei: unknown juShu name %q", name)
		}
		ziweiStartPos[js] = pos
	}

	// luCunTable
	for stemName, pos := range data.LuCun {
		stem, err := ganzhi.ParseGan(stemName)
		if err != nil {
			return err
		}
		luCunTable[int(stem)-1] = pos
	}

	// tianKuiTable
	for stemName, pos := range data.TianKui {
		stem, err := ganzhi.ParseGan(stemName)
		if err != nil {
			return err
		}
		tianKuiTable[int(stem)-1] = pos
	}

	// miaoWangTable
	var mwData struct {
		MiaoWang [][]string `json:"miao_wang"`
	}
	if err := json.Unmarshal(miaoWangJSON, &mwData); err != nil {
		return err
	}
	for i, row := range mwData.MiaoWang {
		for j, s := range row {
			miaoWangTable[i][j] = brightnessFrom(s)
		}
	}

	// tianMaTable
	for zhiName, pos := range data.TianMa {
		zhi, err := ganzhi.ParseZhi(zhiName)
		if err != nil {
			return err
		}
		tianMaTable[int(zhi)-1] = pos
	}

	return nil
}

func nameToStar(s string) starIndex {
	for si, name := range starNames {
		if name == s {
			return si
		}
	}
	return 0
}
