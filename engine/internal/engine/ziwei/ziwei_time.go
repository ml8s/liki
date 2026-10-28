package ziwei

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// LiuNianTime is the resolved lunar-year period for 流年. Start is the civil
// date of 正月初一; End is the next lunar year's 正月初一 (exclusive).
type LiuNianTime struct {
	LunarYear   int               `json:"lunar_year"`
	Start       tianwen.CivilDate `json:"start"`
	End         tianwen.CivilDate `json:"end"`
	BoundaryDay bool              `json:"boundary_day"`
}

// LiuYueTime is the resolved lunar-month period for 流月. It keeps the target
// lunar date because the flow-month projection distinguishes month halves.
type LiuYueTime struct {
	Date        tianwen.CivilDate `json:"date"`
	Lunar       tianwen.LunarDate `json:"lunar"`
	Start       tianwen.CivilDate `json:"start"`
	End         tianwen.CivilDate `json:"end"`
	BoundaryDay bool              `json:"boundary_day"`
}

// LiuRiTime is the resolved lunar day for 流日. Date is the selector date for
// the chosen basis; LocalDate and BeijingDate are echoed projections.
type LiuRiTime struct {
	Instant     time.Time         `json:"-"`
	Location    *time.Location    `json:"-"`
	Timezone    string            `json:"timezone"`
	Date        tianwen.CivilDate `json:"date"`
	LocalDate   tianwen.CivilDate `json:"local_date"`
	BeijingDate tianwen.CivilDate `json:"beijing_date"`
	Lunar       tianwen.LunarDate `json:"lunar"`
}

// ResolveLiuNianTime resolves the lunar year containing date. The 正月初一
// boundary date belongs to the new year.
func ResolveLiuNianTime(date tianwen.CivilDate) (LiuNianTime, error) {
	return ResolveLiuNianTimeWith(tianwen.DefaultLunarProjector(), date)
}

// ResolveLiuNianTimeWith resolves the lunar year using an explicit projector.
func ResolveLiuNianTimeWith(p tianwen.LunarProjector, date tianwen.CivilDate) (LiuNianTime, error) {
	if err := date.Validate(); err != nil {
		return LiuNianTime{}, err
	}
	lunar := p.SolarDateToLunar(date)
	start, err := lunarDayCivilDateWith(p, lunar.Year, 1, 1, false)
	if err != nil {
		return LiuNianTime{}, err
	}
	end, err := lunarDayCivilDateWith(p, lunar.Year+1, 1, 1, false)
	if err != nil {
		return LiuNianTime{}, err
	}
	return LiuNianTime{
		LunarYear:   lunar.Year,
		Start:       start,
		End:         end,
		BoundaryDay: date == start,
	}, nil
}

// NextLiuNianTime returns the next lunar year's resolved period.
func NextLiuNianTime(t LiuNianTime) (LiuNianTime, error) {
	return NextLiuNianTimeWith(tianwen.DefaultLunarProjector(), t)
}

// NextLiuNianTimeWith returns the next lunar year's resolved period using an
// explicit projector.
func NextLiuNianTimeWith(p tianwen.LunarProjector, t LiuNianTime) (LiuNianTime, error) {
	return ResolveLiuNianTimeWith(p, t.End)
}

// ResolveLiuYueTime resolves the lunar month containing date. The lunar month
// start date belongs to the new month.
func ResolveLiuYueTime(date tianwen.CivilDate) (LiuYueTime, error) {
	return ResolveLiuYueTimeWith(tianwen.DefaultLunarProjector(), date)
}

// ResolveLiuYueTimeWith resolves the lunar month using an explicit projector.
func ResolveLiuYueTimeWith(p tianwen.LunarProjector, date tianwen.CivilDate) (LiuYueTime, error) {
	if err := date.Validate(); err != nil {
		return LiuYueTime{}, err
	}
	lunar := p.SolarDateToLunar(date)
	resolved, err := resolveLiuYueTimeWith(p, lunar)
	if err != nil {
		return LiuYueTime{}, err
	}
	resolved.Date = date
	return resolved, nil
}

