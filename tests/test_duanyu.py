"""单元测试：duanyu（因子快照 evaluate_factors + query 断语查询 + yearly 隔离）。"""
import unittest

from _helpers import mock_factors
import duanyu


class TestEvaluateFactors(unittest.TestCase):
    """因子快照（duanyu.evaluate_factors）——核心逻辑回归保护。"""

    def _fac(self, **shishen):
        f = mock_factors(**shishen)
        f["wuxing"] = {"wang_shuai": {"木": "旺", "火": "相", "土": "休", "金": "囚", "水": "死"},
                       "count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 0}}
        f["yongshen"] = {"fu_yi": {"qiangruo": "身强", "yong": "金", "xi": "土", "ji": "木"}}
        return f

    def test_快照含关键因子(self):
        # 正印 2 个且得令 → 「印星现」(现[正印,偏印]) 与「正印旺」(得令[正印]) 命中
        fac = self._fac(正印={"count": 2, "tou_gan": True, "cang_zhi": True, "de_ling": True, "has_root": True, "wuxing": "水"})
        snap = duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 1)
        self.assertEqual(snap.get("正印旺"), 1)
        self.assertEqual(snap.get("gender"), "male")

    def test_未命中因子为0(self):
        # 无正印 → 印星现=0
        fac = self._fac(正财={"count": 0, "wuxing": "土"})
        snap = duanyu.evaluate_factors(fac, "female", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 0)

    def test_多遍稳定_确定性(self):
        fac = self._fac(正官={"count": 1, "tou_gan": True, "de_ling": False, "wuxing": "木"})
        a = duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        b = duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(a, b)

    def test_shushi_过滤(self):
        fac = self._fac()
        ziwei_snap = duanyu.evaluate_factors(fac, "male", {}, shushi="ziwei")
        # 紫微快照不应含八字因子
        self.assertNotIn("印星现", ziwei_snap)


class TestYearlyIsolation(unittest.TestCase):
    """本命快照查 yearly_* 必须隔离（不得命中纯本命约束的流年断语）。"""

    def test_本命快照查yearly_拒绝(self):
        ben = {"八字": {"财坏印": 1}, "紫微": {}}  # 本命快照（无流年特征标记）
        res = duanyu.query("yearly_liuqin", ben)
        self.assertEqual(res["八字"], [])

    def test_流年快照查yearly_正常命中(self):
        liu = {"_snapshot_type": "liunian", "八字": {"财坏印": 1, "财坏印流年": 1}, "紫微": {}}  # 流年快照
        res = duanyu.query("yearly_liuqin", liu)
        self.assertTrue(any(r.get("id") == "yliu_104" for r in res["八字"]))


class TestRiZhuWuXing(unittest.TestCase):
    """日主五行直通因子（自查 2026-08：曾被条件列 0/1 比较破坏 → 五行性情/外貌断语永久丢失）。"""

    def test_直读任意返回字符串(self):
        # _atomic 收尾不得把「任意」取值模式的字符串布尔化（旧 bug："土" != 1 → 因子恒 0）
        fac = mock_factors()
        fac["ri_gan"] = "己"
        self.assertEqual(duanyu._atomic("直读[ri_gan_wx,任意]", fac, "male", None), "土")
        # 期望值模式不受影响（直读[ri_gan_wx,木] 对土日主 = 0）
        self.assertEqual(duanyu._atomic("直读[ri_gan_wx,木]", fac, "male", None), 0)

    def test_日主五行因子为五行字符串(self):
        fac = mock_factors()
        fac["ri_gan"] = "己"
        snap = duanyu.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("日主五行"), "土")


class TestCaiXingShouKe(unittest.TestCase):
    """财星受克因子（自查 2026-08：旧定义 克[比劫,财] 五行恒真 → 因子恒 1、
    liu_101 父星旺断语永不命中 + lq_301 父寿不永 93% 误报）。"""

    def test_比劫强财弱_受克(self):
        fac = mock_factors(正财={"wuxing": "土"}, 比肩={"wuxing": "木", "de_ling": True, "count": 3})
        fac["ri_gan"] = "甲"
        fac["wuxing"] = {"wang_shuai": {"木": "旺", "土": "死"}}
        snap = duanyu.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("财星受克"), 1)

    def test_财旺_不受克(self):
        # 财星得令有根 → 不受克（旧定义恒 1 会误报）
        fac = mock_factors(正财={"wuxing": "土", "de_ling": True, "has_root": True, "count": 3},
                           比肩={"wuxing": "木", "count": 1})
        fac["ri_gan"] = "甲"
        fac["wuxing"] = {"wang_shuai": {"土": "旺", "木": "休"}}
        snap = duanyu.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("财星受克"), 0)


class TestYueLingGe(unittest.TestCase):
    """月令格因子（自查 2026-08：直读任意取值语义 + 条件列比较 + 枚举缺"格"后缀
    三重问题 → 月令格断语全部永久丢失）。"""

    def test_月令格因子为格局字符串(self):
        fac = mock_factors()
        fac["yongshen"] = {"ge_ju": {"ge_ju": "正财格"}}
        snap = duanyu.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("月令格"), "正财格")

    def test_月令格断语命中(self):
        # 正财格 + 身弱 → ge_302（正财格身弱分支）
        fac = mock_factors(正财={"wuxing": "土", "de_ling": True, "count": 3})
        fac["yongshen"] = {"ge_ju": {"ge_ju": "正财格"}}
        fac["qiangruo"] = "身弱"
        from duanyu import load_table
        from engine import match
        snap = duanyu.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        # 手动构造断语表查询依赖的最小快照（身弱/从杀格）
        snap["身弱"] = 1
        snap["从杀格"] = 0
        hits = [e["id"] for e in match(load_table("bazi_geju.csv"), snap)]
        self.assertIn("ge_302", hits)


class TestXueye201Regression(unittest.TestCase):
    """回归保护：xue_201 条件曾误写 印星旺=0（要求印不旺），断语却是「印星得月令而旺」，
    导致无印盘命中最高档学历断语（feedback ba47240e/issue #25）。修复后 0→1。"""

    def _hits(self, yin_wang, guan_sha_de_ling):
        out = duanyu.query("xueye", {"八字": {"印星旺": yin_wang, "官杀得令": guan_sha_de_ling},
                                     "紫微": {}})
        return [e["id"] for e in out["八字"]]

    def test_印弱盘不命中科甲至顶(self):
        # 复现盘：己土日主亥月、原局无火（印星旺=0/官杀得令=1）——不得再触发 xue_201
        self.assertNotIn("xue_201", self._hits(0, 1))

    def test_印旺官杀得令命中(self):
        # 断语本义：印星得月令而旺 + 官杀得令 → 官印相生科甲至顶
        self.assertIn("xue_201", self._hits(1, 1))

    def test_印旺但官杀不得令不命中(self):
        self.assertNotIn("xue_201", self._hits(1, 0))


if __name__ == "__main__":
    unittest.main()
