package ganzhi

import "testing"

func TestXunKong(t *testing.T) {
	cases := []struct {
		gan  Gan
		zhi  Zhi
		want [2]Zhi
	}{
		{GanJia, ZhiZi, [2]Zhi{11, 12}},   // 甲子旬空戌亥
		{GanBing, ZhiYin, [2]Zhi{11, 12}}, // 丙寅（甲子旬）空戌亥
		{GanGeng, ZhiWu, [2]Zhi{11, 12}},  // 庚午（甲子旬）空戌亥
		{GanJia, ZhiXu, [2]Zhi{9, 10}},    // 甲戌旬空申酉
		{GanJia, ZhiShen, [2]Zhi{7, 8}},   // 甲申旬空午未
		{GanJia, ZhiWu, [2]Zhi{5, 6}},     // 甲午旬空辰巳
		{GanJia, ZhiChen, [2]Zhi{3, 4}},   // 甲辰旬空寅卯
		{GanJia, ZhiYin, [2]Zhi{1, 2}},    // 甲寅旬空子丑
	}
	for _, c := range cases {
		got := XunKong(c.gan, c.zhi)
		if got != c.want {
			t.Errorf("XunKong(%s%s) = [%s %s], want [%s %s]",
				ganNames[c.gan], zhiNames[c.zhi],
				zhiNames[got[0]], zhiNames[got[1]],
				zhiNames[c.want[0]], zhiNames[c.want[1]])
		}
	}
}
