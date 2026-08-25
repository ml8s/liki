package bazhai

// ── 门主灶判断 ──

type LayoutResult struct {
	Group      string `json:"group"`       // 东四宅/西四宅
	MingGuaStr string `json:"ming_gua_str"`
	Door       doorStoveInfo `json:"door"`
	Master     doorStoveInfo `json:"master"`
	Stove      doorStoveInfo `json:"stove"`
}

type doorStoveInfo struct {
	GuaName   string `json:"gua_name"`
	Wuxing    string `json:"wuxing"`
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
var dongSiGua = map[int]bool{1: true, 3: true, 4: true, 9: true}  // 坎震巽离
var xiSiGua  = map[int]bool{2: true, 6: true, 7: true, 8: true}  // 坤乾兑艮

// ComputeLayout analyzes 门主灶 in八宅风水.
func ComputeLayout(chart Chart, doorGua, masterGua, stoveGua string) LayoutResult {
	mg := guaNameToNum(chart.MingGua.Gua.Name)
	mgGroup := "东四宅"
	if xiSiGua[mg] { mgGroup = "西四宅" }

	door := evalPosition(guaNameToNum(doorGua), mg)
	master := evalPosition(guaNameToNum(masterGua), mg)
	stove := evalPosition(guaNameToNum(stoveGua), mg)

	return LayoutResult{
		Group:      mgGroup,
		MingGuaStr: guaNames[mg],
		Door:       door,
		Master:     master,
		Stove:      stove,
	}
}

func evalPosition(guaNum, mingGua int) doorStoveInfo {
	group := "东四宅"
	if xiSiGua[guaNum] { group = "西四宅" }
	isMatch := dongSiGua[guaNum] == dongSiGua[mingGua]
	match := "凶"
	if isMatch { match = "吉" }
	return doorStoveInfo{
		GuaName:   guaNames[guaNum],
		Wuxing:    guaWuxing[guaNum],
		Group:     group,
		Match:     match,
	}
}

