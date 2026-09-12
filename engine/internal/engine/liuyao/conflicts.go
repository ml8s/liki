package liuyao

// Conflict is an engine-owned contradictory signal that interpretation must
// resolve instead of silently choosing one side.
type Conflict struct {
	ID     string   `json:"id"`
	Reason string   `json:"reason"`
	Facts  []string `json:"facts,omitempty"`
}

func computeConflicts(chart *Chart) []Conflict {
	if chart == nil {
		return nil
	}
	relations := make([]string, 0, len(chart.DongYaoRelations))
	has := map[DongYaoRelationType]bool{}
	for _, relation := range chart.DongYaoRelations {
		for _, item := range relation.Relations {
			relations = append(relations, string(item))
			has[item] = true
		}
	}

	var conflicts []Conflict
	if has[RelationShengYong] && has[RelationKeYong] {
		conflicts = append(conflicts, Conflict{
			ID:     "moving-support-opposition",
			Reason: "动爻同时存在生用与克用信号",
			Facts:  append([]string(nil), relations...),
		})
	}
	if has[RelationShengYuan] && has[RelationKeYuan] {
		conflicts = append(conflicts, Conflict{
			ID:     "yuanshen-support-opposition",
			Reason: "原神同时受到生扶与克制信号",
			Facts:  append([]string(nil), relations...),
		})
	}
	strong := chart.YongShen.WangShuai == "旺" || chart.YongShen.WangShuai == "相"
	if strong && chart.YongShen.XunKong {
		conflicts = append(conflicts, Conflict{
			ID:     "strong-but-void",
			Reason: "用神旺相与旬空并存，须辨真假空",
		})
	}
	if strong && chart.YongShen.YuePo {
		conflicts = append(conflicts, Conflict{
			ID:     "strong-but-break",
			Reason: "用神旺相与月破并存，须辨真假破",
		})
	}
	return conflicts
}
