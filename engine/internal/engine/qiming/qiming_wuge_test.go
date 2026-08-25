package qiming

import "testing"

func TestListViableStrokes_Double(t *testing.T) {
	pairs := ListViableStrokes(singleStrokes(9), 2) // 姚=9

	if len(pairs) == 0 {
		t.Fatal("expected at least one viable pair for 姚 double-name")
	}

	for _, p := range pairs {
		wg := computeWuGeFromStrokes(singleStrokes(9), p.S1, p.S2)
		if !isAuspicious(wg.RenGe.Fortune) {
			t.Errorf("s1=%d s2=%d 人格=%d fortune=%s, want auspicious", p.S1, p.S2, wg.RenGe.Stroke, wg.RenGe.Fortune)
		}
		if !isAuspicious(wg.DiGe.Fortune) {
			t.Errorf("s1=%d s2=%d 地格=%d fortune=%s, want auspicious", p.S1, p.S2, wg.DiGe.Stroke, wg.DiGe.Fortune)
		}
		if !isAuspicious(wg.WaiGe.Fortune) {
			t.Errorf("s1=%d s2=%d 外格=%d fortune=%s, want auspicious", p.S1, p.S2, wg.WaiGe.Stroke, wg.WaiGe.Fortune)
		}
		if !isAuspicious(wg.ZongGe.Fortune) {
			t.Errorf("s1=%d s2=%d 总格=%d fortune=%s, want auspicious", p.S1, p.S2, wg.ZongGe.Stroke, wg.ZongGe.Fortune)
		}

		if p.S2 == 0 {
			t.Errorf("double-name pair should have s2 > 0: %+v", p)
		}
	}

	// Check for known valid pair from earlier naming session
	found := false
	for _, p := range pairs {
		if p.S1 == 8 && p.S2 == 16 {
			found = true
			break
		}
	}
	if !found {
		t.Log("(8,16) not in pairs for 姚 — may still be correct")
	}
}

func TestListViableStrokes_Single(t *testing.T) {
	pairs := ListViableStrokes(singleStrokes(9), 1)

	if len(pairs) == 0 {
		t.Fatal("expected at least one viable pair for single name")
	}

	for _, p := range pairs {
		if p.S2 != 0 {
			t.Errorf("single-name pair should have s2=0: %+v", p)
		}

		wg := computeWuGeFromStrokes(singleStrokes(9), p.S1, p.S2)
		if !isAuspicious(wg.RenGe.Fortune) {
			t.Errorf("s1=%d 人格=%d fortune=%s, want auspicious", p.S1, wg.RenGe.Stroke, wg.RenGe.Fortune)
		}
		if !isAuspicious(wg.DiGe.Fortune) {
			t.Errorf("s1=%d 地格=%d fortune=%s, want auspicious", p.S1, wg.DiGe.Stroke, wg.DiGe.Fortune)
		}
		if !isAuspicious(wg.ZongGe.Fortune) {
			t.Errorf("s1=%d 总格=%d fortune=%s, want auspicious", p.S1, wg.ZongGe.Stroke, wg.ZongGe.Fortune)
		}
	}
}

func TestListViableStrokes_MinSurname(t *testing.T) {
	pairs := ListViableStrokes(singleStrokes(1), 2)
	if pairs == nil {
		t.Error("should return empty slice, not nil")
	}
	for _, p := range pairs {
		wg := computeWuGeFromStrokes(singleStrokes(1), p.S1, p.S2)
		if !allWuGeAuspicious(wg) {
			t.Errorf("pair (s1=%d,s2=%d) not auspicious: %+v", p.S1, p.S2, wg)
		}
	}
}
