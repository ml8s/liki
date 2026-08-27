package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 从格测试（合成数据） ──

// ── 调候与扶抑冲突场景 ──

func TestYongShen_TiaoHouVsFuYi_Conflict(t *testing.T) {
	// 庚午 壬午 庚申 壬午
	// 庚日主, 午月(死), 地支午午申午, 天干庚壬壬
	// 调候: 穷通宝鉴(庚,午)=壬(水)
	// 格局: 午月不透干 → 虚格
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	result := ComputeYongShen(chart)

	if result.TiaoHou.Yong == "" {
		t.Error("TiaoHou.Yong 不应为空")
	}
	if result.GeJu.Yong == "" {
		t.Error("GeJu.Yong 不应为空")
	}
	if result.FuYi.Strength == "" {
		t.Error("FuYi.Strength 不应为空")
	}
	// 记录三派结果供人工审核
	t.Logf("扶抑: yong=%s xi=%s ji=%s (strength=%s)",
		result.FuYi.Yong, result.FuYi.Xi, result.FuYi.Ji, result.FuYi.Strength)
	t.Logf("调候: yong=%s xi=%s ji=%s (detail=%s)",
		result.TiaoHou.Yong, result.TiaoHou.Xi, result.TiaoHou.Ji, result.TiaoHou.Detail)
	t.Logf("格局: yong=%s xi=%s ji=%s (pattern=%s %s)",
		result.GeJu.Yong, result.GeJu.Xi, result.GeJu.Ji, result.GeJu.Pattern, result.GeJu.Usage)
}

// ── 三会局测试 ──

