package agent

import (
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// TimePoint is a gregorian time with longitude for solar time correction.
// Used only by the tianwen.time handler.
type TimePoint struct {
	Time      string  `json:"time"`
	Longitude float64 `json:"longitude"`
}

// Timeset converts TimePoint to a tianwen.Timeset.
func (b TimePoint) Timeset() (tianwen.Timeset, error) {
	t, err := time.Parse(time.RFC3339, b.Time)
	if err != nil {
		return tianwen.Timeset{}, fmt.Errorf("invalid time: %w", err)
	}
	_, offset := t.Zone()
	tz := float64(offset) / 3600
	return tianwen.ComputeTimeset(tianwen.GregorianTime(t.In(time.FixedZone("", int(tz*3600)))), b.Longitude), nil
}

// wrapResult wraps engine data in the standard {"_product":"...","data":...} envelope.
func wrapResult[T any](product string, data T) (json.RawMessage, error) {
	return json.Marshal(struct {
		Product string `json:"_product"`
		Data    T      `json:"data"`
	}{Product: product, Data: data})
}

// validateGender checks that the gender is male or female.
func validateGender(g ganzhi.Gender) error {
	if _, err := ganzhi.ParseGender(string(g)); err != nil {
		return fmt.Errorf("gender must be male or female, got %q", g)
	}
	return nil
}
