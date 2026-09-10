package bazhai

import (
	"strings"
	"testing"
)

func TestBazhaiJudgment_EastWestGroups(t *testing.T) {
	tests := []struct {
		name      string
		mingGua   string // 命卦
		doorGua   string
		masterGua string
		stoveGua  string
		wantGroup string // 东四宅/西四宅
		wantDong  string // door match
		wantZhu   string // master match
		wantZao   string // stove match
	}{
		{
			name: "东四命+东四门主灶→吉",
			// 巽命(4, 东四), 坎门(1, 东四), 震主(3, 东四), 离灶(9, 东四)
			mingGua: "巽", doorGua: "坎", masterGua: "震", stoveGua: "离",
			wantGroup: "东四宅",
			wantDong:  "吉", wantZhu: "吉", wantZao: "吉",
		},
		{
			name: "西四命+西四门主灶→吉",
			// 乾命(6, 西四), 坤门(2, 西四), 兑主(7, 西四), 艮灶(8, 西四)
			mingGua: "乾", doorGua: "坤", masterGua: "兑", stoveGua: "艮",
			wantGroup: "西四宅",
			wantDong:  "吉", wantZhu: "吉", wantZao: "吉",
		},
		{
			name: "西四命+东四门→凶",
			// 乾命(6, 西四), 离门(9, 东四) → 不匹配
			mingGua: "乾", doorGua: "离", masterGua: "乾", stoveGua: "乾",
			wantGroup: "西四宅",
			wantDong:  "凶", wantZhu: "吉", wantZao: "吉",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeLayout(tt.mingGua, tt.doorGua, tt.masterGua, tt.stoveGua)
			if err != nil {
				t.Fatalf("ComputeLayout() error = %v", err)
			}
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
		})
	}
}

func TestBazhaiJudgment_YouXing(t *testing.T) {
	tests := []struct {
		mingGua string
		target  string
		youxing string
		rating  string
	}{
		{"兑", "乾", "生气", "大吉"},
		{"坎", "巽", "生气", "大吉"},
	}
	for _, tt := range tests {
		result, err := ComputeLayout(tt.mingGua, tt.target, tt.target, tt.target)
		if err != nil {
			t.Fatalf("ComputeLayout() error = %v", err)
		}
		for _, item := range []doorStoveInfo{result.Door, result.Master, result.Stove} {
			if item.YouXing != tt.youxing || item.Rating != tt.rating {
				t.Fatalf("%s命见%s: got %s(%s), want %s(%s)", tt.mingGua, tt.target, item.YouXing, item.Rating, tt.youxing, tt.rating)
			}
		}
	}
}

func TestBazhaiJudgment_RejectsInvalidGua(t *testing.T) {
	_, err := ComputeLayout("兑", "门", "乾", "艮")
	if err == nil || !strings.Contains(err.Error(), `invalid door_gua "门"`) {
		t.Fatalf("error = %v, want invalid door_gua", err)
	}
}
