package qiming

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestDomainOracle_SurnameCandidates(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/qiming_surnames.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID            string `json:"id"`
			SourceSurname string `json:"source_surname"`
			MaxCandidates int    `json:"max_candidates"`
			Expected      struct {
				Strategy string `json:"strategy"`
				First    *struct {
					Surname    string `json:"surname"`
					Tone       int    `json:"tone"`
					MatchLevel string `json:"match_level"`
				} `json:"first"`
				ContainsSurnames []string `json:"contains_surnames"`
				Surnames         []string `json:"surnames"`
				MatchLevels      []string `json:"match_levels"`
			} `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 13 {
		t.Fatalf("cases = %d, want 13", len(doc.Cases))
	}

	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			got, err := MatchSurnames(tc.SourceSurname, tc.MaxCandidates)
			if err != nil {
				t.Fatalf("MatchSurnames() error = %v", err)
			}
			if got.Strategy != tc.Expected.Strategy {
				t.Fatalf("strategy = %q, want %q", got.Strategy, tc.Expected.Strategy)
			}
			if len(got.Candidates) == 0 {
				t.Fatal("candidates are empty")
			}
			if tc.Expected.First != nil {
				first := got.Candidates[0]
				want := tc.Expected.First
				if first.Surname != want.Surname {
					t.Fatalf("first surname = %q, want %q", first.Surname, want.Surname)
				}
				if want.Tone != 0 && first.Tone != want.Tone {
					t.Fatalf("first tone = %d, want %d", first.Tone, want.Tone)
				}
				if first.MatchLevel != want.MatchLevel {
					t.Fatalf("first match level = %q, want %q", first.MatchLevel, want.MatchLevel)
				}
				if len(first.Basis) == 0 {
					t.Fatal("first candidate has no basis")
				}
			}
			surnames := make([]string, 0, len(got.Candidates))
			levels := make([]string, 0, len(got.Candidates))
			for _, candidate := range got.Candidates {
				surnames = append(surnames, candidate.Surname)
				levels = append(levels, candidate.MatchLevel)
			}
			for _, surname := range tc.Expected.ContainsSurnames {
				if !slices.Contains(surnames, surname) {
					t.Fatalf("candidates %v do not contain %q", surnames, surname)
				}
			}
			if tc.Expected.Surnames != nil && !slices.Equal(surnames, tc.Expected.Surnames) {
				t.Fatalf("surnames = %v, want %v", surnames, tc.Expected.Surnames)
			}
			if tc.Expected.MatchLevels != nil {
				for _, level := range tc.Expected.MatchLevels {
					if !slices.Contains(levels, level) {
						t.Fatalf("levels %v do not contain %q", levels, level)
					}
				}
			}
		})
	}
}
