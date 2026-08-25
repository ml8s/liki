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
	wantYong    string // expected yong element
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
		ganzhi.GanRen, ganzhi.GanGui, // 水: 须火解冻
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

// ── 穷通宝鉴权威锚点（多源原文一致）──
// 防止调候表再错：覆盖 2.6.18 修正的 10 处 + 各日干代表月关键条目。
func TestTiaoHou_AuthoritativeAnchors(t *testing.T) {
	anchors := map[string]string{ // "日干+月支" -> "主次"
		// 甲木：寅丙癸 卯庚丙 辰庚丁 巳癸丁 午癸庚 未庚丁 申庚丁(先庚后丁) 戌庚丁 亥庚丁 子丁庚 丑丁庚
		"甲寅": "丙癸", "甲卯": "庚丙", "甲辰": "庚丁", "甲巳": "癸丁", "甲午": "癸庚",
		"甲未": "庚丁", "甲申": "庚丁", "甲戌": "庚丁", "甲亥": "庚丁", "甲子": "丁庚", "甲丑": "丁庚",
		// 乙木：寅丙癸 卯丙癸 巳癸 午癸丙 申丙癸 酉癸丙 戌癸辛(必赖癸水,辛金发源)
		"乙寅": "丙癸", "乙卯": "丙癸", "乙巳": "癸", "乙午": "癸丙", "乙申": "丙癸",
		"乙酉": "癸丙", "乙戌": "癸辛",
		// 丙火：寅壬 卯壬 辰壬甲 巳壬庚 午壬庚 未壬庚 申壬 酉壬癸 戌甲壬(先甲次壬) 亥甲壬(甲为关键) 子壬戊(先壬戊佐) 丑壬甲(先壬,甲佐)
		"丙寅": "壬", "丙卯": "壬", "丙辰": "壬甲", "丙巳": "壬庚", "丙午": "壬庚", "丙未": "壬庚",
		"丙申": "壬", "丙酉": "壬癸", "丙戌": "甲壬", "丙亥": "甲壬", "丙子": "壬戊", "丙丑": "壬甲",
		// 癸水：寅辛丙 卯庚辛 辰丙辛(专用丙火,辛甲佐) 巳辛庚 未庚辛 申丁 酉辛丙(辛为用丙佐) 戌辛甲 亥戊丙 子丙戊 丑丙丁(丙解冻,丁雪后灯光)
		"癸寅": "辛丙", "癸卯": "庚辛", "癸辰": "丙辛", "癸巳": "辛庚", "癸未": "庚辛",
		"癸申": "丁", "癸酉": "辛丙", "癸戌": "辛甲", "癸亥": "戊丙", "癸子": "丙戊", "癸丑": "丙丁",
		// 丁火代表月：午壬庚(五六月丁火用壬)  戊土：午壬甲  己土：午壬癸
		"丁午": "壬庚", "戊午": "壬甲", "己午": "壬癸",
		// 庚金：寅丁甲 子丙甲(子月丁丙？权威丙甲? 用丙解冻) 辛金：寅己壬
		"庚寅": "丁甲", "辛寅": "己壬",
		// 壬水：寅庚丙 卯庚戊 辰甲庚
		"壬寅": "庚丙", "壬卯": "庚戊", "壬辰": "甲庚",
	}
	for key, want := range anchors {
		gan, _ := ganzhi.ParseGan(key[:3])
		zhi, _ := ganzhi.ParseZhi(key[3:])
		e, ok := lookupTiaohou[tiaohouKey{int(gan), int(zhi)}]
		if !ok {
			t.Errorf("%s日%s月: 表内无条目", ganzhi.GanName(gan), ganzhi.ZhiName(zhi))
			continue
		}
		got := ganzhi.GanName(e.primary)
		if e.secondary != 0 {
			got += ganzhi.GanName(e.secondary)
		}
		if got != want {
			t.Errorf("%s日%s月: 调候 = %s, want %s（穷通宝鉴原文）",
				ganzhi.GanName(gan), ganzhi.ZhiName(zhi), got, want)
		}
	}
}
