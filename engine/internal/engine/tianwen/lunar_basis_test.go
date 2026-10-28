package tianwen

import (
	"testing"
	"time"
)

func TestLunarProjectorBases(t *testing.T) {
	ny := mustLoadLocation(t, "America/New_York")
	at := time.Date(2024, 2, 9, 23, 30, 0, 0, ny)

	official, err := NewLunarProjector(LunarBasisOfficial, ny)
	if err != nil {
		t.Fatal(err)
	}
	localLabel, err := NewLunarProjector(LunarBasisLocalLabel, ny)
	if err != nil {
		t.Fatal(err)
	}
	localAstro, err := NewLunarProjector(LunarBasisLocalAstronomical, ny)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := official.SelectorDate(at), NewCivilDate(2024, 2, 10); got != want {
		t.Fatalf("official selector date = %s, want %s", got, want)
	}
	if got, want := localLabel.SelectorDate(at), NewCivilDate(2024, 2, 9); got != want {
		t.Fatalf("local_label selector date = %s, want %s", got, want)
	}
	if got, want := localAstro.SelectorDate(at), NewCivilDate(2024, 2, 9); got != want {
		t.Fatalf("local_astronomical selector date = %s, want %s", got, want)
	}

	if got, want := official.LunarAt(at), (LunarDate{Year: 2024, Month: 1, Day: 1}); got != want {
		t.Fatalf("official lunar = %+v, want %+v", got, want)
	}
	if got, want := localLabel.LunarAt(at), (LunarDate{Year: 2023, Month: 12, Day: 30}); got != want {
		t.Fatalf("local_label lunar = %+v, want %+v", got, want)
	}
	if got, want := localAstro.LunarAt(at), (LunarDate{Year: 2024, Month: 1, Day: 1}); got != want {
		t.Fatalf("local_astronomical lunar = %+v, want %+v", got, want)
	}

	if got, want := official.SelectorInstant(NewCivilDate(2024, 2, 10)), time.Date(2024, 2, 10, 12, 0, 0, 0, cstLocation); !got.Equal(want) {
		t.Fatalf("official selector instant = %s, want %s", got, want)
	}
	if got, want := localLabel.SelectorInstant(NewCivilDate(2024, 2, 9)), time.Date(2024, 2, 9, 12, 0, 0, 0, ny); !got.Equal(want) {
		t.Fatalf("local_label selector instant = %s, want %s", got, want)
	}
	if got, want := localAstro.SelectorInstant(NewCivilDate(2024, 2, 9)), time.Date(2024, 2, 9, 12, 0, 0, 0, ny); !got.Equal(want) {
		t.Fatalf("local_astronomical selector instant = %s, want %s", got, want)
	}

	if got, want := localLabel.OutputBasis(), "local_label"; got != want {
		t.Fatalf("local_label output basis = %q, want %q", got, want)
	}
	if got, want := official.OutputBasis(), "official"; got != want {
		t.Fatalf("official output basis = %q, want %q", got, want)
	}
	if got := localAstro.OutputBasis(); got != "" {
		t.Fatalf("local_astronomical output basis = %q, want empty", got)
	}
}

func TestParseLunarBasis(t *testing.T) {
	for _, input := range []string{"", "official", "local_label", "local_astronomical"} {
		if _, err := ParseLunarBasis(input); err != nil {
			t.Fatalf("ParseLunarBasis(%q): %v", input, err)
		}
	}
	if _, err := ParseLunarBasis("solar"); err == nil {
		t.Fatal("ParseLunarBasis(solar) should fail")
	}
}
