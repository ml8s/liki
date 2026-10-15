package agent

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

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
