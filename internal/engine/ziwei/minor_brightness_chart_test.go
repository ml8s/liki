package ziwei

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// TestMinorBrightnessInChart：验证 chart 输出层给文昌/文曲赋的亮度 == iztro golden（覆盖 complete_test 150 用例）
func TestMinorBrightnessInChart(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "minor_brightness_golden.json"))
	if err != nil {
		t.Fatalf("读 golden: %v", err)
	}
	var golden struct {
		WenChang map[string]string `json:"文昌"`
		WenQu    map[string]string `json:"文曲"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("解析 golden: %v", err)
	}
	cases := loadCases(t)
	checked, fail := 0, 0
	for _, tc := range cases {
		lt := parseLT(tc)
		gender := ganzhi.Female
		if tc.Gender == "男" {
			gender = ganzhi.Male
		}
		chart := ComputeChart(lt, gender)
		fc := ComputeFullChart(chart, 0, 0)
		for i := range fc.GongWei {
			zhi := fc.GongWei[i].Zhi.String()
			for _, s := range fc.GongWei[i].Stars {
				var want string
				switch s.Star {
				case WenChang:
					want = golden.WenChang[zhi]
				case WenQu:
					want = golden.WenQu[zhi]
				default:
					continue
				}
				checked++
				if s.Brightness != want {
					t.Errorf("%s %s宫: got %s want %s", starName(s.Star), zhi, s.Brightness, want)
					fail++
				}
			}
		}
	}
	if checked == 0 {
		t.Error("未检查到任何文昌/文曲——测试无效")
	}
	t.Logf("文昌/文曲亮度检查 %d 处，失败 %d", checked, fail)
}
