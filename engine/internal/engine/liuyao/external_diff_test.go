package liuyao

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

type externalDiffCases struct {
	SchemaVersion string `json:"schema_version"`
	Cases         []struct {
		CaseID string `json:"case_id"`
		Input  struct {
			SolarTime string `json:"solar_time"`
			Yaos      [6]int `json:"yaos"`
			YongShen  string `json:"yong_shen"`
		} `json:"input"`
		Expected struct {
			Name         string   `json:"name"`
			BenGua       string   `json:"ben_gua"`
			BianGua      *string  `json:"bian_gua"`
			Palace       string   `json:"palace"`
			PalaceWuxing string   `json:"palace_wuxing"`
			DongYao      []int    `json:"dong_yao"`
			DayGan       string   `json:"day_gan"`
			DayZhi       string   `json:"day_zhi"`
			MonthBranch  string   `json:"month_branch"`
			XunKong      []string `json:"xun_kong"`
			LineBranches []string `json:"line_branches"`
			LineLiuQin   []string `json:"line_liu_qin"`
			ShiPosition  int      `json:"shi_position"`
			YingPosition int      `json:"ying_position"`
			YongPosition int      `json:"yong_shen_position"`
		} `json:"expected"`
		Scope string `json:"conclusion_scope"`
	} `json:"cases"`
}

func TestExternalDiffFixtures_HardFacts(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/liuyao/external_diff_cases.json")
	if err != nil {
		t.Skipf("external diff fixtures unavailable: %v", err)
	}
	var cases externalDiffCases
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode fixtures: %v", err)
	}
	for _, tc := range cases.Cases {
		t.Run(tc.CaseID, func(t *testing.T) {
			local, err := time.Parse(time.RFC3339, tc.Input.SolarTime)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.GregorianToSolar(local, 116.4, 8)
			yong, err := ParseYongShen(tc.Input.YongShen)
			if err != nil {
				t.Fatal(err)
			}
			got := ComputeChart(st, yong, tc.Input.Yaos)

			if got.Name != tc.Expected.Name ||
				got.Palace != tc.Expected.Palace ||
				got.PalaceWuxing.String() != tc.Expected.PalaceWuxing ||
				got.RiGan.String() != tc.Expected.DayGan ||
				got.RiZhi.String() != tc.Expected.DayZhi ||
				got.YueZhi.String() != tc.Expected.MonthBranch ||
				got.YongShen.Position != tc.Expected.YongPosition {
				t.Fatalf("core hard facts differ: got name=%s palace=%s/%s day=%s%s month=%s yong=%d",
					got.Name, got.Palace, got.PalaceWuxing.String(), got.RiGan.String(),
					got.RiZhi.String(), got.YueZhi.String(), got.YongShen.Position)
			}
			if len(got.DongYao) != len(tc.Expected.DongYao) ||
				len(got.Lines) != len(tc.Expected.LineBranches) ||
				len(got.XunKong) != len(tc.Expected.XunKong) {
				t.Fatalf("array lengths differ: dong=%v xunkong=%v", got.DongYao, got.XunKong)
			}
			for i, line := range got.Lines {
				if line.Zhi.String() != tc.Expected.LineBranches[i] ||
					line.LiuQin.String() != tc.Expected.LineLiuQin[i] ||
					line.ShiYing != shiYingAt(i+1, tc.Expected.ShiPosition, tc.Expected.YingPosition) {
					t.Fatalf("line %d differs: %+v", i+1, line)
				}
			}
		})
	}
}

func shiYingAt(position, world, other int) string {
	switch position {
	case world:
		return "世"
	case other:
		return "应"
	default:
		return ""
	}
}
