"""单元测试：factors 因子求值 + duanyu 断语查询 / yearly 隔离。"""
import unittest

from _helpers import mock_factors
import duanyu
import factors


class TestEvaluateFactors(unittest.TestCase):
    """因子快照（factors.evaluate_factors）——核心逻辑回归保护。"""

    def _fac(self, **shishen):
        f = mock_factors(**shishen)
        f["wuxing"] = {"wang_shuai": {"木": "旺", "火": "相", "土": "休", "金": "囚", "水": "死"},
                       "count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 0}}
        f["yongshen"] = {"fu_yi": {"qiangruo": "身强", "yong": "金", "xi": "土", "ji": "木"}}
        return f

    def test_快照含关键因子(self):
        # 正印 2 个且得令 → 「印星现」(现[正印,偏印]) 与「正印旺」(得令[正印]) 命中
        fac = self._fac(正印={"count": 2, "tou_gan": True, "cang_zhi": True, "de_ling": True, "has_root": True, "wuxing": "水"})
        snap = factors.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 1)
        self.assertEqual(snap.get("正印旺"), 1)
        self.assertNotIn("性别", snap)

    def test_未命中因子为0(self):
        # 无正印 → 印星现=0
        fac = self._fac(正财={"count": 0, "wuxing": "土"})
        snap = factors.evaluate_factors(fac, "female", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 0)

    def test_多遍稳定_确定性(self):
        fac = self._fac(正官={"count": 1, "tou_gan": True, "de_ling": False, "wuxing": "木"})
        a = factors.evaluate_factors(fac, "male", {}, shushi="bazi")
        b = factors.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(a, b)

    def test_shushi_过滤(self):
        fac = self._fac()
        ziwei_snap = factors.evaluate_factors(fac, "male", {}, shushi="ziwei")
        # 紫微快照不应含八字因子
        self.assertNotIn("印星现", ziwei_snap)


class TestFactorCleanupRegression(unittest.TestCase):
    """因子清理后的关键契约：标量枚举、正因子、程度区分与羊刃成格。"""

    def _fac(self, **shishen):
        fac = mock_factors(**shishen)
        fac["wuxing"] = {"wang_shuai": {}, "count": {"木": 2, "火": 1, "土": 3, "金": 1, "水": 1}}
        return fac

    def _chart(self, ri_gan="甲", yue_zhi="卯", state="帝旺"):
        return {
            "gender": "male",
            "chart": {"yue": {"zhi": yue_zhi}, "ri": {"gan": ri_gan, "zhi": "辰"}},
            "full": {"chang_sheng": [{"name": state, "index": yue_zhi}]},
        }

    def test_互斥枚举使用标量因子(self):
        fac = self._fac()
        snap = factors.evaluate_factors(fac, "male", self._chart("甲"), shushi="bazi")
        self.assertEqual(snap["日主"], "甲")
        self.assertEqual(snap["日主长生状态"], "帝旺")

    def test_性别是排盘上下文而非因子(self):
        fac = self._fac()
        pan = {"fac": fac, "gender": "female", "full": {}, "chart": {}, "ziwei": {}}
        result = factors.make_factors(pan)
        self.assertEqual(result["context"], {"性别": "female"})
        self.assertNotIn("性别", result["八字"])
        self.assertNotIn("性别", result["紫微"])

    def test_配偶星现按性别解析(self):
        male = self._fac(正财={"count": 1, "wuxing": "土"})
        snap = factors.evaluate_factors(male, "male", {"gender": "male"}, shushi="bazi")
        self.assertEqual(snap["配偶星现"], 1)

        female = self._fac(正官={"count": 1, "wuxing": "金"})
        snap = factors.evaluate_factors(female, "female", {"gender": "female"}, shushi="bazi")
        self.assertEqual(snap["配偶星现"], 1)

    def test_数量程度词有区分(self):
        fac = self._fac(
            正印={"count": 3, "wuxing": "水"}, 偏印={"count": 0, "wuxing": "水"},
            比肩={"count": 2, "wuxing": "木"}, 劫财={"count": 2, "wuxing": "木"},
        )
        snap = factors.evaluate_factors(fac, "male", {"gender": "male"}, shushi="bazi")
        self.assertEqual(snap["印多"], 1)
        self.assertEqual(snap["印重"], 0)
        self.assertEqual(snap["土多"], 1)
        self.assertEqual(snap["比劫重"], 1)

    def test_土多计数阈值(self):
        two = self._fac()
        two["wuxing"]["count"]["土"] = 2
        snap = factors.evaluate_factors(two, "male", {"gender": "male"}, shushi="bazi")
        self.assertEqual(snap["土多"], 1)

        one = self._fac()
        one["wuxing"]["count"]["土"] = 1
        snap = factors.evaluate_factors(one, "male", {"gender": "male"}, shushi="bazi")
        self.assertEqual(snap["土多"], 0)

    def test_月令格由引擎权威标量化(self):
        yang = self._fac()
        yang["yongshen"] = {"ge_ju": {"ge_ju": "月刃格"}}
        snap = factors.evaluate_factors(yang, "male", self._chart("甲"), shushi="bazi")
        self.assertEqual(snap["月令格"], "月刃格")


