package ziwei

func computeSiHua(nianGan Gan) siHuaResult {
	stars, ok := siHuaTable[nianGan]
	if !ok {
		return nil
	}
	return siHuaResult{
		stars[0]: HuaLu, stars[1]: HuaQuan, stars[2]: HuaKe, stars[3]: HuaJi,
	}
}
