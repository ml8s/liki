package mcpcontract

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/ziwei"
)

// Longitude is the canonical wire encoding for an event or birth longitude.
// The domain still receives float64; this named type carries schema bounds.
type Longitude float64

// NatalSubject is the calibrated natal identity produced by tianwen
// normalization. StandardTime is local standard time (fixed offset, no DST),
// not true solar time. Longitude is retained for true-solar-time correction.
type NatalSubject struct {
	StandardTime StandardTime `json:"standard_time" jsonschema:"当地标准时间（fixed offset，无 DST），RFC3339 带 offset；来自用户明确事实或 subject_candidates 候选，不得编造。"`
	Longitude    Longitude    `json:"longitude" jsonschema:"地点经度（-180..180）；来自用户确认地点、地点检索结果或既有事实。"`
	Gender       GenderName   `json:"gender" jsonschema:"性别；来自用户明确事实。"`
}

func (subject NatalSubject) standardTime() (time.Time, error) {
	standard, _, err := parseStandardTime(subject.StandardTime)
	return standard, err
}

func (subject NatalSubject) gender() (ganzhi.Gender, error) {
	return ganzhi.ParseGender(string(subject.Gender))
}

func (subject NatalSubject) validate() (time.Time, ganzhi.Gender, error) {
	if subject.Longitude < -180 || subject.Longitude > 180 {
		return time.Time{}, ganzhi.Male, fmt.Errorf("subject.longitude must be within [-180,180], got %v", subject.Longitude)
	}
	gender, err := subject.gender()
	if err != nil {
		return time.Time{}, ganzhi.Male, fmt.Errorf("subject.gender: %w", err)
	}
	standard, err := subject.standardTime()
	if err != nil {
		return time.Time{}, ganzhi.Male, err
	}
	return standard, gender, nil
}

func (subject NatalSubject) baziChart() (bazi.Chart, error) {
	standard, gender, err := subject.validate()
	if err != nil {
		return bazi.Chart{}, err
	}
	return bazi.ComputeChartAt(standard, float64(subject.Longitude), gender), nil
}

func (subject NatalSubject) baziCore() (bazi.Core, error) {
	standard, gender, err := subject.validate()
	if err != nil {
		return bazi.Core{}, err
	}
	return bazi.ComputeCoreAt(standard, float64(subject.Longitude), gender), nil
}

func (subject NatalSubject) ziweiChart() (ziwei.Chart, error) {
	standard, gender, err := subject.validate()
	if err != nil {
		return ziwei.Chart{}, err
	}
	return ziwei.ComputeChartAt(standard, float64(subject.Longitude), gender)
}
