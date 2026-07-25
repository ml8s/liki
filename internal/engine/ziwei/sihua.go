package ziwei

func computeSiHua(yearGan Gan) siHuaResult {
	stars, ok := siHuaTable[yearGan]
	if !ok {
		return nil
	}
	return siHuaResult{
		stars[0]: HuaLu, stars[1]: HuaQuan, stars[2]: HuaKe, stars[3]: HuaJi,
	}
}
