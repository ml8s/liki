package liuyao

import (
	"strings"
	"testing"
)

func TestNewCoinsCasting_NormalizesAndAudits(t *testing.T) {
	rounds := [][]string{
		{"正", "反", "反"}, // 7 少阳
		{"正", "正", "反"}, // 8 少阴
		{"正", "正", "正"}, // 9 老阳
		{"反", "反", "反"}, // 6 老阴
		{"正", "正", "反"},
		{"正", "反", "反"},
	}
	got, err := NewCoinsCasting(rounds)
	if err != nil {
		t.Fatalf("NewCoinsCasting() error = %v", err)
	}
	want := [6]int{7, 8, 9, 6, 8, 7}
	if got.Yaos != want {
		t.Fatalf("Yaos = %v, want %v", got.Yaos, want)
	}
	if len(got.DongYao) != 2 || got.DongYao[0] != 3 || got.DongYao[1] != 4 {
		t.Fatalf("DongYao = %v, want [3 4]", got.DongYao)
	}
	if len(got.Rounds) != 6 || got.Rounds[2].Label != "老阳" || !got.Rounds[2].Changing {
		t.Fatalf("unexpected rounds: %+v", got.Rounds)
	}
	if got.CastingID == "" {
		t.Fatal("CastingID is empty")
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestNewCoinsCasting_RejectsBadInput(t *testing.T) {
	tests := []struct {
		name   string
		rounds [][]string
		want   string
	}{
		{"too few", make([][]string, 5), "exactly 6"},
		{
			"bad coin count",
			[][]string{{"正", "反"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}},
			"exactly 3",
		},
		{
			"bad coin",
			[][]string{{"正", "反", "x"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"}},
			"正 or 反",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCoinsCasting(tt.rounds)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want contains %q", err, tt.want)
			}
		})
	}
}

func TestNewValuesCasting_RejectsInvalidYao(t *testing.T) {
	_, err := NewValuesCasting([6]int{7, 7, 7, 7, 7, 5})
	if err == nil || !strings.Contains(err.Error(), "must be 6-9") {
		t.Fatalf("error = %v, want 6-9", err)
	}
}

func TestCastingValidate_RejectsMissingOrTamperedCastingID(t *testing.T) {
	got, err := NewValuesCasting([6]int{7, 7, 7, 7, 7, 7})
	if err != nil {
		t.Fatal(err)
	}
	castingID := got.CastingID
	got.CastingID = ""
	if err := got.Validate(); err == nil || !strings.Contains(err.Error(), "casting_id is required") {
		t.Fatalf("missing casting_id error = %v", err)
	}
	got.CastingID = castingID + "00"
	if err := got.Validate(); err == nil || !strings.Contains(err.Error(), "casting_id mismatch") {
		t.Fatalf("tampered casting_id error = %v", err)
	}
}

func TestCastingValidate_RejectsCrossModeAuditFacts(t *testing.T) {
	values, err := NewValuesCasting([6]int{7, 7, 7, 7, 7, 7})
	if err != nil {
		t.Fatal(err)
	}
	values.Rounds = []CoinRound{{
		Position: 1, Coins: []string{"正", "反", "反"},
		Value: 7, Label: "少阳",
	}}
	if err := values.Validate(); err == nil || !strings.Contains(err.Error(), "yaos casting") {
		t.Fatalf("yaos with rounds error = %v", err)
	}

	coins, err := NewCoinsCasting([][]string{
		{"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"},
		{"正", "反", "反"}, {"正", "反", "反"}, {"正", "反", "反"},
	})
	if err != nil {
		t.Fatal(err)
	}
	coins.Rounds = nil
	if err := coins.Validate(); err == nil || !strings.Contains(err.Error(), "coins casting requires") {
		t.Fatalf("coins without rounds error = %v", err)
	}
}

func TestCastingValidate_RejectsDongYaoConflict(t *testing.T) {
	got, err := NewValuesCasting([6]int{9, 7, 7, 7, 7, 7})
	if err != nil {
		t.Fatalf("NewValuesCasting() error = %v", err)
	}
	got.DongYao = []int{2}
	if err := got.Validate(); err == nil || !strings.Contains(err.Error(), "dong_yao mismatch") {
		t.Fatalf("Validate() error = %v, want dong_yao mismatch", err)
	}
}
