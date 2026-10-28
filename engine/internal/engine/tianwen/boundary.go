package tianwen

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// JieQiNames are the 24 solar-term names in JieQiLongitudes order (0=立春).
var JieQiNames = [24]string{
	"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
	"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
	"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
	"立冬", "小雪", "大雪", "冬至", "小寒", "大寒",
}

// Window is a half-open civil instant window [Start, End). A degenerate window
// (End <= Start) represents an exact instant.
type Window struct {
	Start time.Time
	End   time.Time
}

// IsPoint reports whether the window is degenerate (exact instant).
func (w Window) IsPoint() bool { return !w.End.After(w.Start) }

// BoundaryKind identifies a calendar boundary that can split a time window.
type BoundaryKind string

const (
	// BoundaryDay is the civil midnight of the selected lunar reference. It
	// changes the lunar date (and therefore 紫微 day-based stars).
	BoundaryDay BoundaryKind = "day"
	// BoundarySolarDay is the midnight of the true-solar reading. It changes the
	// 八字/六爻 day pillar (which is derived from the solar date).
	BoundarySolarDay BoundaryKind = "solar_day"
	// BoundaryShichen is a true-solar shichen boundary (odd solar hours).
	BoundaryShichen BoundaryKind = "shichen"
	// BoundarySolarTerm is a solar-term instant (交节).
	BoundarySolarTerm BoundaryKind = "solar_term"
	// BoundaryLunarYear is the lunar new year's civil midnight (正月初一) in the
	// selected lunar reference. It changes the 紫微 year stem/branch.
	BoundaryLunarYear BoundaryKind = "lunar_year"
)

// Boundary is a calendar boundary instant.
type Boundary struct {
	Kind  BoundaryKind // 边界类型
	At    time.Time    // 边界瞬时
	Label string       // 人类可读（如 "立春 01-02 16:26:53"）
}

// BoundaryOptions carries the location inputs needed by solar-frame boundaries.
type BoundaryOptions struct {
	Longitude  float64        // 出生地经度
	Timezone   *time.Location // 输入时区（真太阳时换算与展示）；nil → CST
	LunarBasis LunarBasis     // 农历基准；零值 → local_astronomical
}

var cstLocation = time.FixedZone("CST", 8*3600)

func (o BoundaryOptions) zone() *time.Location {
	if o.Timezone != nil {
		return o.Timezone
	}
	return cstLocation
}

func (o BoundaryOptions) lunarProjector() LunarProjector {
	basis, err := ParseLunarBasis(string(o.LunarBasis))
	if err != nil {
		basis = LunarBasisLocalAstronomical
	}
	projector, err := NewLunarProjector(basis, o.zone())
	if err != nil {
		return DefaultLunarProjector()
	}
	return projector
}

func (o BoundaryOptions) dayLocation() *time.Location {
	basis, err := ParseLunarBasis(string(o.LunarBasis))
	if err == nil && basis == LunarBasisOfficial {
		return cstLocation
	}
	return o.zone()
}

// AllBoundaryKinds is the union watched by tianwen.resolve: every boundary any
// system cares about (see doc comment on each kind).
var AllBoundaryKinds = []BoundaryKind{
	BoundaryDay, BoundarySolarDay, BoundaryShichen, BoundarySolarTerm, BoundaryLunarYear,
}

