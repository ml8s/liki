package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestEdgeDates(t *testing.T) {
	// 边界日期：节气交节日、闰年、午夜、冬/夏至、年初/年末
	dates := []string{
		"2024-02-04", // 立春
		"2024-06-21", // 夏至
		"2024-12-21", // 冬至
		"2024-02-29", // 闰日
		"2000-01-01", // 世纪之交
		"2023-12-31", // 年末
		"2024-03-20", // 春分
		"2024-09-22", // 秋分
		"1900-01-01", // 旧历
		"2099-12-31", // 远未来
	}
	for _, d := range dates {
		bt, err := time.ParseInLocation("2006-01-02", d, time.FixedZone("CST", 8*3600))
		if err != nil {
			t.Fatalf("解析日期 %s: %v", d, err)
		}
		// 测两个时辰（子时/午时）
		for _, h := range []int{0, 12} {
			st := tianwen.GregorianToSolar(bt.Add(time.Duration(h)*time.Hour), 116.4, 8)
			ch := ComputeChart(st, ShiQiMen)
			if ch.Pan.Jushu < 1 || ch.Pan.Jushu > 9 {
				t.Errorf("%s %d时: 局数越界 %d", d, h, ch.Pan.Jushu)
			}
			if ch.Pan.RiGan == 0 || ch.Pan.RiZhi == 0 {
				t.Errorf("%s %d时: 干支缺失", d, h)
			}
			// 用神符号冒烟
			for _, s := range []string{"生门", "戊", "天辅", "六合"} {
				sym, err := ParseYongShen(s)
				if err != nil {
					t.Fatalf("%s %d时 解析 %s: %v", d, h, s, err)
				}
				_ = ComputeYongShen(ch, []YongShenSymbol{sym})
			}
		}
	}
}
