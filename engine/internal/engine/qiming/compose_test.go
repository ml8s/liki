package qiming

import (
	"strings"
	"testing"
)

func TestComposeNames_DoubleName(t *testing.T) {
	result, err := ComposeNames(ComposeRequest{
		First:    []string{"林"},
		Second:   []string{"炎"},
		MaxNames: 10,
	})
	if err != nil {
		t.Fatalf("ComposeNames() error = %v", err)
	}
	if result.TotalPossible != 1 || len(result.Names) != 1 {
		t.Fatalf("ComposeNames() = %+v", result)
	}
	if result.Names[0] != "林炎" {
		t.Fatalf("composed name = %q, want 林炎", result.Names[0])
	}
}

func TestComposeNames_SingleName(t *testing.T) {
	result, err := ComposeNames(ComposeRequest{First: []string{"林", "桐"}, MaxNames: 10})
	if err != nil {
		t.Fatalf("ComposeNames() error = %v", err)
	}
	if result.TotalPossible != 2 || len(result.Names) != 2 {
		t.Fatalf("ComposeNames() = %+v", result)
	}
	if result.Names[0] != "林" || result.Names[1] != "桐" {
		t.Fatalf("names = %+v", result.Names)
	}
}

func TestComposeNames_Truncates(t *testing.T) {
	result, err := ComposeNames(ComposeRequest{
		First:    []string{"林", "桐"},
		Second:   []string{"炎", "煜"},
		MaxNames: 3,
	})
	if err != nil {
		t.Fatalf("ComposeNames() error = %v", err)
	}
	if result.TotalPossible != 4 || len(result.Names) != 3 {
		t.Fatalf("ComposeNames() = %+v", result)
	}
}

func TestComposeNames_DefaultMaxNames(t *testing.T) {
	pool, err := PickChars("木", "", 1)
	if err != nil {
		t.Fatalf("PickChars() error = %v", err)
	}
	if len(pool.Pools[0].Chars) < 101 {
		t.Fatal("wood pool does not contain enough characters")
	}
	result, err := ComposeNames(ComposeRequest{
		First: pool.Pools[0].Chars[:101],
	})
	if err != nil {
		t.Fatalf("ComposeNames() error = %v", err)
	}
	if result.TotalPossible != 101 || len(result.Names) != 100 {
		t.Fatalf("ComposeNames() = %+v, want 100 of 101 names", result)
	}
}

func TestComposeNames_Errors(t *testing.T) {
	tests := []struct {
		name    string
		request ComposeRequest
		want    string
	}{
		{"unknown char", ComposeRequest{First: []string{"龍"}}, "not found"},
		{"negative char", ComposeRequest{First: []string{"病"}}, "excluded"},
		{"duplicate first", ComposeRequest{First: []string{"林", "林"}}, "duplicate"},
		{"duplicate second", ComposeRequest{First: []string{"林"}, Second: []string{"炎", "炎"}}, "duplicate"},
		{"invalid max names", ComposeRequest{First: []string{"林"}, MaxNames: 1001}, "max_names"},
		{"empty first", ComposeRequest{}, "first must contain"},
		{"multi-rune character", ComposeRequest{First: []string{"林炎"}}, "single character"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ComposeNames(test.request)
			if err == nil {
				t.Fatal("ComposeNames() error = nil, want error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}
