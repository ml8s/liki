package bazi

import "testing"

type baziGoldenFile struct {
	Name   string
	File   string
	Checks goldenChecks
}

var baziGoldenFiles = []baziGoldenFile{
	{Name: "core_and_san_yuan", File: "testdata/bazi_golden.json", Checks: goldenChecks{MingGong: true, ShenGong: true, TaiYuan: true}},
	{Name: "edges", File: "testdata/bazi_golden2.json", Checks: goldenChecks{MingGong: true}},
	{Name: "all_layers", File: "testdata/bazi_golden3.json", Checks: goldenChecks{MingGong: true}},
	{Name: "calendar_and_shichen", File: "testdata/bazi_golden4.json", Checks: goldenChecks{}},
	{Name: "lunar_and_overseas", File: "testdata/bazi_golden5.json", Checks: goldenChecks{
		MingGong:      true,
		ShenGong:      true,
		TaiYuan:       true,
		QiYunMonthDay: true,
	}},
	{Name: "qiyun_precision", File: "testdata/bazi_golden6.json", Checks: goldenChecks{QiYunMonthDay: true}},
}

func TestBaziGoldenFiles(t *testing.T) {
	for _, source := range baziGoldenFiles {
		t.Run(source.Name, func(t *testing.T) {
			for _, gc := range loadBaziGoldenFile(t, source.File) {
				t.Run(gc.Name, func(t *testing.T) {
					verifyBaziGoldenCase(t, gc, source.Checks)
				})
			}
		})
	}
}
