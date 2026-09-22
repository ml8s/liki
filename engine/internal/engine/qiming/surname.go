package qiming

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed data/surnames.csv
var surnamesCSV []byte

const (
	MatchPinyinExact        = "pinyin_exact"
	MatchRomanizationExact  = "romanization_exact"
	MatchPhoneticClose      = "phonetic_close"
	MatchFallbackBaijiaxing = "fallback_baijiaxing"

	StrategyPhonetic           = "phonetic"
	StrategyBaijiaxingFallback = "baijiaxing_fallback"

	classicBaijiaxingEntries = 504
)

type SurnameRecord struct {
	Surname       string
	Pinyin        string
	PlainPinyin   string
	BaijiaxingIdx int
	Aliases       []string
}

type SurnameCandidate struct {
	Surname         string   `json:"surname"`
	Pinyin          string   `json:"pinyin"`
	Tone            int      `json:"tone"`
	BaijiaxingIndex int      `json:"baijiaxing_index"`
	MatchLevel      string   `json:"match_level"`
	Basis           []string `json:"basis"`
}

type SurnameResult struct {
	SourceSurname string             `json:"source_surname"`
	Strategy      string             `json:"strategy"`
	Candidates    []SurnameCandidate `json:"candidates"`
}

var (
	surnameOnce    sync.Once
	surnameRecords []SurnameRecord
	surnameErr     error
)

func loadSurnames() ([]SurnameRecord, error) {
	surnameOnce.Do(func() {
		reader := csv.NewReader(strings.NewReader(string(surnamesCSV)))
		reader.FieldsPerRecord = 5
		header, err := reader.Read()
		if err != nil {
			surnameErr = err
			return
		}
		columns := make(map[string]int, len(header))
		for i, column := range header {
			columns[column] = i
		}
		seenSurnames := make(map[string]bool)
		seenIndexes := make(map[int]bool)
		line := 1
		for {
			rec, readErr := reader.Read()
			if readErr == io.EOF {
				return
			}
			if readErr != nil {
				surnameErr = readErr
				return
			}
			line++
			surname := strings.TrimSpace(rec[columns["surname"]])
			pinyin := strings.TrimSpace(rec[columns["pinyin"]])
			plain := foldLatin(rec[columns["plain_pinyin"]])
			if surname == "" || pinyin == "" || plain == "" {
				surnameErr = fmt.Errorf("surnames.csv row %d: empty surname or pinyin", line)
				return
			}
			if runeCount := len([]rune(surname)); runeCount < 1 || runeCount > 2 {
				surnameErr = fmt.Errorf("surnames.csv row %d: invalid surname %q", line, surname)
				return
			}
			if strings.Contains(plain, " ") || !isLatinLower(plain) {
				surnameErr = fmt.Errorf("surnames.csv row %d: invalid plain pinyin %q", line, plain)
				return
			}
			if toneFromPinyin(pinyin) == 5 {
				surnameErr = fmt.Errorf("surnames.csv row %d: pinyin %q must contain a tone mark", line, pinyin)
				return
			}
			baijiaxingIndex, err := strconv.Atoi(strings.TrimSpace(rec[columns["baijiaxing_index"]]))
			if err != nil || baijiaxingIndex != line-1 || baijiaxingIndex > classicBaijiaxingEntries {
				surnameErr = fmt.Errorf("surnames.csv row %d: invalid baijiaxing index", line)
				return
			}
			if seenIndexes[baijiaxingIndex] {
				surnameErr = fmt.Errorf("surnames.csv row %d: duplicate baijiaxing index %d", line, baijiaxingIndex)
				return
			}
			seenIndexes[baijiaxingIndex] = true

			var aliases []string
			seenAliases := make(map[string]bool)
			for _, alias := range strings.Split(rec[columns["aliases"]], ",") {
				alias = foldLatin(alias)
				if alias == "" {
					continue
				}
				if !isLatinLower(alias) || strings.Contains(alias, " ") {
					surnameErr = fmt.Errorf("surnames.csv row %d: invalid alias %q", line, alias)
					return
				}
				if alias == plain || seenAliases[alias] {
					surnameErr = fmt.Errorf("surnames.csv row %d: duplicate alias %q", line, alias)
					return
				}
				seenAliases[alias] = true
				aliases = append(aliases, alias)
			}

			if seenSurnames[surname] {
				continue
			}
			seenSurnames[surname] = true

			surnameRecords = append(surnameRecords, SurnameRecord{
				Surname:       surname,
				Pinyin:        pinyin,
				PlainPinyin:   plain,
				BaijiaxingIdx: baijiaxingIndex,
				Aliases:       aliases,
			})
		}
	})
	return surnameRecords, surnameErr
}

