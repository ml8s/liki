package ziwei

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 十干四化对照《紫微斗数全书》通行本（iztro 口径）。
// 注意壬干化科 = 左辅（不是中州派的天府）——曾混用，已固化防回归。
func TestSiHua_StandardPassbookTable(t *testing.T) {
	want := map[ganzhi.Gan][4]starIndex{
		ganzhi.GanJia:  {LianZhen, PoJun, WuQu, TaiYang},
		ganzhi.GanYi:   {TianJi, TianLiang, ZiWei, TaiYin},
		ganzhi.GanBing: {TianTong, TianJi, WenChang, LianZhen},
		ganzhi.GanDing: {TaiYin, TianTong, TianJi, JuMen},
		ganzhi.GanWu:   {TanLang, TaiYin, YouBi, TianJi},
		ganzhi.GanJi:   {WuQu, TanLang, TianLiang, WenQu},
		ganzhi.GanGeng: {TaiYang, WuQu, TaiYin, TianTong},
		ganzhi.GanXin:  {JuMen, TaiYang, WenQu, WenChang},
		ganzhi.GanRen:  {TianLiang, ZiWei, ZuoFu, WuQu},
		ganzhi.GanGui:  {PoJun, JuMen, TaiYin, TanLang},
	}
	for g, stars := range want {
		got := siHuaTable[g]
		if got != stars {
			t.Errorf("干 %d 四化 = %v, want %v（通行本）", g, got, stars)
		}
	}
}