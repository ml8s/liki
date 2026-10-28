package bazi

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// LiuRiTime is the resolved day-level time for 流日: a civil date and its
// day pillar. Computation consumes this value; it must not re-derive the date.
type LiuRiTime struct {
	Standard    time.Time         `json:"-"`
	Location    *time.Location    `json:"-"`
	Longitude   float64           `json:"longitude"`
	Date        tianwen.CivilDate `json:"date"`
	LocalDate   tianwen.CivilDate `json:"local_date"`
	BeijingDate tianwen.CivilDate `json:"beijing_date"`
	RiGan       ganzhi.Gan        `json:"ri_gan"`
	RiZhi       ganzhi.Zhi        `json:"ri_zhi"`
}

// ResolveLiuRiTime resolves the day pillar of a civil date. The day pillar is
// the almanac day pillar of that date; apparent-solar-time boundaries belong to
// 流时, not to the day-level projection.
func ResolveLiuRiTime(date tianwen.CivilDate) (LiuRiTime, error) {
	if err := date.Validate(); err != nil {
		return LiuRiTime{}, err
	}
	zhu := tianwen.RiZhu(date.Time())
	return LiuRiTime{
		Date:        date,
		LocalDate:   date,
		BeijingDate: date,
		RiGan:       zhu.Gan,
		RiZhi:       zhu.Zhi,
	}, nil
}

// ResolveLiuRiTimeAt resolves the day pillar for a standard time. The day
// pillar uses the local LAST date; LocalDate and BeijingDate are echoed for
// audit. The observer's local date is not substituted for the LAST date.
func ResolveLiuRiTimeAt(standard time.Time, loc *time.Location, longitude float64) (LiuRiTime, error) {
	if longitude < -180 || longitude > 180 {
		return LiuRiTime{}, fmt.Errorf("liuri: longitude must be within [-180,180], got %v", longitude)
	}
	if loc == nil {
		loc = time.UTC
	}
	beijing := standard.UTC().In(time.FixedZone("CST", 8*3600))
	last := tianwen.ApparentSolarFromStandard(standard, longitude)
	lastDate := tianwen.CivilDateOf(last)
	pillar := tianwen.RiZhu(lastDate.Time())
	return LiuRiTime{
		Standard:    standard,
		Location:    loc,
		Longitude:   longitude,
		Date:        lastDate,
		LocalDate:   tianwen.StandardDateAt(standard, loc),
		BeijingDate: tianwen.CivilDateOf(beijing),
		RiGan:       pillar.Gan,
		RiZhi:       pillar.Zhi,
	}, nil
}

// NextLiuRiTime returns the next day's resolved time. Standard-time values
// advance by 24h and re-derive the LAST date; date-based values advance by one
// civil date.
func NextLiuRiTime(t LiuRiTime) (LiuRiTime, error) {
	if t.Standard.IsZero() {
		return ResolveLiuRiTime(t.Date.AddDays(1))
	}
	if t.Location == nil {
		return LiuRiTime{}, fmt.Errorf("liuri: location is required for series")
	}
	return ResolveLiuRiTimeAt(t.Standard.Add(24*time.Hour), t.Location, t.Longitude)
}

// LiuNianTime is the resolved year-level time for 流年. Start is the 立春
// civil date of the year; End is the next 立春 date (exclusive).
type LiuNianTime struct {
	LichunYear   int               `json:"lichun_year"`
	NianGan      ganzhi.Gan        `json:"nian_gan"`
	NianZhi      ganzhi.Zhi        `json:"nian_zhi"`
	Start        tianwen.CivilDate `json:"start"`
	End          tianwen.CivilDate `json:"end"`
	StartInstant time.Time         `json:"-"`
	BoundaryDay  bool              `json:"boundary_day"`
}

