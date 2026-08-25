package liuyao

import "fmt"

// YaoType names for text↔symbol conversion.
var yaoTypeNames = [4]struct {
	N int
	S string
}{
	{6, "老阴"}, {7, "少阳"}, {8, "少阴"}, {9, "老阳"},
}

// YaoTypeString returns the Chinese name for a yao type.
func YaoTypeString(y YaoType) string {
	for _, e := range yaoTypeNames {
		if y == YaoType(e.N) {
			return e.S
		}
	}
	return "?"
}

// ParseYaoType converts a Chinese yao name to a YaoType value.
func ParseYaoType(s string) (YaoType, error) {
	for _, e := range yaoTypeNames {
		if e.S == s {
			return YaoType(e.N), nil
		}
	}
	return 0, fmt.Errorf("unknown yao type: %q", s)
}

// ParseLiuQin converts a Chinese liuqin name to a LiuQin value.
func ParseLiuQin(s string) (LiuQin, error) {
	for i, name := range liuQinNames {
		if name == s {
			return LiuQin(i), nil
		}
	}
	return -1, fmt.Errorf("unknown liuqin: %q", s)
}

// ParseLiuShou converts a Chinese liushou name to a LiuShou value.
func ParseLiuShou(s string) (LiuShou, error) {
	for i, name := range liuShouNames {
		if name == s {
			return LiuShou(i), nil
		}
	}
	return -1, fmt.Errorf("unknown liushou: %q", s)
}

// ParseYongShen converts a Chinese yongshen name to a YongShen value.
func ParseYongShen(s string) (YongShen, error) {
	for i, name := range yongShenNames {
		if name == s {
			return YongShen(i), nil
		}
	}
	return -1, fmt.Errorf("unknown yongshen: %q", s)
}