func MatchSurnames(sourceSurname string, maxCandidates int) (*SurnameResult, error) {
	if maxCandidates == 0 {
		maxCandidates = 6
	}
	if maxCandidates < 1 || maxCandidates > 12 {
		return nil, fmt.Errorf("max_candidates must be between 1 and 12")
	}
	source := strings.TrimSpace(sourceSurname)
	if source == "" {
		return nil, fmt.Errorf("source_surname is required")
	}
	if len([]rune(source)) > 64 {
		return nil, fmt.Errorf("source_surname must contain at most 64 characters")
	}
	tokens := surnameSourceTokens(source)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("source_surname must contain Latin letters")
	}

	records, err := loadSurnames()
	if err != nil {
		return nil, fmt.Errorf("load surnames: %w", err)
	}

	type surnameMatch struct {
		candidate SurnameCandidate
		tokenSize int
	}
	var matches []surnameMatch
	for _, record := range records {
		if level, basis, tokenSize, ok := matchSurname(record, tokens); ok {
			matches = append(matches, surnameMatch{
				candidate: newSurnameCandidate(record, level, basis),
				tokenSize: tokenSize,
			})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		li, lj := matchOrder(matches[i].candidate.MatchLevel), matchOrder(matches[j].candidate.MatchLevel)
		if li != lj {
			return li < lj
		}
		if matches[i].tokenSize != matches[j].tokenSize {
			return matches[i].tokenSize > matches[j].tokenSize
		}
		return matches[i].candidate.BaijiaxingIndex < matches[j].candidate.BaijiaxingIndex
	})

	result := &SurnameResult{SourceSurname: source, Strategy: StrategyPhonetic}
	if len(matches) != 0 {
		if len(matches) > maxCandidates {
			matches = matches[:maxCandidates]
		}
		result.Candidates = make([]SurnameCandidate, len(matches))
		for i, match := range matches {
			result.Candidates[i] = match.candidate
		}
		return result, nil
	}

	result.Strategy = StrategyBaijiaxingFallback
	for _, record := range records {
		if len(result.Candidates) >= maxCandidates {
			break
		}
		result.Candidates = append(result.Candidates, newSurnameCandidate(
			record,
			MatchFallbackBaijiaxing,
			[]string{"classic_baijiaxing_order"},
		))
	}
	return result, nil
}

func newSurnameCandidate(record SurnameRecord, level string, basis []string) SurnameCandidate {
	return SurnameCandidate{
		Surname:         record.Surname,
		Pinyin:          record.Pinyin,
		Tone:            toneFromPinyin(record.Pinyin),
		BaijiaxingIndex: record.BaijiaxingIdx,
		MatchLevel:      level,
		Basis:           basis,
	}
}

func matchSurname(record SurnameRecord, tokens []string) (string, []string, int, bool) {
	for _, token := range tokens {
		if token == record.PlainPinyin {
			return MatchPinyinExact, []string{"mandarin_pinyin=" + token}, len(token), true
		}
	}
	for _, token := range tokens {
		for _, alias := range record.Aliases {
			if token == alias {
				return MatchRomanizationExact, []string{"romanization=" + alias}, len(token), true
			}
		}
	}
	for _, token := range tokens {
		if latinPhoneticClose(token, record.PlainPinyin) {
			return MatchPhoneticClose, []string{"latin=" + token, "pinyin=" + record.PlainPinyin}, len(token), true
		}
	}
	return "", nil, 0, false
}

func matchOrder(level string) int {
	switch level {
	case MatchPinyinExact:
		return 0
	case MatchRomanizationExact:
		return 1
	case MatchPhoneticClose:
		return 2
	default:
		return 3
	}
}

func surnameSourceTokens(source string) []string {
	raw := strings.ToLower(strings.TrimSpace(source))
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ' ' || r == '-' || r == '\'' || r == '.' || r == ',' || !unicode.IsLetter(r)
	})
	folded := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		token := foldLatin(part)
		// Pinyin lü/nü are conventionally typed lv/nv. Apply this only to a
		// complete token so foreign spellings containing lü keep the generic fold.
		if hasCombiningDiaeresis(part) {
			switch token {
			case "lu":
				token = "lv"
			case "nu":
				token = "nv"
			}
		}
		if !isLatinLower(token) {
			return nil
		}
		folded = append(folded, token)
	}
	tokens := folded
	if len(tokens) > 1 {
		tokens = append(tokens, strings.Join(tokens, ""))
	}
	return tokens
}

func isLatinLower(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func latinPhoneticClose(source, plain string) bool {
	if len(source) < 2 || len(plain) < 2 {
		return false
	}
	if strings.HasPrefix(source, plain) || strings.HasPrefix(plain, source) {
		return true
	}
	if source[0] != plain[0] {
		return false
	}
	return firstVowel(source) == firstVowel(plain)
}

func firstVowel(s string) rune {
	for _, r := range s {
		switch r {
		case 'a', 'e', 'i', 'o', 'u', 'v':
			return r
		}
	}
	return 0
}

func foldLatin(value string) string {
	lowered := strings.ToLower(strings.TrimSpace(value))
	var out strings.Builder
	for _, r := range norm.NFD.String(lowered) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if replacement, ok := nonDecomposingLatin[r]; ok {
			out.WriteString(replacement)
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

var nonDecomposingLatin = map[rune]string{
	'æ': "ae",
	'œ': "oe",
	'ß': "ss",
	'ø': "o",
	'đ': "d",
	'ð': "d",
	'ł': "l",
	'þ': "th",
	'ı': "i",
	'ħ': "h",
	'ŋ': "n",
	'ſ': "s",
}

func hasCombiningDiaeresis(value string) bool {
	for _, r := range norm.NFD.String(value) {
		if r == '\u0308' {
			return true
		}
	}
	return false
}

func toneFromPinyin(pinyin string) int {
	toneByVowel := map[rune]int{
		'ā': 1, 'á': 2, 'ǎ': 3, 'à': 4,
		'ē': 1, 'é': 2, 'ě': 3, 'è': 4,
		'ī': 1, 'í': 2, 'ǐ': 3, 'ì': 4,
		'ō': 1, 'ó': 2, 'ǒ': 3, 'ò': 4,
		'ū': 1, 'ú': 2, 'ǔ': 3, 'ù': 4,
		'ǖ': 1, 'ǘ': 2, 'ǚ': 3, 'ǜ': 4,
	}
	for _, r := range pinyin {
		if tone, ok := toneByVowel[r]; ok {
			return tone
		}
	}
	return 5
}
