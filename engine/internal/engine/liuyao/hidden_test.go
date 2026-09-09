package liuyao

import "testing"

func TestComputeHiddenLines_CoversAllMissingRelatives(t *testing.T) {
	chart := factsChart(t, [6]int{7, 7, 7, 7, 7, 7})
	visible := map[string]bool{}
	for _, line := range chart.Lines {
		visible[line.LiuQin.String()] = true
	}
	for _, hidden := range chart.HiddenLines {
		if visible[hidden.LiuQin] {
			t.Fatalf("visible relative %s emitted as hidden", hidden.LiuQin)
		}
		if hidden.FeiFuRelation == "" || hidden.ExitCondition == "" {
			t.Fatalf("incomplete hidden line: %+v", hidden)
		}
	}
}

func TestBranchRelations_MarksTargets(t *testing.T) {
	chart := factsChart(t, [6]int{7, 7, 7, 7, 7, 7})
	for _, fact := range chart.BranchRelationFacts {
		if len(fact.Relations) == 0 || fact.LeftPosition == fact.RightPosition {
			t.Fatalf("invalid branch relation: %+v", fact)
		}
	}
}
