package qimen

import "liki-engine/internal/engine/ganzhi"

// zhiPalace returns the gong index (1-9) where a branch sits.
func zhiPalace(z ganzhi.Zhi) GongIndex {
	switch int(z) {
	case 1: // 子 → 坎
		return GongKan
	case 2, 3: // 丑, 寅 → 艮
		return GongGen
	case 4: // 卯 → 震
		return GongZhen
	case 5, 6: // 辰, 巳 → 巽
		return GongXun
	case 7: // 午 → 离
		return GongLi
	case 8, 9: // 未, 申 → 坤
		return GongKun
	case 10: // 酉 → 兑
		return GongDui
	case 11, 12: // 戌, 亥 → 乾
		return GongQian
	}
	return GongKan
}
// palaceZhi returns the principal branch of a gong.
func palaceZhi(p GongIndex) ganzhi.Zhi {
	switch p {
	case GongKan:
		return ganzhi.ZhiZi // 子
	case GongKun:
		return ganzhi.ZhiWei // 未
	case GongZhen:
		return ganzhi.ZhiMao // 卯
	case GongXun:
		return ganzhi.ZhiSi // 巳
	case GongQian:
		return ganzhi.ZhiXu // 戌
	case GongDui:
		return ganzhi.ZhiYou // 酉
	case GongGen:
		return ganzhi.ZhiYin // 寅
	case GongLi:
		return ganzhi.ZhiWu // 午
	}
	return ganzhi.ZhiZi
}

// palaceWuxing returns the five-element of a gong.
func palaceWuxing(p GongIndex) ganzhi.Wuxing {
	switch p {
	case GongKan:
		return ganzhi.WxShui
	case GongKun, GongZhong, GongGen:
		return ganzhi.WxTu
	case GongZhen, GongXun:
		return ganzhi.WxMu
	case GongQian, GongDui:
		return ganzhi.WxJin
	case GongLi:
		return ganzhi.WxHuo
	}
	return ganzhi.WxTu
}

// starHomePalace returns the home gong index (0-based internal) for a star.
func starHomePalace(s StarIndex) int {
	switch s {
	case StarTianPeng:
		return 0 // 坎
	case StarTianRui:
		return 1 // 坤
	case StarTianChong:
		return 2 // 震
	case StarTianFu:
		return 3 // 巽
	case StarTianQin:
		return 4 // 中
	case StarTianXin:
		return 5 // 乾
	case StarTianZhu:
		return 6 // 兑
	case StarTianRen:
		return 7 // 艮
	case StarTianYing:
		return 8 // 离
	}
	return 4
}
