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
