package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func chartForTest() Chart {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st, ShiQiMen)
}

// TestParseQianShi 占事类型解析
func TestParseQianShi(t *testing.T) {
	tests := []struct{ in, want string }{
		{"事业", "事业"}, {"求财", "求财"}, {"婚姻", "婚姻"},
		{"健康", "健康"}, {"诉讼", "诉讼"}, {"学业", "学业"},
		{"出行", "出行"}, {"隐藏", "隐藏"}, {"综合", "综合"},
	}
	for _, tt := range tests {
		got, err := ParseQianShi(tt.in)
		if err != nil || string(got) != tt.want {
			t.Errorf("ParseQianShi(%q) = %v,%v; want %q", tt.in, got, err, tt.want)
		}
	}
	// 未知类型报错
	if _, err := ParseQianShi("火星"); err == nil {
		t.Error("未知占事应报错")
	}
}

// TestQianShiBody 占事→事象用神映射
func TestQianShiBody(t *testing.T) {
	// 事业 → 开门 + 天心
	body := qianshiBody(QianShiye, "")
	if body.Door == nil || *body.Door != DoorKai {
		t.Errorf("事业用神门 = %v, want 开门", body.Door)
	}
	if body.Star == nil || *body.Star != StarTianXin {
		t.Errorf("事业用神星 = %v, want 天心", body.Star)
	}
	// 求财 → 生门 + 戊
	body = qianshiBody(QianQiucai, "")
	if body.Door == nil || *body.Door != DoorSheng {
		t.Errorf("求财用神门 = %v, want 生门", body.Door)
	}
	if body.Stem == nil || *body.Stem != ganzhi.GanWu {
		t.Errorf("求财用神干 = %v, want 戊", body.Stem)
	}
	// 婚姻 → 六合神 + 庚（男）/ 乙（女）
	body = qianshiBody(QianHunyin, "male")
	if body.Spirit == nil || *body.Spirit != SpiritLiuHe {
		t.Errorf("婚姻用神神 = %v, want 六合", body.Spirit)
	}
	if body.Stem == nil || *body.Stem != ganzhi.GanGeng {
		t.Errorf("男婚姻用神干 = %v, want 庚", body.Stem)
	}
	body = qianshiBody(QianHunyin, "female")
	if body.Stem == nil || *body.Stem != ganzhi.GanYi {
		t.Errorf("女婚姻用神干 = %v, want 乙", body.Stem)
	}
	// 健康 → 死门 + 天芮
	body = qianshiBody(QianJiankang, "")
	if body.Door == nil || *body.Door != DoorSi || body.Star == nil || *body.Star != StarTianRui {
		t.Errorf("健康用神 = 死门+天芮, got door=%v star=%v", body.Door, body.Star)
	}
}

// TestComputeYongShen 用神聚合（日干/时干/事象用神落宫）
func TestComputeYongShen(t *testing.T) {
	chart := chartForTest()
	ys := ComputeYongShen(chart, QianShiye, "")

	if ys.Name != "事业" {
		t.Errorf("name = %q, want 事业", ys.Name)
	}
	// 日干落宫（求测人）
	if ys.RiGanPalace == 0 {
		t.Error("日干落宫应为非零")
	}
	// 时干落宫（所问之事）
	if ys.ShiGanPalace == 0 {
		t.Error("时干落宫应为非零")
	}
	// 事象用神落宫
	if ys.BodyPalace == 0 {
		t.Error("事象用神落宫应为非零")
	}
	// 日时生克
	if ys.RiShiShengKe == "" {
		t.Error("日时生克不应为空")
	}
}

// TestComputeYongShen_NianGan 年命干聚合（需 birth_year）
func TestComputeYongShen_NianGan(t *testing.T) {
	chart := chartForTest()
	// 有 birth_year → 年命干落宫（1985 乙丑年，年干=乙，盘内有乙）
	ys := ComputeYongShenWithBirth(chart, QianZonghe, "", 1985)
	if ys.NianGanPalace == nil {
		t.Fatal("有 birth_year 时年命干落宫不应为空")
	}
	if *ys.NianGanPalace <= 0 {
		t.Errorf("年命干落宫应>0，got %d", *ys.NianGanPalace)
	}
	// 无 birth_year → 年命干为空
	ys2 := ComputeYongShen(chart, QianZonghe, "")
	if ys2.NianGanPalace != nil {
		t.Error("无 birth_year 时年命干落宫应为空")
	}
}

// TestComputeYongShen_KongWangMaXing 用神落宫空亡/马星
func TestComputeYongShen_KongWangMaXing(t *testing.T) {
	chart := chartForTest()
	ys := ComputeYongShen(chart, QianShiye, "")
	// 用神落宫是否为盘的空亡/马星宫
	if ys.BodyPalace > 0 {
		// 验证 kong_wang/ma_xing 与盘一致
		for _, k := range chart.Pan.KongWang {
			if k == ys.BodyPalace {
				if !ys.KongWang {
					t.Errorf("用神落宫%d为空亡宫，kong_wang应为true", ys.BodyPalace)
				}
			}
		}
		if ys.BodyPalace == chart.Pan.MaXing && !ys.MaXing {
			t.Errorf("用神落宫%d为马星宫，ma_xing应为true", ys.BodyPalace)
		}
	}
}
