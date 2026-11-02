package tianwen

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// ResolveInput is a birth-time normalization request. Exactly one of the four
// time forms (Time, Shichen, Range) must be provided.
type ResolveInput struct {
	Calendar  string // "solar" | "lunar"（壁钟形态必填）
	Date      string // YYYY-MM-DD（calendar=solar）
	Lunar     LunarDate
	Timezone  string  // IANA 名（壁钟形态必填）
	Longitude float64 // 出生地经度（必填，-180..180）

	// 地点范围（可选）：城市/行政区的东西经度范围，用于边界不确定性。
	LongitudeRange *LongitudeRange

	Time    string        // HH:MM 或 HH:MM:SS（当地壁钟）
	Shichen string        // 真太阳时时辰
	Range   *ResolveRange // 同日民用区间
}

// LongitudeRange is the east-west span of the recorded place.
type LongitudeRange struct {
	West float64
	East float64
}

// ResolveRange is a same-day civil time range.
type ResolveRange struct {
	Start string
	End   string
}

// Reason domains.
const (
	reasonInput      = "input"
	reasonShichen    = "时辰"
	reasonSolarDay   = "日界"
	reasonSolarTerm  = "交节"
	reasonTimeSystem = "时制"
)

var chartBoundaryKinds = []BoundaryKind{
	BoundarySolarTerm, BoundarySolarDay, BoundaryShichen,
}

// ResolveReason explains why a candidate exists.
type ResolveReason struct {
	Domain string `json:"domain"`
	Detail string `json:"detail"`
}

// ResolveCandidate is one possible chart time point with its reasons.
type ResolveCandidate struct {
	StandardTime time.Time       `json:"-"`
	Reasons      []ResolveReason `json:"reasons"`
}

// ResolveResult is the candidate set for one time fact.
type ResolveResult struct {
	Candidates []ResolveCandidate
}

type resolveForm int

const (
	formTime resolveForm = iota
	formShichen
	formRange
)

// ResolveStandard normalizes a time fact into natal-subject candidates. It
// deliberately removes DST offsets and uses the zone's standard offset.
func ResolveStandard(in ResolveInput) (ResolveResult, error) {
	return resolveCandidates(in)
}

func resolveCandidates(in ResolveInput) (ResolveResult, error) {
	form, err := resolveFormOf(in)
	if err != nil {
		return ResolveResult{}, err
	}
	if in.Longitude < -180 || in.Longitude > 180 {
		return ResolveResult{}, fmt.Errorf("longitude must be within [-180,180]")
	}
	if err := validateLongitudeRange(in.Longitude, in.LongitudeRange); err != nil {
		return ResolveResult{}, err
	}

	zone, err := resolveLocation(in.Timezone)
	if err != nil {
		return ResolveResult{}, err
	}

	year, month, day, err := resolveBaseDate(in, form)
	if err != nil {
		return ResolveResult{}, err
	}

	standard, err := standardOffsetForDate(year, month, day, zone)
	if err != nil {
		return ResolveResult{}, err
	}
	offsets := []int{standard}

	type candidate struct {
		value ResolveCandidate
		input bool
	}
	var all []candidate

	for _, offset := range offsets {
		fixed := time.FixedZone("", offset)
		base, err := buildPrecisionWindow(in, form, year, month, day, fixed)
		if err != nil {
			return ResolveResult{}, err
		}
		anchor := resolveAnchor(in, form, base)
		expanded, err := expandWindow(base, in.Longitude, in.LongitudeRange)
		if err != nil {
			return ResolveResult{}, err
		}
		bounds := BoundariesInWindow(expanded, chartBoundaryKinds,
			BoundaryOptions{Longitude: in.Longitude, Timezone: fixed})
		cells := SplitWindow(expanded, bounds)
		inputIndex := cellIndex(cells, anchor)

		for index, cell := range cells {
			isAnchorCell := index == inputIndex
			instant := cellRepresentative(cell)
			if isAnchorCell {
				instant = anchor
			}
			if isAnchorCell && offset == standard {
				all = append(all, candidate{
					value: ResolveCandidate{
						StandardTime: instant.In(fixed),
						Reasons: []ResolveReason{{
							Domain: reasonInput,
							Detail: inputDetail(anchor, bounds),
						}},
					},
					input: true,
				})
				continue
			}
			reasons := make([]ResolveReason, 0, 3)
			if offset != standard {
				detail := "历史时制"
				if instant.In(zone).IsDST() {
					detail = "夏令时"
				}
				reasons = append(reasons, ResolveReason{Domain: reasonTimeSystem, Detail: detail})
			}
			reasons = append(reasons, crossedReasons(cells, bounds, inputIndex, index, instant, in.Longitude, fixed)...)
			all = append(all, candidate{
				value: ResolveCandidate{StandardTime: instant.In(fixed), Reasons: dedupeReasons(reasons)},
			})
		}
	}

	sort.SliceStable(all, func(i, j int) bool {
		if all[i].input != all[j].input {
			return all[i].input
		}
		return all[i].value.StandardTime.Before(all[j].value.StandardTime)
	})
	result := ResolveResult{Candidates: make([]ResolveCandidate, 0, len(all))}
	for _, item := range all {
		result.Candidates = append(result.Candidates, item.value)
	}
	return result, nil
}

