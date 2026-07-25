package qiming

// ComposeNames builds name strings from character pools.
// When pairs is non-nil, only the specified (s1, s2) stroke pairs are combined.
// When pairs is nil, all stroke pairs in chars1 × chars2 are combined (full Cartesian product).
func ComposeNames(surname string, chars1, chars2 map[int][]CharLite, pairs []StrokePair) []string {
	seen := make(map[string]bool)
	var names []string

	if len(chars2) == 0 {
		for _, chars := range chars1 {
			for _, c := range chars {
				name := surname + c.Char
				if seen[name] {
					continue
				}
				names = append(names, name)
				seen[name] = true
			}
		}
		return names
	}

	if len(pairs) > 0 {
		for _, p := range pairs {
			pool1, ok1 := chars1[p.S1]
			pool2, ok2 := chars2[p.S2]
			if !ok1 || !ok2 {
				continue
			}
			names = appendDoubleNames(surname, pool1, pool2, seen, names)
		}
		return names
	}

	for _, pool1 := range chars1 {
		for _, pool2 := range chars2 {
			names = appendDoubleNames(surname, pool1, pool2, seen, names)
		}
	}
	return names
}

func appendDoubleNames(surname string, pool1, pool2 []CharLite, seen map[string]bool, names []string) []string {
	for _, c1 := range pool1 {
		for _, c2 := range pool2 {
			name := surname + c1.Char + c2.Char
			if seen[name] {
				continue
			}
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}

// computeWuGeFromStrokes computes the five-grid analysis from raw stroke counts.
func computeWuGeFromStrokes(surnameStroke, s1, s2 int) WuGe {
	tian := surnameStroke + 1
	ren := surnameStroke + s1
	di := s1 + s2
	if s2 == 0 {
		di = s1 + 1 // 单字名地格 = 名笔画 + 1
	}
	zong := surnameStroke + s1 + s2
	wai := zong - ren + 1
	if wai < 1 {
		wai = 1
	}
	return WuGe{
		TianGe: strokeResult(tian),
		RenGe:  strokeResult(ren),
		DiGe:   strokeResult(di),
		WaiGe:  strokeResult(wai),
		ZongGe: strokeResult(zong),
	}
}

