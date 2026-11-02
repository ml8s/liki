package bazi

import (
	"reflect"
	"testing"
)

func TestNatalRelationsCoverLegacyRelationArrays(t *testing.T) {
	cases := []struct {
		name    string
		pillars []string
	}{
		{name: "adjacent-gan-he-and-zhi-he", pillars: []string{"甲子", "己丑", "戊辰", "庚午"}},
		{name: "candidates-and-partial-xing", pillars: []string{"甲子", "丙寅", "己巳", "庚午"}},
		{name: "complete-sanhe", pillars: []string{"甲申", "甲子", "甲辰", "戊子"}},
		{name: "complete-sanhui", pillars: []string{"甲寅", "乙卯", "丙辰", "戊午"}},
		{name: "self-xing", pillars: []string{"甲子", "甲午", "甲子", "甲午"}},
		{name: "hai-po-anhe", pillars: []string{"甲子", "甲未", "甲酉", "甲戌"}},
		{name: "complete-xing-and-pairs", pillars: []string{"甲寅", "甲巳", "甲申", "甲子"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart := ComputeChartFromCore(pillarsCoreFromStrings(t, tc.pillars))
			byID := make(map[string]NatalRelationFact, len(chart.GuanXi))
			for _, fact := range chart.GuanXi {
				if fact.ID != natalRelationFactID(fact) {
					t.Fatalf("relation ID %q does not follow kind:group@members", fact.ID)
				}
				if len(fact.Members) == 0 {
					t.Fatalf("relation %s has no members", fact.ID)
				}
				byID[fact.ID] = fact
			}
			requireRelation := func(want NatalRelationFact) {
				t.Helper()
				want.ID = natalRelationFactID(want)
				got, ok := byID[want.ID]
				if !ok {
					t.Fatalf("missing canonical relation %#v; got %#v", want, chart.GuanXi)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("canonical relation = %#v, want %#v", got, want)
				}
			}

			for _, relation := range chart.GanHe {
				kind := "gan_he_candidate"
				if relation.Position == "adjacent" {
					kind = "gan_he"
				}
				group := ganHeGroupLabel(relation.GanA, relation.GanB)
				requireRelation(NatalRelationFact{
					ID: kind + ":" + group, Kind: kind, Group: group,
					Members: []NatalRelationMember{
						{Pillar: zhuLabels[relation.PillarA], Gan: relation.GanA},
						{Pillar: zhuLabels[relation.PillarB], Gan: relation.GanB},
					},
					Wuxing: relation.HeElement, Position: relation.Position, Contested: relation.Contested,
				})
			}
			for _, relation := range chart.GanChong {
				kind := "gan_chong_candidate"
				if relation.Position == "adjacent" {
					kind = "gan_chong"
				}
				group := ganChongGroupLabel(relation.GanA, relation.GanB)
				requireRelation(NatalRelationFact{
					ID: kind + ":" + group, Kind: kind, Group: group,
					Members: []NatalRelationMember{
						{Pillar: zhuLabels[relation.PillarA], Gan: relation.GanA},
						{Pillar: zhuLabels[relation.PillarB], Gan: relation.GanB},
					},
					Position: relation.Position,
				})
			}

			pair := func(kind, zhiA string, pillarA int, zhiB string, pillarB int, wuxing string) {
				t.Helper()
				var group string
				if kind == "liu_xing" {
					group = canonicalXingGroup(zhiA, zhiB)
				} else {
					group = canonicalZhiPairGroup(kind, zhiA, zhiB)
				}
				requireRelation(NatalRelationFact{
					ID: kind + ":" + group, Kind: kind, Group: group,
					Members: []NatalRelationMember{
						{Pillar: zhuLabels[pillarA], Zhi: zhiA},
						{Pillar: zhuLabels[pillarB], Zhi: zhiB},
					},
					Wuxing: wuxing,
				})
			}
			for _, relation := range chart.ZhiLiuHe {
				pair("zhi_liu_he", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, relation.Element)
			}
			for _, relation := range chart.LiuChong {
				pair("liu_chong", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, "")
			}
			for _, relation := range chart.LiuHai {
				pair("liu_hai", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, "")
			}
			for _, relation := range chart.LiuXing {
				pair("liu_xing", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, "")
			}
			for _, relation := range chart.LiuPo {
				pair("liu_po", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, "")
			}
			for _, relation := range chart.AnHe {
				pair("an_he", relation.ZhiA, relation.PillarA, relation.ZhiB, relation.PillarB, "")
			}

			triple := func(kind string, relations []TripleGroup) {
				t.Helper()
				for _, relation := range relations {
					group := branchGroupLabel(parseZhiNames(relation.Branches))
					members := make([]NatalRelationMember, 0, len(relation.Branches))
					for _, branch := range relation.Branches {
						members = append(members, NatalRelationMember{
							Pillar: pillarWithZhi(chart.ToBazi(), branch), Zhi: branch,
						})
					}
					requireRelation(NatalRelationFact{
						ID: kind + ":" + group, Kind: kind, Group: group,
						Members: members, Wuxing: relation.Element,
					})
				}
			}
			triple("san_he", chart.SanHe)
			triple("san_he_partial", chart.SanHePartial)
			triple("san_hui", chart.SanHui)
		})
	}
}
