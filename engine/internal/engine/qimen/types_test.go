package qimen

import (
	"encoding/json"
	"testing"
)

func TestPalaceIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input GongIndex
		want  string
	}{
		{GongKan, `"坎"`},
		{GongKun, `"坤"`},
		{GongLi, `"离"`},
		{GongQian, `"乾"`},
		{0, `"?"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.input)
		if err != nil {
			t.Errorf("MarshalJSON(%d) = %v", tt.input, err)
			continue
		}
		if string(b) != tt.want {
			t.Errorf("MarshalJSON(%d) = %s, want %s", tt.input, string(b), tt.want)
		}
	}
}

func TestPalaceIndex_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  GongIndex
		err   bool
	}{
		{`"坎"`, GongKan, false},
		{`"离"`, GongLi, false},
		{`"中"`, GongZhong, false},
		{`"乾"`, GongQian, false},
		{`1`, 0, true},
		{`0`, 0, true},
		{`"invalid"`, 0, true},
		{`""`, 0, true},
	}
	for _, tt := range tests {
		var p GongIndex
		err := json.Unmarshal([]byte(tt.input), &p)
		if tt.err {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) = nil, want error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) = %v, want %d", tt.input, err, tt.want)
			} else if p != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, p, tt.want)
			}
		}
	}
}

func TestStarIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input StarIndex
		want  string
	}{
		{StarTianPeng, `"天蓬"`},
		{StarTianRui, `"天芮"`},
		{StarTianYing, `"天英"`},
		{0, `"?"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.input)
		if err != nil {
			t.Errorf("MarshalJSON(%d) = %v", tt.input, err)
			continue
		}
		if string(b) != tt.want {
			t.Errorf("MarshalJSON(%d) = %s, want %s", tt.input, string(b), tt.want)
		}
	}
}

func TestStarIndex_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  StarIndex
		err   bool
	}{
		{`"天蓬"`, StarTianPeng, false},
		{`"天芮"`, StarTianRui, false},
		{`"天英"`, StarTianYing, false},
		{`1`, 0, true},
		{`"天"`, 0, true},
	}
	for _, tt := range tests {
		var s StarIndex
		err := json.Unmarshal([]byte(tt.input), &s)
		if tt.err {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) = nil, want error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) = %v, want %d", tt.input, err, tt.want)
			} else if s != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, s, tt.want)
			}
		}
	}
}

func TestDoorIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input DoorIndex
		want  string
	}{
		{DoorXiu, `"休门"`},
		{DoorSheng, `"生门"`},
		{DoorKai, `"开门"`},
		{0, `"?门"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.input)
		if err != nil {
			t.Errorf("MarshalJSON(%d) = %v", tt.input, err)
			continue
		}
		if string(b) != tt.want {
			t.Errorf("MarshalJSON(%d) = %s, want %s", tt.input, string(b), tt.want)
		}
	}
}

func TestDoorIndex_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  DoorIndex
		err   bool
	}{
		{`"休门"`, DoorXiu, false},
		{`"生门"`, DoorSheng, false},
		{`"开门"`, DoorKai, false},
		{`"休"`, DoorXiu, false}, // 没门字也接受
		{`1`, 0, true},
	}
	for _, tt := range tests {
		var d DoorIndex
		err := json.Unmarshal([]byte(tt.input), &d)
		if tt.err {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) = nil, want error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) = %v, want %d", tt.input, err, tt.want)
			} else if d != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, d, tt.want)
			}
		}
	}
}
