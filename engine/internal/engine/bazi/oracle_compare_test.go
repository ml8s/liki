package bazi

import (
	"encoding/json"
	"fmt"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestOracleComparison_19810826(t *testing.T) {
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}
	chart := canonicalChart(bz, ganzhi.Male, 1981)
	full := ComputeFullChart(chart)
	pillars := [4]fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi}

	// 1. Pillars
	wantPillars := []string{"辛酉", "丙申", "丙子", "己丑"}
	for i, p := range pillars {
		got := fmt.Sprintf("%s%s", ganzhi.GanName(p.Gan), ganzhi.ZhiName(p.Zhi))
		if got != wantPillars[i] {
			t.Errorf("pillar[%d]=%s want %s", i, got, wantPillars[i])
		}
	}

	// 2. Hidden stems
	wantHidden := [][]string{{"辛"}, {"庚", "壬", "戊"}, {"癸"}, {"己", "癸", "辛"}}
	for i, p := range pillars {
		got := hiddenLabels(p.CangGan)
		if !equalStrings(got, wantHidden[i]) {
			t.Errorf("hidden[%d]=%v want %v", i, got, wantHidden[i])
		}
	}

	// 3. Ten gods (hidden)
	wantTG := [][]string{{"正财"}, {"偏财", "七杀", "食神"}, {"正官"}, {"伤官", "正官", "正财"}}
	for i, p := range pillars {
		got := hiddenTenGodLabels(p.ShiShens)
		if !equalStrings(got, wantTG[i]) {
			t.Errorf("ten_gods[%d]=%v want %v", i, got, wantTG[i])
		}
	}

	// 4. Trend
	wantTrend := []string{"死", "病", "胎", "养"}
	for i, p := range pillars {
		got := dayMasterTrendLabel(full, p.Zhi)
		if got != wantTrend[i] {
			t.Errorf("trend[%d]=%s want %s", i, got, wantTrend[i])
		}
	}

	// 5. Self-sitting
	wantSS := []string{"临官", "病", "胎", "墓"}
	for i, p := range pillars {
		if p.SelfSitting != wantSS[i] {
			t.Errorf("self_sitting[%d]=%s want %s", i, p.SelfSitting, wantSS[i])
		}
	}

	// 6. Void
	wantVoid := []bool{true, true, false, false}
	for i, p := range pillars {
		if p.IsVoid != wantVoid[i] {
			t.Errorf("void[%d]=%v want %v", i, p.IsVoid, wantVoid[i])
		}
	}

	// 7. Gan He (丙辛合 x2)
	if len(full.GanHe) < 2 {
		t.Errorf("gan_he count=%d want>=2", len(full.GanHe))
	}

	// 8. Shensha union check
	oracleGods := map[string]bool{
		"天乙贵人": true, "太极贵人": true, "德秀贵人": true, "桃花": true,
		"文昌": true, "亡神": true, "阴差阳错": true, "天厨贵人": true,
		"福星贵人": true, "绞神": true, "童子煞": true, "飞刃": true,
		"天喜": true, "国印": true, "华盖": true,
	}
	engineGods := make(map[string]bool)
	for _, p := range pillars {
		for _, s := range p.ShenSha {
			engineGods[s.Name] = true
		}
	}
	for name := range oracleGods {
		if !engineGods[name] {
			t.Errorf("missing shensha %q (oracle has it)", name)
		}
	}

	// 9. Nayin
	wantNayin := []string{"石榴木", "山下火", "涧下水", "霹雳火"}
	for i, p := range pillars {
		if p.NaYin != wantNayin[i] {
			t.Errorf("nayin[%d]=%s want %s", i, p.NaYin, wantNayin[i])
		}
	}

	// 10. Six Po (子酉破)
	hehui := ComputeHeHui(chart)
	foundPo := false
	for _, po := range hehui.LiuPo {
		if po.ZhiA == "子" && po.ZhiB == "酉" || po.ZhiA == "酉" && po.ZhiB == "子" {
			foundPo = true
		}
	}
	if !foundPo {
		t.Error("liu_po: 子酉破 not detected (oracle detects it)")
	}

	// 11. Month season states: oracle pro_decl for 申月 is 金旺/水相/土休/火囚/木死.
	wantSeason := map[string]string{"金": "旺", "水": "相", "土": "休", "火": "囚", "木": "死"}
	for _, state := range full.ElementStates {
		if state.Season != wantSeason[state.Wuxing] {
			t.Errorf("element %s season = %s, want %s", state.Wuxing, state.Season, wantSeason[state.Wuxing])
		}
	}

	out, _ := json.MarshalIndent(map[string]any{
		"gan_he_count": len(full.GanHe),
		"liu_po":       hehui.LiuPo,
	}, "", "  ")
	t.Logf("summary:\n%s", out)
}

func canonicalChart(bz ganzhi.Bazi, gender ganzhi.Gender, year int) Chart {
	pillars := bz.Slice()
	zhuInfos := [4]zhuInfo{}
	for i, p := range pillars {
		zhuInfos[i] = zhuInfo{Zhu: p, NaYin: ganzhi.NayinLabel(p.Gan, p.Zhi)}
	}
	return Chart{
		Nian: zhuInfos[0], Yue: zhuInfos[1], Ri: zhuInfos[2], Shi: zhuInfos[3],
		Gender: gender, BirthYear: year,
	}
}
