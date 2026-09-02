package qiming

import (
	"fmt"
)

// PickChars returns character pools for the requested naming slots.
func PickChars(wuxing1, wuxing2 string, count int) (*PickResult, error) {
	elem1 := wuxingFromChinese(wuxing1)
	if elem1 == 0 {
		return nil, fmt.Errorf("invalid wuxing1 %q", wuxing1)
	}
	if count != 1 && count != 2 {
		return nil, fmt.Errorf("count must be 1 or 2")
	}
	if count == 1 && wuxing2 != "" {
		return nil, fmt.Errorf("wuxing2 is not allowed when count is 1")
	}
	if wuxing2 == "" {
		wuxing2 = wuxing1
	}
	elem2 := wuxingFromChinese(wuxing2)
	if elem2 == 0 {
		return nil, fmt.Errorf("invalid wuxing2 %q", wuxing2)
	}

	first := charactersByElement(elem1)
	if len(first) == 0 {
		return nil, fmt.Errorf("no characters available for wuxing %q", wuxing1)
	}
	result := &PickResult{
		Wuxing1: elem1.String(),
		Pools: []CandidatePool{{
			Slot:  "first",
			Chars: characterNames(first),
		}},
	}
	if count == 1 {
		return result, nil
	}

	second := charactersByElement(elem2)
	if len(second) == 0 {
		return nil, fmt.Errorf("no characters available for wuxing %q", wuxing2)
	}
	if elem2 != elem1 {
		result.Wuxing2 = elem2.String()
	}
	result.Pools = append(result.Pools, CandidatePool{
		Slot:  "second",
		Chars: characterNames(second),
	})
	return result, nil
}

func characterNames(chars []Character) []string {
	out := make([]string, 0, len(chars))
	for _, char := range chars {
		out = append(out, char.Char)
	}
	return out
}
