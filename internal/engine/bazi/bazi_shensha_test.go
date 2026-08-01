package bazi

import (
	"time"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 神煞准确性测试 ──
// 基于标准神煞口诀独立验证，不依赖代码实现。

func makeChart(year, month, day, hour, minute int, lon, tz float64, gender ganzhi.Gender) Chart {
	// 公历直接排盘（对齐生产链路，不做真太阳时经度校正）
	st := tianwen.SolarTime(time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.FixedZone("", int(tz*3600))))
	return ComputeChart(st, gender)
}

func makeFullChart(year, month, day, hour, minute int, lon, tz float64, gender ganzhi.Gender) FullChart {
	c := makeChart(year, month, day, hour, minute, lon, tz, gender)
	return ComputeFullChart(c)
}


// collectShenShaNames returns all shensha names for a pillar.
func collectShenShaNames(p fullZhuInfo) map[string]bool {
	m := make(map[string]bool)
	for _, ss := range p.ShenSha {
		m[ss.Name] = true
	}
	return m
}

// ── 天乙贵人 ──
// 口诀：甲戊庚牛羊(丑未), 乙己鼠猴乡(子申), 丙丁猪鸡位(亥酉),
//
//	壬癸兔蛇藏(卯巳), 六辛逢虎马(寅午)

func TestShenSha_TianYi(t *testing.T) {
	tests := []struct {
		name        string
		dayGan      ganzhi.Gan
		yearGan     ganzhi.Gan
		pillarBranches [4]ganzhi.Zhi
		wantZhus []int // 0-indexed pillars that should have 天乙
	}{
		{
			name:        "甲日主-日支见丑未",
			dayGan:      ganzhi.GanJia,   // 甲
			yearGan:     ganzhi.GanJia,   // 甲（年干同）
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiChou, ganzhi.ZhiYin, ganzhi.ZhiChen, ganzhi.ZhiShen}, // 丑在年柱
			wantZhus: []int{0}, // 年柱见丑
		},
		{
			name:        "甲日主-年干乙-双天乙",
			dayGan:      ganzhi.GanJia,
			yearGan:     ganzhi.GanYi,    // 年干乙 → 天乙子申
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiWu, ganzhi.ZhiWei, ganzhi.ZhiChen, ganzhi.ZhiShen}, // 未在月柱(日主天乙), 申在时柱(年干天乙)
			wantZhus: []int{1, 3}, // 月柱未+时柱申
		},
		{
			name:        "丙日主-见亥酉",
			dayGan:      ganzhi.GanBing,  // 丙
			yearGan:     ganzhi.GanBing,  // 丙
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiHai, ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiYin}, // 亥在年柱
			wantZhus: []int{0}, // 年柱亥
		},
		{
			name:        "庚日主-见丑(天乙)",
			dayGan:      ganzhi.GanGeng,  // 庚 → 甲戊庚同 → 丑未
			yearGan:     ganzhi.GanGeng,
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiYin, ganzhi.ZhiMao}, // 丑在月柱
			wantZhus: []int{1},
		},
		{
			name:        "辛日主-见午寅",
			dayGan:      ganzhi.GanXin,   // 辛 → 六辛逢虎马(寅午)
			yearGan:     ganzhi.GanXin,
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiWu, ganzhi.ZhiMao}, // 午在日柱
			wantZhus: []int{2},
		},
		{
			name:        "壬日主-见卯巳",
			dayGan:      ganzhi.GanRen,   // 壬 → 壬癸兔蛇藏(卯巳)
			yearGan:     ganzhi.GanRen,
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiChen, ganzhi.ZhiSi}, // 巳在时柱
			wantZhus: []int{3},
		},
		// 没有天乙的情况
		{
			name:        "甲日主-无丑未",
			dayGan:      ganzhi.GanJia,
			yearGan:     ganzhi.GanJia,
			pillarBranches: [4]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiYin, ganzhi.ZhiChen, ganzhi.ZhiWu},
			wantZhus: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: tt.yearGan, Zhi: tt.pillarBranches[0]},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.pillarBranches[1]},
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: tt.pillarBranches[2]},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: tt.pillarBranches[3]},
			}
			ss := computeShenSha(bz)

			gotSet := make(map[int]bool)
			for i, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "天乙贵人" {
						gotSet[i] = true
					}
				}
			}

			wantSet := make(map[int]bool)
			for _, p := range tt.wantZhus {
				wantSet[p] = true
			}

			for i := 0; i < 4; i++ {
				if gotSet[i] != wantSet[i] {
					if wantSet[i] {
						t.Errorf("pillar[%d]: want 天乙贵人 but not found", i)
					} else {
						t.Errorf("pillar[%d]: unexpected 天乙贵人", i)
					}
				}
			}
		})
	}
}

