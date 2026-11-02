// Package ganzhitest contains oracle helpers shared by tests in other engine
// packages. It is intentionally separate from the production ganzhi API.
package ganzhitest

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// ParseWangShuai converts a Chinese prosperity label used by oracle fixtures.
func ParseWangShuai(s string) (ganzhi.WangShuai, error) {
	switch s {
	case "旺":
		return ganzhi.WSWang, nil
	case "相":
		return ganzhi.WSXiang, nil
	case "休":
		return ganzhi.WSXiu, nil
	case "囚":
		return ganzhi.WSQiu, nil
	case "死":
		return ganzhi.WSSi, nil
	default:
		return 0, fmt.Errorf("unknown wangshuai: %q", s)
	}
}
