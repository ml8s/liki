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

//go:embed data/negative_chars.txt
var negativeCharsTxt []byte

var charByElement map[Wuxing][]Character
var charByRune = make(map[rune]Character)
var negativeChars = make(map[string]bool)

func init() {
	if err := loadNaming(); err != nil {
		log.Fatalf("qiming: load qiming data: %v", err)
	}
	buildElementPools()
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

func buildElementPools() {
	namingChars := make([]Character, 0, len(charByRune))
	for _, character := range charByRune {
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
}