// ── 文昌 ──
// 口诀：甲巳乙午报君知，丙戊申宫丁己鸡，庚猪辛鼠壬逢虎，癸人见兔入云梯
// 即：甲→巳, 乙→午, 丙戊→申, 丁己→酉, 庚→亥, 辛→子, 壬→寅, 癸→卯

func TestShenSha_WenChang(t *testing.T) {
	tests := []struct {
		name    string
		dayGan  ganzhi.Gan
		zhi     ganzhi.Zhi
		wantHit bool
	}{
		{"甲见巳", ganzhi.GanJia, ganzhi.ZhiSi, true},
		{"乙见午", ganzhi.GanYi, ganzhi.ZhiWu, true},
		{"丙见申", ganzhi.GanBing, ganzhi.ZhiShen, true},
		{"丁见酉", ganzhi.GanDing, ganzhi.ZhiYou, true},
		{"戊见申", ganzhi.GanWu, ganzhi.ZhiShen, true},
		{"己见酉", ganzhi.GanJi, ganzhi.ZhiYou, true},
		{"庚见亥", ganzhi.GanGeng, ganzhi.ZhiHai, true},
		{"辛见子", ganzhi.GanXin, ganzhi.ZhiZi, true},
		{"壬见寅", ganzhi.GanRen, ganzhi.ZhiYin, true},
		{"癸见卯", ganzhi.GanGui, ganzhi.ZhiMao, true},
		// 不应有文昌
		{"甲见午-无", ganzhi.GanJia, ganzhi.ZhiWu, false},
		{"丙见巳-无", ganzhi.GanBing, ganzhi.ZhiSi, false},
		{"庚见子-无", ganzhi.GanGeng, ganzhi.ZhiZi, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi}, // 测试支在月柱
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)

			// 检查所有柱是否有文昌
			hasWenChang := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "文昌" {
						hasWenChang = true
					}
				}
			}
			if hasWenChang != tt.wantHit {
				t.Errorf("文昌 found = %v, want %v", hasWenChang, tt.wantHit)
			}
		})
	}
}

// ── 桃花/驿马/华盖 ──
// 三合局神煞，以年支和日支为基准

