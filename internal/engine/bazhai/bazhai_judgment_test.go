package bazhai

import "testing"

func TestBazhaiJudgment_EastWestGroups(t *testing.T) {
	tests := []struct {
		name      string
		mingGua   int // 命卦
		doorGua   int
		masterGua int
		stoveGua  int
		wantGroup string // 东四宅/西四宅
		wantDong  string // door match
		wantZhu   string // master match
		wantZao   string // stove match
	}{
		{
			name: "东四命+东四门主灶→吉",
			// 巽命(4, 东四), 坎门(1, 东四), 震主(3, 东四), 离灶(9, 东四)
			mingGua: 4, doorGua: 1, masterGua: 3, stoveGua: 9,
			wantGroup: "东四宅",
			wantDong: "吉", wantZhu: "吉", wantZao: "吉",
		},
		{
			name: "西四命+西四门主灶→吉",
			// 乾命(6, 西四), 坤门(2, 西四), 兑主(7, 西四), 艮灶(8, 西四)
			mingGua: 6, doorGua: 2, masterGua: 7, stoveGua: 8,
			wantGroup: "西四宅",
			wantDong: "吉", wantZhu: "吉", wantZao: "吉",
		},
		{
			name: "西四命+东四门→凶",
			// 乾命(6, 西四), 离门(9, 东四) → 不匹配
			mingGua: 6, doorGua: 9, masterGua: 6, stoveGua: 6,
			wantGroup: "西四宅",
			wantDong: "凶", wantZhu: "吉", wantZao: "吉",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := Chart{
				MingGua: MingGua{GuaNumber: tt.mingGua},
			}
			result := ComputeJudgment(chart, tt.doorGua, tt.masterGua, tt.stoveGua)
			if result.Group != tt.wantGroup {
				t.Errorf("group=%q, want %q", result.Group, tt.wantGroup)
			}
			if result.Door.Match != tt.wantDong {
				t.Errorf("door.match=%q, want %q", result.Door.Match, tt.wantDong)
			}
			if result.Master.Match != tt.wantZhu {
				t.Errorf("master.match=%q, want %q", result.Master.Match, tt.wantZhu)
			}
			if result.Stove.Match != tt.wantZao {
				t.Errorf("stove.match=%q, want %q", result.Stove.Match, tt.wantZao)
			}
			if tt.wantDong == "吉" && tt.wantZhu == "吉" && tt.wantZao == "吉" && result.Rating != "吉" {
				t.Errorf("rating=%q, want 吉(全吉)", result.Rating)
			}
		})
	}
}


func TestBazhaiJudgment_WuXingShengKe(t *testing.T) {
	// 生克验证: 门生主→加分, 主克灶→减分
	// 木(震3,巽4)→生→火(离9), 木(震3)→克→土(坤2)
	tests := []struct {
		name       string
		doorGua    int
		masterGua  int
		stoveGua   int
		wantRating string // 吉=门生主匹配, 凶=主克灶
	}{
		{
			name: "门生主+同组→吉",
			// 坎命(1, 东四), 震门(3, 木), 离灶(9, 火)
			// 门生主: 震木生离火→吉. 主克灶: 离火克?离=火,灶=?
			// Let's test: 门=坎(水), 主=震(木), 水→木: 门生主✅
			// 坎命(1, 东四), 坎门(1, 水), 震主(3, 木), 离灶(9, 火)
			// 门生主: 水生木✅, 主与灶: 木生火(主生灶也吉)
			doorGua: 1, masterGua: 3, stoveGua: 9,
			wantRating: "吉",
		},
		{
			name: "主克灶→减分",
			// 坎命(1, 东四), 离门(9, 火), 震主(3, 木), 坤灶(2, 土)
			// 门是同组? 离=东四✅
			// 门生主: 火生木✅→加分
			// 主克灶: 木克土✅→减分
			// net: +1 match + 1 sheng - 1 ke = 1 → 平
			doorGua: 9, masterGua: 3, stoveGua: 2,
			wantRating: "平",
		},
		{
			name: "门克主+不同组→凶",
			// 乾命(6, 西四), 离门(9, 火=东四), 兑主(7, 金=西四), 艮灶(8, 土=西四)
			// 门不同组→凶. 门克主: 火克金→凶
			// door: 不同组→凶, master: 同组→吉, stove: 同组→吉
			// net: -1 + 1 + 1 = +1 → 平
			// 再加主克灶: 金克木?灶=艮土, 金克木(不对), 土生金→吉
			// 木克土?不对, 主=兑金, 灶=艮土, 土生金(主被灶生→吉)
			doorGua: 9, masterGua: 7, stoveGua: 8,
			wantRating: "平",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := Chart{
				MingGua: MingGua{GuaNumber: 1}, // 坎命东四宅
			}
			result := ComputeJudgment(chart, tt.doorGua, tt.masterGua, tt.stoveGua)
			if result.Rating != tt.wantRating {
				t.Errorf("rating=%q, want %q (door=%s/master=%s/stove=%s, score=%d)",
					result.Rating, tt.wantRating,
					result.Door.Match, result.Master.Match, result.Stove.Match,
					scoreFromRating(result.Rating))
			}
		})
	}
}

func scoreFromRating(r string) int {
	switch r {
	case "吉": return 3
	case "平": return 1
	default: return 0
	}
}
