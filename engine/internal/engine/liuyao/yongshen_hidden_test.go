package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// Regression for feedback: 天风姤 has no visible 妻财. A 妻财 appearing in the
// changed layer must not be promoted to the primary 用神 and must not inherit
// the first visible line's state.
func TestComputeChart_HiddenYongShenDoesNotUseBianOrFlyingLine(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	st := tianwen.SolarTime(time.Date(2026, 9, 9, 17, 28, 0, 0, cst))
	chart := ComputeChart(st, YongQiCai, [6]int{8, 7, 9, 9, 7, 7})

	if chart.Name != "天风姤" || chart.Palace != "乾" {
		t.Fatalf("chart = %s/%s, want 天风姤/乾", chart.Name, chart.Palace)
	}
	if chart.YongShen.Position != 0 || !chart.YongShen.IsHidden {
		t.Fatalf("yong_shen = %+v, want hidden position 0", chart.YongShen)
	}
	if chart.YongShen.FuShen == nil {
		t.Fatal("fu_shen is nil")
	}
	if chart.YongShen.FuShen.Position != 2 ||
		chart.YongShen.FuShen.LiuQin.String() != "妻财" ||
		chart.YongShen.FuShen.Zhi != "寅" {
		t.Fatalf("fu_shen = %+v, want position 2/妻财/寅", chart.YongShen.FuShen)
	}
	if chart.YongShen.WangShuai != "死" {
		t.Errorf("wang_shuai = %q, want 死 from fu-shen branch 寅", chart.YongShen.WangShuai)
	}
	if chart.YongShen.LiuShou.String() != "青龙" {
		t.Errorf("liu_shou = %q, want omitted; fu-shen must not inherit flying liu-shou", chart.YongShen.LiuShou)
	}
	if chart.ForceChain == nil || !chart.ForceChain.IsHidden {
		t.Fatalf("force_chain = %+v, want hidden force chain", chart.ForceChain)
	}
	if chart.ForceChain.YongElement != "木" || chart.ForceChain.Position != 2 {
		t.Errorf("force_chain = %+v, want yong=wood at fu position 2", chart.ForceChain)
	}
}
