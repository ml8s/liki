package bazi

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// Core is the narrow natal fact set consumed by derived Bazi methods.
// It intentionally excludes derived chart projections; only the facts needed
// to place a flow period or compare two natal subjects are carried forward.
type Core struct {
	Pillars   [4]ganzhi.Zhu
	NaYin     [4]string
	DaYun     *DaYun
	Gender    ganzhi.Gender
	BirthYear int
}

func (core Core) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: core.Pillars[0],
		Yue:  core.Pillars[1],
		Ri:   core.Pillars[2],
		Shi:  core.Pillars[3],
	}
}

// Core projects the chart back to the narrow fact set consumed by
// derived methods. The DaYun pointer is shared; derived methods treat it as
// read-only input.
func (c Chart) Core() Core {
	return Core{
		Pillars:   [4]ganzhi.Zhu{c.Nian.Zhu, c.Yue.Zhu, c.Ri.Zhu, c.Shi.Zhu},
		NaYin:     c.NaYinArray(),
		DaYun:     c.DaYun,
		Gender:    c.Gender,
		BirthYear: c.BirthYear,
	}
}

func computeCore(gt time.Time, st time.Time, bz ganzhi.Bazi, gender ganzhi.Gender) Core {
	return Core{
		Pillars:   bz.Slice(),
		NaYin:     computeNaYin(bz),
		DaYun:     computeDaYun(gt, bz, gender),
		Gender:    gender,
		BirthYear: gt.Year(),
	}
}
