package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// ── tianwen.resolve input ──

// TianwenResolveLunarInput is a lunar calendar date (with leap month).
type TianwenResolveLunarInput struct {
	Year  int  `json:"year" jsonschema:"农历年（1900-2100）。"`
	Month int  `json:"month" jsonschema:"农历月（1-12）。"`
	Day   int  `json:"day" jsonschema:"农历日（1-30）。"`
	Leap  bool `json:"leap,omitempty" jsonschema:"是否闰月。"`
}

// TianwenResolveRangeInput is a same-day civil time range.
type TianwenResolveRangeInput struct {
	Start string `json:"start" jsonschema:"区间起点 HH:MM（含）。"`
	End   string `json:"end" jsonschema:"区间终点 HH:MM（不含）。"`
}

// TianwenResolveLongitudeRangeInput is the east-west span of the recorded place.
type TianwenResolveLongitudeRangeInput struct {
	West float64 `json:"west" jsonschema:"地点西界经度（-180..180）。"`
	East float64 `json:"east" jsonschema:"地点东界经度（-180..180）。"`
}

// TianwenResolveInput is a time-fact normalization request. Exactly one of
// gregorian_time / time / shichen / range must be provided.
type TianwenResolveInput struct {
	Calendar  string                    `json:"calendar,omitempty" jsonschema:"壁钟形态必填：solar 或 lunar。"`
	Date      string                    `json:"date,omitempty" jsonschema:"公历日期 YYYY-MM-DD（calendar 为 solar 时必填）。"`
	Lunar     *TianwenResolveLunarInput `json:"lunar,omitempty" jsonschema:"农历日期（calendar 为 lunar 时必填，含闰月标记 leap）。"`
	Timezone  string                    `json:"timezone,omitempty" jsonschema:"记录时间所在时区（IANA）；壁钟/农历形态必填。"`
	Longitude float64                   `json:"longitude" jsonschema:"出生地经度（-180..180）；真太阳时换算必需。"`

	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选，用于边界不确定性。"`

	Instant Instant                   `json:"instant,omitempty" jsonschema:"精确瞬间 RFC3339；canonical 是 UTC 且以 Z 结尾，兼容 numeric offset；与 time/shichen/range 四选一。"`
	Time    string                    `json:"time,omitempty" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS（分钟/秒精度）；与 gregorian_time/shichen/range 四选一。"`
	Shichen string                    `json:"shichen,omitempty" jsonschema:"已知时辰（真太阳时基，2 小时）；与 gregorian_time/time/range 四选一。"`
	Range   *TianwenResolveRangeInput `json:"range,omitempty" jsonschema:"同日民用时间区间；与 gregorian_time/time/shichen 四选一。"`
}

// ── tianwen.resolve input variants (public oneOf) ──

type TianwenUTCForm struct {
	Instant        Instant                            `json:"instant" jsonschema:"精确瞬间 RFC3339；canonical 是 UTC 且以 Z 结尾，兼容 numeric offset。"`
	Timezone       string                             `json:"timezone" jsonschema:"记录时间所在 IANA 时区。"`
	Longitude      float64                            `json:"longitude" jsonschema:"出生地经度（-180..180）；真太阳时换算必需。"`
	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选，用于边界不确定性。"`
}

