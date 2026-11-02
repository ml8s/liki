package bazi

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// NatalCore is the narrow natal fact set consumed by derived Bazi methods.
// It intentionally excludes derived chart projections; only the facts needed
// to place a flow period or compare two natal subjects are carried forward.
type NatalCore struct {
	Pillars   [4]ganzhi.Zhu
	NaYin     [4]string
	DaYun     *DaYun
	Gender    ganzhi.Gender
	BirthYear int
}

func (core NatalCore) ToBazi() ganzhi.Bazi {
	return ganzhi.Bazi{
		Nian: core.Pillars[0],
		Yue:  core.Pillars[1],
		Ri:   core.Pillars[2],
		Shi:  core.Pillars[3],
	}
}

func computeNatalCore(gt time.Time, st time.Time, bz ganzhi.Bazi, gender ganzhi.Gender) NatalCore {
	return NatalCore{
		Pillars:   bz.Slice(),
		NaYin:     computeNaYin(bz),
		DaYun:     computeDaYun(gt, bz, gender),
		Gender:    gender,
		BirthYear: gt.Year(),
	}
}