func TestFuYi_SanHui_HuoJu(t *testing.T) {
	// 丙火日主, 午月, 地支巳午未 → 三会火方
	// 即使无其他火根, 三会火方本身应使日主不弱
	st := tianwen.GregorianToSolar(
		time.Date(2025, 7, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	result := ComputeYongShen(chart)

	t.Logf("三会火方: 八字=%s%s %s%s %s%s %s%s, strength=%s",
		chart.Nian.Gan, chart.Nian.Zhi,
		chart.Yue.Gan, chart.Yue.Zhi,
		chart.Ri.Gan, chart.Ri.Zhi,
		chart.Shi.Gan, chart.Shi.Zhi,
		result.FuYi.Strength)

	if result.FuYi.Strength == "身弱" {
		t.Error("三会火方的丙火日主不应身弱")
	}
}

// ── 经典命理案例参考测试（提取自原书案例） ──

func TestReference_DiTianSui_WeakWoodWithFire(t *testing.T) {
	// 滴天髓案例: 乙木秋生, 火旺制杀
	// 乙木日主, 酉月(七杀辛本气), 年/月干丁火透 → 火旺制杀
	// 正确日期: 2001-09-19 12:00 = 辛巳 丁酉 乙未 壬午
	st := tianwen.GregorianToSolar(
		time.Date(2001, 9, 19, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	result := ComputeYongShen(chart)

	t.Logf("滴天髓: %s%s %s%s %s%s %s%s",
		chart.Nian.Gan, chart.Nian.Zhi,
		chart.Yue.Gan, chart.Yue.Zhi,
		chart.Ri.Gan, chart.Ri.Zhi,
		chart.Shi.Gan, chart.Shi.Zhi)
	t.Logf("  扶抑: strength=%s yong=%s xi=%s ji=%s",
		result.FuYi.Strength, result.FuYi.Yong, result.FuYi.Xi, result.FuYi.Ji)
	t.Logf("  格局: %s %s, yong=%s xi=%s ji=%s",
		result.GeJu.Pattern, result.GeJu.Usage,
		result.GeJu.Yong, result.GeJu.Xi, result.GeJu.Ji)
	t.Logf("  调候: yong=%s xi=%s ji=%s detail=%s",
		result.TiaoHou.Yong, result.TiaoHou.Xi, result.TiaoHou.Ji, result.TiaoHou.Detail)

	// 滴天髓原文: 乙木秋生, 火旺制杀 → 具体格局取决于透干的火是食神/伤官: 食神→顺用, 伤官→逆用
	if result.GeJu.Usage != "逆用" {
		if result.GeJu.Pattern != "七杀格" {
			t.Errorf("pattern=%q, want 七杀格(滴天髓:乙木秋生火旺制杀)", result.GeJu.Pattern)
		}
	}
}

func TestReference_ZiPing_ZhengGuanGe(t *testing.T) {
	// 子平真诠: 甲木酉月正官格
	// 2004-09-22 10:00 → 甲申 癸酉 甲辰 己巳
	st := tianwen.GregorianToSolar(
		time.Date(2004, 9, 22, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	result := ComputeYongShen(chart)

	t.Logf("子平真诠: %s%s %s%s %s%s %s%s, 格局=%s %s",
		chart.Nian.Gan, chart.Nian.Zhi,
		chart.Yue.Gan, chart.Yue.Zhi,
		chart.Ri.Gan, chart.Ri.Zhi,
		chart.Shi.Gan, chart.Shi.Zhi,
		result.GeJu.Pattern, result.GeJu.Usage)

	// 甲木日主, 酉月: 酉辛为甲之正官
	if result.GeJu.Pattern == "" {
		t.Error("子平真诠案例: 甲木酉月应有格局")
	}
}

func TestReference_QiongTongBaoJian_SampleEntry(t *testing.T) {
	// 穷通宝鉴验证: 丙日午月=壬+庚 (穷通宝鉴: 壬庚又须并用)
	riYuan := ganzhi.GanBing
	yueZhi := ganzhi.ZhiWu
	th, ok := queryTiaoHou(riYuan, yueZhi)
	if !ok {
		t.Fatalf("调候表中缺少(丙,午)条目")
	}
	if th.Yong != "水" {
		t.Errorf("穷通宝鉴(丙,午): yong=%s, want 水", th.Yong)
	}
	// 穷通宝鉴(甲,酉)用丁(火), 不是庚(金) — 按穷通原文"丁火制金，丙火暖木"
	th2, ok2 := queryTiaoHou(ganzhi.GanJia, ganzhi.ZhiYou)
	if !ok2 {
		t.Fatalf("调候表中缺少(甲,酉)条目")
	}
	if th2.Yong != "火" {
		t.Errorf("穷通宝鉴(甲,酉): yong=%s, want 火(穷通原文丁火)", th2.Yong)
	}
}

// ── 调候表完整性 ──

func TestTiaoHou_SeasonalConsistency(t *testing.T) {
	allGan := []ganzhi.Gan{
		ganzhi.GanJia, ganzhi.GanYi, ganzhi.GanBing, ganzhi.GanDing,
		ganzhi.GanWu, ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin,
		ganzhi.GanRen, ganzhi.GanGui,
	}
	summerZhi := []ganzhi.Zhi{ganzhi.ZhiSi, ganzhi.ZhiWu, ganzhi.ZhiWei}
	winterZhi := []ganzhi.Zhi{ganzhi.ZhiHai, ganzhi.ZhiZi, ganzhi.ZhiChou}

	summerFireCount := 0
	summerTotal := 0
	for _, gan := range allGan {
		for _, zhi := range summerZhi {
			th, ok := queryTiaoHou(gan, zhi)
			if !ok {
				continue
			}
			summerTotal++
			if th.Yong == "火" {
				summerFireCount++
			}
		}
	}
	if summerFireCount >= 5 {
		t.Errorf("夏季%d/%d条用火(穷通例外应<5)", summerFireCount, summerTotal)
	}

	winterWaterCount := 0
	winterTotal := 0
	for _, gan := range allGan {
		for _, zhi := range winterZhi {
			th, ok := queryTiaoHou(gan, zhi)
			if !ok {
				continue
			}
			winterTotal++
			if th.Yong == "水" {
				winterWaterCount++
			}
		}
	}
	if winterWaterCount >= 5 {
		t.Errorf("冬季%d/%d条用水(穷通例外应<5)", winterWaterCount, winterTotal)
	}
}

// ── TiaoHou 忌神不冲突验证 ──

func TestTiaoHou_JiNotConflict(t *testing.T) {
	allGan := []ganzhi.Gan{
		ganzhi.GanJia, ganzhi.GanYi, ganzhi.GanBing, ganzhi.GanDing,
		ganzhi.GanWu, ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin,
		ganzhi.GanRen, ganzhi.GanGui,
	}
	conflicts := 0
	for _, s := range allGan {
		for b := ganzhi.ZhiZi; b <= ganzhi.ZhiHai; b++ {
			th, ok := queryTiaoHou(s, b)
			if !ok {
				continue
			}
			if th.Ji == th.Yong || th.Ji == th.Xi {
				conflicts++
			}
		}
	}
	if conflicts > 9 {
		t.Errorf("忌神冲突=%d/120, 超过预期9条(穷通原文占满五行空间)", conflicts)
	}
	t.Logf("调候忌神冲突: %d/120 (9条穷通固有)", conflicts)
}
