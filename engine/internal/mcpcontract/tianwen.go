package mcpcontract

import (
	"context"
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// TianwenTimeInput requests true-solar and lunar calendar facts.
type TianwenTimeInput struct {
	Time      RFC3339Time `json:"time" jsonschema:"公历时间（RFC3339），如 1984-02-04T06:00:00+08:00。"`
	Longitude float64     `json:"longitude" jsonschema:"出生地经度，用于真太阳时校正。北京≈116.4。"`
}

// TianwenLunarOutput is the lunar date plus traditional double-hour.
type TianwenLunarOutput struct {
	Year    int    `json:"year" jsonschema:"农历年。"`
	Month   int    `json:"month" jsonschema:"农历月。"`
	Day     int    `json:"day" jsonschema:"农历日。"`
	Leap    bool   `json:"leap" jsonschema:"是否闰月。"`
	Shichen string `json:"shichen" jsonschema:"传统时辰。"`
}

// TianwenHistoricalHintOutput explains a historical time-rule ambiguity.
type TianwenHistoricalHintOutput struct {
	Kind            string      `json:"kind" jsonschema:"历史时制提示类型。"`
	Year            int         `json:"year" jsonschema:"年份。"`
	RecordedTime    RFC3339Time `json:"recorded_time" jsonschema:"记录使用的时间。"`
	StandardTime    RFC3339Time `json:"standard_time" jsonschema:"标准时间解释。"`
	DaylightTime    RFC3339Time `json:"daylight_time" jsonschema:"夏令时解释。"`
	RequiresConfirm bool        `json:"requires_confirm" jsonschema:"调用方是否必须确认口径。"`
}

// TianwenTimeOutput returns Gregorian, true-solar, and lunar representations.
type TianwenTimeOutput struct {
	Gregorian           RFC3339Time                   `json:"gregorian" jsonschema:"公历时间（RFC3339）。"`
	Solar               RFC3339Time                   `json:"solar" jsonschema:"真太阳时（RFC3339）。"`
	Lunar               TianwenLunarOutput            `json:"lunar" jsonschema:"农历信息。"`
	HistoricalTimeHints []TianwenHistoricalHintOutput `json:"historical_time_hints,omitempty" jsonschema:"历史时制歧义提示（如 1986-1991 中国夏令时）。"`
}

// TianwenTime computes deterministic calendar facts for one time/longitude.
func TianwenTime(_ context.Context, input TianwenTimeInput) (TianwenTimeOutput, error) {
	t, err := time.Parse(time.RFC3339, string(input.Time))
	if err != nil {
		return TianwenTimeOutput{}, fmt.Errorf("tianwen.time: invalid time: %w", err)
	}
	_, offset := t.Zone()
	tz := float64(offset) / 3600
	timeset := tianwen.ComputeTimeset(
		tianwen.GregorianTime(t.In(time.FixedZone("", int(tz*3600)))),
		input.Longitude,
	)

	output := TianwenTimeOutput{
		Gregorian: RFC3339Time(time.Time(timeset.Gregorian).Format(time.RFC3339Nano)),
		Solar:     RFC3339Time(time.Time(timeset.Solar).Format(time.RFC3339Nano)),
		Lunar: TianwenLunarOutput{
			Year:    timeset.Lunar.Year,
			Month:   timeset.Lunar.Month,
			Day:     timeset.Lunar.Day,
			Leap:    timeset.Lunar.Leap,
			Shichen: timeset.Lunar.Shichen.String(),
		},
	}
	for _, hint := range timeset.HistoricalTimeHints {
		output.HistoricalTimeHints = append(output.HistoricalTimeHints, TianwenHistoricalHintOutput{
			Kind:            hint.Kind,
			Year:            hint.Year,
			RecordedTime:    RFC3339Time(hint.RecordedTime),
			StandardTime:    RFC3339Time(hint.StandardTime),
			DaylightTime:    RFC3339Time(hint.DaylightTime),
			RequiresConfirm: hint.RequiresConfirm,
		})
	}
	return output, nil
}