func resolveLiuYueTimeWith(p tianwen.LunarProjector, lunar tianwen.LunarDate) (LiuYueTime, error) {
	if _, err := p.ValidateLunarDate(lunar); err != nil {
		return LiuYueTime{}, err
	}
	start, err := lunarDayCivilDateWith(p, lunar.Year, lunar.Month, 1, lunar.Leap)
	if err != nil {
		return LiuYueTime{}, err
	}
	next, err := p.NextLunarMonth(lunar.Year, lunar.Month, lunar.Leap)
	if err != nil {
		return LiuYueTime{}, err
	}
	end, err := lunarDayCivilDateWith(p, next.Year, next.Month, 1, next.Leap)
	if err != nil {
		return LiuYueTime{}, err
	}
	return LiuYueTime{
		Lunar:       lunar,
		Start:       start,
		End:         end,
		BoundaryDay: lunar.Day == 1,
	}, nil
}

// NextLiuYueTime returns the next lunar month's resolved period. The target
// day is preserved and clamped to the next month's length.
func NextLiuYueTime(t LiuYueTime) (LiuYueTime, error) {
	return NextLiuYueTimeWith(tianwen.DefaultLunarProjector(), t)
}

// NextLiuYueTimeWith returns the next lunar month's resolved period using an
// explicit projector.
func NextLiuYueTimeWith(p tianwen.LunarProjector, t LiuYueTime) (LiuYueTime, error) {
	next, err := p.NextLunarMonth(t.Lunar.Year, t.Lunar.Month, t.Lunar.Leap)
	if err != nil {
		return LiuYueTime{}, err
	}
	day := t.Lunar.Day
	if day > next.DayCount {
		day = next.DayCount
	}
	nextLunar := tianwen.LunarDate{
		Year:  next.Year,
		Month: next.Month,
		Day:   day,
		Leap:  next.Leap,
	}
	resolved, err := resolveLiuYueTimeWith(p, nextLunar)
	if err != nil {
		return LiuYueTime{}, err
	}
	date, err := p.CivilDateFromLunar(nextLunar)
	if err != nil {
		return LiuYueTime{}, err
	}
	resolved.Date = date
	return resolved, nil
}

// ResolveLiuRiTime resolves the lunar day containing date.
func ResolveLiuRiTime(date tianwen.CivilDate) (LiuRiTime, error) {
	return ResolveLiuRiTimeWith(tianwen.DefaultLunarProjector(), date)
}

// ResolveLiuRiTimeWith resolves the lunar day using an explicit projector.
func ResolveLiuRiTimeWith(p tianwen.LunarProjector, date tianwen.CivilDate) (LiuRiTime, error) {
	if err := date.Validate(); err != nil {
		return LiuRiTime{}, err
	}
	return LiuRiTime{
		Date:        date,
		LocalDate:   date,
		BeijingDate: date,
		Lunar:       p.SolarDateToLunar(date),
	}, nil
}

// ResolveLiuRiTimeAt resolves the lunar day for an absolute instant. The flow
// day is selected by the projector's lunar basis.
func ResolveLiuRiTimeAt(instant time.Time, tz *time.Location, tzName string) (LiuRiTime, error) {
	return ResolveLiuRiTimeAtWith(tianwen.DefaultLunarProjector(), instant, tz, tzName)
}

// ResolveLiuRiTimeAtWith resolves the lunar day for an instant using an
// explicit projector.
func ResolveLiuRiTimeAtWith(p tianwen.LunarProjector, instant time.Time, tz *time.Location, tzName string) (LiuRiTime, error) {
	if tzName == "" {
		return LiuRiTime{}, fmt.Errorf("liuri: timezone is required")
	}
	if tz == nil {
		tz = time.UTC
	}
	localDate := tianwen.StandardDateAt(instant, tz)
	selectorDate := p.SelectorDate(instant)
	beijingDate := tianwen.BeijingDateAt(instant)
	return LiuRiTime{
		Instant:     instant,
		Location:    tz,
		Timezone:    tzName,
		Date:        selectorDate,
		LocalDate:   localDate,
		BeijingDate: beijingDate,
		Lunar:       p.SolarDateToLunar(selectorDate),
	}, nil
}

// NextLiuRiTime returns the next civil day's lunar date. Converting through
// the civil date handles 29/30-day months and leap months.
func NextLiuRiTime(t LiuRiTime) (LiuRiTime, error) {
	return NextLiuRiTimeWith(tianwen.DefaultLunarProjector(), t)
}

