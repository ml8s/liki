package tianwen

import "testing"

func civilDateFromLunarForTest(t *testing.T, lunar LunarDate) CivilDate {
	t.Helper()
	date, err := ShanghaiLunarReference().CivilDateFromLunar(lunar)
	if err != nil {
		t.Fatalf("CivilDateFromLunar(%+v): %v", lunar, err)
	}
	return date
}
