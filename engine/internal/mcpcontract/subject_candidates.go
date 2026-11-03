package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── subject_candidates input ──

// TianwenSubjectCandidatesLunarInput is a lunar calendar date (with leap month).
type TianwenSubjectCandidatesLunarInput struct {
	Year  int  `json:"year" jsonschema:"农历年（1900-2100）。"`
	Month int  `json:"month" jsonschema:"农历月（1-12）。"`
	Day   int  `json:"day" jsonschema:"农历日（1-30）。"`
	Leap  bool `json:"leap,omitempty" jsonschema:"是否闰月。"`
}

// TianwenSubjectCandidatesRangeInput is a same-day civil time range.
type TianwenSubjectCandidatesRangeInput struct {
	Start string `json:"start" jsonschema:"区间起点 HH:MM（含）。"`
	End   string `json:"end" jsonschema:"区间终点 HH:MM（不含）。"`
}

// TianwenSubjectCandidatesLongitudeRangeInput is the east-west span of the recorded place.
type TianwenSubjectCandidatesLongitudeRangeInput struct {
	West float64 `json:"west" jsonschema:"地点西界经度（-180..180）。"`
	East float64 `json:"east" jsonschema:"地点东界经度（-180..180）。"`
}

// TianwenSubjectCandidatesInput is a time-fact normalization request. Exactly one of
// recorded_time / recorded_shichen / recorded_time_range must be provided; recorded_date or recorded_lunar supplies the day.
type TianwenSubjectCandidatesInput struct {
	CalendarKind   string                              `json:"calendar_kind,omitempty" jsonschema:"壁钟形态必填：solar 或 lunar。"`
	RecordedDate   string                              `json:"recorded_date,omitempty" jsonschema:"公历日期 YYYY-MM-DD（calendar 为 solar 时必填）。"`
	RecordedLunar  *TianwenSubjectCandidatesLunarInput `json:"recorded_lunar,omitempty" jsonschema:"农历日期（calendar 为 lunar 时必填，含闰月标记 leap）。"`
	RecordTimezone string                              `json:"record_timezone,omitempty" jsonschema:"记录时间所在时区（IANA）；壁钟/农历形态必填。"`
	PlaceLongitude float64                             `json:"place_longitude" jsonschema:"地点经度（-180..180）；真太阳时换算必需。"`
	SubjectGender  GenderName                          `json:"subject_gender" jsonschema:"性别。"`

	PlaceLongitudeRange *TianwenSubjectCandidatesLongitudeRangeInput `json:"place_longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选，用于边界不确定性。"`

	RecordedTime      string                              `json:"recorded_time,omitempty" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS（分钟/秒精度）；与 recorded_shichen/recorded_time_range 二选一；recorded_date/recorded_lunar 提供日期。"`
	RecordedShichen   string                              `json:"recorded_shichen,omitempty" jsonschema:"已知时辰（真太阳时基，2 小时）；与 recorded_time/recorded_time_range 二选一。"`
	RecordedTimeRange *TianwenSubjectCandidatesRangeInput `json:"recorded_time_range,omitempty" jsonschema:"同日民用时间区间；与 recorded_time/recorded_shichen 二选一。"`
}

// ── subject_candidates input variants (public oneOf) ──

