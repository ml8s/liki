package bazhai

import (
	"fmt"

	"liki-engine/internal/engine/luoshu"
)

// ── 门主灶判断 ──

type LayoutResult struct {
	Group   string        `json:"group"` // 东四命/西四命
	MingGua string        `json:"ming_gua"`
	Door    doorStoveInfo `json:"door"`
	Master  doorStoveInfo `json:"master"`
	Stove   doorStoveInfo `json:"stove"`
}

type doorStoveInfo struct {
	Direction string `json:"direction"`
	GuaName   string `json:"gua_name"`
	Wuxing    string `json:"wuxing"`
	YouXing   string `json:"youxing"`
	Rating    string `json:"rating"`
	Group     string `json:"group"` // 东四卦/西四卦
	Match     string `json:"match"` // 吉/凶(与命卦同组不同组)
}

var dongSiGua = map[int]bool{1: true, 3: true, 4: true, 9: true} // 坎震巽离
var xiSiGua = map[int]bool{2: true, 6: true, 7: true, 8: true}   // 坤乾兑艮

// ComputeLayout analyzes 门主灶 in八宅风水.
func ComputeLayout(mingGua, doorGua, masterGua, stoveGua string) (LayoutResult, error) {
	mg := luoshu.TrigramNumber(mingGua)
	if mg == 0 {
		return LayoutResult{}, fmt.Errorf("invalid ming_gua %q", mingGua)
	}
	for slot, gua := range map[string]string{
		"door_gua":   doorGua,
		"master_gua": masterGua,
		"stove_gua":  stoveGua,
	} {
		if luoshu.TrigramNumber(gua) == 0 {
			return LayoutResult{}, fmt.Errorf("invalid %s %q", slot, gua)
		}
	}
	mgGroup := "东四命"
	if xiSiGua[mg] {
		mgGroup = "西四命"
	}

	door := evalPosition(luoshu.TrigramNumber(doorGua), mg)
	master := evalPosition(luoshu.TrigramNumber(masterGua), mg)
	stove := evalPosition(luoshu.TrigramNumber(stoveGua), mg)

	return LayoutResult{
		Group:   mgGroup,
		MingGua: luoshu.PalaceTable[mg].Name,
		Door:    door,
		Master:  master,
		Stove:   stove,
	}, nil
}

func evalPosition(guaNum, mingGua int) doorStoveInfo {
	group := "东四卦"
	if xiSiGua[guaNum] {
		group = "西四卦"
	}
	isMatch := dongSiGua[guaNum] == dongSiGua[mingGua]
	match := "凶"
	if isMatch {
		match = "吉"
	}
	youxing, rating := youxingForGua(mingGua, guaNum)
	return doorStoveInfo{
		Direction: luoshu.PalaceTable[guaNum].Direction,
		GuaName:   luoshu.PalaceTable[guaNum].Name,
		Wuxing:    luoshu.PalaceTable[guaNum].Element.String(),
		YouXing:   youxing,
		Rating:    rating,
		Group:     group,
		Match:     match,
	}
}

func youxingForGua(mingGua, guaNum int) (string, string) {
	p, ok := eightMansionPatterns[mingGua]
	if !ok {
		return "", ""
	}
	switch guaNum {
	case p.shengQi:
		return "生气", "大吉"
	case p.tianYi:
		return "天医", "吉"
	case p.yanNian:
		return "延年", "吉"
	case p.fuWei:
		return "伏位", "平"
	case p.huoHai:
		return "祸害", "凶"
	case p.wuGui:
		return "五鬼", "凶"
	case p.liuSha:
		return "六煞", "凶"
	case p.jueMing:
		return "绝命", "大凶"
	default:
		return "", ""
	}
}
