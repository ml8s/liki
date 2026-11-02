package ziwei

import "liki-engine/internal/engine/ganzhi"

// ZiweiCore is the narrow natal fact set consumed by derived Ziwei methods.
// It intentionally excludes palace projections; derived methods must not take
// the complete chart.
type ZiweiCore struct {
	Gender    ganzhi.Gender
	NianGan   Gan
	JuShu     juShu
	BirthYear int
}

// ZiweiCore projects the chart to the narrow facts consumed by derived methods.
func (c Chart) ZiweiCore() ZiweiCore {
	return ZiweiCore{
		Gender:    c.Gender,
		NianGan:   c.NianGan,
		JuShu:     c.JuShu,
		BirthYear: c.BirthYear,
	}
}

// ZiweiFlowContext carries only the natal facts required by flow month/day/hour
// and annual projection. It prevents derived Ziwei methods from consuming the
// complete Chart.
type ZiweiFlowContext struct {
	Core            ZiweiCore
	BirthLunarMonth int
	BirthLunarDay   int
	BirthIsLeap     bool
	ShiZhi          Zhi
	MingGongZhi     Zhi
	MingGongName    string
	SiHuaPalaces    map[starIndex]gongIndex
}

// ZiweiFlowContext extracts the narrow facts used by flow methods.
func (c Chart) ZiweiFlowContext() ZiweiFlowContext {
	siHuaPalaces := make(map[starIndex]gongIndex)
	for _, palace := range c.GongWei {
		for _, star := range palace.Stars {
			if _, exists := siHuaPalaces[star.Star]; !exists {
				siHuaPalaces[star.Star] = palace.Index
			}
		}
	}
	return ZiweiFlowContext{
		Core:            c.ZiweiCore(),
		BirthLunarMonth: c.BirthLunarMonth,
		BirthLunarDay:   c.LunarDay,
		BirthIsLeap:     c.BirthIsLeap,
		ShiZhi:          c.ShiZhi,
		MingGongZhi:     c.GongWei[c.MingGong].Zhi,
		MingGongName:    c.GongWei[c.MingGong].Name,
		SiHuaPalaces:    siHuaPalaces,
	}
}

// ZiweiBondPalaceFacts carries only the palace facts consumed by bond output.
type ZiweiBondPalaceFacts struct {
	Name       string
	MajorStars []string
}

// ZiweiBondSideFacts carries only the natal palace facts consumed by bond.
type ZiweiBondSideFacts struct {
	MingGong ZiweiBondPalaceFacts
	FuQiGong ZiweiBondPalaceFacts
	ZiNvGong ZiweiBondPalaceFacts
}

// ZiweiBondSideFacts extracts the palace facts consumed by bond.
func (c Chart) ZiweiBondSideFacts() ZiweiBondSideFacts {
	facts := func(idx int) ZiweiBondPalaceFacts {
		majorStars := make([]string, 0, len(c.GongWei[idx].Stars))
		for _, star := range c.GongWei[idx].Stars {
			if star.IsMajor {
				majorStars = append(majorStars, star.Name)
			}
		}
		return ZiweiBondPalaceFacts{Name: c.GongWei[idx].Name, MajorStars: majorStars}
	}
	return ZiweiBondSideFacts{
		MingGong: facts(int(c.MingGong)),
		FuQiGong: facts(2),
		ZiNvGong: facts(3),
	}
}
