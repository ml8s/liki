package liuyao

import "fmt"

// ParseYongShen converts a Chinese yongshen name to a YongShen value.
func ParseYongShen(s string) (YongShen, error) {
	for i, name := range yongShenNames {
		if name == s {
			return YongShen(i), nil
		}
	}
	return -1, fmt.Errorf("unknown yongshen: %q", s)
}
