package qiming

import "fmt"

// Combo 是 pick 的一个组合：一组笔画对应的字池。
// first=第1字池，second=第2字池（双名才有）。
type Combo struct {
	ID     int      `json:"id"`
	First  []string `json:"first"`
	Second []string `json:"second,omitempty"`
}

// PickResult 是 pick 的输出：所有组合。
type PickResult struct {
	Combos []Combo `json:"combos"`
}

// PickChars 起名取字：合并 wuge（算吉笔画）与 pick（取字）。
// surname: 姓氏（wuge=true 时必填，算五格）
// wuxing1: 第1字五行
// wuxing2: 第2字五行（默认同 wuxing1）
// count: 1=单名 2=双名（恒生效，决定单/双字池）
// wuge: true=按吉笔画过滤取字 / false=按笔画拆 id 取字
func PickChars(surname, wuxing1, wuxing2 string, count int, wuge bool) ([]Combo, error) {
	if !wuge {
		chars1, err := GetChars(wuxing1)
		if err != nil {
			return nil, fmt.Errorf("pick: wuxing1: %w", err)
		}
		if wuxing2 == "" {
			wuxing2 = wuxing1
		}
		chars2, err := GetChars(wuxing2)
		if err != nil {
			return nil, fmt.Errorf("pick: wuxing2: %w", err)
		}
		// 不考虑五格：按笔画拆 id（chars1 的笔画键）
		combos := make([]Combo, 0)
		id := 0
		for stroke := range chars1 {
			combos = append(combos, Combo{
				ID:    id,
				First: charLiteToChars(chars1[stroke]),
			})
			if count == 2 {
				c2, ok := chars2[stroke]
				if !ok || len(c2) == 0 {
					combos = combos[:len(combos)-1] // 撤掉刚 append 的
					continue
				}
				combos[id].Second = charLiteToChars(c2)
			}
			id++
		}
		return combos, nil
	}

	// 考虑五格：算吉笔画对，按笔画取字
	chars1, err := GetWugeChars(wuxing1)
	if err != nil {
		return nil, fmt.Errorf("pick: wuxing1: %w", err)
	}
	if wuxing2 == "" {
		wuxing2 = wuxing1
	}
	chars2, err := GetWugeChars(wuxing2)
	if err != nil {
		return nil, fmt.Errorf("pick: wuxing2: %w", err)
	}
	ss, err := SurnameStrokesOf(surname)
	if err != nil {
		return nil, fmt.Errorf("pick: %w", err)
	}
	pairs := ListViableStrokes(ss, count)
	if count == 2 {
		pairs = FilterSancai(ss, pairs)
	}

	combos := make([]Combo, 0, len(pairs))
	id := 0
	for _, p := range pairs {
		c := Combo{ID: id}
		if c1, ok := chars1[p.S1]; ok {
			c.First = charLiteToChars(c1)
		}
		if len(c.First) == 0 {
			continue // 无字可用的笔画组合丢弃
		}
		if count == 2 {
			c2, ok := chars2[p.S2]
			if !ok || len(c2) == 0 {
				continue // 第2字无字，丢弃
			}
			c.Second = charLiteToChars(c2)
		}
		combos = append(combos, c)
		id++
	}
	return combos, nil
}

// charLiteToChars 提取 CharLite 的纯 char 数组。
func charLiteToChars(chars []CharLite) []string {
	out := make([]string, 0, len(chars))
	for _, c := range chars {
		out = append(out, c.Char)
	}
	return out
}
