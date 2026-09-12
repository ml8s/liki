package liuyao

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func parseOracleLiuQin(t *testing.T, name string) LiuQin {
	t.Helper()
	switch name {
	case "父母":
		return QinFumu
	case "兄弟":
		return QinXiongDi
	case "官鬼":
		return QinGuanGui
	case "妻财":
		return QinQiCai
	case "子孙":
		return QinZiSun
	default:
		t.Fatalf("unknown liu_qin %q", name)
		return -1
	}
}

func oracleWuxing(name string) ganzhi.Wuxing {
	switch name {
	case "木":
		return ganzhi.WxMu
	case "火":
		return ganzhi.WxHuo
	case "土":
		return ganzhi.WxTu
	case "金":
		return ganzhi.WxJin
	case "水":
		return ganzhi.WxShui
	default:
		return 0
	}
}

func mustOracleZhi(t *testing.T, name string) ganzhi.Zhi {
	t.Helper()
	value, err := ganzhi.ParseZhi(name)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestDomainOracle_LiuyaoPatternSemantics(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/liuyao_pattern_semantics.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		TombCases []struct {
			ID                 string   `json:"id"`
			TargetWuxing       string   `json:"target_wuxing"`
			TargetZhi          string   `json:"target_zhi"`
			TargetIsMoving     bool     `json:"target_is_moving"`
			DayZhi             string   `json:"day_zhi"`
			MovingTombPosition int      `json:"moving_tomb_position"`
			MovingTombZhi      string   `json:"moving_tomb_zhi"`
			ChangedZhi         string   `json:"changed_zhi"`
			ExpectedTypes      []string `json:"expected_types"`
		} `json:"tomb_cases"`
		SelectionCases []struct {
			ID       string `json:"id"`
			YongShen string `json:"yong_shen"`
			Lines    []struct {
				Position  int    `json:"position"`
				Zhi       string `json:"zhi"`
				LiuQin    string `json:"liu_qin"`
				WangShuai string `json:"wang_shuai"`
				XunKong   bool   `json:"xun_kong"`
				YuePo     bool   `json:"yue_po"`
				MuKu      bool   `json:"mu_ku"`
				DongSelf  bool   `json:"dong_self"`
				ShiYing   string `json:"shi_ying"`
			} `json:"lines"`
			ExpectedPosition int `json:"expected_position"`
		} `json:"selection_cases"`
		PatternCases []struct {
			ID     string `json:"id"`
			Target string `json:"target"`
			Lines  []struct {
				Position  int    `json:"position"`
				Zhi       string `json:"zhi"`
				LiuQin    string `json:"liu_qin"`
				WangShuai string `json:"wang_shuai"`
				YuePo     bool   `json:"yue_po"`
			} `json:"lines"`
			DayRelations []struct {
				Relations []string `json:"relations"`
			} `json:"day_relations"`
			MovingPositions       []int    `json:"moving_positions"`
			LinesZhi              []string `json:"lines_zhi"`
			ChangedLinesZhi       []string `json:"changed_lines_zhi"`
			ExpectedSubtype       string   `json:"expected_subtype"`
			ExpectedAbsentSubtype string   `json:"expected_absent_subtype"`
			ExpectedIsTrue        bool     `json:"expected_is_true"`
		} `json:"pattern_cases"`
		ComputedCases []struct {
			ID              string `json:"id"`
			Time            string `json:"time"`
			Yaos            [6]int `json:"yaos"`
			YongShen        string `json:"yong_shen"`
			ExpectedPattern struct {
				Type    string `json:"type"`
				SubType string `json:"sub_type"`
				IsTrue  bool   `json:"is_true"`
			} `json:"expected_pattern"`
		} `json:"computed_cases"`
		WholeGuaCases []struct {
			ID              string `json:"id"`
			Time            string `json:"time"`
			Yaos            [6]int `json:"yaos"`
			ExpectedName    string `json:"expected_name"`
			ExpectedSubtype string `json:"expected_subtype"`
		} `json:"whole_gua_cases"`
		WholeGuaClosure struct {
			SixClash       []string `json:"six_clash"`
			SixCombination []string `json:"six_combination"`
		} `json:"whole_gua_closure"`
		SanheCases []struct {
			ID    string `json:"id"`
			Lines []struct {
				Position int    `json:"position"`
				Zhi      string `json:"zhi"`
			} `json:"lines"`
			RiZhi            string   `json:"ri_zhi"`
			YueZhi           string   `json:"yue_zhi"`
			Element          string   `json:"element"`
			ExpectedComplete bool     `json:"expected_complete"`
			ExpectedMissing  []string `json:"expected_missing"`
		} `json:"sanhe_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.TombCases) != 4 || len(doc.SelectionCases) != 4 || len(doc.PatternCases) != 3 || len(doc.ComputedCases) != 1 || len(doc.WholeGuaCases) != 5 || len(doc.SanheCases) != 2 {
		t.Fatalf("cases = %d/%d/%d/%d/%d/%d, want 4/4/3/1/5/2", len(doc.TombCases), len(doc.SelectionCases), len(doc.PatternCases), len(doc.ComputedCases), len(doc.WholeGuaCases), len(doc.SanheCases))
	}
	for _, tc := range doc.TombCases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := Chart{RiZhi: ganzhi.ZhiZi}
			if tc.DayZhi != "" {
				chart.RiZhi = mustOracleZhi(t, tc.DayZhi)
			}
			targetType := YaoType(ShaoYin)
			if tc.TargetIsMoving {
				targetType = LaoYang
			}
			targetZhi := ganzhi.ZhiYin
			if tc.TargetZhi != "" {
				targetZhi = mustOracleZhi(t, tc.TargetZhi)
			}
			chart.Lines[0] = Line{Position: 1, Type: targetType, Zhi: targetZhi, Wuxing: oracleWuxing(tc.TargetWuxing)}
			if tc.MovingTombPosition > 0 {
				chart.DongYao = append(chart.DongYao, tc.MovingTombPosition)
				chart.Lines[tc.MovingTombPosition-1] = Line{
					Position: tc.MovingTombPosition, Type: LaoYang,
					Zhi: mustOracleZhi(t, tc.MovingTombZhi), Wuxing: ganzhi.WxTu,
				}
			}
			if tc.TargetIsMoving {
				chart.DongYao = append(chart.DongYao, 1)
				changed := targetZhi
				if tc.ChangedZhi != "" {
					changed = mustOracleZhi(t, tc.ChangedZhi)
				}
				chart.BianLines[0] = Line{Position: 1, Type: ShaoYin, Zhi: changed, Wuxing: ganzhi.ZhiWuxing(changed)}
			}
			computeLineDerived(&chart)
			if len(chart.Lines[0].MuKuTypes) != len(tc.ExpectedTypes) {
				t.Fatalf("mu_ku_types = %#v, want %#v", chart.Lines[0].MuKuTypes, tc.ExpectedTypes)
			}
			for index := range tc.ExpectedTypes {
				if chart.Lines[0].MuKuTypes[index] != tc.ExpectedTypes[index] {
					t.Fatalf("mu_ku_types = %#v, want %#v", chart.Lines[0].MuKuTypes, tc.ExpectedTypes)
				}
			}
		})
	}
	for _, tc := range doc.SelectionCases {
		t.Run(tc.ID, func(t *testing.T) {
			typ, err := ParseYongShen(tc.YongShen)
			if err != nil {
				t.Fatal(err)
			}
			chart := Chart{}
			for _, item := range tc.Lines {
				if item.Position < 1 || item.Position > 6 {
					t.Fatalf("invalid position %d", item.Position)
				}
				zhi, err := ganzhi.ParseZhi(item.Zhi)
				if err != nil {
					t.Fatal(err)
				}
				ws, err := ganzhi.ParseWangShuai(item.WangShuai)
				if err != nil {
					t.Fatal(err)
				}
				chart.Lines[item.Position-1] = Line{
					Position: item.Position, Zhi: zhi, LiuQin: parseOracleLiuQin(t, item.LiuQin),
					XunKong: item.XunKong, YuePo: item.YuePo, MuKu: item.MuKu,
					DongSelf: item.DongSelf, ShiYing: item.ShiYing,
				}
				chart.WangShuai[item.Position-1] = ws
			}
			if got := chart.findYongShen(typ); got != tc.ExpectedPosition {
				t.Fatalf("position = %d, want %d", got, tc.ExpectedPosition)
			}
		})
	}
	for _, tc := range doc.PatternCases {
		t.Run(tc.ID, func(t *testing.T) {
			typ, err := ParseYongShen(tc.Target)
			if err != nil {
				t.Fatal(err)
			}
			chart := Chart{DongYao: tc.MovingPositions}
			if len(tc.Lines) > 0 {
				for _, item := range tc.Lines {
					zhi, err := ganzhi.ParseZhi(item.Zhi)
					if err != nil {
						t.Fatal(err)
					}
					ws, err := ganzhi.ParseWangShuai(item.WangShuai)
					if err != nil {
						t.Fatal(err)
					}
					chart.Lines[item.Position-1] = Line{
						Position: item.Position, Type: ShaoYin, Zhi: zhi,
						LiuQin: parseOracleLiuQin(t, item.LiuQin), YuePo: item.YuePo,
					}
					chart.WangShuai[item.Position-1] = ws
				}
				for index, relation := range tc.DayRelations {
					chart.DayRelations[index] = DayRelation{Relations: relation.Relations}
				}
				patterns := computeYuePo(&chart, typ)
				if len(patterns) != 1 || patterns[0].SubType != tc.ExpectedSubtype || patterns[0].IsTrue != tc.ExpectedIsTrue {
					t.Fatalf("patterns = %#v, want %s", patterns, tc.ExpectedSubtype)
				}
				return
			}
			for index, name := range tc.LinesZhi {
				from, err := ganzhi.ParseZhi(name)
				if err != nil {
					t.Fatal(err)
				}
				to, err := ganzhi.ParseZhi(tc.ChangedLinesZhi[index])
				if err != nil {
					t.Fatal(err)
				}
				lineType := YaoType(ShaoYin)
				changedType := YaoType(ShaoYin)
				for _, position := range tc.MovingPositions {
					if position == index+1 {
						lineType = LaoYang
						changedType = ShaoYin
					}
				}
				chart.Lines[index] = Line{Position: index + 1, Type: lineType, Zhi: from, LiuQin: parseOracleLiuQin(t, tc.Target)}
				chart.BianLines[index] = Line{Position: index + 1, Type: changedType, Zhi: to}
			}
			patterns := computeFanYin(&chart, typ)
			if tc.ExpectedAbsentSubtype != "" {
				for _, pattern := range patterns {
					if pattern.SubType == tc.ExpectedAbsentSubtype {
						t.Fatalf("subtype %q must be absent: %#v", tc.ExpectedAbsentSubtype, patterns)
					}
				}
				return
			}
			found := false
			for _, pattern := range patterns {
				if pattern.SubType == tc.ExpectedSubtype && pattern.IsTrue == tc.ExpectedIsTrue {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("patterns = %#v, want %s", patterns, tc.ExpectedSubtype)
			}
		})
	}
	for _, tc := range doc.ComputedCases {
		t.Run(tc.ID, func(t *testing.T) {
			value, err := time.Parse(time.RFC3339, tc.Time)
			if err != nil {
				t.Fatal(err)
			}
			typ, err := ParseYongShen(tc.YongShen)
			if err != nil {
				t.Fatal(err)
			}
			chart := ComputeChart(tianwen.SolarTime(value), typ, tc.Yaos)
			found := false
			for _, pattern := range chart.Patterns {
				if string(pattern.Type) == tc.ExpectedPattern.Type &&
					pattern.SubType == tc.ExpectedPattern.SubType &&
					pattern.IsTrue == tc.ExpectedPattern.IsTrue {
					found = true
				}
			}
			if !found {
				t.Fatalf("pattern absent: %#v", chart.Patterns)
			}
		})
	}
	for _, tc := range doc.WholeGuaCases {
		t.Run(tc.ID, func(t *testing.T) {
			value, err := time.Parse(time.RFC3339, tc.Time)
			if err != nil {
				t.Fatal(err)
			}
			chart := ComputeChart(tianwen.SolarTime(value), YongGuanGui, tc.Yaos)
			if chart.Name != tc.ExpectedName {
				t.Fatalf("name = %q, want %q", chart.Name, tc.ExpectedName)
			}
			found := []string{}
			for _, pattern := range chart.Patterns {
				if pattern.Type == PatternChongHe {
					found = append(found, pattern.SubType)
					if !pattern.IsTrue {
						t.Fatalf("whole-gua pattern must be true: %+v", pattern)
					}
				}
			}
			if tc.ExpectedSubtype == "" {
				if len(found) != 0 {
					t.Fatalf("whole-gua chong/he must be absent, got %v", found)
				}
				return
			}
			hasExpected := false
			for _, subtype := range found {
				if subtype == tc.ExpectedSubtype {
					hasExpected = true
				}
			}
			if !hasExpected {
				t.Fatalf("whole-gua subtype = %v, want member [%s]", found, tc.ExpectedSubtype)
			}
		})
	}
	t.Run("whole-gua-closure", func(t *testing.T) {
		want := map[string][]string{}
		for _, name := range doc.WholeGuaClosure.SixClash {
			want[name] = append(want[name], "六冲")
		}
		for _, name := range doc.WholeGuaClosure.SixCombination {
			want[name] = append(want[name], "六合")
		}
		if len(want) != 18 {
			t.Fatalf("closure hexagrams = %d, want 18", len(want))
		}
		for index, meta := range guaTable {
			got := []string{}
			gua := guaIndex(index)
			if gua.isLiuChong() {
				got = append(got, "六冲")
			}
			if gua.isLiuHe() {
				got = append(got, "六合")
			}
			expected := want[meta.Name]
			if expected == nil {
				expected = []string{}
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("%s whole-gua relation = %v, want %v", meta.Name, got, expected)
			}
		}
	})
	for _, tc := range doc.SanheCases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := Chart{}
			ri, err := ganzhi.ParseZhi(tc.RiZhi)
			if err != nil {
				t.Fatal(err)
			}
			yue, err := ganzhi.ParseZhi(tc.YueZhi)
			if err != nil {
				t.Fatal(err)
			}
			chart.RiZhi, chart.YueZhi = ri, yue
			for _, item := range tc.Lines {
				zhi, err := ganzhi.ParseZhi(item.Zhi)
				if err != nil {
					t.Fatal(err)
				}
				chart.Lines[item.Position-1] = Line{Position: item.Position, Zhi: zhi}
			}
			for _, candidate := range computeSanHeCandidates(&chart) {
				if candidate.Element != tc.Element {
					continue
				}
				if candidate.Complete != tc.ExpectedComplete {
					t.Fatalf("complete = %v, want %v: %+v", candidate.Complete, tc.ExpectedComplete, candidate)
				}
				if len(candidate.Missing) != len(tc.ExpectedMissing) {
					t.Fatalf("missing = %v, want %v: %+v", candidate.Missing, tc.ExpectedMissing, candidate)
				}
				for index, want := range tc.ExpectedMissing {
					if candidate.Missing[index] != want {
						t.Fatalf("missing = %v, want %v: %+v", candidate.Missing, tc.ExpectedMissing, candidate)
					}
				}
				return
			}
			t.Fatalf("water candidate absent: %#v", computeSanHeCandidates(&chart))
		})
	}
}
