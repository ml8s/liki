package qiming

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
)

//go:embed data/naming_characters.csv
var namingCharactersCSV []byte

var namingCharacterColumns = []string{"char", "pinyin", "radical", "stroke", "wuxing", "tone"}

//go:embed data/kangxi_character_strokes.csv
var kangxiCharacterStrokesCSV []byte

//go:embed data/negative_chars.txt
var negativeCharsTxt []byte

var charByElement map[Wuxing][]Character
var charByRune = make(map[rune]Character)
var kangxiByRune = make(map[rune]kangxiStroke)
var negativeChars = make(map[string]bool)

type kangxiStroke struct {
	Stroke int
	Form   string
}

func init() {
	if err := loadNaming(); err != nil {
		log.Fatalf("qiming: load qiming data: %v", err)
	}
	if err := loadKangxiStrokes(); err != nil {
		log.Fatalf("qiming: load kangxi strokes: %v", err)
	}
	if err := applyKangxiStrokes(); err != nil {
		log.Fatalf("qiming: apply kangxi strokes: %v", err)
	}
}

func loadNaming() error {
	reader := csv.NewReader(bytes.NewReader(namingCharactersCSV))
	reader.FieldsPerRecord = len(namingCharacterColumns)
	header, err := reader.Read()
	if err != nil {
		return err
	}
	columns := make(map[string]int, len(header))
	for i, column := range header {
		columns[column] = i
	}
	for _, column := range namingCharacterColumns {
		if _, ok := columns[column]; !ok {
			return fmt.Errorf("naming_characters.csv: missing column %q", column)
		}
	}
	line := 1
	for {
		rec, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
		line++
		word := rec[columns["char"]]
		if len([]rune(word)) != 1 {
			return fmt.Errorf("naming_characters.csv row %d: invalid character %q", line, word)
		}
		radical := rec[columns["radical"]]
		if radical == "NULL" {
			radical = ""
		}
		stroke, parseErr := strconv.Atoi(rec[columns["stroke"]])
		if parseErr != nil || stroke <= 0 {
			return fmt.Errorf("naming_characters.csv row %d: invalid stroke for %q", line, word)
		}
		tone, parseErr := strconv.Atoi(rec[columns["tone"]])
		if parseErr != nil || tone < 1 || tone > 5 {
			return fmt.Errorf("naming_characters.csv row %d: invalid tone for %q", line, word)
		}
		pinyin := strings.TrimSpace(rec[columns["pinyin"]])
		if pinyin == "" {
			return fmt.Errorf("naming_characters.csv row %d: missing pinyin for %q", line, word)
		}

		elem := wuxingFromChinese(rec[columns["wuxing"]])
		if elem == 0 {
			return fmt.Errorf("naming_characters.csv row %d: missing naming element for %q", line, word)
		}
		charRune := []rune(word)[0]
		if _, exists := charByRune[charRune]; exists {
			return fmt.Errorf("naming_characters.csv row %d: duplicate character %q", line, word)
		}
		// Take the first reading when multiple pinyin values are present.
		if idx := strings.IndexByte(pinyin, ','); idx >= 0 {
			pinyin = pinyin[:idx]
		}
		pinyin = strings.TrimRight(pinyin, "0123456789·")
		character := Character{
			Char:    word,
			Element: elem,
			Stroke:  stroke,
			Radical: radical,
			Pinyin:  pinyin,
			Tone:    tone,
		}
		charByRune[charRune] = character
	}

	// Load characters excluded from naming pools.
	for _, excluded := range strings.Split(strings.TrimSpace(string(negativeCharsTxt)), "\n") {
		excluded = strings.TrimSpace(excluded)
		if excluded == "" {
			continue
		}
		if len([]rune(excluded)) != 1 {
			return fmt.Errorf("negative_chars.txt: invalid character %q", excluded)
		}
		if _, ok := charByRune[[]rune(excluded)[0]]; !ok {
			return fmt.Errorf("negative_chars.txt: unknown character %q", excluded)
		}
		if negativeChars[excluded] {
			return fmt.Errorf("negative_chars.txt: duplicate character %q", excluded)
		}
		negativeChars[excluded] = true
	}

	return nil
}

func loadKangxiStrokes() error {
	r := csv.NewReader(bytes.NewReader(kangxiCharacterStrokesCSV))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("kangxi_character_strokes.csv: empty table")
	}
	columns := map[string]int{}
	for i, column := range records[0] {
		columns[column] = i
	}
	for _, required := range []string{"char", "kangxi_form", "kangxi_stroke"} {
		if _, ok := columns[required]; !ok {
			return fmt.Errorf("kangxi_character_strokes.csv: missing column %q", required)
		}
	}
	for _, rec := range records[1:] {
		if len(rec) <= columns["char"] || len(rec) <= columns["kangxi_form"] || len(rec) <= columns["kangxi_stroke"] {
			return fmt.Errorf("kangxi_character_strokes.csv: missing columns")
		}
		char := rec[columns["char"]]
		if len([]rune(char)) != 1 {
			return fmt.Errorf("kangxi_character_strokes.csv: invalid character %q", char)
		}
		stroke, err := strconv.Atoi(rec[columns["kangxi_stroke"]])
		if err != nil || stroke <= 0 {
			return fmt.Errorf("kangxi_character_strokes.csv: invalid stroke for %q", char)
		}
		form := rec[columns["kangxi_form"]]
		if len([]rune(form)) != 1 {
			return fmt.Errorf("kangxi_character_strokes.csv: invalid Kangxi form for %q", char)
		}
		r := []rune(char)[0]
		if _, exists := kangxiByRune[r]; exists {
			return fmt.Errorf("kangxi_character_strokes.csv: duplicate character %q", char)
		}
		kangxiByRune[r] = kangxiStroke{Stroke: stroke, Form: form}
	}
	return nil
}

func applyKangxiStrokes() error {
	namingChars := make([]Character, 0, len(charByRune))
	for r, character := range charByRune {
		entry, ok := kangxiByRune[r]
		if !ok {
			return fmt.Errorf("character %q has no kangxi stroke", r)
		}
		character.KangxiStroke = entry.Stroke
		character.KangxiForm = entry.Form
		charByRune[r] = character
		if !negativeChars[character.Char] {
			namingChars = append(namingChars, character)
		}
	}
	sort.Slice(namingChars, func(i, j int) bool {
		return namingChars[i].Char < namingChars[j].Char
	})
	charByElement = make(map[Wuxing][]Character)
	for _, character := range namingChars {
		charByElement[character.Element] = append(charByElement[character.Element], character)
	}
	return nil
}
