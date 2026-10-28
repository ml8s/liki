package tianwen

import (
	"strings"
	"testing"
	"time"
)

// DST fold：同一壁钟时刻存在两个绝对时刻 → 恰好两个候选，且按时间排序。
func TestResolveCivilCandidatesDSTFold(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Resolve(ResolveInput{
		Calendar:  "solar",
		Date:      "2026-11-01",
		Time:      "01:30",
		Timezone:  loc.String(),
		Longitude: -74.0064,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates = %v, want DST fold pair", result.Candidates)
	}
	// 新 API 候选按角色排序（input 在先）；fold 语义 = 两个不同绝对时刻（EDT/EST）。
	a, b := result.Candidates[0].StandardTime, result.Candidates[1].StandardTime
	if a.Equal(b) {
		t.Fatal("fold candidates must be distinct absolute times")
	}
	if a.UTC().Sub(b.UTC()) == 0 {
		t.Fatal("fold candidates must differ in UTC offset")
	}
}

// DST gap：不存在的壁钟时刻 → 显式报错（fail closed）。
func TestResolveCivilCandidatesDSTGap(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Resolve(ResolveInput{
		Calendar:  "solar",
		Date:      "2026-03-08",
		Time:      "02:30",
		Timezone:  loc.String(),
		Longitude: -74.0064,
	})
	if err == nil {
		t.Fatal("DST gap must fail closed")
	}
	if !strings.Contains(err.Error(), "DST gap") {
		t.Fatalf("error %q should mention DST gap", err)
	}
}
