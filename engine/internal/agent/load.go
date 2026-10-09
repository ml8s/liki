package agent

import (
	"liki-engine/internal/engine/bazhai"
	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/huangli"
	"liki-engine/internal/engine/liuyao"
	"liki-engine/internal/engine/qimen"
	"liki-engine/internal/engine/xuankong"
	"liki-engine/internal/engine/ziwei"
)

// Load initializes every embedded engine data table. It is idempotent and must
// be called (from main, before serving) before any tool runs. Each domain Load
// pulls in its own dependencies, so listing the leaves + domains is enough.
func Load() error {
	for _, load := range []func() error{
		ganzhi.Load,
		bazi.Load,
		ziwei.Load,
		liuyao.Load,
		qimen.Load,
		huangli.Load,
		bazhai.Load,
		xuankong.Load,
	} {
		if err := load(); err != nil {
			return err
		}
	}
	return nil
}
