package mcpcontract

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/ziwei"
)

// NatalSubject is the minimal resolved identity needed by natal methods.
// BirthTime is local standard time (fixed offset, no DST), not true solar time.
// NatalSubject is the calibrated natal identity produced by tianwen
// normalization. BirthTime is local standard time (fixed offset, no DST) and
// BirthLongitude is retained for true-solar-time correction.
type NatalSubject struct {
	BirthTime      StandardTime `json:"birth_time" jsonschema:"出生地标准时间（local standard time，无 DST），RFC3339 带 offset。"`
	BirthLongitude float64      `json:"birth_longitude" jsonschema:"出生地经度（-180..180）。"`
	Gender         GenderName   `json:"gender" jsonschema:"性别。"`
}

type SubjectSide struct {
	Subject NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

func (subject NatalSubject) standardTime() (time.Time, error) {
	standard, _, err := parseStandardTime(subject.BirthTime)
	return standard, err
}

func (subject NatalSubject) gender() (ganzhi.Gender, error) {
	return ganzhi.ParseGender(string(subject.Gender))
}

func (subject NatalSubject) validate() (time.Time, ganzhi.Gender, error) {
	if subject.BirthLongitude < -180 || subject.BirthLongitude > 180 {
		return time.Time{}, ganzhi.Male, fmt.Errorf("subject.birth_longitude must be within [-180,180], got %v", subject.BirthLongitude)
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
	return bazi.ComputeChartAt(standard, subject.BirthLongitude, gender), nil
}

func (subject NatalSubject) baziCore() (bazi.Core, error) {
	standard, gender, err := subject.validate()
	if err != nil {
		return bazi.Core{}, err
	}
	return bazi.ComputeCoreAt(standard, subject.BirthLongitude, gender), nil
}

func (subject NatalSubject) ziweiChart() (ziwei.Chart, error) {
	standard, gender, err := subject.validate()
	if err != nil {
		return ziwei.Chart{}, err
	}
	return ziwei.ComputeChartAt(standard, subject.BirthLongitude, gender)
}
