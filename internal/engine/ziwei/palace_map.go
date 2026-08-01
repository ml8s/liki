package ziwei

// ── 宫名真相源 ────────────────────────────────────────────────────
// 宫名是「地支」的标签，不是独立的顺序数组。
// 命盘确定后，命宫支 = 命宫，其余宫名按固定方向分布在其余地支上。
//
// 根表达：
//   palaceLabels[i] 是 palaceIndex（命宫序）的固定标签：
//     0=命宫 1=兄弟 2=夫妻 3=子女 4=财帛 5=疾厄
//     6=迁移 7=仆役 8=官禄 9=田宅 10=福德 11=父母
//   palaceIndex 语义 = 「从命宫支起，逆时针（地支递减）走 i 步」。
//
// 推导（顺逆是遍历方向，不是存储）：
//   zhiZM1(palaceIndex) = (mingZM1 - palaceIndex + 12) % 12   // 逆时针
//   palaceIndex(zhiZM1) = (mingZM1 - zhiZM1 + 12) % 12
// 流盘（顺时针/display）只需把「地支 → palaceIndex → 标签」方向反过来。
//
// 没有「宫名顺序数组」——宫名分布由上述公式 + 固定标签完全确定。

// palaceLabels maps palaceIndex (0=命宫) to its fixed label.
// 这不是宫名分布数组，而是 palaceIndex 的定义（命宫序标签）。
var palaceLabels = [12]string{
	"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
	"迁移", "仆役", "官禄", "田宅", "福德", "父母",
}

// zhiIdxToPalaceIndex returns the palace index (0=命宫) whose earth branch
// is zhiIdx (0=子), anchored at the 命宫 branch. Direction: 逆时针.
func zhiIdxToPalaceIndex(mingZhiIdx, zhiIdx int) palaceIndex {
	return palaceIndex(((mingZhiIdx - zhiIdx) % 12 + 12) % 12)
}

// palaceIndexToZhiIdx returns the earth branch index (zhiIdx, 0=子) of a palace
// index, anchored at the 命宫 branch. Direction: 逆时针.
func palaceIndexToZhiIdx(mingZhiIdx int, pi palaceIndex) int {
	return ((mingZhiIdx - int(pi)) % 12 + 12) % 12
}


