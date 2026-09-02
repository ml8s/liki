package qiming

import (
	"fmt"
)

const (
	composeMaxCandidates = 256
	composeDefaultNames  = 100
	composeMaxNames      = 1000
)

// ComposeNames combines pool-filtered characters into candidate given names.
func ComposeNames(request ComposeRequest) (*ComposeResult, error) {
	if len(request.First) == 0 || len(request.First) > composeMaxCandidates {
		return nil, fmt.Errorf("first must contain 1 to %d characters", composeMaxCandidates)
	}
	doubleName := len(request.Second) != 0
	if doubleName && len(request.Second) > composeMaxCandidates {
		return nil, fmt.Errorf("second must contain 1 to %d characters", composeMaxCandidates)
	}
	if err := rejectDuplicates("first", request.First); err != nil {
		return nil, err
	}
	if err := rejectDuplicates("second", request.Second); err != nil {
		return nil, err
	}
	maxNames := request.MaxNames
	if maxNames == 0 {
		maxNames = composeDefaultNames
	}
	if maxNames < 0 || maxNames > composeMaxNames {
		return nil, fmt.Errorf("max_names must be between 1 and %d", composeMaxNames)
	}

	first, err := lookupCharacters(request.First)
	if err != nil {
		return nil, err
	}
	second, err := lookupCharacters(request.Second)
	if err != nil {
		return nil, err
	}

	totalPossible := len(first)
	if doubleName {
		totalPossible *= len(second)
	}
	result := &ComposeResult{
		TotalPossible: totalPossible,
		Names:         make([]string, 0, min(totalPossible, maxNames)),
	}
	for _, firstChar := range first {
		if len(result.Names) >= maxNames {
			break
		}
		if !doubleName {
			result.Names = append(result.Names, firstChar.Char)
			continue
		}
		for _, secondChar := range second {
			if len(result.Names) >= maxNames {
				break
			}
			result.Names = append(result.Names, firstChar.Char+secondChar.Char)
		}
	}
	return result, nil
}

func lookupCharacters(chars []string) ([]Character, error) {
	out := make([]Character, 0, len(chars))
	for _, char := range chars {
		character, err := lookupCharacter(char)
		if err != nil {
			return nil, err
		}
		out = append(out, character)
	}
	return out, nil
}

func rejectDuplicates(slot string, chars []string) error {
	seen := make(map[string]bool, len(chars))
	for _, char := range chars {
		if seen[char] {
			return fmt.Errorf("%s contains duplicate character %q", slot, char)
		}
		seen[char] = true
	}
	return nil
}