class TestYearlyIsolation(unittest.TestCase):
    """query() 拒绝流年域；流年查询走 query_yearly（yearly_range 内部调用）。"""

    def test_本命快照查yearly_拒绝(self):
        pan = {"fac": {}}  # pan 直通
        with self.assertRaises(ValueError):
            duanyu.query("yearly_family", pan)

    def test_本命快照查yingqi_拒绝(self):
        pan = {"fac": {}}
        with self.assertRaises(ValueError):
            duanyu.query("yingqi", pan)

    def test_流年快照查命理域_正常命中(self):
        liu = {"_snapshot_type": "liunian", "八字": {"流年财坏印": 1}, "紫微": {}}
        res = duanyu.query_yearly("年十神", liu)   # yliu_104(财坏印)归入年十神域
        self.assertTrue(any(r.get("id") == "yliu_104" for r in res["八字"]))


class TestRiZhuWuXing(unittest.TestCase):
    """日主五行直通因子（自查 2026-08：曾被条件列 0/1 比较破坏 → 五行性情/外貌断语永久丢失）。"""

    def test_直读任意返回字符串(self):
        # _atomic 收尾不得把「任意」取值模式的字符串布尔化（旧 bug："土" != 1 → 因子恒 0）
        fac = mock_factors()
        fac["ri_gan"] = "己"
        self.assertEqual(factors._atomic("直读[ri_gan_wx,任意]", fac, "male", None), "土")
        # 期望值模式不受影响（直读[ri_gan_wx,木] 对土日主 = 0）
        self.assertEqual(factors._atomic("直读[ri_gan_wx,木]", fac, "male", None), 0)

    def test_日主五行因子为五行字符串(self):
        fac = mock_factors()
        fac["ri_gan"] = "己"
        snap = factors.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("日主五行"), "土")


class TestBiDuoCai(unittest.TestCase):
    """比劫夺财：要求财星存在且弱，不再把“财星不现”误作受克。"""

    def test_比劫旺而财星弱(self):
        fac = mock_factors(
            正财={"wuxing": "土", "count": 1, "tou_gan": True},
            比肩={"wuxing": "木", "de_ling": True, "count": 3},
        )
        fac["wuxing"] = {"wang_shuai": {"木": "旺", "土": "死"}}
        snap = factors.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("比劫夺财"), 1)

    def test_财星不现不构成夺财(self):
        fac = mock_factors(
            正财={"wuxing": "土", "count": 0},
            比肩={"wuxing": "木", "de_ling": True, "count": 3},
        )
        fac["wuxing"] = {"wang_shuai": {"木": "旺", "土": "死"}}
        snap = factors.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("比劫夺财"), 0)


class TestYueLingGe(unittest.TestCase):
    """月令格因子（自查 2026-08：直读任意取值语义 + 条件列比较 + 枚举缺"格"后缀
    三重问题 → 月令格断语全部永久丢失）。"""

    def test_月令格因子为格局字符串(self):
        fac = mock_factors()
        fac["yongshen"] = {"ge_ju": {"ge_ju": "正财格"}}
        snap = factors.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        self.assertEqual(snap.get("月令格"), "正财格")

    def test_月令格断语命中(self):
        # 正财格 + 身弱 → ge_302（正财格身弱分支）
        fac = mock_factors(正财={"wuxing": "土", "de_ling": True, "count": 3})
        fac["yongshen"] = {"ge_ju": {"ge_ju": "正财格"}}
        from duanyu import load_table
        from duanyu import match
        snap = factors.evaluate_factors(fac, "male", {"chart": {}}, shushi="bazi")
        # 手动构造断语表查询依赖的最小快照（身弱/从杀格）
        snap["身强弱"] = "身弱"
        hits = [e["id"] for e in match(load_table("bazi_格局.csv"), snap)]
        self.assertIn("ge_302", hits)


class TestXueye201Regression(unittest.TestCase):
    """回归保护：xue_201 条件曾误写 印星旺=0（要求印不旺），断语却是「印星得月令而旺」，
    导致无印盘命中最高档学历断语（feedback ba47240e/issue #25）。修复后 0→1。"""

    def _hits(self, yin_wang, guan_sha_de_ling):
        out = duanyu._match_rule("十神", {"八字": {"印星旺": yin_wang, "官杀得令": guan_sha_de_ling},
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
