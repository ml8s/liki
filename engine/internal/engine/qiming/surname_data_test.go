package qiming

import (
	"encoding/csv"
	"strconv"
	"strings"
	"testing"
)

func TestSurnameRuntimeTableIsCompleteAndUnique(t *testing.T) {
	records, err := loadSurnames()
	if err != nil {
		t.Fatalf("load surnames: %v", err)
	}
	if len(records) != 502 {
		t.Fatalf("runtime surname records = %d, want 502", len(records))
	}
	if records[0].BaijiaxingIdx != 1 || records[len(records)-1].BaijiaxingIdx != 504 {
		t.Fatalf("classic index bounds = %d..%d, want 1..504",
			records[0].BaijiaxingIdx, records[len(records)-1].BaijiaxingIdx)
	}
	seenSurnames := make(map[string]bool, len(records))
	seenIndexes := make(map[int]bool, len(records))
	compound := 0
	for _, record := range records {
		if seenSurnames[record.Surname] || seenIndexes[record.BaijiaxingIdx] {
			t.Fatalf("duplicate runtime surname %q / index %d", record.Surname, record.BaijiaxingIdx)
		}
		seenSurnames[record.Surname] = true
		seenIndexes[record.BaijiaxingIdx] = true
		if len([]rune(record.Surname)) > 1 {
			compound++
		}
	}
	if compound != 60 {
		t.Fatalf("compound surnames = %d, want 60", compound)
	}
}

func TestSurnameSourceTableKeepsClassicPositions(t *testing.T) {
	reader := csv.NewReader(strings.NewReader(string(surnamesCSV)))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read surname source: %v", err)
	}
	if len(rows) != 505 {
		t.Fatalf("source rows including header = %d, want 505", len(rows))
	}
	seenSurnames := make(map[string]int)
	compound := 0
	for number, row := range rows[1:] {
		if row[3] != strconv.Itoa(number+1) {
			t.Fatalf("source row %d has classic index %q, want %d", number+1, row[3], number+1)
		}
		if len([]rune(row[0])) > 1 {
			compound++
		}
		seenSurnames[row[0]]++
	}
	if compound != 60 {
		t.Fatalf("source compound surnames = %d, want 60", compound)
	}
	if len(seenSurnames) != 502 {
		t.Fatalf("source unique surnames = %d, want 502", len(seenSurnames))
	}
	for surname, count := range seenSurnames {
		if count != 1 && surname != "郁" && surname != "后" {
			t.Fatalf("unexpected repeated surname %q occurs %d times", surname, count)
		}
	}
}

func TestFoldLatinSupportsAllMandarinToneMarks(t *testing.T) {
	tests := map[string]string{
		"Zhāng":  "zhang",
		"Zháng":  "zhang",
		"Zhǎng":  "zhang",
		"Zhàng":  "zhang",
		"Lǖ":     "lu",
		"Lǘ":     "lu",
		"Lǚ":     "lu",
		"Lǜ":     "lu",
		"Lü":     "lu",
		"Nü":     "nu",
		"Müller": "muller",
		"Šilhan": "silhan",
		"Dvořák": "dvorak",
		"Østrom": "ostrom",
		"Łódź":   "lodz",
		"Straße": "strasse",
		"Þór":    "thor",
		"Ærø":    "aero",
	}
	for source, want := range tests {
		if got := foldLatin(source); got != want {
			t.Errorf("foldLatin(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestSurnameSourceTokensApplyPinyinUmlautConventionOnlyToCompleteToken(t *testing.T) {
	tests := map[string][]string{
		"Lü":     {"lv"},
		"Lǚ":     {"lv"},
		"Nü":     {"nv"},
		"Nǚ":     {"nv"},
		"Lühl":   {"luhl"},
		"Müller": {"muller"},
		"Šilhan": {"silhan"},
		"Dvořák": {"dvorak"},
		"Østrom": {"ostrom"},
		"Łódź":   {"lodz"},
		"Straße": {"strasse"},
	}
	for source, want := range tests {
		got := surnameSourceTokens(source)
		if len(got) != len(want) {
			t.Errorf("surnameSourceTokens(%q) = %v, want %v", source, got, want)
			continue
		}
		for index := range want {
			if got[index] != want[index] {
				t.Errorf("surnameSourceTokens(%q) = %v, want %v", source, got, want)
				break
			}
		}
	}
}

func TestMatchSurnamesAcceptsAccentedLatinSurnames(t *testing.T) {
	for _, source := range []string{
		"Šilhan", "Dvořák", "Østrom", "Łódź", "Straße", "Þór", "Ærø",
	} {
		result, err := MatchSurnames(source, 3)
		if err != nil {
			t.Errorf("MatchSurnames(%q) error = %v", source, err)
			continue
		}
		if result.Strategy == "" || len(result.Candidates) == 0 {
			t.Errorf("MatchSurnames(%q) = %+v, want candidates", source, result)
		}
	}
}

func TestVietnameseRomanizationsReachCuratedSurnames(t *testing.T) {
	tests := map[string]string{
		"Nguyễn": "阮",
		"Phạm":   "范",
		"Trần":   "陈",
		"Lê":     "黎",
		"Lý":     "李",
		"Hồ":     "胡",
		"Đặng":   "邓",
		"Ngô":    "吴",
		"Dương":  "杨",
		"Đỗ":     "杜",
		"Bùi":    "裴",
		"Tạ":     "谢",
		"Tôn":    "孙",
		"Hoàng":  "黄",
		"Huỳnh":  "黄",
		"Phan":   "潘",
		"Quách":  "郭",
		"Vũ":     "武",
		"Võ":     "武",
	}
	for source, surname := range tests {
		result, err := MatchSurnames(source, 12)
		if err != nil {
			t.Errorf("MatchSurnames(%q) error = %v", source, err)
			continue
		}
		found := false
		for _, candidate := range result.Candidates {
			if candidate.Surname != surname {
				continue
			}
			found = true
			expectedBasis := "romanization=" + foldLatin(source)
			if candidate.MatchLevel != MatchRomanizationExact || !containsBasis(candidate.Basis, expectedBasis) {
				t.Errorf(
					"MatchSurnames(%q) %s candidate = %+v, want basis %q",
					source, surname, candidate, expectedBasis,
				)
			}
			break
		}
		if !found {
			t.Errorf("MatchSurnames(%q) candidates = %+v, want %s", source, result.Candidates, surname)
		}
	}
}

func containsBasis(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
