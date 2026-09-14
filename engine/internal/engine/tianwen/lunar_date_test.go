package tianwen

import "testing"

func TestLunarMonthMeta_LeapMonth(t *testing.T) {
	// 2025 has a leap sixth month. This is an external calendar fact used as
	// an oracle for boundary handling.
	meta, err := LunarMonthMeta(2025, 6, true)
	if err != nil {
		t.Fatalf("LunarMonthMeta(2025 leap 6): %v", err)
	}
	if meta.Year != 2025 || meta.Month != 6 || !meta.Leap {
		t.Fatalf("meta = %+v, want 2025 leap 6", meta)
	}
	if meta.DayCount != 29 && meta.DayCount != 30 {
		t.Fatalf("day count = %d, want 29 or 30", meta.DayCount)
	}

	for _, day := range []int{1, 15, 16, meta.DayCount} {
		if _, err := ValidateLunarDate(LunarDate{Year: 2025, Month: 6, Day: day, Leap: true}); err != nil {
			t.Fatalf("ValidateLunarDate(day=%d): %v", day, err)
		}
	}
	for _, day := range []int{0, -1, meta.DayCount + 1} {
		if _, err := ValidateLunarDate(LunarDate{Year: 2025, Month: 6, Day: day, Leap: true}); err == nil {
			t.Fatalf("ValidateLunarDate(day=%d) succeeded, want error", day)
		}
	}
}

func TestLunarMonthMeta_MissingLeapMonth(t *testing.T) {
	// 2024 does not have a leap sixth month.
	if _, err := LunarMonthMeta(2024, 6, true); err == nil {
		t.Fatal("missing leap month succeeded, want error")
	}
}

func TestNextLunarMonth(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		month     int
		leap      bool
		wantYear  int
		wantMonth int
		wantLeap  bool
	}{
		{name: "ordinary to leap", year: 2025, month: 6, leap: false, wantYear: 2025, wantMonth: 6, wantLeap: true},
		{name: "leap to ordinary", year: 2025, month: 6, leap: true, wantYear: 2025, wantMonth: 7, wantLeap: false},
		{name: "year wrap", year: 2024, month: 12, leap: false, wantYear: 2025, wantMonth: 1, wantLeap: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextLunarMonth(tt.year, tt.month, tt.leap)
			if err != nil {
				t.Fatalf("NextLunarMonth: %v", err)
			}
			if got.Year != tt.wantYear || got.Month != tt.wantMonth || got.Leap != tt.wantLeap {
				t.Fatalf("next = %+v, want %d-%02d leap=%t", got, tt.wantYear, tt.wantMonth, tt.wantLeap)
			}
		})
	}
}
