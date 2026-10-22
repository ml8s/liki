package tianwen

import (
	"errors"
	"testing"
	"time"
)

func TestResolveCivilCandidatesDSTFold(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	candidates, err := ResolveCivilCandidates(2026, 11, 1, 1, 30, 0, 0, loc)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidates = %v, want DST fold pair", candidates)
	}
	if !candidates[0].Before(candidates[1]) {
		t.Fatal("fold candidates must be ordered")
	}
}

func TestResolveCivilCandidatesDSTGap(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveCivilCandidates(2026, 3, 8, 2, 30, 0, 0, loc); !errors.Is(err, ErrCivilTimeGap) {
		t.Fatalf("error = %v, want ErrCivilTimeGap", err)
	}
}
