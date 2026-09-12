package bazi

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func parseOracleZhu(t *testing.T, name string) ganzhi.Zhu {
	t.Helper()
	gan, err := ganzhi.ParseGan(string([]rune(name)[0]))
	if err != nil {
		t.Fatal(err)
	}
	zhi, err := ganzhi.ParseZhi(string([]rune(name)[1]))
	if err != nil {
		t.Fatal(err)
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}

func hasShenShaName(rows [4][]shenShaEntry, name string) bool {
	for _, pillar := range rows {
		for _, item := range pillar {
			if item.Name == name {
				return true
			}
		}
	}
	return false
}

func TestDomainOracle_BaziClassicCorrections(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_classic_corrections.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		YangRen []struct {
			DayGan string `json:"day_gan"`
			Branch string `json:"branch"`
			Hit    bool   `json:"hit"`
		} `json:"yang_ren"`
		YueRenGe []struct {
			DayGan  string `json:"day_gan"`
			Month   string `json:"month"`
			Pattern string `json:"pattern"`
			Usage   string `json:"usage"`
		} `json:"yue_ren_ge"`
		FeedbackCharts []struct {
			ID                      string   `json:"id"`
			Pillars                 []string `json:"pillars"`
			YangRen                 bool     `json:"yang_ren"`
			Pattern                 string   `json:"pattern"`
			PatternGod              string   `json:"pattern_god"`
			PatternGodSource        string   `json:"pattern_god_source"`
			Usage                   string   `json:"usage"`
			Yong                    string   `json:"yong"`
			Xi                      string   `json:"xi"`
			Ji                      string   `json:"ji"`
			Strength                string   `json:"strength"`
			TiaoHouPrimary          string   `json:"tiao_hou_primary"`
			TiaoHouPrimaryHidden    bool     `json:"tiao_hou_primary_hidden"`
			TiaoHouSecondary        string   `json:"tiao_hou_secondary"`
			TiaoHouSecondaryPresent bool     `json:"tiao_hou_secondary_present"`
		} `json:"feedback_charts"`
		Season []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Month string `json:"month"`
			Day   string `json:"day"`
			Hit   bool   `json:"hit"`
		} `json:"season_shen_sha"`
		TianLuo []struct {
			ID     string `json:"id"`
			Year   string `json:"year"`
			Branch string `json:"branch"`
			Name   string `json:"name"`
		} `json:"tian_luo_di_wang"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.YangRen) != 10 || len(doc.YueRenGe) != 10 || len(doc.FeedbackCharts) != 1 || len(doc.Season) != 12 || len(doc.TianLuo) != 6 {
		t.Fatalf("coverage = %d/%d/%d/%d/%d, want 10/10/1/12/6", len(doc.YangRen), len(doc.YueRenGe), len(doc.FeedbackCharts), len(doc.Season), len(doc.TianLuo))
	}
	for _, tc := range doc.YangRen {
		t.Run("yang_ren/"+tc.DayGan, func(t *testing.T) {
			gan, err := ganzhi.ParseGan(tc.DayGan)
			if err != nil {
				t.Fatal(err)
			}
			zhi, err := ganzhi.ParseZhi(tc.Branch)
			if err != nil {
				t.Fatal(err)
			}
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: zhi},
				Ri:   ganzhi.Zhu{Gan: gan, Zhi: ganzhi.ZhiChen},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			if got := hasShenShaName(computeShenSha(bz, ganzhi.Male), "羊刃"); got != tc.Hit {
				t.Fatalf("羊刃 = %v, want %v", got, tc.Hit)
			}
		})
	}
	for _, tc := range doc.Season {
		t.Run(tc.ID, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: mustOracleBaziZhi(t, tc.Month)},
				Ri:   parseOracleZhu(t, tc.Day),
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},
			}
			if got := hasShenShaName(computeShenSha(bz, ganzhi.Male), tc.Name); got != tc.Hit {
				t.Fatalf("%s = %v, want %v", tc.Name, got, tc.Hit)
			}
		})
	}
	for _, tc := range doc.YueRenGe {
		t.Run("yue_ren_ge/"+tc.DayGan, func(t *testing.T) {
			gan, err := ganzhi.ParseGan(tc.DayGan)
			if err != nil {
				t.Fatal(err)
			}
			zhi, err := ganzhi.ParseZhi(tc.Month)
			if err != nil {
				t.Fatal(err)
			}
			chart := mkGeJuChart(gan, ganzhi.GanBing, ganzhi.GanGui, ganzhi.GanRen, zhi)
			got := computeGeJu(chart, nil)
			if got.Pattern != tc.Pattern || got.Usage != tc.Usage {
				t.Fatalf("geju = %+v, want %s/%s", got, tc.Pattern, tc.Usage)
			}
			if got.PatternGodSource == "" ||
				got.Structure.Pattern.Wuxing == "" || len(got.Structure.Pattern.Roots) == 0 {
				t.Fatalf("geju structural fields = %+v", got)
			}
			if tc.Pattern == "月刃格" && got.PatternGodSource != "month_blade" {
				t.Fatalf("month blade source = %s", got.PatternGodSource)
			}
		})
	}
	for _, tc := range doc.FeedbackCharts {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			full := ComputeFullChart(chart)
			if hasShenShaName([4][]shenShaEntry{full.Nian.ShenSha, full.Yue.ShenSha, full.Ri.ShenSha, full.Shi.ShenSha}, "羊刃") != tc.YangRen {
				t.Fatalf("羊刃 fact is inconsistent with classic yin-stem boundary")
			}
			ge := full.YongShen.GeJu
			if ge.Pattern != tc.Pattern || ge.PatternGod != tc.PatternGod ||
				ge.PatternGodSource != tc.PatternGodSource || ge.Usage != tc.Usage ||
				ge.Yong != tc.Yong || ge.Xi != tc.Xi || ge.Ji != tc.Ji {
				t.Fatalf("geju = %+v, want %+v", ge, tc)
			}
			if full.YongShen.FuYi.Strength != tc.Strength {
				t.Fatalf("strength = %s, want %s", full.YongShen.FuYi.Strength, tc.Strength)
			}
			th := full.YongShen.TiaoHou
			if th.Primary.Stem != tc.TiaoHouPrimary || th.Primary.Hidden != tc.TiaoHouPrimaryHidden ||
				th.Secondary == nil || th.Secondary.Stem != tc.TiaoHouSecondary ||
				(th.Secondary.Transparent || th.Secondary.Hidden) != tc.TiaoHouSecondaryPresent {
				t.Fatalf("tiaohou availability = primary:%+v secondary:%+v", th.Primary, th.Secondary)
			}
		})
	}
	for _, tc := range doc.TianLuo {
		t.Run(tc.ID, func(t *testing.T) {
			bz := ganzhi.Bazi{
				Nian: parseOracleZhu(t, tc.Year),
				Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: mustOracleBaziZhi(t, tc.Branch)},
				Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu},
				Shi:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiShen},
			}
			if got := hasShenShaName(computeShenSha(bz, ganzhi.Male), tc.Name); got != (tc.Name != "") {
				t.Fatalf("hit=%v, want %v", got, tc.Name != "")
			}
		})
	}
	if hourZhiIndex(time.Date(2026, 1, 1, 23, 0, 0, 0, cstLocation)) != 0 {
		t.Fatal("23:00 must be late Zi, not Hai")
	}
	full := ComputeFullChart(atomicChartFromStrings(t, []string{"甲子", "丙寅", "戊午", "壬子"}))
	if full.ShenShaSchool.Policy != "union_of_year_and_day_references" || len(full.ShenShaSchool.DualReference) != 7 {
		t.Fatalf("shen sha school = %+v", full.ShenShaSchool)
	}
	if full.YongShen.FuYi.Model != "support_control_with_day_master_strength" ||
		full.YongShen.FuYi.Basis.RootType == "" || len(full.YongShen.FuYi.Basis.DayMasterRoots) == 0 {
		t.Fatalf("fu yi model = %+v", full.YongShen.FuYi)
	}
	if full.YongShen.TiaoHou.Model != "qiongtong_primary_secondary_table" ||
		(!full.YongShen.TiaoHou.Primary.Transparent && !full.YongShen.TiaoHou.Primary.Hidden) {
		t.Fatalf("tiao hou model = %+v", full.YongShen.TiaoHou)
	}
}

func mustOracleBaziZhi(t *testing.T, name string) ganzhi.Zhi {
	t.Helper()
	zhi, err := ganzhi.ParseZhi(name)
	if err != nil {
		t.Fatal(err)
	}
	return zhi
}
