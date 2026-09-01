package qiming

import (
	"fmt"
	"sort"
)

// Character is a naming character with modern and Kangxi stroke semantics.
type Character struct {
	Char         string `json:"char"`
	Element      Wuxing `json:"wuxing"`
	Stroke       int    `json:"stroke"`
	KangxiStroke int    `json:"kangxi_stroke"`
	Radical      string `json:"radical"`
	Pinyin       string `json:"pinyin"`
	Tone         int    `json:"tone"`
	Traditional  string `json:"traditional,omitempty"`
	KangxiForm   string `json:"kangxi_form,omitempty"`
}

// CharLite is a lightweight character view for the HTTP chars endpoint.
type CharLite struct {
	Char         string `json:"char"`
	Wuxing       string `json:"wuxing"`
	Stroke       int    `json:"stroke"`
	KangxiStroke int    `json:"kangxi_stroke"`
	Radical      string `json:"radical"`
	Pinyin       string `json:"pinyin"`
	Tone         int    `json:"tone"`
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

func lookupKangxiStroke(char string) (int, error) {
	rs := []rune(char)
	if len(rs) == 0 {
		return 0, fmt.Errorf("character is empty")
	}
	entry, ok := kangxiByRune[rs[0]]
	if !ok || entry.Stroke <= 0 {
		return 0, fmt.Errorf("character %q not found in kangxi stroke table", char)
	}
	return entry.Stroke, nil
}

func lookupSurnameKangxiStroke(char string) (int, error) {
	rs := []rune(char)
	if len(rs) == 0 {
		return 0, fmt.Errorf("character is empty")
	}
	if entry, ok := surnameKangxiByRune[rs[0]]; ok && entry.Stroke > 0 {
		return entry.Stroke, nil
	}
	return lookupKangxiStroke(char)
}

// SurnameStrokesOf 计算姓氏的 Kangxi 笔画信息（单姓/复姓）。
// Total=全部笔画之和；Last=最后一字笔画；Compound=是否复姓。
func SurnameStrokesOf(surname string) (SurnameStrokes, error) {
	rs := []rune(surname)
	if len(rs) == 0 {
		return SurnameStrokes{}, fmt.Errorf("surname is empty")
	}
	ss := SurnameStrokes{Compound: len(rs) > 1}
	for i, r := range rs {
		stroke, err := lookupSurnameKangxiStroke(string(r))
		if err != nil {
			return SurnameStrokes{}, fmt.Errorf("surname %q not found in kangxi stroke table", surname)
		}
		ss.Total += stroke
		if i == len(rs)-1 {
			ss.Last = stroke
		}
	}
	return ss, nil
}

// getCharsByElement groups characters by modern stroke or by Kangxi stroke.
// Negative characters are filtered at this level — they never appear in the char pool.
func getCharsByElement(elem Wuxing, useKangxiStroke bool) map[int][]CharLite {
	chars := charByElement[elem]
	result := make(map[int][]CharLite)
	for _, c := range chars {
		if negativeChars[c.Char] {
			continue
		}
		stroke := c.Stroke
		if useKangxiStroke {
			stroke = c.KangxiStroke
		}
		result[stroke] = append(result[stroke], CharLite{
			Char:         c.Char,
			Wuxing:       c.Element.String(),
			Stroke:       c.Stroke,
			KangxiStroke: c.KangxiStroke,
			Radical:      c.Radical,
			Pinyin:       c.Pinyin,
			Tone:         c.Tone,
		})
	}
	for _, v := range result {
		sort.Slice(v, func(i, j int) bool { return v[i].Char < v[j].Char })
	}
	return result
}
