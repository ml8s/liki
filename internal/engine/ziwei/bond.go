package ziwei

// Bond holds the compatibility analysis between two charts.
type Bond struct {
	// A入B: A's ming gong zhi in B's palace index.
	AIntoB palaceIndex `json:"a_into_b"`
	// B入A: B's ming gong zhi in A's palace index.
	BIntoA palaceIndex `json:"b_into_a"`
	// StarCross: A's main stars mapped to B's palaces.
	StarCross []bondStar `json:"star_cross"`
	// SiHuaCross: A's sihua stars falling into B's palaces.
	SiHuaCross []bondSiHua `json:"sihua_cross"`
}

// bondStar records one star from chart A landing in chart B's palace.
type bondStar struct {
	Star      starIndex   `json:"star"`
	FromA     palaceIndex `json:"from_a"` // A's palace index
	IntoB     palaceIndex `json:"into_b"` // B's palace index (same zhi)
}

// bondSiHua records one sihua star from A landing in B.
type bondSiHua struct {
	Star  starIndex   `json:"star"`
	Type  siHuaType   `json:"type"`
	IntoB palaceIndex `json:"into_b"`
}

// computeBond builds compatibility between two charts.
func ComputeBond(a, b Chart) Bond {
	// 宫位互入
	aMingZhi := a.Palaces[0].Zhi
	bMingZhi := b.Palaces[0].Zhi
	aIntoB := findPalaceByZhi(b.Palaces, aMingZhi)
	bIntoA := findPalaceByZhi(a.Palaces, bMingZhi)

	// 星曜互入: A's stars → which of B's palaces have the same zhi
	var starCross []bondStar
	for i := range a.Palaces {
		z := a.Palaces[i].Zhi
		bIdx := findPalaceByZhi(b.Palaces, z)
		for _, s := range a.Palaces[i].Stars {
			if s.IsMajor {
				starCross = append(starCross, bondStar{
					Star: s.Star, FromA: palaceIndex(i), IntoB: bIdx,
				})
			}
		}
	}

	// 四化互引: A's sihua stars → B
	var sihuaCross []bondSiHua
	for star, h := range a.SiHua {
		// Find the star in A's chart to get its zhi, then map to B
		for i := range a.Palaces {
			for _, s := range a.Palaces[i].Stars {
				if s.Star == star {
					bIdx := findPalaceByZhi(b.Palaces, a.Palaces[i].Zhi)
					sihuaCross = append(sihuaCross, bondSiHua{
						Star: star, Type: h, IntoB: bIdx,
					})
				}
			}
		}
	}

	return Bond{
		AIntoB: aIntoB, BIntoA: bIntoA,
		StarCross: starCross, SiHuaCross: sihuaCross,
	}
}

func findPalaceByZhi(palaces [12]palace, zhi Zhi) palaceIndex {
	for i, p := range palaces {
		if p.Zhi == zhi {
			return palaceIndex(i)
		}
	}
	return 0
}