func resolveFormOf(in ResolveInput) (resolveForm, error) {
	forms := 0
	form := formTime
	if strings.TrimSpace(in.Time) != "" {
		forms++
		form = formTime
	}
	if strings.TrimSpace(in.Shichen) != "" {
		forms++
		form = formShichen
	}
	if in.Range != nil {
		forms++
		form = formRange
	}
	if forms != 1 {
		return 0, fmt.Errorf("gregorian/time/shichen/range 必须且只能提供一个")
	}
	return form, nil
}

func validateLongitudeRange(longitude float64, span *LongitudeRange) error {
	if span == nil {
		return nil
	}
	if span.West > span.East {
		return fmt.Errorf("longitude_range.west must not exceed east")
	}
	if span.West < -180 || span.East > 180 {
		return fmt.Errorf("longitude_range must be within [-180,180]")
	}
	if longitude < span.West || longitude > span.East {
		return fmt.Errorf("longitude must be within longitude_range")
	}
	return nil
}

func resolveLocation(name string) (*time.Location, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("timezone is required for wall-clock forms")
	}
	loc, err := time.LoadLocation(trimmed)
	if err != nil {
		return nil, fmt.Errorf("timezone must be an IANA name (e.g. Asia/Shanghai): %w", err)
	}
	return loc, nil
}

func resolveBaseDate(in ResolveInput, form resolveForm) (int, time.Month, int, error) {
	calendar := strings.TrimSpace(in.Calendar)
	if calendar == "" {
		return 0, 0, 0, fmt.Errorf("calendar is required for wall-clock forms")
	}
	switch calendar {
	case "solar":
		t, err := time.Parse("2006-01-02", strings.TrimSpace(in.Date))
		if err != nil {
			return 0, 0, 0, fmt.Errorf("date must be YYYY-MM-DD: %w", err)
		}
		return t.Year(), t.Month(), t.Day(), nil
	case "lunar":
		loc, err := resolveLocation(in.Timezone)
		if err != nil {
			return 0, 0, 0, err
		}
		projector, err := NewLunarProjector(LunarBasisLocalAstronomical, loc)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid lunar basis: %w", err)
		}
		date, err := projector.CivilDateFromLunar(in.Lunar)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid lunar date: %w", err)
		}
		return date.Year, time.Month(date.Month), date.Day, nil
	default:
		return 0, 0, 0, fmt.Errorf("calendar must be solar or lunar")
	}
}