type TianwenDateTimeForm struct {
	RecordedDate        string                                       `json:"recorded_date" jsonschema:"公历日期 YYYY-MM-DD。"`
	RecordedTime        string                                       `json:"recorded_time" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS（分钟/秒精度）。"`
	RecordTimezone      string                                       `json:"record_timezone" jsonschema:"记录时间所在时区（IANA）。"`
	PlaceLongitude      float64                                      `json:"place_longitude" jsonschema:"地点经度（-180..180）。"`
	SubjectGender       GenderName                                   `json:"subject_gender" jsonschema:"性别。"`
	PlaceLongitudeRange *TianwenSubjectCandidatesLongitudeRangeInput `json:"place_longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenLunarForm struct {
	RecordedLunar       *TianwenSubjectCandidatesLunarInput          `json:"recorded_lunar" jsonschema:"农历日期（含闰月标记 leap）。"`
	RecordedTime        string                                       `json:"recorded_time" jsonschema:"当地壁钟时间 HH:MM 或 HH:MM:SS。"`
	RecordTimezone      string                                       `json:"record_timezone" jsonschema:"记录时间所在时区（IANA）。"`
	PlaceLongitude      float64                                      `json:"place_longitude" jsonschema:"地点经度（-180..180）。"`
	SubjectGender       GenderName                                   `json:"subject_gender" jsonschema:"性别。"`
	PlaceLongitudeRange *TianwenSubjectCandidatesLongitudeRangeInput `json:"place_longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenShichenForm struct {
	RecordedDate        string                                       `json:"recorded_date" jsonschema:"公历日期 YYYY-MM-DD。"`
	RecordedShichen     string                                       `json:"recorded_shichen" jsonschema:"已知时辰（真太阳时基，2 小时）。"`
	RecordTimezone      string                                       `json:"record_timezone" jsonschema:"记录时间所在时区（IANA）。"`
	PlaceLongitude      float64                                      `json:"place_longitude" jsonschema:"地点经度（-180..180）。"`
	SubjectGender       GenderName                                   `json:"subject_gender" jsonschema:"性别。"`
	PlaceLongitudeRange *TianwenSubjectCandidatesLongitudeRangeInput `json:"place_longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

type TianwenRangeForm struct {
	RecordedDate        string                                       `json:"recorded_date" jsonschema:"公历日期 YYYY-MM-DD。"`
	RecordedTimeRange   *TianwenSubjectCandidatesRangeInput          `json:"recorded_time_range" jsonschema:"同日民用时间区间。"`
	RecordTimezone      string                                       `json:"record_timezone" jsonschema:"记录时间所在时区（IANA）。"`
	PlaceLongitude      float64                                      `json:"place_longitude" jsonschema:"地点经度（-180..180）。"`
	SubjectGender       GenderName                                   `json:"subject_gender" jsonschema:"性别。"`
	PlaceLongitudeRange *TianwenSubjectCandidatesLongitudeRangeInput `json:"place_longitude_range,omitempty" jsonschema:"地点行政范围（东西经度）；可选。"`
}

func tianwenResolveFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenDateTimeForm](options)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenLunarForm](options)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenShichenForm](options)
			return raw, err
		},
		func() (json.RawMessage, error) {
			_, raw, err := inputSchemaFor[TianwenRangeForm](options)
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

// ── subject_candidates output ──

// TianwenSubjectCandidatesReasonOutput explains why a candidate exists.
type TianwenSubjectCandidatesReasonOutput struct {
	ReasonKind  string `json:"reason_kind" jsonschema:"原因领域：input / 时辰 / 日界 / 交节 / 时制。"`
	Explanation string `json:"explanation" jsonschema:"领域术语说明（如 输入时间、丑时、立春后、夏令时）。"`
}

// TianwenSubjectCandidateOutput is one possible natal subject.
//
// BirthTime is local standard time (fixed offset, no DST), not true solar time.
// BirthLongitude remains explicit for domain-specific solar corrections.
type TianwenSubjectCandidateOutput struct {
	Subject NatalSubject                           `json:"subject" jsonschema:"候选出生主体。"`
	Reasons []TianwenSubjectCandidatesReasonOutput `json:"reasons" jsonschema:"候选原因；input 候选恒定存在。"`
}

// TianwenSubjectCandidatesOutput is the candidate set for one natal subject.
type TianwenSubjectCandidatesOutput struct {
	Candidates []TianwenSubjectCandidateOutput `json:"candidates" jsonschema:"候选列表：永远包含一个 input 候选；其余为边界派生。"`
}

// TianwenResolve normalizes a time fact into candidates.
func TianwenSubjectCandidates(_ context.Context, input TianwenSubjectCandidatesInput) (TianwenSubjectCandidatesOutput, error) {
	in := tianwen.ResolveInput{
		Calendar:  input.CalendarKind,
		Date:      input.RecordedDate,
		Timezone:  input.RecordTimezone,
		Longitude: input.PlaceLongitude,
		Time:      input.RecordedTime,
		Shichen:   input.RecordedShichen,
	}
	if input.RecordedLunar != nil {
		in.Lunar = tianwen.LunarDate{
			Year: input.RecordedLunar.Year, Month: input.RecordedLunar.Month,
			Day: input.RecordedLunar.Day, Leap: input.RecordedLunar.Leap,
		}
	}
	if input.RecordedTimeRange != nil {
		in.Range = &tianwen.ResolveRange{Start: input.RecordedTimeRange.Start, End: input.RecordedTimeRange.End}
	}
	if input.PlaceLongitudeRange != nil {
		in.LongitudeRange = &tianwen.LongitudeRange{
			West: input.PlaceLongitudeRange.West,
			East: input.PlaceLongitudeRange.East,
		}
	}

	if _, err := ganzhi.ParseGender(string(input.SubjectGender)); err != nil {
		return TianwenSubjectCandidatesOutput{}, fmt.Errorf("tianwen.subject_candidates: gender: %w", err)
	}

	result, err := tianwen.ResolveStandard(in)
	if err != nil {
		return TianwenSubjectCandidatesOutput{}, fmt.Errorf("tianwen.subject_candidates: %w", err)
	}
	output := TianwenSubjectCandidatesOutput{
		Candidates: make([]TianwenSubjectCandidateOutput, 0, len(result.Candidates)),
	}
	zone, err := loadTimezone(input.RecordTimezone)
	if err != nil {
		return TianwenSubjectCandidatesOutput{}, fmt.Errorf("tianwen.subject_candidates: %w", err)
	}
	for _, candidate := range result.Candidates {
		reasons := make([]TianwenSubjectCandidatesReasonOutput, 0, len(candidate.Reasons))
		for _, reason := range candidate.Reasons {
			reasons = append(reasons, TianwenSubjectCandidatesReasonOutput{
				ReasonKind:  reason.Domain,
				Explanation: reason.Detail,
			})
		}
		output.Candidates = append(output.Candidates, TianwenSubjectCandidateOutput{
			Subject: NatalSubject{
				BirthTime:      StandardTime(standardTimeString(candidate.StandardTime, zone)),
				BirthLongitude: input.PlaceLongitude,
				Gender:         input.SubjectGender,
			},
			Reasons: reasons,
		})
	}
	return output, nil
}
