package bazi

import (
	"testing"
)

func TestRelationLinksPointToCanonicalGuanXi(t *testing.T) {
	cases := [][]string{
		{"甲子", "己丑", "戊辰", "庚午"},
		{"甲子", "丙寅", "己巳", "庚午"},
		{"甲申", "甲子", "甲辰", "戊子"},
		{"甲寅", "乙卯", "丙辰", "戊午"},
		{"甲子", "甲午", "甲子", "甲午"},
		{"甲子", "甲未", "甲酉", "甲戌"},
		{"甲寅", "甲巳", "甲申", "甲子"},
		{"辛巳", "癸酉", "甲卯", "庚辰"},
	}
	for _, pillars := range cases {
		chart := ComputeChartFromCore(pillarsCoreFromStrings(t, pillars))
		ids := make(map[string]bool, len(chart.GuanXi))
		for _, fact := range chart.GuanXi {
			ids[fact.ID] = true
		}
		var links []RelationLink
		links = append(links, chart.FuYi.Basis.GuanXiLinks...)
		links = append(links, chart.GeJu.Structure.GuanXiLinks...)
		links = append(links, chart.TiaoHou.Primary.GuanXiLinks...)
		if chart.TiaoHou.Secondary != nil {
			links = append(links, chart.TiaoHou.Secondary.GuanXiLinks...)
		}
		for _, link := range links {
			if !ids[link.ID] {
				t.Fatalf("pillars %v link %q does not point to guan_xi; ids=%v", pillars, link.ID, ids)
			}
			if len(link.Targets) == 0 {
				t.Fatalf("pillars %v link %q has no targets", pillars, link.ID)
			}
		}
	}
}