// ResolveLiuNianTime resolves the 立春 year containing date and its year
// pillar. The 立春 boundary date belongs to the new year.
func ResolveLiuNianTime(date tianwen.CivilDate) (LiuNianTime, error) {
	ly, err := tianwen.LichunYearOf(date)
	if err != nil {
		return LiuNianTime{}, err
	}
	// The year pillar is constant inside the resolved 立春 year; the mid-year
	// The standard time is only an implementation detail of NianZhu.
	pillar := tianwen.NianZhu(tianwen.NewCivilDate(ly.Year, 6, 15).Time())
	return LiuNianTime{
		LichunYear:   ly.Year,
		NianGan:      pillar.Gan,
		NianZhi:      pillar.Zhi,
		Start:        ly.Start,
		End:          ly.End,
		StartInstant: tianwen.SolarTermTime(ly.Year, 315),
		BoundaryDay:  ly.BoundaryDay,
	}, nil
}

// NextLiuNianTime returns the next 立春 year's resolved time.
func NextLiuNianTime(t LiuNianTime) (LiuNianTime, error) {
	return ResolveLiuNianTime(t.End)
}

// LiuYueTime is the resolved month-level time for 流月. JieqiMonth is 1=寅 ..
// 12=丑; Start is the 节 civil date and End is the next 节 date (exclusive).
type LiuYueTime struct {
	Date         tianwen.CivilDate `json:"date"`
	LichunYear   int               `json:"lichun_year"`
	JieqiMonth   int               `json:"jieqi_month"`
	YueGan       ganzhi.Gan        `json:"yue_gan"`
	YueZhi       ganzhi.Zhi        `json:"yue_zhi"`
	Start        tianwen.CivilDate `json:"start"`
	End          tianwen.CivilDate `json:"end"`
	StartInstant time.Time         `json:"-"`
	EndInstant   time.Time         `json:"-"`
	BoundaryDay  bool              `json:"boundary_day"`
}

// ResolveLiuYueTime resolves the solar-term month containing date and its
// month pillar. The 节 boundary date belongs to the new month.
func ResolveLiuYueTime(date tianwen.CivilDate) (LiuYueTime, error) {
	window, err := tianwen.JieqiMonthOf(date)
	if err != nil {
		return LiuYueTime{}, err
	}
	nian, err := ResolveLiuNianTime(date)
	if err != nil {
		return LiuYueTime{}, err
	}
	pillar := tianwen.YueZhuFromYearGan(nian.NianGan, window.Index)
	return LiuYueTime{
		Date:         date,
		LichunYear:   nian.LichunYear,
		JieqiMonth:   window.Index + 1,
		YueGan:       pillar.Gan,
		YueZhi:       pillar.Zhi,
		Start:        window.Start,
		End:          window.End,
		StartInstant: window.StartInstant,
		EndInstant:   window.EndInstant,
		BoundaryDay:  window.BoundaryDay,
	}, nil
}

// NextLiuYueTime returns the next solar-term month's resolved time.
func NextLiuYueTime(t LiuYueTime) (LiuYueTime, error) {
	return ResolveLiuYueTime(t.End)
}

// ResolveLiuNianTimeAt resolves the 立春 year containing a standard time.
// The year boundary is compared as an instant; Start/End are local dates.
func ResolveLiuNianTimeAt(standard time.Time, loc *time.Location) (LiuNianTime, error) {
	if loc == nil {
		loc = time.UTC
	}
	ly, err := tianwen.LichunYearAt(standard)
	if err != nil {
		return LiuNianTime{}, err
	}
	pillar := tianwen.NianZhu(tianwen.NewCivilDate(ly.Year, 6, 15).Time())
	start := tianwen.StandardDateAt(ly.Start, loc)
	end := tianwen.StandardDateAt(ly.End, loc)
	return LiuNianTime{
		LichunYear:   ly.Year,
		NianGan:      pillar.Gan,
		NianZhi:      pillar.Zhi,
		Start:        start,
		End:          end,
		StartInstant: ly.Start,
		BoundaryDay:  tianwen.StandardDateAt(standard, loc) == start,
	}, nil
}

// NextLiuNianTimeAt returns the next 立春 year's resolved time from an
// standard-time resolved period.
func NextLiuNianTimeAt(t LiuNianTime, loc *time.Location) (LiuNianTime, error) {
	return ResolveLiuNianTimeAt(tianwen.SolarTermTime(t.LichunYear+1, 315), loc)
}

