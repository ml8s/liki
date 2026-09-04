"""单元测试：factors 核心算子（现/透/藏/得令/克/旺/弱/缺 + 流年大运算子）。"""
import unittest

from _helpers import mock_base_context
from operators_liunian import _liu_op
from operators_natal import _op


class TestOperators(unittest.TestCase):
    def test_现(self):
        base = mock_base_context(正财={"count": 1, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], "male", base), 1)
        base = mock_base_context(正财={"count": 0, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], "male", base), 0)

    def test_透(self):
        base = mock_base_context(正财={"tou_gan": True, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], "male", base), 1)
        base = mock_base_context(正财={"tou_gan": False, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], "male", base), 0)

    def test_得令(self):
        base = mock_base_context(正财={"de_ling": True, "wuxing": "土"})
        self.assertEqual(_op("得令", ["正财"], "male", base), 1)

    def test_克_五行生克(self):
        # 财（土）克印（水）——土克水 → 1
        base = mock_base_context(正财={"wuxing": "土", "count": 1}, 正印={"wuxing": "水", "count": 1})
        self.assertEqual(_op("克", ["财星", "印星"], "male", base), 1)

    def test_克_不克(self):
        # 财（土）不克官（木）——土不克木 → 0
        base = mock_base_context(正财={"wuxing": "土", "count": 1}, 正官={"wuxing": "木", "count": 1})
        self.assertEqual(_op("克", ["财星", "官杀"], "male", base), 0)


class TestOperatorsExtended(unittest.TestCase):
    """补测核心算子：藏/有根/旺/弱/缺。"""

    def test_藏(self):
        base = mock_base_context(正印={"cang_zhi": True, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], "male", base), 1)
        base = mock_base_context(正印={"cang_zhi": False, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], "male", base), 0)

    def test_有根(self):
        base = mock_base_context(正财={"has_root": True, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], "male", base), 1)
        base = mock_base_context(正财={"has_root": False, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], "male", base), 0)

    def test_旺_五行直读(self):
        base = mock_base_context()
        base["wuxing"] = {"wang_shuai": {"木": "旺", "水": "休"}}
        self.assertEqual(_op("旺", ["木"], "male", base), 1)
        self.assertEqual(_op("旺", ["水"], "male", base), 0)

    def test_弱_十神三条件(self):
        # 失令 + 不透 + 无根 → 弱
        base = mock_base_context(正财={"wuxing": "土", "tou_gan": False, "has_root": False})
        base["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], "male", base), 1)
        # 透干则不算弱
        base = mock_base_context(正财={"wuxing": "土", "tou_gan": True, "has_root": False})
        base["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], "male", base), 0)

    def test_缺_五行(self):
        # 缺[土]：五行 counts 中土=0 → 1
        base = mock_base_context()
        base["wuxing"] = {"count": {"木": 2, "火": 1, "土": 0, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], "male", base), 1)
        base = mock_base_context()
        base["wuxing"] = {"count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], "male", base), 0)


class TestDaYunOps_YearRange(unittest.TestCase):
    """2.6.15 大运公历年段算子（start_year/end_year 直判，免虚岁换算）。"""

    def _fac(self):
        f = mock_base_context()
        f["dayun_steps"] = [
            {"name": "丁卯", "shi_shen": "偏印运", "start_year": 1990, "end_year": 2000},
            {"name": "戊辰", "shi_shen": "正财运", "start_year": 2001, "end_year": 2010},
            {"name": "己巳", "shi_shen": "比肩运", "start_year": 2011, "end_year": 2020},
        ]
        return f

    def test_大运窗口流年_年份段内(self):
        ctx = {"year": 2005, "liunian": {"nian_zhi": "酉", "nian_gan": "乙", "shi_shen": "偏财"}}
        self.assertEqual(_liu_op("大运窗口流年", ["配偶星"], "male", self._fac(), ctx), 1)

    def test_大运窗口流年_窗口外(self):
        ctx = {"year": 1989, "liunian": {}}
        self.assertEqual(_liu_op("大运窗口流年", ["配偶星"], "male", self._fac(), ctx), 0)

    def test_大运窗口流年_非配偶星大运(self):
        # 丁卯=偏印运（非配偶星）——1995 窗口内但 shi_shen 不含配偶星 → 0
        base = self._fac()
        ctx = {"year": 1995, "liunian": {}}
        self.assertEqual(_liu_op("大运窗口流年", ["配偶星"], "male", base, ctx), 0)

    def test_换运流年_首年(self):
        ctx = {"year": 2001, "liunian": {}}
        self.assertEqual(_liu_op("换运流年", ["配偶星"], "male", self._fac(), ctx), 1)

    def test_换运流年_非首年(self):
        ctx = {"year": 2005, "liunian": {}}
        self.assertEqual(_liu_op("换运流年", ["配偶星"], "male", self._fac(), ctx), 0)

    # 三刑必须三方齐备。
    def _chart(self, nian, yue, ri, shi):
        return {"chart": {"nian": {"zhi": nian}, "yue": {"zhi": yue},
                          "ri": {"zhi": ri}, "shi": {"zhi": shi}}}

    def test_三刑_整组齐备(self):
        # 丑戌未三刑齐备 → 1
        ctx = {"liunian": {"nian_zhi": "丑"}}
        ch = self._chart("戌", "未", "子", "午")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 1)

    def test_三刑_缺一组支(self):
        # 命局丑戌、流年寅——丑戌未缺未 → 0。
        ctx = {"liunian": {"nian_zhi": "寅"}}
        ch = self._chart("戌", "丑", "子", "午")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 0)

    def test_三刑_流年支补全(self):
        # 命局仅 寅巳，流年申 → 寅巳申齐 → 1
        ctx = {"liunian": {"nian_zhi": "申"}}
        ch = self._chart("寅", "巳", "子", "午")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 1)

    def test_三刑_自刑双字(self):
        # 辰辰自刑（辰午酉亥自刑组需同字≥2）→ 1
        ctx = {"liunian": {"nian_zhi": "辰"}}
        ch = self._chart("辰", "子", "子", "午")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 1)

    def test_三刑_自刑单字(self):
        # 仅一个辰（无同字成双）→ 0；自刑不成立。
        ctx = {"liunian": {"nian_zhi": "午"}}
        ch = self._chart("辰", "丑", "寅", "巳")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 0)

    def test_三刑_无刑(self):
        # 寅午辰戌 + 流年亥——丑戌未缺丑/未、寅巳申缺巳/申、辰午酉亥无同字成双 → 0
        ctx = {"liunian": {"nian_zhi": "亥"}}
        ch = self._chart("寅", "午", "辰", "戌")
        self.assertEqual(_liu_op("三刑", ["流年支"], "male", ch, ctx), 0)


