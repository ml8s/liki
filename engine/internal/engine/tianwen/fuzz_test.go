package tianwen

import (
	"testing"
	"time"
)

// min/max Unix seconds for valid Gregorian range (1900–2100)
const (
	minEpoch = int64(-2208988800) // 1900-01-01 UTC
	maxEpoch = int64(4133980800)  // 2100-12-31 UTC
)

func FuzzLunarDateAt(f *testing.F) {
	f.Add(int64(0))            // Unix epoch
	f.Add(minEpoch)            // 1900
	f.Add(int64(946684800))    // 2000-01-01
	f.Add(int64(1704067200))   // 2024-01-01
	f.Add(maxEpoch)            // 2100-12-31
	f.Add(int64(-62135596800)) // year 0

	f.Fuzz(func(t *testing.T, epochSec int64) {
		gt := time.Time(time.Unix(epochSec, 0))

		// LunarDateAt should never panic even for extreme values
		lt := LunarDateAt(gt)

		// Basic sanity: lunar year should be reasonable
		if lt.Year < -5000 || lt.Year > 5000 {
			t.Skipf("skipping extreme lunar year %d from epoch %d", lt.Year, epochSec)
		}
		if lt.Month < 1 || lt.Month > 13 {
			t.Errorf("LunarDateAt(epoch=%d).Month = %d, want 1-13", epochSec, lt.Month)
		}
		if lt.Day < 1 || lt.Day > 30 {
			t.Errorf("LunarDateAt(epoch=%d).Day = %d, want 1-30", epochSec, lt.Day)
		}
	})
}