// BoundariesInWindow returns the boundaries of the requested kinds strictly
// inside the window (Start < At < End), sorted by instant.
func BoundariesInWindow(w Window, kinds []BoundaryKind, opts BoundaryOptions) []Boundary {
	if !w.End.After(w.Start) {
		return nil
	}
	tz := opts.zone()
	wanted := make(map[BoundaryKind]bool, len(kinds))
	for _, kind := range kinds {
		wanted[kind] = true
	}
	inside := func(at time.Time) bool { return at.After(w.Start) && at.Before(w.End) }
	var out []Boundary

	if wanted[BoundaryDay] {
		dayLoc := opts.dayLocation()
		label := "午夜（当地） "
		if dayLoc == cstLocation {
			label = "午夜（CST） "
		}
		for d := midnightIn(dayLoc, w.Start); !d.After(w.End); d = d.AddDate(0, 0, 1) {
			if inside(d) {
				out = append(out, Boundary{Kind: BoundaryDay, At: d,
					Label: label + d.In(dayLoc).Format("01-02 15:04")})
			}
		}
	}
	if wanted[BoundaryLunarYear] {
		projector := opts.lunarProjector()
		year := projector.LunarAt(w.Start).Year
		for _, lunarYear := range []int{year - 1, year, year + 1} {
			at := LunarNewYearInstantIn(lunarYear, projector)
			if inside(at) {
				out = append(out, Boundary{Kind: BoundaryLunarYear, At: at,
					Label: fmt.Sprintf("农历年界 %d年正月初一", lunarYear)})
			}
		}
	}
	if wanted[BoundarySolarTerm] {
		startYear := w.Start.In(tz).Year()
		endYear := w.End.In(tz).Year()
		for year := startYear - 1; year <= endYear+1; year++ {
			for i, lon := range JieQiLongitudes {
				at := SolarTermTime(year, lon)
				if inside(at) {
					out = append(out, Boundary{Kind: BoundarySolarTerm, At: at,
						Label: JieQiNames[i] + " " + at.In(tz).Format("01-02 15:04:05")})
				}
			}
		}
	}
	if wanted[BoundaryShichen] || wanted[BoundarySolarDay] {
		base := midnightIn(tz, w.Start)
		tzHours := offsetHours(tz, w.Start)
		if wanted[BoundaryShichen] {
			for _, at := range crossings(w, base, opts.Longitude, tzHours, 120, 60) {
				out = append(out, Boundary{Kind: BoundaryShichen, At: at,
					Label: "时辰界 " + at.In(tz).Format("15:04")})
			}
		}
		if wanted[BoundarySolarDay] {
			for _, at := range crossings(w, base, opts.Longitude, tzHours, 1440, 0) {
				out = append(out, Boundary{Kind: BoundarySolarDay, At: at,
					Label: "太阳日界 " + at.In(tz).Format("01-02 15:04")})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}

// SplitWindow splits w at the given boundaries into half-open cells covering the
// whole window. Boundaries outside (Start, End) are ignored.
func SplitWindow(w Window, boundaries []Boundary) []Window {
	if !w.End.After(w.Start) {
		return []Window{w}
	}
	points := make([]time.Time, 0, len(boundaries))
	for _, b := range boundaries {
		if b.At.After(w.Start) && b.At.Before(w.End) {
			points = append(points, b.At)
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Before(points[j]) })

	cells := make([]Window, 0, len(points)+1)
	start := w.Start
	for _, p := range points {
		if !p.After(start) {
			continue
		}
		cells = append(cells, Window{Start: start, End: p})
		start = p
	}
	if start.Before(w.End) {
		cells = append(cells, Window{Start: start, End: w.End})
	}
	if len(cells) == 0 {
		return []Window{w}
	}
	return cells
}

// LunarNewYearInstant returns the civil midnight (CST) of the first day of the
// given lunar year under the legacy official reference.
func LunarNewYearInstant(lunarYear int) time.Time {
	return LunarNewYearInstantIn(lunarYear, DefaultLunarProjector())
}

// LunarNewYearInstantIn returns the first-day midnight of the lunar year in the
// projector's date frame.
func LunarNewYearInstantIn(lunarYear int, projector LunarProjector) time.Time {
	date, err := projector.CivilDateFromLunar(LunarDate{Year: lunarYear, Month: 1, Day: 1})
	if err != nil {
		return time.Time{}
	}
	return projector.DayStart(date)
}

func midnightIn(loc *time.Location, t time.Time) time.Time {
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func offsetHours(loc *time.Location, t time.Time) float64 {
	_, offset := t.In(loc).Zone()
	return float64(offset) / 3600
}

// solarMinutesSince returns continuous true-solar minutes elapsed since base
// (a civil midnight in the same fixed-offset zone as t). The longitude shift and
// equation of time are applied here, exactly as in ApparentSolarFromCivil.
func solarMinutesSince(t, base time.Time, longitude, timezone float64) float64 {
	lonOffset := 4 * (longitude - 15*timezone)
	return t.Sub(base).Minutes() + lonOffset + equationOfTime(t)
}

// crossings scans the window for instants where the solar reading crosses
// values congruent to residue modulo modulus (shichen: odd solar hours,
// modulus=120, residue=60; solar day: modulus=1440, residue=0).
func crossings(w Window, base time.Time, longitude, timezone, modulus, residue float64) []time.Time {
	const step = time.Minute
	var out []time.Time
	prevT := w.Start
	prev := math.Mod(solarMinutesSince(prevT, base, longitude, timezone)-residue, modulus)
	if prev < 0 {
		prev += modulus
	}
	for t := w.Start.Add(step); !t.After(w.End); t = t.Add(step) {
		cur := math.Mod(solarMinutesSince(t, base, longitude, timezone)-residue, modulus)
		if cur < 0 {
			cur += modulus
		}
		if cur < prev {
			out = append(out, bisectCrossing(prevT, t, base, longitude, timezone, residue, modulus))
		}
		prevT, prev = t, cur
	}
	return out
}

func bisectCrossing(a, b time.Time, base time.Time, longitude, timezone, residue, modulus float64) time.Time {
	va := solarMinutesSince(a, base, longitude, timezone)
	k := math.Floor((va-residue)/modulus) + 1
	target := residue + modulus*k
	lo, hi := a, b
	for hi.Sub(lo) > time.Second {
		mid := lo.Add(hi.Sub(lo) / 2)
		if solarMinutesSince(mid, base, longitude, timezone) < target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi
}

// civilForSolarMinutes inverts solarMinutesSince: returns the civil instant
// whose solar reading equals target (minutes since base).
func civilForSolarMinutes(base time.Time, target, longitude, timezone float64) time.Time {
	lo := base.Add(-12 * time.Hour)
	hi := base.Add(36 * time.Hour)
	for hi.Sub(lo) > time.Second {
		mid := lo.Add(hi.Sub(lo) / 2)
		if solarMinutesSince(mid, base, longitude, timezone) < target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi
}
