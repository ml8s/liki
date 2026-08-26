package qiming

import (
	"fmt"
	"sort"
)

// Character is a naming character from the ben-hua general standard Chinese table.
type Character struct {
	Char        string `json:"char"`
	Element     Wuxing `json:"wuxing"`
	Stroke      int    `json:"stroke"`
	Radical     string `json:"radical"`
	Pinyin      string `json:"pinyin"`
	Tone        int    `json:"tone"`
	Traditional string `json:"traditional,omitempty"`
}

// CharLite is a lightweight character view for the HTTP chars endpoint.
type CharLite struct {
	Char    string `json:"char"`
	Wuxing  string `json:"wuxing"`
	Stroke  int    `json:"stroke"`
	Radical string `json:"radical"`
	Pinyin  string `json:"pinyin"`
	Tone    int    `json:"tone"`
}

func elementYAMLToChinese(e string) string {
	switch e {
	case "wood":
		return "木"
	case "fire":
		return "火"
	case "earth":
		return "土"
	case "metal":
		return "金"
	case "water":
		return "水"
	}
	return e
}

// lookupKangxiStroke returns the Kangxi dictionary stroke count for a character.
func lookupKangxiStroke(char string) int {
	rs := []rune(char)
	if len(rs) > 0 {
		if ce, ok := charByRune[rs[0]]; ok {
			return ce.Stroke
		}
	}
	return 0
}

// SurnameStrokesOf 计算姓氏的康熙笔画信息（单姓/复姓）。
// Total=全部笔画之和；Last=最后一字笔画；Compound=是否复姓。
func SurnameStrokesOf(surname string) (SurnameStrokes, error) {
	rs := []rune(surname)
	if len(rs) == 0 {
		return SurnameStrokes{}, fmt.Errorf("surname is empty")
	}
	ss := SurnameStrokes{Compound: len(rs) > 1}
	for i, r := range rs {
		ce, ok := charByRune[r]
		if !ok {
			return SurnameStrokes{}, fmt.Errorf("surname %q not found in Kangxi dictionary", surname)
		}
		ss.Total += ce.Stroke
		if i == len(rs)-1 {
			ss.Last = ce.Stroke
		}
	}
	return ss, nil
}

// getCharsByElement returns all characters of the given element, grouped by stroke.
// Negative characters are filtered at this level — they never appear in the char pool.
func getCharsByElement(elem Wuxing) map[int][]CharLite {
	chars := charByElement[elem]
	result := make(map[int][]CharLite)
	for _, c := range chars {
		if negativeChars[c.Char] {
			continue
		}
		result[c.Stroke] = append(result[c.Stroke], CharLite{
			Char:    c.Char,
			Wuxing:  c.Element.String(),
			Stroke:  c.Stroke,
			Radical: c.Radical,
			Pinyin:  c.Pinyin,
			Tone:    c.Tone,
		})
	}
	for _, v := range result {
		sort.Slice(v, func(i, j int) bool { return v[i].Char < v[j].Char })
	}
	return result
}