// ResolveLiuYueTimeAt resolves the solar-term month containing an absolute
// standard time. The 节 boundary is compared as an instant; dates are local.
func ResolveLiuYueTimeAt(standard time.Time, loc *time.Location) (LiuYueTime, error) {
	if loc == nil {
		loc = time.UTC
	}
	window, err := tianwen.JieqiMonthAt(standard, loc)
	if err != nil {
		return LiuYueTime{}, err
	}
	nian, err := ResolveLiuNianTimeAt(standard, loc)
	if err != nil {
		return LiuYueTime{}, err
	}
	pillar := tianwen.YueZhuFromYearGan(nian.NianGan, window.Index)
	return LiuYueTime{
		Date:         tianwen.StandardDateAt(standard, loc),
		LichunYear:   nian.LichunYear,
		JieqiMonth:   window.Index + 1,
		YueGan:       pillar.Gan,
		YueZhi:       pillar.Zhi,
		Start:        window.Start,
		End:          window.End,
		StartInstant: window.StartInstant,
		EndInstant:   window.EndInstant,
		BoundaryDay:  window.BoundaryDay,
	}, nil
}

// NextLiuYueTimeAt returns the next solar-term month from a standard-time
// resolved period.
func NextLiuYueTimeAt(t LiuYueTime, loc *time.Location) (LiuYueTime, error) {
	return ResolveLiuYueTimeAt(t.EndInstant, loc)
}

// LiuShiTime is the resolved hour-level time for 流时. Standard, Location and
// Longitude are retained so the series can advance one hour at a time.
type LiuShiTime struct {
	Standard  time.Time         `json:"-"`
	Location  *time.Location    `json:"-"`
	Longitude float64           `json:"longitude"`
	LocalDate tianwen.CivilDate `json:"local_date"`
	LastDate  tianwen.CivilDate `json:"last_date"`
	Shichen   ganzhi.Zhi        `json:"shichen"`
	ShiGan    ganzhi.Gan        `json:"shi_gan"`
	ShiZhi    ganzhi.Zhi        `json:"shi_zhi"`
}

// ResolveLiuShiTime derives the LAST hour pillar for a standard time. The hour
// gan follows 五鼠遁; the late 子时 (23:00-24:00 LAST) uses the next day's day
// gan, matching the repository's zi_shi_rule.
func ResolveLiuShiTime(standard time.Time, loc *time.Location, longitude float64) (LiuShiTime, error) {
	if longitude < -180 || longitude > 180 {
		return LiuShiTime{}, fmt.Errorf("liushi: longitude must be within [-180,180], got %v", longitude)
	}
	if loc == nil {
		loc = time.UTC
	}

	last := tianwen.ApparentSolarFromStandard(standard, longitude)
	localDate := tianwen.StandardDateAt(standard, loc)
	lastDate := tianwen.CivilDateOf(last)
	shichen := tianwen.ShichenOf(last)

	dayGanDate := lastDate
	if last.Hour() >= 23 {
		dayGanDate = lastDate.AddDays(1)
	}
	dayPillar := tianwen.RiZhu(dayGanDate.Time())
	shiGan := ganzhi.Gan(((int(dayPillar.Gan)*2 + int(shichen) - 2) % 10))
	if shiGan == 0 {
		shiGan = 10
	}

	return LiuShiTime{
		Standard:  standard,
		Location:  loc,
		Longitude: longitude,
		LocalDate: localDate,
		LastDate:  lastDate,
		Shichen:   shichen,
		ShiGan:    shiGan,
		ShiZhi:    shichen,
	}, nil
}

// NextLiuShiTime advances one 时辰 (two clock hours).
func NextLiuShiTime(t LiuShiTime) (LiuShiTime, error) {
	if t.Location == nil {
		return LiuShiTime{}, fmt.Errorf("liushi: location is required for series")
	}
	return ResolveLiuShiTime(t.Standard.Add(2*time.Hour), t.Location, t.Longitude)
}
