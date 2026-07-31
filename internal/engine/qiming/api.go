// Package qiming provides起名 computation.
//
// Types
//
//	Wuxing
//	Evaluation
//	WuGe, SanCai
//	Character, CharLite, Phonetic
//
// Functions
//
//	SurnameStroke(surname) → (int, error)
//	GetChars(wuxing) → (map[int][]CharLite, error)
//	EvaluateNames(surname, names, yong, xi, ji) → ([]Evaluation, error)
package qiming

import (
	"fmt"
)

// SurnameStroke returns the Kangxi stroke count for a surname, or an error if not found.
func SurnameStroke(surname string) (int, error) {
	n := lookupKangxiStroke(surname)
	if n == 0 {
		return 0, fmt.Errorf("surname %q not found in Kangxi dictionary", surname)
	}
	return n, nil
}

// GetChars returns naming characters of the given element, grouped by stroke.
func GetChars(wuxingName string) (map[int][]CharLite, error) {
	elem := wuxingFromChinese(wuxingName)
	if elem == 0 {
		return nil, fmt.Errorf("chars: invalid wuxing %q", wuxingName)
	}
	return getCharsByElement(elem), nil
}

// LookupChar looks up a single character's wuxing, stroke count, and other attributes.
// Returns nil if the character is not in the naming database.
func LookupChar(char string) *Character {
	rs := []rune(char)
	if len(rs) != 1 {
		return nil
	}
	ce, ok := charByRune[rs[0]]
	if !ok {
		return nil
	}
	return &ce
}

// EvaluateNames evaluates a batch of given names with full wuxing analysis.
func EvaluateNames(surname string, givenNames []string, yongShen string, xiShen, jiShen []string) ([]Evaluation, error) {
	surnameStrokes, err := SurnameStroke(surname)
	if err != nil {
		return nil, fmt.Errorf("evaluate names: %w", err)
	}

	yongElem := Wuxing(0)
	if yongShen != "" {
		yongElem = wuxingFromChinese(yongShen)
	}
	xiElems := make([]Wuxing, len(xiShen))
	for i, xs := range xiShen {
		xiElems[i] = wuxingFromChinese(xs)
	}
	jiElems := make([]Wuxing, len(jiShen))
	for i, js := range jiShen {
		jiElems[i] = wuxingFromChinese(js)
	}

	var results []Evaluation
	for _, given := range givenNames {
		rs := []rune(given)
		if len(rs) < 1 || len(rs) > 2 {
			continue
		}

		var charEntries []Character
		for _, r := range rs {
			ce, ok := charByRune[r]
			if !ok {
				continue
			}
			charEntries = append(charEntries, ce)
		}
		if len(charEntries) != len(rs) {
			continue
		}

		s1, s2 := charEntries[0].Stroke, 0
		if len(charEntries) > 1 {
			s2 = charEntries[1].Stroke
		}
		wg := computeWuGeFromStrokes(surnameStrokes, s1, s2)
		sc := computeSanCai(wg.TianGe.Element, wg.RenGe.Element, wg.DiGe.Element)
		phon := analyzePhonetic(charEntries)

		wuxingMatch := false
		if yongElem != 0 {
			for _, ce := range charEntries {
				if ce.Element == yongElem {
					wuxingMatch = true
					break
				}
			}
		}

		ev := Evaluation{
			Name:        surname + given,
			Surname:     surname,
			GivenName:   given,
			Characters:  charEntries,
			WuGe:        wg,
			SanCai:      sc,
			Phonetic:    phon,
			WuxingMatch: wuxingMatch,
		}

		if yongShen != "" {
			wx := &struct {
				Yong bool `json:"yong"`
				Xi   bool `json:"xi,omitempty"`
				Ji   bool `json:"ji,omitempty"`
			}{}
			wx.Yong = wuxingMatch
			if len(xiElems) > 0 {
				for _, ce := range charEntries {
					for _, xe := range xiElems {
						if ce.Element == xe {
							wx.Xi = true
							break
						}
					}
					if wx.Xi {
						break
					}
				}
			}
			if len(jiElems) > 0 {
				for _, ce := range charEntries {
					for _, je := range jiElems {
						if ce.Element == je {
							wx.Ji = true
							break
						}
					}
					if wx.Ji {
						break
					}
				}
			}
			ev.Wuxing = wx
		}

		results = append(results, ev)
	}

	return results, nil
}
