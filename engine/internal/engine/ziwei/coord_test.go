package ziwei

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSortFlowStarsUsesCanonicalDisplayOrder(t *testing.T) {
	got := sortFlowStars([]string{"月曲", "月喜", "月马", "月钺", "月魁", "月陀", "月羊", "月禄", "月鸾", "月昌"})
	want := []string{"月禄", "月羊", "月陀", "月魁", "月钺", "月马", "月鸾", "月喜", "月昌", "月曲"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildFlowPalacesHasDeterministicStarOrder(t *testing.T) {
	// The input is intentionally unsorted; map iteration must not leak into output.
	stars := map[int][]string{
		7: {"月曲", "月禄", "月昌", "月羊"},
		8: {"流喜", "流禄", "流陀"},
	}

	first := buildFlowPalaces(0, stars)
	second := buildFlowPalaces(0, stars)
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("flow palace JSON is not deterministic")
	}

	want := []string{"月禄", "月羊", "月昌", "月曲"}
	if !reflect.DeepEqual(first[7].Stars, want) {
		t.Fatalf("stars = %v, want %v", first[7].Stars, want)
	}
	want = []string{"流禄", "流陀", "流喜"}
	if !reflect.DeepEqual(first[8].Stars, want) {
		t.Fatalf("stars = %v, want %v", first[8].Stars, want)
	}
}

func TestBuildFlowPalacesEmitsEmptyStarArrays(t *testing.T) {
	palaces := buildFlowPalaces(0, map[int][]string{})
	for _, palace := range palaces {
		if palace.Stars == nil || !reflect.DeepEqual(palace.Stars, []string{}) {
			t.Fatalf("palace %s stars = %#v, want non-nil empty array", palace.Name, palace.Stars)
		}
	}

	data, err := json.Marshal(palaces)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("null")) {
		t.Fatalf("flow palace JSON contains null: %s", data)
	}
}
