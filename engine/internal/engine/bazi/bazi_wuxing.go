package bazi

import "liki-engine/internal/engine/ganzhi"

// elementThatGenerates returns the element that generates (生) the given element.
func elementThatGenerates(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Sheng(i, e) {
			return i
		}
	}
	return 0
}

// elementThatControls returns the element that controls (克) the given element.
func elementThatControls(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Ke(i, e) {
			return i
		}
	}
	return 0
}

// elementThatDrains returns the element that the given element generates
// (泄, 生). E.g., Wood drains to Fire (木生火).
func elementThatDrains(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Sheng(e, i) {
			return i
		}
	}
	return 0
}

// elementControlledBy returns the element that the given element controls
// (克). E.g., Wood controls Earth (木克土).
func elementControlledBy(e ganzhi.Wuxing) ganzhi.Wuxing {
	for i := ganzhi.WxMu; i <= ganzhi.WxShui; i++ {
		if ganzhi.Ke(e, i) {
			return i
		}
	}
	return 0
}
