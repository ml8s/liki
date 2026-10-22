package mcpcontract

import (
	"testing"
	"time"
)

func TestParseInstantAcceptsEquivalentRFC3339Forms(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  string
	}{
		{value: "1990-05-20T04:00:00Z", want: "1990-05-20T04:00:00Z"},
		{value: "1990-05-20T12:00:00+08:00", want: "1990-05-20T04:00:00Z"},
		{value: "1990-05-19T23:00:00-05:00", want: "1990-05-20T04:00:00Z"},
		{value: "1990-05-20T04:00:00+00:00", want: "1990-05-20T04:00:00Z"},
	} {
		at, err := parseInstant(tc.value)
		if err != nil {
			t.Fatalf("parseInstant(%q): %v", tc.value, err)
		}
		if got := at.UTC().Format(time.RFC3339); got != tc.want {
			t.Fatalf("parseInstant(%q) = %s, want %s", tc.value, got, tc.want)
		}
	}
}

func TestParseInstantRejectsNaiveTimeAndUnknownOffset(t *testing.T) {
	for _, value := range []string{
		"1990-05-20T04:00:00",
		"1990-05-20 04:00:00Z",
		"1990-05-20T04:00:00+0800",
		"1990-05-20T04:00:00-00:00",
	} {
		if _, err := parseInstant(value); err == nil {
			t.Fatalf("parseInstant(%q) accepted invalid input", value)
		}
	}
}

func TestLoadTimezoneAcceptsIANA(t *testing.T) {
	for _, value := range []string{"UTC", "Asia/Shanghai", "America/New_York"} {
		loc, err := loadTimezone(value)
		if err != nil {
			t.Fatalf("loadTimezone(%q): %v", value, err)
		}
		if loc.String() != value {
			t.Fatalf("loadTimezone(%q).String() = %q", value, loc.String())
		}
	}
}

func TestLoadTimezoneRejectsImplicitAndInvalidZones(t *testing.T) {
	for _, value := range []string{"", "Local", "+08:00", "UTC+8", "CST", "Not/AZone"} {
		if _, err := loadTimezone(value); err == nil {
			t.Fatalf("loadTimezone(%q) accepted invalid timezone", value)
		}
	}
}
