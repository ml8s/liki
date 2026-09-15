package xuankong

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFlyingStarMatrixGolden_AllYunAndOppositePairs(t *testing.T) {
	raw, err := os.ReadFile("testdata/flying_star_matrix_golden.json")
	if err != nil {
		t.Fatalf("read flying-star matrix golden: %v", err)
	}
	var doc struct {
		Version int `json:"version"`
		Policy  struct {
			Plate   string `json:"plate"`
			SitFace string `json:"sit_face"`
		} `json:"policy"`
		Scope struct {
			YunCount           int `json:"yun_count"`
			SitFacePairsPerYun int `json:"sit_face_pairs_per_yun"`
			CaseCount          int `json:"case_count"`
		} `json:"scope"`
		Cases []struct {
			ID            string `json:"id"`
			Yun           int    `json:"yun"`
			Year          int    `json:"year"`
			Sit           int    `json:"sit"`
			Face          int    `json:"face"`
			PeriodStars   []int  `json:"period_stars"`
			MountainStars []int  `json:"mountain_stars"`
			FacingStars   []int  `json:"facing_stars"`
			FourSituation string `json:"four_situation"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode flying-star matrix golden: %v", err)
	}
	if doc.Version != 1 || doc.Policy.Plate != "lower_plate_forward_direction" ||
		doc.Policy.SitFace != "opposite_24_mountain" || len(doc.Cases) != doc.Scope.CaseCount {
		t.Fatalf("invalid fixture metadata: %+v", doc)
	}
	if doc.Scope.YunCount != 9 || doc.Scope.SitFacePairsPerYun != 12 || len(doc.Cases) != 108 {
		t.Fatalf("coverage = %+v, cases=%d", doc.Scope, len(doc.Cases))
	}

	for _, golden := range doc.Cases {
		t.Run(golden.ID, func(t *testing.T) {
			chart := computeChart(golden.Sit, golden.Face, golden.Year)
			if chart.Yun.YunNumber != golden.Yun {
				t.Fatalf("yun = %d, want %d", chart.Yun.YunNumber, golden.Yun)
			}
			if chart.FourSituation.Name != golden.FourSituation {
				t.Fatalf("four situation = %s, want %s", chart.FourSituation.Name, golden.FourSituation)
			}
			for i, palace := range chart.Palaces {
				if palace.PeriodStar.Number != golden.PeriodStars[i] ||
					palace.MountainStar.Number != golden.MountainStars[i] ||
					palace.FacingStar.Number != golden.FacingStars[i] {
					t.Fatalf("palace %d stars = %d/%d/%d, want %d/%d/%d",
						i+1,
						palace.PeriodStar.Number, palace.MountainStar.Number, palace.FacingStar.Number,
						golden.PeriodStars[i], golden.MountainStars[i], golden.FacingStars[i],
					)
				}
			}
		})
	}
}