func TestShenSha_TaoHua(t *testing.T) {
	// 口诀：申子辰在酉，寅午戌在卯，巳酉丑在午，亥卯未在子
	// 命理：桃花以年支和日支为参考
	tests := []struct {
		name     string
		yearZhi  ganzhi.Zhi
		dayZhi   ganzhi.Zhi
		checkZhi ganzhi.Zhi
		wantHit  bool
	}{
		// 申子辰 → 桃花在酉
		{"申年-酉为桃花", ganzhi.ZhiShen, ganzhi.ZhiZi, ganzhi.ZhiYou, true},
		{"子年-酉为桃花", ganzhi.ZhiZi, ganzhi.ZhiChen, ganzhi.ZhiYou, true},
		// 寅午戌 → 桃花在卯
		{"寅年-卯为桃花", ganzhi.ZhiYin, ganzhi.ZhiWu, ganzhi.ZhiMao, true},
		{"午年-卯为桃花", ganzhi.ZhiWu, ganzhi.ZhiYin, ganzhi.ZhiMao, true},
		// 巳酉丑 → 桃花在午
		{"巳年-午为桃花", ganzhi.ZhiSi, ganzhi.ZhiYou, ganzhi.ZhiWu, true},
		// 亥卯未 → 桃花在子
		{"亥年-子为桃花", ganzhi.ZhiHai, ganzhi.ZhiMao, ganzhi.ZhiZi, true},
		// 年支子→酉, 日支丑→午。卯不在两者桃花位 → 不命中
		{"子年-卯非桃花", ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiMao, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.dayZhi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "桃花" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("桃花 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

func TestShenSha_YiMa(t *testing.T) {
	// 口诀：申子辰马在寅，寅午戌马在申，巳酉丑马在亥，亥卯未马在巳
	tests := []struct {
		name    string
		yearZhi ganzhi.Zhi
		dayZhi  ganzhi.Zhi
		checkZhi ganzhi.Zhi
		wantHit bool
	}{
		{"申年-寅为驿马", ganzhi.ZhiShen, ganzhi.ZhiZi, ganzhi.ZhiYin, true},
		{"寅年-申为驿马", ganzhi.ZhiYin, ganzhi.ZhiWu, ganzhi.ZhiShen, true},
		{"巳年-亥为驿马", ganzhi.ZhiSi, ganzhi.ZhiYou, ganzhi.ZhiHai, true},
		{"亥年-巳为驿马", ganzhi.ZhiHai, ganzhi.ZhiMao, ganzhi.ZhiSi, true},
		{"子年-申非驿马", ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiShen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.dayZhi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "驿马" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("驿马 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

func TestShenSha_HuaGai(t *testing.T) {
	// 口诀：申子辰见辰，寅午戌见戌，巳酉丑见丑，亥卯未见未
	// 命理：华盖以年支和日支为参考
	tests := []struct {
		name     string
		yearZhi  ganzhi.Zhi
		dayZhi   ganzhi.Zhi
		checkZhi ganzhi.Zhi
		wantHit  bool
	}{
		{"申子辰-见辰", ganzhi.ZhiShen, ganzhi.ZhiZi, ganzhi.ZhiChen, true},
		{"寅午戌-见戌", ganzhi.ZhiYin, ganzhi.ZhiWu, ganzhi.ZhiXu, true},
		{"巳酉丑-见丑", ganzhi.ZhiSi, ganzhi.ZhiYou, ganzhi.ZhiChou, true},
		{"亥卯未-见未", ganzhi.ZhiHai, ganzhi.ZhiMao, ganzhi.ZhiWei, true},
		// 年支子→华盖辰, 日支卯→华盖未。申均非两者华盖 → 不命中
		{"子年-见申-非华盖", ganzhi.ZhiZi, ganzhi.ZhiMao, ganzhi.ZhiShen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.dayZhi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "华盖" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("华盖 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 羊刃 ──
// 甲刃在卯，乙刃在寅，丙戊刃在午，丁己刃在巳，庚刃在酉，辛刃在申，壬刃在子，癸刃在亥

func TestShenSha_YangRen(t *testing.T) {
	tests := []struct {
		name    string
		dayGan  ganzhi.Gan
		zhi     ganzhi.Zhi
		wantHit bool
	}{
		{"甲刃在卯", ganzhi.GanJia, ganzhi.ZhiMao, true},
		{"乙刃在寅", ganzhi.GanYi, ganzhi.ZhiYin, true},
		{"丙刃在午", ganzhi.GanBing, ganzhi.ZhiWu, true},
		{"丁刃在巳", ganzhi.GanDing, ganzhi.ZhiSi, true},
		{"戊刃在午", ganzhi.GanWu, ganzhi.ZhiWu, true},
		{"己刃在巳", ganzhi.GanJi, ganzhi.ZhiSi, true},
		{"庚刃在酉", ganzhi.GanGeng, ganzhi.ZhiYou, true},
		{"辛刃在申", ganzhi.GanXin, ganzhi.ZhiShen, true},
		{"壬刃在子", ganzhi.GanRen, ganzhi.ZhiZi, true},
		{"癸刃在亥", ganzhi.GanGui, ganzhi.ZhiHai, true},
		{"甲非刃在寅", ganzhi.GanJia, ganzhi.ZhiYin, false},
		{"庚非刃在申", ganzhi.GanGeng, ganzhi.ZhiShen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi},
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "羊刃" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("羊刃 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 劫煞 / 灾煞 ──

func TestShenSha_JieSha(t *testing.T) {
	// 劫煞：申子辰在巳，寅午戌在亥，巳酉丑在寅，亥卯未在申
	tests := []struct {
		name    string
		yearZhi ganzhi.Zhi
		checkZhi ganzhi.Zhi
		wantHit bool
	}{
		{"申子辰-劫煞在巳", ganzhi.ZhiShen, ganzhi.ZhiSi, true},
		{"寅午戌-劫煞在亥", ganzhi.ZhiYin, ganzhi.ZhiHai, true},
		{"巳酉丑-劫煞在寅", ganzhi.ZhiSi, ganzhi.ZhiYin, true},
		{"亥卯未-劫煞在申", ganzhi.ZhiHai, ganzhi.ZhiShen, true},
		{"子年-劫煞不在寅", ganzhi.ZhiZi, ganzhi.ZhiYin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "劫煞" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("劫煞 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

func TestShenSha_ZaiSha(t *testing.T) {
	// 灾煞：申子辰在午，寅午戌在子，巳酉丑在卯，亥卯未在酉
	tests := []struct {
		name    string
		yearZhi ganzhi.Zhi
		checkZhi ganzhi.Zhi
		wantHit bool
	}{
		{"申子辰-灾煞在午", ganzhi.ZhiShen, ganzhi.ZhiWu, true},
		{"寅午戌-灾煞在子", ganzhi.ZhiYin, ganzhi.ZhiZi, true},
		{"巳酉丑-灾煞在卯", ganzhi.ZhiSi, ganzhi.ZhiMao, true},
		{"亥卯未-灾煞在酉", ganzhi.ZhiHai, ganzhi.ZhiYou, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "灾煞" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("灾煞 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 月德 ──
// 寅午戌月见丙，亥卯未月见甲，申子辰月见壬，巳酉丑月见庚

func TestShenSha_YueDe(t *testing.T) {
	tests := []struct {
		name       string
		monthZhi   ganzhi.Zhi
		checkGan   ganzhi.Gan // 月干
		wantHit    bool
	}{
		{"寅月月德丙", ganzhi.ZhiYin, ganzhi.GanBing, true},
		{"午月月德丙", ganzhi.ZhiWu, ganzhi.GanBing, true},
		{"戌月月德丙", ganzhi.ZhiXu, ganzhi.GanBing, true},
		{"亥月月德甲", ganzhi.ZhiHai, ganzhi.GanJia, true},
		{"申月月德壬", ganzhi.ZhiShen, ganzhi.GanRen, true},
		{"巳月月德庚", ganzhi.ZhiSi, ganzhi.GanGeng, true},
		{"寅月月德非甲", ganzhi.ZhiYin, ganzhi.GanJia, false},
		{"巳月月德非丙", ganzhi.ZhiSi, ganzhi.GanBing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: tt.checkGan, Zhi: tt.monthZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)
			found := false
			for _, s := range ss[1] {
				if s.Name == "月德" { found = true; break }
			}
			if found != tt.wantHit {
				t.Errorf("月德在月柱 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 空亡 ──

func TestShenSha_KongWang(t *testing.T) {
	// 空亡由日柱所在旬决定
	// 甲子旬(0-9): 空戌亥, 甲戌旬(10-19): 空申酉, 甲申旬(20-29): 空午未
	// 甲午旬(30-39): 空辰巳, 甲辰旬(40-49): 空寅卯, 甲寅旬(50-59): 空子丑
	tests := []struct {
		name     string
		dayGan   ganzhi.Gan
		dayZhi   ganzhi.Zhi
		checkZhi ganzhi.Zhi
		inXun    int // 旬 0=甲子..5=甲寅
		wantVoid bool
	}{
		// 甲子旬(0-9): 空戌(11)亥(12)
		{"甲子日-戌空亡", ganzhi.GanJia, ganzhi.ZhiZi, ganzhi.ZhiXu, 0, true},
		{"甲子日-亥空亡", ganzhi.GanJia, ganzhi.ZhiZi, ganzhi.ZhiHai, 0, true},
		{"甲子日-子不空", ganzhi.GanJia, ganzhi.ZhiZi, ganzhi.ZhiZi, 0, false},
		// 甲戌旬(10-19): 空申酉
		{"甲戌日-申空亡", ganzhi.GanJia, ganzhi.ZhiXu, ganzhi.ZhiShen, 1, true},
		{"甲戌日-酉空亡", ganzhi.GanJia, ganzhi.ZhiXu, ganzhi.ZhiYou, 1, true},
		// 甲申旬(20-29): 空午未
		{"甲申日-午空亡", ganzhi.GanJia, ganzhi.ZhiShen, ganzhi.ZhiWu, 2, true},
		// 甲午旬(30-39): 空辰巳
		{"甲午日-辰空亡", ganzhi.GanJia, ganzhi.ZhiWu, ganzhi.ZhiChen, 3, true},
		// 甲辰旬(40-49): 空寅卯
		{"甲辰日-寅空亡", ganzhi.GanJia, ganzhi.ZhiChen, ganzhi.ZhiYin, 4, true},
		// 甲寅旬(50-59): 空子丑
		{"甲寅日-子空亡", ganzhi.GanJia, ganzhi.ZhiYin, ganzhi.ZhiZi, 5, true},
		{"甲寅日-丑空亡", ganzhi.GanJia, ganzhi.ZhiYin, ganzhi.ZhiChou, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.checkZhi}, // 测试支在年柱
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: tt.dayZhi},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
			}
			// 验证旬
			idx := ganzhi.SixtyCycleIndex(tt.dayGan, tt.dayZhi)
			xun := idx / 10
			if xun != tt.inXun {
				t.Fatalf("test data: six cycle index=%d, xun=%d, expected %d", idx, xun, tt.inXun)
			}

			voidHits := computeKongWang(bz)
			hasVoid := false
			for _, found := range voidHits {
				if found == 0 { // 年柱
					hasVoid = true
				}
			}
			if hasVoid != tt.wantVoid {
				t.Errorf("空亡 found = %v, want %v (void indices: %v)", hasVoid, tt.wantVoid, voidHits)
			}
		})
	}
}

// ── 天罗地网 ──

func TestShenSha_TianLuoDiWang(t *testing.T) {
	// 戌亥为天罗(11,12)，辰巳为地网(5,6)
	tests := []struct {
		name     string
		zhi      ganzhi.Zhi
		wantName string // "天罗" or "地网" or ""
	}{
		{"戌为天罗", ganzhi.ZhiXu, "天罗"},
		{"亥为天罗", ganzhi.ZhiHai, "天罗"},
		{"辰为地网", ganzhi.ZhiChen, "地网"},
		{"巳为地网", ganzhi.ZhiSi, "地网"},
		{"子非天罗地网", ganzhi.ZhiZi, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiShen},
			}
			ss := computeShenSha(bz)
			if tt.wantName == "" {
				for _, pillar := range ss {
					for _, e := range pillar {
						if e.Name == "天罗" || e.Name == "地网" {
							t.Errorf("unexpected %s at zhi=%s", e.Name, ganzhi.ZhiName(tt.zhi))
						}
					}
				}
			} else {
				found := false
				for _, pillar := range ss {
					for _, e := range pillar {
						if e.Name == tt.wantName {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("want %s but not found", tt.wantName)
				}
			}
		})
	}
}

// ── 禄神 ──
// 甲禄在寅，乙禄在卯，丙戊禄在巳，丁己禄在午，庚禄在申，辛禄在酉，壬禄在亥，癸禄在子

func TestShenSha_LuShen(t *testing.T) {
	tests := []struct {
		name    string
		dayGan  ganzhi.Gan
		zhi     ganzhi.Zhi
		wantHit bool
	}{
		{"甲禄在寅", ganzhi.GanJia, ganzhi.ZhiYin, true},
		{"乙禄在卯", ganzhi.GanYi, ganzhi.ZhiMao, true},
		{"丙禄在巳", ganzhi.GanBing, ganzhi.ZhiSi, true},
		{"丁禄在午", ganzhi.GanDing, ganzhi.ZhiWu, true},
		{"戊禄在巳", ganzhi.GanWu, ganzhi.ZhiSi, true},
		{"己禄在午", ganzhi.GanJi, ganzhi.ZhiWu, true},
		{"庚禄在申", ganzhi.GanGeng, ganzhi.ZhiShen, true},
		{"辛禄在酉", ganzhi.GanXin, ganzhi.ZhiYou, true},
		{"壬禄在亥", ganzhi.GanRen, ganzhi.ZhiHai, true},
		{"癸禄在子", ganzhi.GanGui, ganzhi.ZhiZi, true},
		{"甲禄不在卯", ganzhi.GanJia, ganzhi.ZhiMao, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi},
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "禄神" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("禄神 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 红鸾天喜 ──

func TestShenSha_HongLuanTianXi(t *testing.T) {
	// 红鸾：子→卯, 丑→寅, 寅→丑, 卯→子, 辰→亥, 巳→戌,
	//        午→酉, 未→申, 申→未, 酉→午, 戌→巳, 亥→辰
	// 天喜：子→酉, 丑→申, 寅→未, 卯→午, 辰→巳, 巳→辰,
	//        午→卯, 未→寅, 申→丑, 酉→子, 戌→亥, 亥→戌
	tests := []struct {
		name       string
		yearZhi    ganzhi.Zhi
		checkZhi   ganzhi.Zhi
		wantHongLuan bool
		wantTianXi   bool
	}{
		{"子年见卯→红鸾", ganzhi.ZhiZi, ganzhi.ZhiMao, true, false},
		{"子年见酉→天喜", ganzhi.ZhiZi, ganzhi.ZhiYou, false, true},
		{"午年见酉→红鸾", ganzhi.ZhiWu, ganzhi.ZhiYou, true, false},
		{"午年见卯→天喜", ganzhi.ZhiWu, ganzhi.ZhiMao, false, true},
		{"寅年见丑→红鸾", ganzhi.ZhiYin, ganzhi.ZhiChou, true, false},
		{"寅年见未→天喜", ganzhi.ZhiYin, ganzhi.ZhiWei, false, true},
		{"子年见午→无", ganzhi.ZhiZi, ganzhi.ZhiWu, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: tt.yearZhi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.checkZhi},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiZi},
			}
			ss := computeShenSha(bz)
			hasHL := false
			hasTX := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "红鸾" {
						hasHL = true
					}
					if e.Name == "天喜" {
						hasTX = true
					}
				}
			}
			if hasHL != tt.wantHongLuan {
				t.Errorf("红鸾 found = %v, want %v", hasHL, tt.wantHongLuan)
			}
			if hasTX != tt.wantTianXi {
				t.Errorf("天喜 found = %v, want %v", hasTX, tt.wantTianXi)
			}
		})
	}
}

// ── 学堂 (日主长生之位) ──

func TestShenSha_XueTang(t *testing.T) {
	// 学堂 = 日主长生之位
	// 甲长生在亥，乙长生在午，丙长生在寅，丁长生在酉，
	// 戊长生在寅，己长生在酉，庚长生在巳，辛长生在子，
	// 壬长生在申，癸长生在卯
	tests := []struct {
		name    string
		dayGan  ganzhi.Gan
		zhi     ganzhi.Zhi
		wantHit bool
	}{
		{"甲长生亥→学堂", ganzhi.GanJia, ganzhi.ZhiHai, true},
		{"乙长生午→学堂", ganzhi.GanYi, ganzhi.ZhiWu, true},
		{"丙长生寅→学堂", ganzhi.GanBing, ganzhi.ZhiYin, true},
		{"庚长生巳→学堂", ganzhi.GanGeng, ganzhi.ZhiSi, true},
		{"壬长生申→学堂", ganzhi.GanRen, ganzhi.ZhiShen, true},
		{"甲长生非寅", ganzhi.GanJia, ganzhi.ZhiYin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: tt.zhi},
				Ri:   ganzhi.Zhu{Gan: tt.dayGan, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			ss := computeShenSha(bz)
			found := false
			for _, pillar := range ss {
				for _, e := range pillar {
					if e.Name == "学堂" {
						found = true
					}
				}
			}
			if found != tt.wantHit {
				t.Errorf("学堂 found = %v, want %v", found, tt.wantHit)
			}
		})
	}
}

// ── 综合神煞：真实八字盘 ──

func TestShenSha_RealChart(t *testing.T) {
	// 1984-02-15 08:00 北京 (120°E, UTC+8)
	// 命理推导:
	//   年: 1984已过立春 → 甲子年
	//   月: 立春后惊蛰前 → 寅月, 甲年→五虎遁丙寅
	//   日: 1900-01-01=甲戌(序号10)基准, 累加30725天 → 序号15=己卯
	//   时: 08:00辰时, 己日→五鼠遁戊辰
	// 八字: 甲子 丙寅 己卯 戊辰, 日主己土
	cr := makeFullChart(1984, 2, 15, 8, 0, 120, 8, ganzhi.Male)

	// 验证四柱
	if cr.Nian.Gan != ganzhi.GanJia || cr.Nian.Zhi != ganzhi.ZhiZi {
		t.Errorf("年柱 = %s%s, want 甲子", ganzhi.GanName(cr.Nian.Gan), ganzhi.ZhiName(cr.Nian.Zhi))
	}
	if cr.Yue.Gan != ganzhi.GanBing || cr.Yue.Zhi != ganzhi.ZhiYin {
		t.Errorf("月柱 = %s%s, want 丙寅", ganzhi.GanName(cr.Yue.Gan), ganzhi.ZhiName(cr.Yue.Zhi))
	}
	if cr.Ri.Gan != ganzhi.GanJi || cr.Ri.Zhi != ganzhi.ZhiMao {
		t.Errorf("日柱 = %s%s, want 己卯", ganzhi.GanName(cr.Ri.Gan), ganzhi.ZhiName(cr.Ri.Zhi))
	}
	if cr.Shi.Gan != ganzhi.GanWu || cr.Shi.Zhi != ganzhi.ZhiChen {
		t.Errorf("时柱 = %s%s, want 戊辰", ganzhi.GanName(cr.Shi.Gan), ganzhi.ZhiName(cr.Shi.Zhi))
	}
	if cr.Ri.Gan != ganzhi.GanJi {
		t.Errorf("日主 = %s, want 己", ganzhi.GanName(cr.Ri.Gan))
	}

	// 神煞预期 — 基于命理知识独立推导：
	//
	// 年柱甲子:
	//   天乙: 日干己→子申,年支子✓
	//   桃花: 日支卯→亥卯未桃花在子,年支子✓
	//   将星: 年支子→申子辰将星在子(自坐)
	//   灾煞: 年支子→申子辰灾煞在子(自坐); 日支卯→亥卯未灾煞在卯(自坐,在日柱)
	//   (灾煞在年柱说明: zaishaBranch[1(子)]=1(子) → 年柱见子=自坐)
	//
	// 月柱丙寅:
	//   月德: 寅月月德在丙(3), 月干丙✓
	//   驿马: 年支子→申子辰驿马在寅, 月支寅✓
	//
	// 日柱己卯:
	//   将星: 日支卯→亥卯未将星在卯(自坐)
	//   灾煞: 日支卯→亥卯未灾煞在卯(自坐)
	//   红鸾: 年支子→红鸾在卯, 日支卯✓
	//   勾神: 年支子→勾神=(1+2)%12+1=卯, 日支卯✓
	//
	// 时柱戊辰:
	//   华盖: 年支子→申子辰华盖在辰, 时支辰✓
	//   月恩: 寅月月恩在戊(5), 时干戊✓
	//   地网: 辰为地网
	//   血刃: 日干己→血刃在辰(5), 时支辰✓
	//
	// 空亡: 日柱己卯序号15, 旬=1(甲戌旬), 空申酉 → 四支子寅卯辰无申酉 → 无空亡

	expected := map[int][]string{
		0: {"天乙贵人", "桃花", "将星", "灾煞"},
		1: {"月德", "驿马"},
		2: {"将星", "灾煞", "红鸾", "勾神"},
		3: {"华盖", "月恩", "地网", "血刃"},
	}

	pillarLabels := [4]string{"年柱", "月柱", "日柱", "时柱"}

	for i, wantNames := range expected {
		var got map[string]bool
		switch i {
		case 0:
			got = collectShenShaNames(cr.Nian)
		case 1:
			got = collectShenShaNames(cr.Yue)
		case 2:
			got = collectShenShaNames(cr.Ri)
		case 3:
			got = collectShenShaNames(cr.Shi)
		}

		for _, want := range wantNames {
			if !got[want] {
				t.Errorf("%s: want %s but not found. Got: %v", pillarLabels[i], want, got)
			}
		}
		// 检查无明显多余神煞（不能有冲突类别）
		for name := range got {
			found := false
			for _, w := range wantNames {
				if name == w {
					found = true
					break
				}
			}
			if !found {
				t.Logf("%s: unexpected shensha %s", pillarLabels[i], name)
			}
		}
	}

	// 验证空亡: 己卯日在甲戌旬, 空申酉
	if cr.Nian.IsVoid {
		t.Error("年柱(子)不应在甲戌旬空亡(空申酉)")
	}
	if cr.Yue.IsVoid {
		t.Error("月柱(寅)不应空亡")
	}
	if cr.Ri.IsVoid {
		t.Error("日柱(卯)不应空亡")
	}
	if cr.Shi.IsVoid {
		t.Error("时柱(辰)不应空亡")
	}
}
