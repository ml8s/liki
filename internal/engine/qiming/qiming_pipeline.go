package qiming

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

