package liuyao

import (
	"encoding/json"
	"testing"
)

func TestGuaIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input guaIndex
		want  string
	}{
		{0, `"乾"`},
		{1, `"姤"`},
		{2, `"遁"`},
		{3, `"否"`},
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

func TestGuaIndex_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  guaIndex
		err   bool
	}{
		{`"乾"`, 0, false},
		{`"姤"`, 1, false},
		{`"否"`, 3, false},
		{`0`, 0, true},
		{`64`, 0, true},
		{`"invalid"`, 0, true},
	}
	for _, tt := range tests {
		var g guaIndex
		err := json.Unmarshal([]byte(tt.input), &g)
		if tt.err {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) = nil, want error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) = %v, want %d", tt.input, err, tt.want)
			} else if g != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, g, tt.want)
			}
		}
	}
}
