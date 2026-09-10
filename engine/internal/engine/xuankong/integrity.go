package xuankong

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// calculateChartDigest hashes the canonical JSON encoding of a chart. The
// digest field itself is cleared before marshalling, so a returned chart can be
// re-validated without recursive dependency on its own digest.
func calculateChartDigest(chart Chart) string {
	chart.ChartDigest = ""
	raw, err := json.Marshal(chart)
	if err != nil {
		// Chart consists of JSON-centric engine data; marshal failure is a
		// programming error and must never silently produce a usable digest.
		panic(fmt.Sprintf("xuankong: marshal chart digest: %v", err))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (c *Chart) refreshDigest() {
	c.ChartDigest = calculateChartDigest(*c)
}

// ValidateChartDigest rejects a caller-modified or hand-assembled chart.
func (c Chart) ValidateChartDigest() error {
	if c.ChartDigest == "" {
		return fmt.Errorf("xuankong chart lacks chart_digest")
	}
	if c.ChartDigest != calculateChartDigest(c) {
		return fmt.Errorf("xuankong chart digest mismatch")
	}
	return nil
}
