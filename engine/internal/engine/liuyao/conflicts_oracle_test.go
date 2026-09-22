package liuyao

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDomainOracle_LiuyaoConflicts(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/liuyao_core.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID          string   `json:"id"`
			Relations   []string `json:"relations"`
			WangShuai   string   `json:"wang_shuai"`
			XunKong     bool     `json:"xun_kong"`
			YuePo       bool     `json:"yue_po"`
			ExpectedIDs []string `json:"expected_ids"`
		} `json:"conflict_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 6 {
		t.Fatalf("cases = %d, want 6", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := Chart{
				YongShen: YongShenResult{
					WangShuai: tc.WangShuai,
					XunKong:   tc.XunKong,
					YuePo:     tc.YuePo,
				},
			}
			relations := make([]DongYaoRelationType, 0, len(tc.Relations))
			for _, name := range tc.Relations {
				relations = append(relations, relationTypeFromOracle(t, name))
			}
			chart.DongYaoRelations = []DongYaoRelation{{Position: 1, Relations: relations}}
			got := make([]string, 0)
			for _, conflict := range computeConflicts(&chart) {
				got = append(got, conflict.ID)
				if strings.HasSuffix(conflict.ID, "opposition") && len(conflict.Facts) != len(tc.Relations) {
					t.Errorf("%s facts = %v, want %v", conflict.ID, conflict.Facts, tc.Relations)
				}
			}
			if !reflect.DeepEqual(got, tc.ExpectedIDs) {
				t.Errorf("conflicts = %v, want %v", got, tc.ExpectedIDs)
			}
		})
	}
}

func relationTypeFromOracle(t *testing.T, name string) DongYaoRelationType {
	t.Helper()
	switch name {
	case "生用":
		return RelationShengYong
	case "克用":
		return RelationKeYong
	case "比和":
		return RelationBiHe
	case "冲用":
		return RelationChongYong
	case "生原神":
		return RelationShengYuan
	case "克原神":
		return RelationKeYuan
	case "生忌神":
		return RelationShengJi
	case "克忌神":
		return RelationKeJi
	default:
		t.Fatalf("unknown dong-yao relation %q", name)
		return ""
	}
}
