package qiming

import (
	"testing"
)

func TestPickChars_DoubleName(t *testing.T) {
	result, err := PickChars("木", "火", 2)
	if err != nil {
		t.Fatalf("PickChars() error = %v", err)
	}
	if result.Wuxing1 != "木" || result.Wuxing2 != "火" {
		t.Fatalf("PickChars() = %+v", result)
	}
	if len(result.Pools) != 2 {
		t.Fatalf("pools = %+v, want two pools", result.Pools)
	}
	if result.Pools[0].Slot != "first" || len(result.Pools[0].Chars) == 0 {
		t.Fatalf("first pool = %+v", result.Pools[0])
	}
	if result.Pools[1].Slot != "second" || len(result.Pools[1].Chars) == 0 {
		t.Fatalf("second pool = %+v", result.Pools[1])
	}
	for _, pool := range result.Pools {
		seen := make(map[string]bool)
		for i, char := range pool.Chars {
			if len([]rune(char.Char)) != 1 {
				t.Fatalf("pool contains non-single character %q", char.Char)
			}
			switch char.Frequency {
			case "common", "standard", "rare":
			default:
				t.Fatalf("character %q has invalid frequency %q", char.Char, char.Frequency)
			}
			if seen[char.Char] {
				t.Fatalf("pool contains duplicate character %q", char.Char)
			}
			seen[char.Char] = true
			if i != 0 && pool.Chars[i-1].Char >= char.Char {
				t.Fatalf("pool is not sorted: %q >= %q", pool.Chars[i-1].Char, char.Char)
			}
		}
	}
	for _, char := range result.Pools[0].Chars {
		character := LookupChar(char.Char)
		if character == nil || character.Element.String() != "木" {
			t.Fatalf("first pool character %q has unexpected element", char.Char)
		}
	}
	for _, char := range result.Pools[1].Chars {
		character := LookupChar(char.Char)
		if character == nil || character.Element.String() != "火" {
			t.Fatalf("second pool character %q has unexpected element", char.Char)
		}
	}
}

func TestPickChars_SingleName(t *testing.T) {
	result, err := PickChars("木", "", 1)
	if err != nil {
		t.Fatalf("PickChars() error = %v", err)
	}
	if result.Wuxing2 != "" || len(result.Pools) != 1 || result.Pools[0].Slot != "first" {
		t.Fatalf("PickChars() = %+v, want one first pool", result)
	}
}

func TestPickChars_DefaultSecondElement(t *testing.T) {
	result, err := PickChars("木", "", 2)
	if err != nil {
		t.Fatalf("PickChars() error = %v", err)
	}
	if result.Wuxing2 != "" || len(result.Pools) != 2 {
		t.Fatalf("PickChars() = %+v, want two wood pools", result)
	}
}

func TestPickChars_ExcludesNegativeCharacters(t *testing.T) {
	result, err := PickChars("水", "", 1)
	if err != nil {
		t.Fatalf("PickChars() error = %v", err)
	}
	for _, char := range result.Pools[0].Chars {
		if char.Char == "病" {
			t.Fatal("pool contains excluded character 病")
		}
	}
}

func TestPickChars_Errors(t *testing.T) {
	tests := []struct {
		name    string
		wuxing1 string
		wuxing2 string
		count   int
	}{
		{"invalid wuxing1", "土土", "火", 2},
		{"invalid wuxing2", "木", "土土", 2},
		{"invalid count", "木", "火", 3},
		{"unexpected wuxing2", "木", "火", 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := PickChars(test.wuxing1, test.wuxing2, test.count); err == nil {
				t.Fatal("PickChars() error = nil, want error")
			}
		})
	}
}
