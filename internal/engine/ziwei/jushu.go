package ziwei

import "liki-engine/internal/engine/ganzhi"

// determineJuShu returns the bureau number from ming gong stem and branch.
func determineJuShu(mingGan Gan, mingZhi Zhi) juShu {
	nayinName := ganzhi.NayinLabel(mingGan, mingZhi)
	wx := ganzhi.NayinWuxing(nayinName)
	if wx == 0 {
		return 0
	}
	return juShuFromWuxing(wx)
}
