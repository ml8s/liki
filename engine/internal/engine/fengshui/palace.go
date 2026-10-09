package fengshui

import "liki-engine/internal/engine/luoshu"

// Palace is the shared 洛书九宫 palace; the canonical definition lives in the
// neutral luoshu package (fengshui, bazhai, xuankong and qimen share it).
type Palace = luoshu.Palace

// PalaceTable is the shared 洛书九宫 table indexed by palace number (1-9).
var PalaceTable = luoshu.PalaceTable

// PalaceByNumber returns the palace for a given number (1-9).
func PalaceByNumber(n int) Palace {
	return luoshu.ByNumber(n)
}
