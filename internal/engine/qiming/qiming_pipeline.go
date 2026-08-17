package qiming

// computeWuGeFromStrokes computes the five-grid analysis from raw stroke counts.
// 单姓：天格=Total+1；复姓：天格=Total（不加 1）。人格=姓氏最后一字+名第一字。
func computeWuGeFromStrokes(s SurnameStrokes, s1, s2 int) WuGe {
	tian := s.Total
	if !s.Compound {
		tian += 1
	}
	ren := s.Last + s1
	di := s1 + s2
	if s2 == 0 {
		di = s1 + 1 // 单字名地格 = 名笔画 + 1
	}
	zong := s.Total + s1 + s2
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