// standardOffsetForDate finds the zone's non-DST offset around a civil date.
// It intentionally probes nearby days because a requested wall clock may fall
// inside a DST gap.
func standardOffsetForDate(year int, month time.Month, day int, zone *time.Location) (int, error) {
	probes := []time.Time{
		time.Date(year, 1, 15, 12, 0, 0, 0, zone),
		time.Date(year, 7, 15, 12, 0, 0, 0, zone),
		time.Date(year, month, day, 12, 0, 0, 0, zone),
	}
	standard := 0
	foundNonDST := false
	for _, probe := range probes {
		_, offset := probe.In(zone).Zone()
		if !probe.In(zone).IsDST() {
			return offset, nil
		}
		if !foundNonDST || absInt(offset) < absInt(standard) {
			standard = offset
			foundNonDST = true
		}
	}
	if !foundNonDST {
		return 0, fmt.Errorf("cannot determine standard offset in %s", zone.String())
	}
	return standard, nil
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// buildPrecisionWindow returns the civil window implied by the input precision:
// HH:MM -> one minute; HH:MM:SS -> one second;
// shichen/range -> their civil windows.
func buildPrecisionWindow(in ResolveInput, form resolveForm, year int, month time.Month, day int, zone *time.Location) (Window, error) {
	date := time.Date(year, month, day, 0, 0, 0, 0, zone)

	switch form {
	case formTime:
		start, end, err := parseTimeWindow(in.Time)
		if err != nil {
			return Window{}, err
		}
		return Window{Start: date.Add(start), End: date.Add(end)}, nil
	case formRange:
		start, err := parseMinuteOfDay(in.Range.Start)
		if err != nil {
			return Window{}, err
		}
		end, err := parseMinuteOfDay(in.Range.End)
		if err != nil {
			return Window{}, err
		}
		if end <= start {
			return Window{}, fmt.Errorf("range.end must be after range.start")
		}
		return Window{Start: date.Add(start), End: date.Add(end)}, nil
	case formShichen:
		zhi, err := ganzhi.ParseZhi(strings.TrimSpace(in.Shichen))
		if err != nil {
			return Window{}, fmt.Errorf("shichen must be one of 子丑寅卯辰巳午未申酉戌亥: %w", err)
		}
		startMin, endMin := shichenSolarBounds(zhi)
		tzHours := offsetHours(zone, date)
		start := civilForSolarMinutes(date, startMin, in.Longitude, tzHours)
		end := civilForSolarMinutes(date, endMin, in.Longitude, tzHours)
		if !end.After(start) {
			return Window{}, fmt.Errorf("failed to map shichen %s to a civil window", zhi)
		}
		return Window{Start: start, End: end}, nil
	default:
		return Window{}, fmt.Errorf("unsupported time form")
	}
}

func resolveAnchor(in ResolveInput, form resolveForm, base Window) time.Time {
	if form == formTime {
		return base.Start
	}
	return base.Start.Add(base.End.Sub(base.Start) / 2)
}

// expandWindow widens the civil window by the east-west longitude span. The
// true-solar reading is monotonic in both instant and longitude, so the
// reachable civil instants are the window shifted by the longitude range.
func expandWindow(base Window, longitude float64, span *LongitudeRange) (Window, error) {
	if span == nil {
		return base, nil
	}
	westShift := time.Duration(4 * (longitude - span.West) * float64(time.Minute))
	eastShift := time.Duration(4 * (span.East - longitude) * float64(time.Minute))
	return Window{Start: base.Start.Add(-westShift), End: base.End.Add(eastShift)}, nil
}

func cellIndex(cells []Window, instant time.Time) int {
	for index, cell := range cells {
		if instant.Before(cell.Start) {
			continue
		}
		if instant.Before(cell.End) || cell.IsPoint() || index == len(cells)-1 {
			return index
		}
	}
	return len(cells) - 1
}

func inputDetail(anchor time.Time, bounds []Boundary) string {
	nearest, distance := nearestBoundary(anchor, bounds)
	if nearest == nil {
		return "输入时间"
	}
	minutes := int(math.Round(distance.Minutes()))
	if minutes <= 0 {
		return fmt.Sprintf("输入时间（恰在%s边界）", boundaryDomain(nearest.Kind))
	}
	return fmt.Sprintf("输入时间（距%s %d 分钟）", boundaryDomain(nearest.Kind), minutes)
}

func nearestBoundary(anchor time.Time, bounds []Boundary) (*Boundary, time.Duration) {
	var nearest *Boundary
	var distance time.Duration
	for index := range bounds {
		diff := bounds[index].At.Sub(anchor)
		if diff < 0 {
			diff = -diff
		}
		if nearest == nil || diff < distance {
			nearest = &bounds[index]
			distance = diff
		}
	}
	return nearest, distance
}

func boundaryDomain(kind BoundaryKind) string {
	switch kind {
	case BoundarySolarTerm:
		return reasonSolarTerm
	case BoundarySolarDay:
		return reasonSolarDay
	case BoundaryShichen:
		return reasonShichen
	default:
		return string(kind)
	}
}

func cellRepresentative(cell Window) time.Time {
	if cell.IsPoint() {
		return cell.Start
	}
	return cell.Start.Add(cell.End.Sub(cell.Start) / 2)
}

// crossedReasons returns the reasons for the boundaries between the input cell
// and the candidate cell. The candidate representative is used for side details
// so boundary rounding does not leak the neighboring branch.
func crossedReasons(cells []Window, bounds []Boundary, inputIndex, index int, representative time.Time, longitude float64, zone *time.Location) []ResolveReason {
	var reasons []ResolveReason
	switch {
	case index > inputIndex:
		for cellIndex := inputIndex + 1; cellIndex <= index; cellIndex++ {
			reasons = append(reasons, reasonsFor(cells[cellIndex].Start, bounds, true, representative, longitude, zone)...)
		}
	case index < inputIndex:
		for cellIndex := index; cellIndex < inputIndex; cellIndex++ {
			reasons = append(reasons, reasonsFor(cells[cellIndex].End, bounds, false, representative, longitude, zone)...)
		}
	}
	return reasons
}

func reasonsFor(at time.Time, bounds []Boundary, after bool, representative time.Time, longitude float64, zone *time.Location) []ResolveReason {
	var reasons []ResolveReason
	for _, boundary := range bounds {
		if !boundary.At.Equal(at) {
			continue
		}
		reasons = append(reasons, ResolveReason{
			Domain: boundaryDomain(boundary.Kind),
			Detail: sideDetail(boundary, after, representative, longitude, zone),
		})
	}
	return reasons
}

func sideDetail(boundary Boundary, after bool, representative time.Time, longitude float64, zone *time.Location) string {
	tzHours := offsetHours(zone, representative)
	solar := ApparentSolarFromCivil(representative, longitude, tzHours)
	switch boundary.Kind {
	case BoundarySolarTerm:
		side := "前"
		if after {
			side = "后"
		}
		return termName(boundary.Label) + side
	case BoundaryShichen:
		return shichenFromApparentSolarMinutes(apparentSolarMinutes(solar)).String() + "时"
	case BoundarySolarDay:
		return solar.Format("2006-01-02")
	default:
		return boundary.Label
	}
}

func termName(label string) string {
	fields := strings.Fields(label)
	if len(fields) == 0 {
		return label
	}
	return fields[0]
}

func dedupeReasons(reasons []ResolveReason) []ResolveReason {
	seen := make(map[string]bool, len(reasons))
	out := reasons[:0]
	for _, reason := range reasons {
		key := reason.Domain + "|" + reason.Detail
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, reason)
	}
	return out
}

func parseTimeWindow(value string) (time.Duration, time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	var layout string
	var precision time.Duration
	switch len(trimmed) {
	case 5:
		layout, precision = "15:04", time.Minute
	case 8:
		layout, precision = "15:04:05", time.Second
	default:
		return 0, 0, fmt.Errorf("time must be HH:MM or HH:MM:SS")
	}
	parsed, err := time.Parse(layout, trimmed)
	if err != nil {
		return 0, 0, fmt.Errorf("time must be HH:MM or HH:MM:SS: %w", err)
	}
	start := time.Duration(parsed.Hour())*time.Hour + time.Duration(parsed.Minute())*time.Minute + time.Duration(parsed.Second())*time.Second
	return start, start + precision, nil
}

func parseMinuteOfDay(value string) (time.Duration, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("time must be HH:MM: %w", err)
	}
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute, nil
}

// shichenSolarBounds returns the [start,end) true-solar minutes for a zhi.
func shichenSolarBounds(zhi ganzhi.Zhi) (float64, float64) {
	start := 120*float64(int(zhi)-1) - 60
	if start < 0 {
		start += 1440
	}
	return start, start + 120
}
