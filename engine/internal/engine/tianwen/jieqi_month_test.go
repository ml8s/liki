package tianwen

import "testing"

func TestYueZhuFromYearGanMatchesYueZhu(t *testing.T) {
	date := NewCivilDate(2026, 6, 15)
	window, err := JieqiMonthAt(date.Time(), cstLocation)
	if err != nil {
		t.Fatal(err)
	}
	nian := NianZhu(date.Time())
	want := YueZhu(date.Time())
	got := YueZhuFromYearGan(nian.Gan, window.Index)
	if got != want {
		t.Fatalf("YueZhuFromYearGan = %+v, want %+v", got, want)
	}
}
