package xuankong

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDomainOracle_XuankongCore(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/fengshui_core.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Periods []struct {
			Year      int    `json:"year"`
			Yuan      string `json:"yuan"`
			YunNumber int    `json:"yun_number"`
			StartYear int    `json:"start_year"`
			EndYear   int    `json:"end_year"`
		} `json:"san_yuan_yun_periods"`
		FourSituations []struct {
			CaseID          string `json:"case_id"`
			Sit             int    `json:"sit_mountain"`
			Face            int    `json:"face_mountain"`
			Year            int    `json:"year"`
			WangShan        bool   `json:"wang_shan"`
			WangXiang       bool   `json:"wang_xiang"`
			DoubleSit       bool   `json:"double_star_sit"`
			DoubleFace      bool   `json:"double_star_face"`
			UphillDownwater bool   `json:"uphill_downwater"`
		} `json:"xuankong_four_situation_anchors"`
		InvalidPairs []struct {
			Sit  int `json:"sit"`
			Face int `json:"face"`
		} `json:"invalid_sit_face_pairs"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	for _, want := range doc.Periods {
		got := ComputeSanYuanYun(want.Year)
		if got.Yuan != want.Yuan || got.YunNumber != want.YunNumber || got.StartYear != want.StartYear || got.EndYear != want.EndYear {
			t.Errorf("year %d = %s/%d/%d-%d, want %s/%d/%d-%d", want.Year, got.Yuan, got.YunNumber, got.StartYear, got.EndYear, want.Yuan, want.YunNumber, want.StartYear, want.EndYear)
		}
	}
	for _, want := range doc.FourSituations {
		got := computeChart(want.Sit, want.Face, want.Year)
		if got.WangShan != want.WangShan || got.WangXiang != want.WangXiang || got.ShanXing != want.DoubleSit ||
			got.XiangXing != want.DoubleFace || got.XiaShui != want.UphillDownwater {
			t.Errorf("%s = %+v, want %+v", want.CaseID, got, want)
		}
	}
	for _, pair := range doc.InvalidPairs {
		if got := computeChart(pair.Sit, pair.Face, 2026); got.Yun.Year != 0 || got.Palaces[4].PeriodStar.Number != 0 {
			t.Errorf("invalid pair %d/%d should return empty chart", pair.Sit, pair.Face)
		}
	}
}
