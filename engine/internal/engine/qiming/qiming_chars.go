// Package qiming provides naming character pools, composition, and evaluation.
package qiming

import "fmt"

// Character is a naming character consumed by the current qiming domain.
type Character struct {
	Char    string `json:"char"`
	Element Wuxing `json:"wuxing"`
	Stroke  int    `json:"stroke"`
	Radical string `json:"radical,omitempty"`
	Pinyin  string `json:"pinyin"`
	Tone    int    `json:"tone"`
}

func lookupCharacter(char string) (Character, error) {
	rs := []rune(char)
	if len(rs) != 1 {
		return Character{}, fmt.Errorf("character %q must be a single character", char)
	}
	character, ok := charByRune[rs[0]]
	if !ok {
		return Character{}, fmt.Errorf("character %q not found in database", char)
	}
	if negativeChars[char] {
		return Character{}, fmt.Errorf("character %q is excluded from naming", char)
	}
	return character, nil
}

// LookupChar returns one character from the qiming database.
func LookupChar(char string) *Character {
	rs := []rune(char)
	if len(rs) != 1 {
		return nil
	}
	character, ok := charByRune[rs[0]]
	if !ok {
		return nil
	}
	return &character
}

func charactersByElement(elem Wuxing) []Character {
	return charByElement[elem]
}
