package bazhai

import "fmt"

// ── 门主灶判断 ──

type LayoutResult struct {
	Group   string        `json:"group"` // 东四宅/西四宅
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
	Group     string `json:"group"` // 东四/西四
	Match     string `json:"match"` // 吉/凶(与命卦同组不同组)
}

var guaNames = [10]string{"", "坎", "坤", "震", "巽", "中", "乾", "兑", "艮", "离"}

// guaNameToNum maps 卦名 to 洛书数 (1-9).
func guaNameToNum(name string) int {
	for i, n := range guaNames {
		if n == name {
			return i
		}
	}
	return 0
}

var guaWuxing = [10]string{"", "水", "土", "木", "木", "土", "金", "金", "土", "火"}
var dongSiGua = map[int]bool{1: true, 3: true, 4: true, 9: true} // 坎震巽离
var xiSiGua = map[int]bool{2: true, 6: true, 7: true, 8: true}   // 坤乾兑艮

// ComputeLayout analyzes 门主灶 in八宅风水.
func ComputeLayout(mingGua, doorGua, masterGua, stoveGua string) (LayoutResult, error) {
	mg := guaNameToNum(mingGua)
	if mg == 0 {
		return LayoutResult{}, fmt.Errorf("invalid ming_gua %q", mingGua)
	}
	for slot, gua := range map[string]string{
		"door_gua":   doorGua,
		"master_gua": masterGua,
		"stove_gua":  stoveGua,
	} {
		if guaNameToNum(gua) == 0 {
			return LayoutResult{}, fmt.Errorf("invalid %s %q", slot, gua)
		}
	}
	mgGroup := "东四宅"
	if xiSiGua[mg] {
		mgGroup = "西四宅"
	}

	door := evalPosition(guaNameToNum(doorGua), mg)
	master := evalPosition(guaNameToNum(masterGua), mg)
	stove := evalPosition(guaNameToNum(stoveGua), mg)

	return LayoutResult{
		Group:   mgGroup,
		MingGua: guaNames[mg],
		Door:    door,
		Master:  master,
		Stove:   stove,
	}, nil
}

func evalPosition(guaNum, mingGua int) doorStoveInfo {
	group := "东四宅"
	if xiSiGua[guaNum] {
		group = "西四宅"
	}
	isMatch := dongSiGua[guaNum] == dongSiGua[mingGua]
	match := "凶"
	if isMatch {
		match = "吉"
	}
	youxing, rating := youxingForGua(mingGua, guaNum)
	return doorStoveInfo{
		Direction: palaceDirs[guaNum],
		GuaName:   guaNames[guaNum],
		Wuxing:    guaWuxing[guaNum],
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
