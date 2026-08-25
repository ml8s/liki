package huangli

import "liki-engine/internal/engine/ganzhi"

// -- 择日体系：喜神/财神/福神/彭祖百忌 -----------------------------------------------
// These belong to the huangli day-selection (择日) domain, NOT bazi fortune-telling.
// They are pure stem→value lookups — no analysis, no scoring.

// -- 喜神方位 (joy god direction by day stem) ---------------------------------------

var xiShenDir = [11]string{
	"", "东北", "西北", "西南", "正南", "东南", // 甲己→东北, 乙庚→西北, 丙辛→西南, 丁壬→南, 戊癸→东南
	"东北", "西北", "西南", "正南", "东南",
}

// dirFromStem returns a direction from a 1-indexed stem lookup table.
func dirFromStem(stem ganzhi.Gan, table [11]string) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return table[stem]
	}
	return ""
}

// xiShenDirection returns the 喜神方位 for a given day stem.
func xiShenDirection(stem ganzhi.Gan) string { return dirFromStem(stem, xiShenDir) }

// -- 财神方位 (wealth god direction by day stem) ------------------------------------

var caiShenDir = [11]string{
	"", "东北", "东北", "正西", "正西", "正北", // 甲, 乙, 丙, 丁, 戊
	"正北", "正东", "正东", "正南", "正南", // 己, 庚, 辛, 壬, 癸
}

// caiShenDirection returns the 财神方位 for a given day stem.
func caiShenDirection(stem ganzhi.Gan) string { return dirFromStem(stem, caiShenDir) }

// -- 福神方位 (blessing god direction by day stem) ----------------------------------

var fuShenDir = [11]string{
	"", "东南", "东南", "西北", "正东", "正南", // 甲, 乙, 丙, 丁, 戊
	"正南", "西南", "西南", "西北", "正西", // 己, 庚, 辛, 壬, 癸
}

// fuShenDirection returns the 福神方位 for a given day stem.
func fuShenDirection(stem ganzhi.Gan) string { return dirFromStem(stem, fuShenDir) }

// -- 彭祖百忌 (Peng Zu daily taboos by stem and branch) -----------------------------

var stemTabooTable = [11]string{
	"", "甲不开仓财物耗散", "乙不栽植千株不长",
	"丙不修灶必见灾殃", "丁不剃头头必生疮",
	"戊不受田田主不祥", "己不破券二比并亡",
	"庚不经络织机虚张", "辛不合酱主人不尝",
	"壬不汲水更难提防", "癸不词讼理弱敌强",
}

var branchTabooTable = [13]string{
	"", "子不问卜自惹祸殃", "丑不冠带主不还乡",
	"寅不祭祀神鬼不尝", "卯不穿井水泉不香",
	"辰不哭泣必主重丧", "巳不远行财物伏藏",
	"午不苫盖屋主更张", "未不服药毒气入肠",
	"申不安床鬼祟入房", "酉不会客醉坐颠狂",
	"戌不吃犬作怪上床", "亥不嫁娶不利新郎",
}

// tabooFromStem returns a Peng Zu taboo for a given day stem.
func tabooFromStem(stem ganzhi.Gan, table [11]string) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return table[stem]
	}
	return ""
}

// tabooFromBranch returns a Peng Zu taboo for a given day branch.
func tabooFromBranch(branch ganzhi.Zhi, table [13]string) string {
	if int(branch) >= 1 && int(branch) <= 12 {
		return table[branch]
	}
	return ""
}

// pengZuStemTaboo returns the Peng Zu taboo for a given day stem.
func pengZuStemTaboo(stem ganzhi.Gan) string { return tabooFromStem(stem, stemTabooTable) }

// pengZuBranchTaboo returns the Peng Zu taboo for a given day branch.
func pengZuBranchTaboo(branch ganzhi.Zhi) string { return tabooFromBranch(branch, branchTabooTable) }
// -- 黄道黑道十二神 (Yellow/Black Path 12 Day Stars) --------------------------------
// Determined by month branch (青龙 start) + day branch offset.
// 黄道 = auspicious (6 stars), 黑道 = inauspicious (6 stars).

// huangDaoStar holds one of the 12 yellow/black path stars.
type huangDaoStar struct {
	Index    int    `json:"index"`    // 0-11
	Name     string `json:"name"`     // e.g. "青龙"
	Path     string `json:"path"`     // "黄道" or "黑道"
	Sequence int    `json:"sequence"` // position in the 12-star cycle (0=青龙)
}

// huangDaoForDay returns the yellow/black path star for a given month branch and day branch.
func huangDaoForDay(monthBranch, dayBranch ganzhi.Zhi) huangDaoStar {
	start, ok := qingLongStart[monthBranch]
	if !ok {
		return huangDaoStar{}
	}
	offset := (int(dayBranch) - int(start) + 12) % 12
	return huangDaoStars[offset]
}


// ShiChenFortune computes the hour-by-hour fortune for a given day.
func computeShiChen(riZhi, yueZhi ganzhi.Zhi, dayJianChu string) []ShiChenFortune {
	zhiNames := [12]string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	timeRanges := [12]string{
		"23:00-01:00", "01:00-03:00", "03:00-05:00", "05:00-07:00",
		"07:00-09:00", "09:00-11:00", "11:00-13:00", "13:00-15:00",
		"15:00-17:00", "17:00-19:00", "19:00-21:00", "21:00-23:00",
	}
	huangDaoNames := [12]string{
		"青龙", "明堂", "天刑", "朱雀", "金匮", "天德",
		"白虎", "玉堂", "天牢", "玄武", "司命", "勾陈",
	}
	huangDaoPath := [12]string{
		"黄道", "黄道", "黑道", "黑道", "黄道", "黄道",
		"黑道", "黄道", "黑道", "黑道", "黄道", "黑道",
	}
	jianchuSeq := [12]string{"建", "除", "满", "平", "定", "执", "破", "危", "成", "收", "开", "闭"}

	// 1. Find QingLong start hour for this riZhi
	// Use package-level qingLongStart map
	qlConfig := qingLongStart
	qlStartZhi, ok := qlConfig[yueZhi]
	if !ok {
		return nil
	}
	qlStart := int(qlStartZhi - 1) // 子=0, 丑=1...
	if qlStart < 0 { qlStart = 0 }

	// 2. Find JianChu start index
	jcStart := 0
	for i, n := range jianchuSeq {
		if n == dayJianChu { jcStart = i; break }
	}

	result := make([]ShiChenFortune, 12)
	for i := 0; i < 12; i++ {
		hsIdx := (i - qlStart + 12) % 12
		jcIdx := (jcStart + i) % 12
		// 黄道=吉, 黑道=凶
		isSuitable := huangDaoPath[hsIdx] == "黄道"
		result[i] = ShiChenFortune{
			Zhi:     zhiNames[i],
			Time:    timeRanges[i],
			HuangDaoStr: huangDaoNames[hsIdx],
			JianChu: jianchuSeq[jcIdx],
			Suitable: isSuitable,
		}
	}
	return result
}
