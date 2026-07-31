package bazhai

// ── 门主灶判断 ──

type JudgmentResult struct {
	Group      string `json:"group"`       // 东四宅/西四宅
	MingGuaStr string `json:"ming_gua_str"`
	Door       doorStoveInfo `json:"door"`
	Master     doorStoveInfo `json:"master"`
	Stove      doorStoveInfo `json:"stove"`
	Rating     string `json:"rating"`     // 吉/平/凶
	Summary    string `json:"summary"`
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

// ComputeJudgment analyzes 门主灶 in八宅风水.
func ComputeJudgment(chart Chart, doorGua, masterGua, stoveGua string) JudgmentResult {
	mg := guaNameToNum(chart.MingGua.Gua.Name)
	mgGroup := "东四宅"
	if xiSiGua[mg] { mgGroup = "西四宅" }

	door := evalPosition(guaNameToNum(doorGua), mg)
	master := evalPosition(guaNameToNum(masterGua), mg)
	stove := evalPosition(guaNameToNum(stoveGua), mg)

	// 综合评定
	score := 0
	if door.Match == "吉" { score++ }
	if master.Match == "吉" { score++ }
	if stove.Match == "吉" { score++ }

	// 门生主
	if guaWuxing[guaNameToNum(doorGua)] == guaWuxing[guaNameToNum(masterGua)] || wuxingSheng(guaWuxing[guaNameToNum(doorGua)], guaWuxing[guaNameToNum(masterGua)]) {
		score++
	}
	// 主克灶为凶
	if wuxingKe(guaWuxing[guaNameToNum(masterGua)], guaWuxing[guaNameToNum(stoveGua)]) {
		score--
	}

	rating := "凶"
	summary := "门主灶配合不佳，宜择吉调整"
	if score >= 3 {
		rating = "吉"
		summary = "门主灶配合吉利，家宅兴旺"
	} else if score >= 1 {
		rating = "平"
		summary = "门主灶配合平平，可进一步优化"
	}

	return JudgmentResult{
		Group:      mgGroup,
		MingGuaStr: guaNames[mg],
		Door:       door,
		Master:     master,
		Stove:      stove,
		Rating:     rating,
		Summary:    summary,
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

func wuxingSheng(a, b string) bool {
	// 生: 木→火→土→金→水→木
	switch a {
	case "木": return b == "火"
	case "火": return b == "土"
	case "土": return b == "金"
	case "金": return b == "水"
	case "水": return b == "木"
	}
	return false
}

func wuxingKe(a, b string) bool {
	// 克: 木→土→水→火→金→木
	switch a {
	case "木": return b == "土"
	case "土": return b == "水"
	case "水": return b == "火"
	case "火": return b == "金"
	case "金": return b == "木"
	}
	return false
}
