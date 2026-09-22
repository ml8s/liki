package liuyao

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_DongYaoDirectPriority(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/liuyao_dong_yao_priority.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID        string   `json:"id"`
			Target    string   `json:"target"`
			YongZhi   string   `json:"yong_zhi"`
			MovingZhi string   `json:"moving_zhi"`
			Expected  []string `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 6 {
		t.Fatalf("cases = %d, want 6", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			typ, err := ParseYongShen(tc.Target)
			if err != nil {
				t.Fatal(err)
			}
			yong, err := ganzhi.ParseZhi(tc.YongZhi)
			if err != nil {
				t.Fatal(err)
			}
			moving, err := ganzhi.ParseZhi(tc.MovingZhi)
			if err != nil {
				t.Fatal(err)
			}
			chart := Chart{
				Lines: [6]Line{
					{Position: 1, Type: ShaoYin, Zhi: yong, Wuxing: ganzhi.ZhiWuxing(yong), LiuQin: yongShenToLiuQin(typ)},
					{},
					{},
					{Position: 4, Type: LaoYang, Zhi: moving, Wuxing: ganzhi.ZhiWuxing(moving), DongSelf: true, LiuQin: QinGuanGui},
					{},
					{},
				},
				DongYao: []int{4},
			}
			got := computeDongYaoRelations(&chart, typ)
			gotNames := []string{}
			if len(got) == 1 {
				for _, item := range got[0].Relations {
					gotNames = append(gotNames, string(item))
				}
			}
			if len(got) != 1 || !reflect.DeepEqual(gotNames, tc.Expected) {
				t.Fatalf("relations = %v, want %v", gotNames, tc.Expected)
			}
		})
	}
}
