package xuankong

import (
	"encoding/json"
	"testing"
)

func TestStar9Index_MarshalJSON(t *testing.T) {
	tests := []struct {
		input Star9Index
		want  string
	}{
		{Star9TanLang, `"贪狼"`},
		{Star9JuMen, `"巨门"`},
		{Star9LuCun, `"禄存"`},
		{Star9WenQu, `"文曲"`},
		{Star9LianZhen, `"廉贞"`},
		{Star9WuQu, `"武曲"`},
		{Star9PoJun, `"破军"`},
		{Star9ZuoFu, `"左辅"`},
		{Star9YouBi, `"右弼"`},
		{0, `""`},
		{10, `""`},
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

func TestStar9Index_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  Star9Index
		err   bool
	}{
		{`"贪狼"`, Star9TanLang, false},
		{`"右弼"`, Star9YouBi, false},
		{`1`, 0, true},
		{`"invalid"`, 0, true},
		{`""`, 0, false},
	}
	for _, tt := range tests {
		var s Star9Index
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
