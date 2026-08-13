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
	earthlyIdxTable = [12]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	nianStars = map[string][12]int{
		"华盖": {4,1,10,7,4,1,10,7,4,1,10,7},
		"咸池": {9,6,3,0,9,6,3,0,9,6,3,0},
		"孤辰": {2,2,5,5,5,8,8,8,11,11,11,2},
		"寡宿": {10,10,1,1,1,4,4,4,7,7,7,10},
		"破碎": {5,1,9,5,1,9,5,1,9,5,1,9},
		"蜚廉": {8,9,10,5,6,7,2,3,4,11,0,1},
		"龙池": {4,5,6,7,8,9,10,11,0,1,2,3},
		"凤阁": {10,9,8,7,6,5,4,3,2,1,0,11},
		"天哭": {6,5,4,3,2,1,0,11,10,9,8,7},
		"天虚": {6,7,8,9,10,11,0,1,2,3,4,5},
		"天空": {1,2,3,4,5,6,7,8,9,10,11,0},
		"天德": {9,10,11,0,1,2,3,4,5,6,7,8},
		"月德": {5,6,7,8,9,10,11,0,1,2,3,4},
	}
	yueStars = map[string][12]int{
		"天姚": {1,2,3,4,5,6,7,8,9,10,11,0},
		"天刑": {9,10,11,0,1,2,3,4,5,6,7,8},
		"阴煞": {2,0,10,8,6,4,2,0,10,8,6,4},
		"解神": {8,8,10,10,0,0,2,2,4,4,6,6},
		"天月": {10,5,4,2,7,3,11,7,2,6,10,2},
		"天巫": {5,8,2,11,5,8,2,11,5,8,2,11},
	}
	shiStars = map[string][12]int{
		"台辅": {6,7,8,9,10,11,0,1,2,3,4,5},
		"封诰": {2,3,4,5,6,7,8,9,10,11,0,1},
	}
	ganStars = map[string][10]int{
		"天厨": {5,6,0,5,6,8,2,6,9,11},
		"天官": {7,4,5,2,3,9,11,9,10,6},
		"天福": {9,8,0,11,3,2,6,5,6,5},
		"截路": {8,6,4,2,0,8,6,4,2,0},
		"空亡": {9,7,5,3,1,9,7,5,3,1},
	}
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

	// tianMaTable — iztro formula if JSON empty
	if len(data.TianMa) == 0 {
		// 寅午戌→申, 申子辰→寅, 巳酉丑→亥, 亥卯未→巳 (zhiIdx)
		tianMaTable = [12]int{2, 11, 8, 5, 2, 11, 8, 5, 2, 11, 8, 5}
	} else {
		for zhiName, pos := range data.TianMa {
			zhi, err := ganzhi.ParseZhi(zhiName)
			if err != nil {
				return err
			}
			tianMaTable[int(zhi)-1] = pos
		}
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