class TestLiuNianOps(unittest.TestCase):
    """流年算子边界。

    关注点：三刑必须三方齐备，任一支在场或部分满足不命中——
    其余流年算子同样需要"不满足必为 0"的边界用例。
    """

    def _fac(self, ri_zhi="子", shishen=None, ri_gan="甲"):
        f = mock_base_context(**(shishen or {}))
        f["ri_gan"] = ri_gan
        f["palace_ri"] = {"zhi": ri_zhi}
        f["dayun_steps"] = [
            {"name": "甲子", "shi_shen": "偏财运", "start_year": 1990, "end_year": 2000},
        ]
        return f

    def _chart(self, ri_gan="甲", ri_zhi="子", shi_zhi="午"):
        return {"chart": {"nian": {"gan": "庚", "zhi": "午"},
                          "yue": {"gan": "壬", "zhi": "午"},
                          "ri": {"gan": ri_gan, "zhi": ri_zhi},
                          "shi": {"gan": "庚", "zhi": shi_zhi}}}

    def test_流年值_命中与不命中(self):
        # 配偶星宫位=日支（palace_ri.zhi=子）——流年支=子 → 1；丑 → 0
        chart = self._chart()
        ctx = {"liunian": {"nian_zhi": "子"}, "chart": chart}
        self.assertEqual(_liu_op("流年值", ["配偶星"], "male", chart, ctx), 1)
        ctx2 = {"liunian": {"nian_zhi": "丑"}, "chart": chart}
        self.assertEqual(_liu_op("流年值", ["配偶星"], "male", chart, ctx2), 0)

    def test_流年克_五行相克与不克(self):
        # 男命配偶星=正财（土）——流年支寅（木克土）→ 1；子（水不克土）→ 0
        base = self._fac(shishen={"正财": {"wuxing": "土"}})
        ctx = {"liunian": {"nian_gan": "庚", "nian_zhi": "寅"}}
        self.assertEqual(_liu_op("流年克", ["配偶星"], "male", base, ctx), 1)
        ctx2 = {"liunian": {"nian_gan": "庚", "nian_zhi": "子"}}
        self.assertEqual(_liu_op("流年克", ["配偶星"], "male", base, ctx2), 0)

    def test_流年克_本命星不现仍可推导目标五行(self):
        # 甲男配偶星=财星土；本命财星虽不现，流年木仍克财星土。
        ctx = {"liunian": {"nian_gan": "甲", "nian_zhi": "寅"}}
        self.assertEqual(_liu_op("流年克", ["配偶星"], "male", self._fac(), ctx), 1)

    def test_旬空_空亡支命中与不命中(self):
        # 日柱甲子（甲子旬空戌亥）——流年支戌 → 1；午 → 0
        chart = self._chart(ri_gan="甲", ri_zhi="子")
        ctx = {"liunian": {"nian_zhi": "戌"}, "chart": chart}
        self.assertEqual(_liu_op("旬空", ["日柱", "流年支"], "male", chart, ctx), 1)
        ctx2 = {"liunian": {"nian_zhi": "午"}, "chart": chart}
        self.assertEqual(_liu_op("旬空", ["日柱", "流年支"], "male", chart, ctx2), 0)

    def test_旬空_非零偏移日柱(self):
        # 己亥日（甲午旬，空亡辰巳）——流年支辰填实旬空 → 1。
        chart = self._chart(ri_gan="己", ri_zhi="亥")
        ctx = {"liunian": {"nian_zhi": "辰"}, "chart": chart}
        self.assertEqual(_liu_op("旬空", ["日柱", "流年支"], "male", chart, ctx), 1)
        ctx2 = {"liunian": {"nian_zhi": "午"}, "chart": chart}
        self.assertEqual(_liu_op("旬空", ["日柱", "流年支"], "male", chart, ctx2), 0)
        # 甲戌日（甲戌旬，空亡申酉）——流年支申 → 1
        chart3 = self._chart(ri_gan="甲", ri_zhi="戌")
        ctx3 = {"liunian": {"nian_zhi": "申"}, "chart": chart3}
        self.assertEqual(_liu_op("旬空", ["日柱", "流年支"], "male", chart3, ctx3), 1)

    def test_干支相等_相同与不同(self):
        # 流年甲子 == 日柱甲子 → 1；日柱乙丑 → 0
        chart = self._chart(ri_gan="甲", ri_zhi="子")
        ctx = {"liunian": {"nian_gan": "甲", "nian_zhi": "子"}, "chart": chart}
        self.assertEqual(_liu_op("干支相等", ["流年", "日柱"], "male", chart, ctx), 1)
        chart2 = self._chart(ri_gan="乙", ri_zhi="丑")
        ctx2 = {"liunian": {"nian_gan": "甲", "nian_zhi": "子"}, "chart": chart2}
        self.assertEqual(_liu_op("干支相等", ["流年", "日柱"], "male", chart2, ctx2), 0)

    def test_引用本命_支持键与未知键(self):
        # 支持的 key 返回 ctx 值；未知 key → 0（死条件防护）
        ctx = {"snapshot": {"食伤旺": 1}}
        self.assertEqual(_liu_op("引用本命", ["食伤旺"], "male", None, ctx), 1)
        self.assertEqual(_liu_op("引用本命", ["未知键"], "male", None, ctx), 0)

    def test_流年宫化_无紫微四化数据(self):
        # 无 zw_liunian 数据 → 0（不得误命中）
        ctx = {"zw_liunian": {}}
        self.assertEqual(_liu_op("流年宫化", ["夫妻", "禄"], "male", None, ctx), 0)

    def test_流年透_配偶星透与不透(self):
        # 男命配偶星=正财/偏财——流年十神为正财 → 1；比肩 → 0
        ctx = {"liunian": {"shi_shen": "正财"}}
        self.assertEqual(_liu_op("流年透", ["配偶星"], "male", None, ctx), 1)
        ctx2 = {"liunian": {"shi_shen": "比肩"}}
        self.assertEqual(_liu_op("流年透", ["配偶星"], "male", None, ctx2), 0)

    def test_流年神煞_命中与不命中(self):
        ctx = {"liunian": {"shensha": [{"name": "红鸾"}, {"name": "天喜"}]}}
        self.assertEqual(_liu_op("流年神煞", ["红鸾"], "male", None, ctx), 1)
        ctx2 = {"liunian": {"shensha": [{"name": "桃花"}]}}
        self.assertEqual(_liu_op("流年神煞", ["红鸾"], "male", None, ctx2), 0)

    def test_年柱干伏吟_相同与不同(self):
        chart = self._chart()  # 年柱庚
        ctx = {"liunian": {"nian_gan": "庚"}, "chart": chart}
        self.assertEqual(_liu_op("年柱干伏吟", [], "male", chart, ctx), 1)
        ctx2 = {"liunian": {"nian_gan": "甲"}, "chart": chart}
        self.assertEqual(_liu_op("年柱干伏吟", [], "male", chart, ctx2), 0)

    def test_忌神干_命中与不命中(self):
        # 忌神=火（fu_yi.ji）——流年干丙（火）→ 1；庚（金）→ 0
        base = self._fac()
        base["yongshen"] = {"fu_yi": {"ji": "火"}}
        ctx = {"liunian": {"nian_gan": "丙"}}
        self.assertEqual(_liu_op("忌神干", [], "male", base, ctx), 1)
        ctx2 = {"liunian": {"nian_gan": "庚"}}
        self.assertEqual(_liu_op("忌神干", [], "male", base, ctx2), 0)

    def test_财坏印流年_印年被克(self):
        # 流年干为印（正印）+ 流年支五行克印干五行 → 1
        base = self._fac(shishen={"正财": {"wuxing": "土"}})
        # 日主甲木，印=水（壬/癸），流年干壬（水=印），流年支寅（木）——木不克水 → 0
        ctx = {"liunian": {"nian_gan": "壬", "nian_zhi": "寅", "shi_shen": "正印"}}
        self.assertEqual(_liu_op("财坏印流年", [], "male", base, ctx), 0)
        # 流年支戌（土）——土克水 → 1（流年干为印被流年支所克）
        ctx2 = {"liunian": {"nian_gan": "壬", "nian_zhi": "戌", "shi_shen": "正印"}}
        self.assertEqual(_liu_op("财坏印流年", [], "male", base, ctx2), 1)


if __name__ == "__main__":
    unittest.main()
