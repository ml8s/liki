package ziwei

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestMinorBrightnessGolden：验证文昌/文曲亮度与 iztro golden 一致（testdata/minor_brightness_golden.json）
func TestMinorBrightnessGolden(t *testing.T) {
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
	zhiNames := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	check := func(star starIndex, want map[string]string, label string) {
		for i, zhiName := range zhiNames {
			got := miaoWang(star, Zhi(i+1)).String()
			if got != want[zhiName] {
				t.Errorf("%s %s: got %s want %s", label, zhiName, got, want[zhiName])
			}
		}
	}
	check(WenChang, golden.WenChang, "文昌")
	check(WenQu, golden.WenQu, "文曲")
}
