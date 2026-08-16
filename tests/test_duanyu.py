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


if __name__ == "__main__":
    unittest.main()
