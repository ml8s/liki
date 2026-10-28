package tianwen

import (
	"fmt"
	"time"
)

// NextApparentSolarDay returns a representative instant in the apparent-solar
// day after standard. LAST days are not exactly 24 standard hours long because
// EOT changes between successive solar midnights.
func NextApparentSolarDay(standard time.Time, longitude float64) (time.Time, error) {
	return nextApparentSolarCell(standard, longitude, BoundarySolarDay, 50*time.Hour)
}

// NextApparentSolarShichen returns a representative instant in the shichen after
// standard. LAST shichen are close to, but not exactly, two standard hours long.
func NextApparentSolarShichen(standard time.Time, longitude float64) (time.Time, error) {
	return nextApparentSolarCell(standard, longitude, BoundaryShichen, 5*time.Hour)
}

func nextApparentSolarCell(
	standard time.Time,
	longitude float64,
	kind BoundaryKind,
	searchDuration time.Duration,
) (time.Time, error) {
	if longitude < -180 || longitude > 180 {
		return time.Time{}, fmt.Errorf(
			"apparent solar cell: longitude must be within [-180,180], got %v", longitude)
	}

	_, offset := standard.Zone()
	zone := time.FixedZone("standard", offset)
	boundaries := BoundariesInWindow(
		Window{Start: standard, End: standard.Add(searchDuration)},
		[]BoundaryKind{kind},
		BoundaryOptions{Longitude: longitude, Timezone: zone},
	)
	if len(boundaries) < 2 {
		return time.Time{}, fmt.Errorf(
			"apparent solar cell: need two %s boundaries after %s", kind, standard)
	}

	// The first boundary opens the next cell. Use the midpoint of that cell, not
	// the boundary itself, so second-level rounding cannot select the old cell.
	start := boundaries[0].At
	end := boundaries[1].At
	return start.Add(end.Sub(start) / 2), nil
}
