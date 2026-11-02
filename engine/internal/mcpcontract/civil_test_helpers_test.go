package mcpcontract

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func civilResolveInput(t *testing.T, instant, timezone string, longitude float64) TianwenResolveInput {
	t.Helper()
	at, err := parseInstant(instant)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := loadTimezone(timezone)
	if err != nil {
		t.Fatal(err)
	}
	civil := at.In(loc)
	return TianwenResolveInput{
		CalendarKind:   "solar",
		RecordedDate:   civil.Format("2006-01-02"),
		RecordedTime:   civil.Format("15:04:05"),
		RecordTimezone: timezone,
		PlaceLongitude: longitude,
		SubjectGender:  "male",
	}
}

func standardInputFromInstant(t testing.TB, instant, timezone string) StandardTime {
	t.Helper()
	at, err := parseInstant(instant)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := loadTimezone(timezone)
	if err != nil {
		t.Fatal(err)
	}
	return StandardTime(tianwen.StandardLocalTime(at, loc).Format(time.RFC3339Nano))
}

func testNatalSubject(t testing.TB, birthTime StandardTime, gender string) NatalSubject {
	t.Helper()
	return NatalSubject{BirthTime: birthTime, BirthLongitude: 116.4, Gender: GenderName(gender)}
}
