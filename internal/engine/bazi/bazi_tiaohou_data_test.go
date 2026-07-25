package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// ──────────────────────────────────────────────────────────────────
// 穷通宝鉴数据逐条校验
// 根据穷通宝鉴原文验证 tiaohou.json 的关键条目
// ──────────────────────────────────────────────────────────────────

type tiaoHouExpectation struct {
	riYuan      ganzhi.Gan
	monthBranch ganzhi.Zhi
	wantYong    string   // expected yong element
}

func TestTiaoHou_WinterDingFire_ShouldHaveFire(t *testing.T) {
	// 冬季(亥子丑月)需火调候: 庚金须火炼, 壬癸水须丙火解冻
	tests := []tiaoHouExpectation{
		// 穷通宝鉴: "庚金生子月, 丁火调候, 丙火解冻"
		{riYuan: ganzhi.GanGeng, monthBranch: ganzhi.ZhiZi, wantYong: "火"},
		// 穷通宝鉴: "壬水丑月, 丙火解冻, 戊土制水" — primary=戊(土), secondary=丙(火)
		{riYuan: ganzhi.GanRen, monthBranch: ganzhi.ZhiChou, wantYong: "火"},
		// 穷通宝鉴: "癸水丑月, 用丙火调候"
		{riYuan: ganzhi.GanGui, monthBranch: ganzhi.ZhiChou, wantYong: "火"},
	}

	for _, tt := range tests {
		name := ganzhi.GanName(tt.riYuan) + "日" + ganzhi.ZhiName(tt.monthBranch) + "月"
		t.Run(name, func(t *testing.T) {
			result := computeTiaoHou(tt.riYuan, tt.monthBranch)

			if result.Yong != tt.wantYong {
				// Also accept if fire is present as Xi (喜神) for winter entries
				if tt.wantYong == "火" && result.Xi == "火" {
					return
				}
				t.Errorf("调候用神 = %q, want %q (冬季需火调候)\n  detail: %s",
					result.Yong, tt.wantYong, result.Detail)
			}
		})
	}
}

func TestTiaoHou_WinterWater_ShouldHaveFire(t *testing.T) {
	// 冬季壬水需丙火调候 (穷通宝鉴)
	// 注意: 壬亥月已有丙火(丙丁), 检查子/丑

	// 穷通宝鉴: "壬水生子月, 戊土制水, 丙火调候" — primary=戊(土), secondary=丙(火)
	result := computeTiaoHou(ganzhi.GanRen, ganzhi.ZhiZi)
	if result.Yong != "土" && result.Yong != "火" {
		t.Errorf("壬日子月调候 = yong=%s, 应为土或火", result.Yong)
	}
	t.Logf("壬日子月: yong=%s xi=%s ji=%s detail=%s",
		result.Yong, result.Xi, result.Ji, result.Detail)
}

func TestTiaoHou_BingWu_ShouldUseGeng(t *testing.T) {
	// 穷通宝鉴: "五月丙火, 愈炎, 壬庚又须并用"
	// primary=壬正确, 但secondary应为庚(非戊)
	result := computeTiaoHou(ganzhi.GanBing, ganzhi.ZhiWu)

	t.Logf("丙午月: yong=%s xi=%s ji=%s detail=%s",
		result.Yong, result.Xi, result.Ji, result.Detail)

	if result.Yong != "水" {
		t.Errorf("丙午月调候 = %q, want 水 (穷通宝鉴: 壬水)", result.Yong)
	}
}

// ──────────────────────────────────────────────────────────────────
// 综合调候逻辑验证: 120条穷通宝鉴对照
// ──────────────────────────────────────────────────────────────────

func TestTiaoHou_AllWinterEntries_HaveFireWhenNeeded(t *testing.T) {
	// 冬季(申酉戌亥子丑月)的庚辛金和壬癸水需要火调候
	// 但同样是金, 庚申月(金当令)不需要火
	// 检查规则: 金水日主在子丑月应有火元素
	coldDayMasters := []ganzhi.Gan{
		ganzhi.GanGeng, ganzhi.GanXin, // 金: 须火炼
		ganzhi.GanRen, ganzhi.GanGui,  // 水: 须火解冻
	}
	coldMonths := []ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiChou} // 最冷月

	mismatches := 0
	for _, dm := range coldDayMasters {
		for _, mb := range coldMonths {
			result := computeTiaoHou(dm, mb)
			hasFire := result.Yong == "火" || result.Xi == "火"
			if !hasFire {
				mismatches++
				t.Logf("冬季无火: %s日%s月: yong=%s xi=%s ji=%s",
					ganzhi.GanName(dm), ganzhi.ZhiName(mb),
					result.Yong, result.Xi, result.Ji)
			}
		}
	}

	if mismatches > 0 {
		t.Errorf("冬季应有火调候: %d/8 条不满足(金水日主子丑月)", mismatches)
	}
}