type TianwenDateTimeForm struct {
	Date           string                             `json:"date" jsonschema:"公历日期 YYYY-MM-DD。"`
	Time           string                             `json:"time" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS（分钟/秒精度）。"`
	Timezone       string                             `json:"timezone" jsonschema:"记录时间所在时区（IANA）。"`
	Longitude      float64                            `json:"longitude" jsonschema:"出生地经度（-180..180）。"`
	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenLunarForm struct {
	Lunar          *TianwenResolveLunarInput          `json:"lunar" jsonschema:"农历日期（含闰月标记 leap）。"`
	Time           string                             `json:"time" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS。"`
	Timezone       string                             `json:"timezone" jsonschema:"记录时间所在时区（IANA）。"`
	Longitude      float64                            `json:"longitude" jsonschema:"出生地经度（-180..180）。"`
	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenShichenForm struct {
	Date           string                             `json:"date" jsonschema:"公历日期 YYYY-MM-DD。"`
	Shichen        string                             `json:"shichen" jsonschema:"已知时辰（真太阳时基，2 小时）。"`
	Timezone       string                             `json:"timezone" jsonschema:"记录时间所在时区（IANA）。"`
	Longitude      float64                            `json:"longitude" jsonschema:"出生地经度（-180..180）。"`
	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenRangeForm struct {
	Date           string                             `json:"date" jsonschema:"公历日期 YYYY-MM-DD。"`
	Range          *TianwenResolveRangeInput          `json:"range" jsonschema:"同日民用时间区间。"`
	Timezone       string                             `json:"timezone" jsonschema:"记录时间所在时区（IANA）。"`
	Longitude      float64                            `json:"longitude" jsonschema:"出生地经度（-180..180）。"`
	LongitudeRange *TianwenResolveLongitudeRangeInput `json:"longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

func tianwenResolveFormsSchema() (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenUTCForm](false, nil)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenDateTimeForm](false, nil)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenLunarForm](false, nil)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenShichenForm](false, nil)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenRangeForm](false, nil)
			return raw, err
		},
	}
	forms := make([]json.RawMessage, 0, len(builders))
	for _, build := range builders {
		raw, err := build()
		if err != nil {
			return nil, err
		}
		forms = append(forms, raw)
	}
	return json.Marshal(map[string]any{"type": "object", "oneOf": forms})
}

// ── tianwen.resolve output ──

// TianwenResolveReasonOutput explains why a candidate exists.
type TianwenResolveReasonOutput struct {
	Domain string `json:"domain" jsonschema:"原因领域：input / 时辰 / 日界 / 交节 / 时制。"`
	Detail string `json:"detail" jsonschema:"领域术语说明（如 输入时间、丑时、立春后、夏令时）。"`
}

// TianwenResolveCandidateOutput is one possible chart time point.
//
// UTC is the canonical instant; Timezone is the IANA zone; LocalTime is the
// civil time with offset for display only. Do not add calendar/true-solar
// fields here: charts derive Lunar/LAST/SolarTerm on demand from the shared L1.
type TianwenResolveCandidateOutput struct {
	Instant  Instant                      `json:"instant" jsonschema:"候选瞬间 UTC（RFC3339，Z）。"`
	Timezone string                       `json:"timezone" jsonschema:"该候选的 IANA 时区。"`
	Reasons  []TianwenResolveReasonOutput `json:"reasons" jsonschema:"候选原因；input 候选恒定存在。"`
}

// TianwenResolveOutput is the candidate set for one time fact. The only
// canonical value is {instant, timezone}; local_time/offset are derived views.
type TianwenResolveOutput struct {
	Candidates []TianwenResolveCandidateOutput `json:"candidates" jsonschema:"候选列表：永远包含一个 input 候选；其余为边界派生。"`
}

// TianwenResolve normalizes a time fact into candidates.
func TianwenResolve(_ context.Context, input TianwenResolveInput) (TianwenResolveOutput, error) {
	in := tianwen.ResolveInput{
		Calendar:  input.Calendar,
		Date:      input.Date,
		Timezone:  input.Timezone,
		Longitude: input.Longitude,
		Time:      input.Time,
		Shichen:   input.Shichen,
	}
	if input.Instant != "" {
		if input.Timezone == "" {
			return TianwenResolveOutput{}, fmt.Errorf("tianwen.resolve: timezone is required")
		}
		zone, err := loadTimezone(input.Timezone)
		if err != nil {
			return TianwenResolveOutput{}, fmt.Errorf("tianwen.resolve: %w", err)
		}
		at, err := parseInstant(string(input.Instant))
		if err != nil {
			return TianwenResolveOutput{}, fmt.Errorf("tianwen.resolve: %w", err)
		}
		local := at.In(zone)
		in.Gregorian = &local
	}
	if input.Lunar != nil {
		in.Lunar = tianwen.LunarDate{
			Year: input.Lunar.Year, Month: input.Lunar.Month,
			Day: input.Lunar.Day, Leap: input.Lunar.Leap,
		}
	}
	if input.Range != nil {
		in.Range = &tianwen.ResolveRange{Start: input.Range.Start, End: input.Range.End}
	}
	if input.LongitudeRange != nil {
		in.LongitudeRange = &tianwen.LongitudeRange{
			West: input.LongitudeRange.West,
			East: input.LongitudeRange.East,
		}
	}

	result, err := tianwen.Resolve(in)
	if err != nil {
		return TianwenResolveOutput{}, fmt.Errorf("tianwen.resolve: %w", err)
	}
	output := TianwenResolveOutput{
		Candidates: make([]TianwenResolveCandidateOutput, 0, len(result.Candidates)),
	}
	for _, candidate := range result.Candidates {
		reasons := make([]TianwenResolveReasonOutput, 0, len(candidate.Reasons))
		for _, reason := range candidate.Reasons {
			reasons = append(reasons, TianwenResolveReasonOutput{
				Domain: reason.Domain,
				Detail: reason.Detail,
			})
		}
		output.Candidates = append(output.Candidates, TianwenResolveCandidateOutput{
			Instant:  Instant(candidate.CivilTime.UTC().UTC().Format(time.RFC3339Nano)),
			Timezone: input.Timezone,
			Reasons:  reasons,
		})
	}
	return output, nil
}
