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
