package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 动爻 → 变卦 数据驱动命理锚点（《增删卜易》变卦规则）：
//   - 乾为天初爻老阳发动 → 变天风姤（乾宫一世卦）
//   - 变卦纳甲按卦体上下经卦（天风姤上乾下巽：内卦巽 → 辛丑辛亥辛酉），非按本宫
//   - 变爻六亲以本卦（主卦）宫五行论：辛丑土 生 乾宫金 → 父母
func TestDongYao_BianGua_Anchors(t *testing.T) {
	cases := []struct {
		name     string
		yaos     [6]int
		wantBen  string
		wantBian string
		wantDong []int
		bian1Gan string
		bian1Zhi string
		bian1Qin string
	}{
		{
			"乾为天 初爻动 → 天风姤",
			[6]int{9, 7, 7, 7, 7, 7}, // 初爻老阳(9 发动变阴)，余少阳(7)
			"乾为天", "姤", []int{1},
			"辛", "丑", "父母", // 内卦巽按经卦纳甲取辛丑
		},
		{
			"乾为天 五爻动 → 火天大有（归魂卦）",
			[6]int{7, 7, 7, 7, 9, 7},
			"乾为天", "大有", []int{5},
			"", "", "", // 变爻纳甲断言由初爻动 case 覆盖
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(2024, 2, 2, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
				116.4, 8)
			chart := ComputeChart(st, YongShiYao, c.yaos)
			if chart.Name != c.wantBen {
				t.Errorf("ben_gua = %s, want %s", chart.Name, c.wantBen)
			}
			if zhouyiTable[chart.BianGua].Name != c.wantBian {
				t.Errorf("bian_gua = %s, want %s", zhouyiTable[chart.BianGua].Name, c.wantBian)
			}
			if len(chart.DongYao) != len(c.wantDong) || chart.DongYao[0] != c.wantDong[0] {
				t.Errorf("dong_yao = %v, want %v", chart.DongYao, c.wantDong)
			}
			// 变卦纳甲 + 六亲（《增删卜易》：变卦纳甲按变卦所属宫，六亲以本卦宫五行论）。
			b0 := chart.BianLines[0]
			if b0.Gan == 0 || b0.Zhi == 0 {
				t.Fatal("bian_yao[0] 无纳甲")
			}
			if c.bian1Gan != "" {
				if ganzhi.GanName(b0.Gan) != c.bian1Gan || ganzhi.ZhiName(b0.Zhi) != c.bian1Zhi {
					t.Errorf("bian_yao[0] 纳甲 = %s%s, want %s%s",
						ganzhi.GanName(b0.Gan), ganzhi.ZhiName(b0.Zhi), c.bian1Gan, c.bian1Zhi)
				}
				if b0.LiuQin.String() != c.bian1Qin {
					t.Errorf("bian_yao[0] 六亲 = %s, want %s（乾金生水→子孙）", b0.LiuQin, c.bian1Qin)
				}
			}
		})
	}
}

// 引擎层防御：非法爻数（非 6-9）返回空盘，不产出错卦。
func TestComputeChart_InvalidYaos_EmptyChart(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2024, 2, 2, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	for _, yaos := range [][6]int{
		{0, 0, 0, 0, 0, 0},
		{1, 2, 3, 4, 5, 6},
		{6, 7, 7, 7, 7, 10},
		{7, 7, 7, 7, 7, 99},
	} {
		chart := ComputeChart(st, YongShiYao, yaos)
		if chart.Name != "" {
			t.Errorf("非法爻数 %v → 卦名 %q, want 空盘", yaos, chart.Name)
		}
	}
}