// NextLiuRiTimeWith returns the next civil day's lunar date using an explicit
// projector.
func NextLiuRiTimeWith(p tianwen.LunarProjector, t LiuRiTime) (LiuRiTime, error) {
	if !t.Instant.IsZero() {
		tz := t.Location
		if tz == nil {
			var err error
			tz, err = time.LoadLocation(t.Timezone)
			if err != nil {
				return LiuRiTime{}, err
			}
		}
		return ResolveLiuRiTimeAtWith(p, t.Instant.Add(24*time.Hour), tz, t.Timezone)
	}
	if _, err := p.ValidateLunarDate(t.Lunar); err != nil {
		return LiuRiTime{}, err
	}
	nextDate, err := p.CivilDateFromLunar(t.Lunar)
	if err != nil {
		return LiuRiTime{}, err
	}
	nextDate = nextDate.AddDays(1)
	return ResolveLiuRiTimeWith(p, nextDate)
}

func lunarDayCivilDateWith(p tianwen.LunarProjector, year, month, day int, leap bool) (tianwen.CivilDate, error) {
	ld := tianwen.LunarDate{Year: year, Month: month, Day: day, Leap: leap}
	if _, err := p.ValidateLunarDate(ld); err != nil {
		return tianwen.CivilDate{}, err
	}
	date, err := p.CivilDateFromLunar(ld)
	if err != nil {
		return tianwen.CivilDate{}, fmt.Errorf("lunar %04d-%02d-%02d has no civil date: %w", year, month, day, err)
	}
	return date, nil
}

// LiuShiTime is the resolved hour-level time for 流时. Instant, Timezone and
// Longitude are retained so the series can advance one 时辰 at a time.
type LiuShiTime struct {
	Instant   time.Time         `json:"-"`
	Location  *time.Location    `json:"-"`
	Timezone  string            `json:"timezone"`
	Longitude float64           `json:"longitude"`
	LocalDate tianwen.CivilDate `json:"local_date"`
	Lunar     tianwen.LunarDate `json:"lunar"`
	Shichen   ganzhi.Zhi        `json:"shichen"`
}

// ResolveLiuShiTime derives the lunar day and LAST shichen from an instant.
func ResolveLiuShiTime(instant time.Time, tz *time.Location, tzName string, longitude float64) (LiuShiTime, error) {
	return ResolveLiuShiTimeWith(tianwen.DefaultLunarProjector(), instant, tz, tzName, longitude)
}

// ResolveLiuShiTimeWith derives the lunar day and LAST shichen using an
// explicit projector.
func ResolveLiuShiTimeWith(p tianwen.LunarProjector, instant time.Time, tz *time.Location, tzName string, longitude float64) (LiuShiTime, error) {
	if longitude < -180 || longitude > 180 {
		return LiuShiTime{}, fmt.Errorf("liushi: longitude must be within [-180,180], got %v", longitude)
	}
	if tzName == "" {
		return LiuShiTime{}, fmt.Errorf("liushi: timezone is required")
	}
	if tz == nil {
		tz = time.UTC
	}
	shichen := tianwen.ShichenAt(instant, tz, longitude)
	return LiuShiTime{
		Instant:   instant,
		Location:  tz,
		Timezone:  tzName,
		Longitude: longitude,
		LocalDate: tianwen.StandardDateAt(instant, tz),
		Lunar:     p.LunarAt(instant),
		Shichen:   shichen,
	}, nil
}

// NextLiuShiTime advances one 时辰 (two clock hours), preserving the existing
// ziwei_liushi series contract.
func NextLiuShiTime(t LiuShiTime) (LiuShiTime, error) {
	return NextLiuShiTimeWith(tianwen.DefaultLunarProjector(), t)
}

// NextLiuShiTimeWith advances one 时辰 using an explicit projector.
func NextLiuShiTimeWith(p tianwen.LunarProjector, t LiuShiTime) (LiuShiTime, error) {
	tz := t.Location
	if tz == nil {
		var err error
		tz, err = time.LoadLocation(t.Timezone)
		if err != nil {
			return LiuShiTime{}, err
		}
	}
	return ResolveLiuShiTimeWith(p, t.Instant.Add(2*time.Hour), tz, t.Timezone, t.Longitude)
}

// ComputeLiuShiAtTime projects a resolved flow-hour time.
func ComputeLiuShiAtTime(chart Chart, t LiuShiTime) (LiuShi, error) {
	return ComputeLiuShiAtTimeWith(chart, t, tianwen.DefaultLunarProjector())
}

// ComputeLiuShiAtTimeWith projects a resolved flow-hour time using an explicit
// projector.
func ComputeLiuShiAtTimeWith(chart Chart, t LiuShiTime, p tianwen.LunarProjector) (LiuShi, error) {
	return ComputeLiuShiWith(chart, t.Lunar, t.Shichen, p)
}
