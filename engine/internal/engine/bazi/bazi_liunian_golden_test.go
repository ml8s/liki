package bazi

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ============================================================================
// 流年数据驱动 golden 测试（命理知识正确性锚点）
// 手算预期（非引擎自证）：流年干支、流年干对日主的十神、流年神煞（动态年支+日支双查、值年病符/丧门/吊客/大耗）
// 命例：beijing-1984（甲子 丙寅 己卯 戊辰，己土日主，男）——同 bazi_golden.json
// ============================================================================

type liuNianGoldenCase struct {
	Name     string   `json:"name"`
	Solar    string   `json:"solar"`
	TZ       int      `json:"tz"`
	Lon      float64  `json:"lon"`
	Gender   string   `json:"gender"`
	Year     int      `json:"year"`
	YearName string   `json:"year_name"`
	ShiShen  string   `json:"shi_shen"`
	ShenSha  []string `json:"shensha"`
}

func loadLiuNianGolden(t *testing.T) []liuNianGoldenCase {
	t.Helper()
	b, err := os.ReadFile("testdata/liunian_golden.json")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var cases []liuNianGoldenCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	return cases
}

func TestLiuNian_Golden(t *testing.T) {
	for _, gc := range loadLiuNianGolden(t) {
		t.Run(gc.Name, func(t *testing.T) {
			tt, err := time.ParseInLocation("2006-01-02T15:04:05", gc.Solar, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatalf("parse solar: %v", err)
			}
			st := tianwen.SolarTime(tt)
			g := ganzhi.Male
			if gc.Gender == "female" {
				g = ganzhi.Female
			}
			chart := ComputeChart(st, g)

			ln, err := ComputeLiuNian(chart, gc.Year)
			if err != nil {
				t.Fatalf("ComputeLiuNian(%d): %v", gc.Year, err)
			}

			// 流年干支（nian_name）
			if ln.YearName != gc.YearName {
				t.Errorf("year_name = %q, want %q", ln.YearName, gc.YearName)
			}
			// 流年干对日主的十神（shi_shen）
			if ln.ShiShen != gc.ShiShen {
				t.Errorf("shi_shen = %q, want %q", ln.ShiShen, gc.ShiShen)
			}
			// 流年神煞（name 集合，忽略顺序）
			got := map[string]bool{}
			for _, s := range ln.ShenSha {
				got[s.Name] = true
			}
			for _, want := range gc.ShenSha {
				if !got[want] {
					t.Errorf("%d 缺神煞 %q, got %v", gc.Year, want, ln.ShenSha)
				}
			}
			// 流年神煞集合必须与基准一致。
			if len(got) != len(gc.ShenSha) {
				t.Errorf("%d 神煞数 = %d (got %v), want %d (%v)", gc.Year, len(got), ln.ShenSha, len(gc.ShenSha), gc.ShenSha)
			}
		})
	}
}
