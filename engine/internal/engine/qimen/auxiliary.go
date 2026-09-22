package qimen

import "liki-engine/internal/engine/ganzhi"

func maXingZhi(leadZhi ganzhi.Zhi) ganzhi.Zhi {
	if leadZhi >= 1 && int(leadZhi) <= len(maXingTable) {
		return maXingTable[int(leadZhi)-1]
	}
	return 0
}

func findMaXing(leadZhi ganzhi.Zhi) BranchPalace {
	branch := maXingZhi(leadZhi)
	return BranchPalace{Branch: branch, Gong: zhiPalace(branch)}
}

func kongWangZhi(leadZhu ganzhi.Zhu) [2]ganzhi.Zhi {
	idx := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi)
	return xunKongTable[idx/10]
}

func findKongWang(leadZhu ganzhi.Zhu) [2]BranchPalace {
	z := kongWangZhi(leadZhu)
	return [2]BranchPalace{
		{Branch: z[0], Gong: zhiPalace(z[0])},
		{Branch: z[1], Gong: zhiPalace(z[1])},
	}
}
