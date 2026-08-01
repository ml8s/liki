package ziwei

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// TestJudgment_PatternsAndSihua 验证 ziwei.judgment 的格局/四化/三方四正输出。
// 用例复用 complete_test.json（iztro golden 数据）。
func TestJudgment_PatternsAndSihua(t *testing.T) {
	data, err := os.ReadFile("testdata/complete_test.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var cases []testCaseRef
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	if len(cases) < 10 {
		t.Fatalf("golden 用例不足: %d", len(cases))
	}

	var pass, fail int
	for i, tc := range cases[:10] {
		lt := parseLT(tc)
		gender := ganzhi.Female
		if tc.Gender == "男" {
			gender = ganzhi.Male
		}
		chart := ComputeChart(lt, gender)
		j := ComputeJudgment(chart)

		// 1. 四化非空且与 chart 一致
		if len(j.SiHua) != len(tc.SiHua) {
			t.Errorf("case %d (%s): 四化数 %d != golden %d", i, tc.Lunar, len(j.SiHua), len(tc.SiHua))
			fail++
			continue
		}
		for _, s := range j.SiHua {
			if s.StarName == "" || s.Type == "" {
				t.Errorf("case %d: 四化星名/类型为空: %+v", i, s)
				fail++
			}
		}

		// 2. 三方四正必有 3 宫（含本宫）
		if len(j.SanFang) != 4 {
			t.Errorf("case %d: 三方四正宫数 %d != 4", i, len(j.SanFang))
			fail++
		}
		for _, sf := range j.SanFang {
			if sf.Name == "" {
				t.Errorf("case %d: 三方宫名空", i)
				fail++
			}
		}

		// 3. 格局数组非 nil
		if j.Patterns == nil {
			t.Errorf("case %d: patterns 为 nil", i)
			fail++
		}

		pass++
	}
	if fail > 0 {
		t.Fatalf("%d/%d 失败", fail, pass+fail)
	}
	fmt.Printf("judgment 10 例验证通过（四化/三方四正/格局）\n")
}
