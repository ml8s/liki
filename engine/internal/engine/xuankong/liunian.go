package xuankong

import (
	"fmt"

	"liki-engine/internal/engine/fengshui"
)

// LiuNianResult is the 流年飞星 result for a year, optionally overlaid on a
// 宅盘 (chart). 全部为确定性计算：流年飞星盘来自共享 fengshui 实现，
// house_overlay 把流年凶星（凶/大凶）落宫对照宅盘该宫三星。
type LiuNianResult struct {
	Year         int                         `json:"year"`
	RuZhong      string                      `json:"ru_zhong"`
	GongWei      []fengshui.AnnualFlyingStar `json:"gong_wei"`
	HouseOverlay []HouseOverlay              `json:"house_overlay,omitempty"`
}

// HouseOverlay highlights a 流年凶星落宫 and the 宅盘 stars in that palace.
type HouseOverlay struct {
	GongNum     int    `json:"gong_num"`
	Star        string `json:"star"`         // 流年星名（如 五黄廉贞）
	StarRating  string `json:"star_rating"`  // 星固有吉凶（凶/大凶）
	PalaceStars string `json:"palace_stars"` // 宅盘该宫 运星/山星/向星
}

// ComputeLiuNian computes the annual flying-star board for a year.
// chart 可选：nil 时只返回流年飞星盘；非 nil 时叠加宅盘凶星落宫对照。
func ComputeLiuNian(year int, chart *Chart) LiuNianResult {
	board := fengshui.ComputeAnnualFlyingStars(year)
	res := LiuNianResult{
		Year:    board.Year,
		RuZhong: board.RuZhong,
		GongWei: board.GongWei,
	}

	if chart == nil {
		return res
	}

	for _, s := range board.GongWei {
		if s.Rating != "凶" && s.Rating != "大凶" {
			continue
		}
		if s.GongNum < 1 || s.GongNum > 9 {
			continue
		}
		pal := chart.Palaces[s.GongNum-1]
		res.HouseOverlay = append(res.HouseOverlay, HouseOverlay{
			GongNum:     s.GongNum,
			Star:        s.XingName,
			StarRating:  s.Rating,
			PalaceStars: fmt.Sprintf("%s/%s/%s", pal.PeriodStar.Name, pal.MountainStar.Name, pal.FacingStar.Name),
		})
	}
	return res
}
