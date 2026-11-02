package tianwen

import (
	"testing"
	"time"
)

func civilDateFromLunarForTest(t *testing.T, lunar LunarDate) CivilDate {
	t.Helper()
	date, err := ShanghaiLunarReference().CivilDateFromLunar(lunar)
	if err != nil {
		t.Fatalf("CivilDateFromLunar(%+v): %v", lunar, err)
	}
	return date
}

func lunarDateAtForTest(ref *LunarReference, at time.Time) LunarDate {
	return ref.SolarDateToLunar(ref.dateOf(at))
}
