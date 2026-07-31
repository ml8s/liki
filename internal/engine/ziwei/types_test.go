package ziwei

import (
	"encoding/json"
	"testing"
)

func TestPalaceIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input palaceIndex
		want  string // expected JSON output
	}{
		{0, `"命宫"`},
		{1, `"兄弟"`},
		{11, `"父母"`},
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
		want  palaceIndex
		err   bool
	}{
		{`"命宫"`, 0, false},
		{`"兄弟"`, 1, false},
		{`"父母"`, 11, false},
		{`0`, 0, true},  // 整数必须拒绝
		{`5`, 0, true},
		{`"invalid"`, 0, true},
		{`""`, 0, true},
	}
	for _, tt := range tests {
		var p palaceIndex
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

func TestJuShu_MarshalJSON(t *testing.T) {
	tests := []struct {
		input juShu
		want  string
	}{
		{0, `""`},     // unknown→empty
		{2, `"水二局"`},
		{3, `"木三局"`},
		{4, `"金四局"`},
		{5, `"土五局"`},
		{6, `"火六局"`},
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

func TestJuShu_UnmarshalJSON_StringOnly(t *testing.T) {
	tests := []struct {
		input string
		want  juShu
		err   bool
	}{
		{`"水二局"`, 2, false},
		{`"木三局"`, 3, false},
		{`"火六局"`, 6, false},
		{`""`, 0, false},   // empty→0
		{`"金七局"`, 0, true}, // 不存在
		{`2`, 0, true},     // 整数拒绝
	}
	for _, tt := range tests {
		var j juShu
		err := json.Unmarshal([]byte(tt.input), &j)
		if tt.err {
			if err == nil {
				t.Errorf("UnmarshalJSON(%s) = nil, want error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("UnmarshalJSON(%s) = %v, want %d", tt.input, err, tt.want)
			} else if j != tt.want {
				t.Errorf("UnmarshalJSON(%s) = %d, want %d", tt.input, j, tt.want)
			}
		}
	}
}

func TestStarIndex_MarshalJSON(t *testing.T) {
	tests := []struct {
		input starIndex
		want  string
	}{
		{ZiWei, `"紫微"`},
		{TianJi, `"天机"`},
		{PoJun, `"破军"`},
		{LuCun, `"禄存"`},
		{DiJie, `"地劫"`},
		{starIndex(99), `""`},
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
		want  starIndex
		err   bool
	}{
		{`"紫微"`, ZiWei, false},
		{`"破军"`, PoJun, false},
		{`"禄存"`, LuCun, false},
		{`1`, 0, true},
		{`"不存在的星"`, 0, true},
		{`""`, 0, true},
	}
	for _, tt := range tests {
		var s starIndex
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
