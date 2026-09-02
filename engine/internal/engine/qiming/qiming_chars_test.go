package qiming

import "testing"

func TestLookupCharInfersElementFromRadical(t *testing.T) {
	character := LookupChar("饰")
	if character == nil {
		t.Fatal("LookupChar() = nil, want character with inferred element")
	}
	if character.Element.String() != "火" {
		t.Fatalf("element = %s, want 火", character.Element)
	}
	if character.Radical != "饣" {
		t.Fatalf("radical = %q, want 饣", character.Radical)
	}
}

func TestLookupCharOmitsMissingRadical(t *testing.T) {
	character := LookupChar("有")
	if character == nil {
		t.Fatal("LookupChar() = nil, want character")
	}
	if character.Radical != "" {
		t.Fatalf("radical = %q, want empty", character.Radical)
	}
}
