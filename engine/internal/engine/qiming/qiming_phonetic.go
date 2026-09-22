package qiming

import (
	"fmt"
	"strings"
)

// Phonetic holds phonetic analysis.
type Phonetic struct {
	Tones string `json:"tones"`
}

// analyzePhonetic returns the tone string for a character sequence.
func analyzePhonetic(chars []Character) Phonetic {
	if len(chars) == 0 {
		return Phonetic{}
	}
	var tones []string
	for _, c := range chars {
		tones = append(tones, fmt.Sprintf("%d", c.Tone))
	}
	return Phonetic{Tones: strings.Join(tones, "-")}
}
