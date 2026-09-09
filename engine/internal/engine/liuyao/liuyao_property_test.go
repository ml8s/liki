package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestComputeChart_All4096YaoCombinations(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 9, 8, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	count := 0
	for a := 6; a <= 9; a++ {
		for b := 6; b <= 9; b++ {
			for c := 6; c <= 9; c++ {
				for d := 6; d <= 9; d++ {
					for e := 6; e <= 9; e++ {
						for f := 6; f <= 9; f++ {
							yaos := [6]int{a, b, c, d, e, f}
							chart := ComputeChart(st, YongShiYao, yaos)
							count++

							if chart.Name == "" || chart.Palace == "" {
								t.Fatalf("yaos %v produced empty core chart: %+v", yaos, chart)
							}
							if len(chart.Lines) != 6 {
								t.Fatalf("yaos %v lines=%d", yaos, len(chart.Lines))
							}
							seenWorld, seenOther := false, false
							for i, line := range chart.Lines {
								if line.Position != i+1 {
									t.Fatalf("yaos %v line %d position=%d", yaos, i, line.Position)
								}
								if int(line.Type) != yaos[i] {
									t.Fatalf("yaos %v line %d type=%d", yaos, i, line.Type)
								}
								if line.LiuQin.String() == "?" || line.LiuShou.String() == "?" {
									t.Fatalf("yaos %v has invalid six relative/spirit: %+v", yaos, line)
								}
								switch line.ShiYing {
								case "世":
									if seenWorld {
										t.Fatalf("yaos %v has duplicate world line", yaos)
									}
									seenWorld = true
								case "应":
									if seenOther {
										t.Fatalf("yaos %v has duplicate other line", yaos)
									}
									seenOther = true
								case "":
								default:
									t.Fatalf("yaos %v invalid shi_ying %q", yaos, line.ShiYing)
								}
							}
							if !seenWorld || !seenOther {
								t.Fatalf("yaos %v missing world/other", yaos)
							}
							if chart.YongShen.Position < 1 || chart.YongShen.Position > 6 {
								t.Fatalf("yaos %v yong position=%d", yaos, chart.YongShen.Position)
							}
							expectedDong := 0
							for _, value := range yaos {
								if value == 6 || value == 9 {
									expectedDong++
								}
							}
							if len(chart.DongYao) != expectedDong {
								t.Fatalf("yaos %v dong=%v want %d lines", yaos, chart.DongYao, expectedDong)
							}
							for _, candidate := range chart.TimingCandidates {
								if candidate.ID == "" || candidate.Mechanism == "" || candidate.Confidence != "candidate" {
									t.Fatalf("yaos %v invalid timing candidate %+v", yaos, candidate)
								}
							}
						}
					}
				}
			}
		}
	}
	if count != 4096 {
		t.Fatalf("combination count=%d, want 4096", count)
	}
}
