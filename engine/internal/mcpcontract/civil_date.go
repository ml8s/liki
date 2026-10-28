package mcpcontract

import "liki-engine/internal/engine/tianwen"

func civilDateString(date tianwen.CivilDate) DateString {
	return DateString(date.String())
}
