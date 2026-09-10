package ziwei

import (
	"reflect"
	"testing"
)

func TestComputePalaceFactsMatchesAuditAndGroups(t *testing.T) {
	chart := Chart{GongWei: [12]gong{
		{
			Name: "命宫",
			Stars: []starInfo{
				{Name: "紫微", Star: ZiWei, IsMajor: true},
				{Name: "文昌", Star: WenChang, Brightness: "庙"},
			},
		},
		{
			Name: "夫妻宫",
			Stars: []starInfo{
				{Name: "贪狼", Star: TanLang, IsMajor: true, SiHua: string(HuaJi)},
			},
		},
		{
			Name: "子女宫",
			Stars: []starInfo{
				{Name: "擎羊", Star: QingYang, Brightness: "陷"},
			},
		},
	}}
	facts := computePalaceFacts(chart)
	want := []PalaceFact{
		{Palace: "命宫", Kind: "star", Target: "紫微", Star: "紫微"},
		{Palace: "命宫", Kind: "star", Target: "紫微主星", Star: "紫微"},
		{Palace: "命宫", Kind: "star", Target: "文昌", Star: "文昌"},
		{Palace: "命宫", Kind: "star", Target: "紫微六吉星", Star: "文昌"},
		{Palace: "命宫", Kind: "star", Target: "紫微文星", Star: "文昌"},
		{Palace: "命宫", Kind: "brightness", Target: "庙", Star: "文昌", Value: "庙"},
		{Palace: "命宫", Kind: "brightness", Target: "庙旺", Star: "文昌", Value: "庙"},
		{Palace: "夫妻宫", Kind: "star", Target: "贪狼", Star: "贪狼"},
		{Palace: "夫妻宫", Kind: "star", Target: "紫微主星", Star: "贪狼"},
		{Palace: "夫妻宫", Kind: "si_hua", Target: "忌", Star: "贪狼"},
		{Palace: "夫妻宫", Kind: "special", Target: "唯一主星", Star: "贪狼", Value: "true"},
		{Palace: "子女宫", Kind: "star", Target: "擎羊", Star: "擎羊"},
		{Palace: "子女宫", Kind: "star", Target: "煞星", Star: "擎羊"},
		{Palace: "子女宫", Kind: "brightness", Target: "陷", Star: "擎羊", Value: "陷"},
		{Palace: "子女宫", Kind: "brightness", Target: "落陷", Star: "擎羊", Value: "陷"},
	}
	for _, want := range want {
		found := false
		for _, got := range facts {
			if reflect.DeepEqual(got, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("fact %#v not found in %#v", want, facts)
		}
	}
	if facts := computePalaceFacts(Chart{}); len(facts) != 12 {
		// 十二个空宫各产生一条无主星事实。
		t.Fatalf("empty chart facts = %d, want 12", len(facts))
	}
}
