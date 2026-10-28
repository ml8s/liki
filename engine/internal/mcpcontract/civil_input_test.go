package mcpcontract

import (
	"strings"
	"testing"
	"time"
)

func TestParseStandardTimeAcceptsKnownRFC3339Offsets(t *testing.T) {
	cases := []struct {
		value       string
		wantInstant time.Time
		wantOffset  int
	}{
		{
			value:       "1990-05-20T04:00:00Z",
			wantInstant: time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC),
			wantOffset:  0,
		},
		{
			value:       "1990-05-20T04:00:00+00:00",
			wantInstant: time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC),
			wantOffset:  0,
		},
		{
			value:       "1990-05-20T12:00:00+08:00",
			wantInstant: time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC),
			wantOffset:  8 * 3600,
		},
		{
			value:       "1990-05-19T23:00:00-05:00",
			wantInstant: time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC),
			wantOffset:  -5 * 3600,
		},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			at, _, err := parseStandardTime(StandardTime(tc.value))
			if err != nil {
				t.Fatal(err)
			}
			if !at.Equal(tc.wantInstant) {
				t.Fatalf("instant = %s, want %s", at.UTC(), tc.wantInstant)
			}
			if _, offset := at.Zone(); offset != tc.wantOffset {
				t.Fatalf("offset = %d, want %d", offset, tc.wantOffset)
			}
		})
	}
}

func TestParseStandardTimeRejectsUnknownOffsetAndNaiveTime(t *testing.T) {
	for _, value := range []string{
		"1990-05-20T04:00:00-00:00",
		"1990-05-20T04:00:00",
		"1990-05-20 04:00:00Z",
	} {
		t.Run(value, func(t *testing.T) {
			_, _, err := parseStandardTime(StandardTime(value))
			if err == nil {
				t.Fatal("expected error")
			}
			if value == "1990-05-20T04:00:00-00:00" && !strings.Contains(err.Error(), "unknown offset") {
				t.Fatalf("error %q should identify unknown offset", err)
			}
		})
	}
}

func TestStandardTimeOffsetSelectsLocalStandardDate(t *testing.T) {
	utc := time.Date(2026, 2, 3, 16, 30, 0, 0, time.UTC)
	plus8, _, err := parseStandardTime("2026-02-04T00:30:00+08:00")
	if err != nil {
		t.Fatal(err)
	}
	if !plus8.Equal(utc) {
		t.Fatalf("+08:00 instant = %s, want %s", plus8.UTC(), utc)
	}
	if got := plus8.Format("2006-01-02"); got != "2026-02-04" {
		t.Fatalf("+08:00 local date = %s, want 2026-02-04", got)
	}

	zulu, _, err := parseStandardTime("2026-02-03T16:30:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if !zulu.Equal(utc) {
		t.Fatalf("Z instant = %s, want %s", zulu.UTC(), utc)
	}
	if got := zulu.Format("2006-01-02"); got != "2026-02-03" {
		t.Fatalf("Z local date = %s, want 2026-02-03", got)
	}
}
