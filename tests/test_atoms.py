"""单元测试：atoms 核心算子（现/透/藏/得令/克/旺/弱/缺 + 流年大运算子）。"""
import unittest

from _helpers import mock_factors
from atoms import _op, _liu_op


class TestOperators(unittest.TestCase):
    def test_现(self):
        fac = mock_factors(正财={"count": 1, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], fac, "male", None), 1)
        fac2 = mock_factors(正财={"count": 0, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], fac2, "male", None), 0)

    def test_透(self):
        fac = mock_factors(正财={"tou_gan": True, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], fac, "male", None), 1)
        fac2 = mock_factors(正财={"tou_gan": False, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], fac2, "male", None), 0)

    def test_得令(self):
        fac = mock_factors(正财={"de_ling": True, "wuxing": "土"})
        self.assertEqual(_op("得令", ["正财"], fac, "male", None), 1)

    def test_克_五行生克(self):
        # 财（土）克印（水）——土克水 → 1
        fac = mock_factors(正财={"wuxing": "土", "count": 1}, 正印={"wuxing": "水", "count": 1})
        self.assertEqual(_op("克", ["财", "印"], fac, "male", None), 1)

    def test_克_不克(self):
        # 财（土）不克官（木）——土不克木 → 0
        fac = mock_factors(正财={"wuxing": "土", "count": 1}, 正官={"wuxing": "木", "count": 1})
        self.assertEqual(_op("克", ["财", "官杀"], fac, "male", None), 0)


class TestOperatorsExtended(unittest.TestCase):
    """补测核心算子：藏/有根/旺/弱/缺。"""

    def test_藏(self):
        fac = mock_factors(正印={"cang_zhi": True, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], fac, "male", None), 1)
        fac2 = mock_factors(正印={"cang_zhi": False, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], fac2, "male", None), 0)

    def test_有根(self):
        fac = mock_factors(正财={"has_root": True, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], fac, "male", None), 1)
        fac2 = mock_factors(正财={"has_root": False, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], fac2, "male", None), 0)

    def test_旺_五行直读(self):
        fac = mock_factors()
        fac["wuxing"] = {"wang_shuai": {"木": "旺", "水": "休"}}
        self.assertEqual(_op("旺", ["木"], fac, "male", None), 1)
        self.assertEqual(_op("旺", ["水"], fac, "male", None), 0)

    def test_弱_十神三条件(self):
        # 失令 + 不透 + 无根 → 弱
        fac = mock_factors(正财={"wuxing": "土", "tou_gan": False, "has_root": False})
        fac["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], fac, "male", None), 1)
        # 透干则不算弱
        fac2 = mock_factors(正财={"wuxing": "土", "tou_gan": True, "has_root": False})
        fac2["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], fac2, "male", None), 0)

    def test_缺_五行(self):
        # 缺[土]：五行 counts 中土=0 → 1
        fac = mock_factors()
        fac["wuxing"] = {"count": {"木": 2, "火": 1, "土": 0, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], fac, "male", None), 1)
        fac2 = mock_factors()
        fac2["wuxing"] = {"count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], fac2, "male", None), 0)


class TestDaYunOps_YearRange(unittest.TestCase):
    """2.6.15 大运公历年段算子（start_year/end_year 直判，免虚岁换算）。"""

    def _fac(self):
        f = mock_factors()
        f["dayun_steps"] = [
            {"name": "丁卯", "shi_shen": "偏印运", "start_year": 1990, "end_year": 2000},
            {"name": "戊辰", "shi_shen": "正财运", "start_year": 2001, "end_year": 2010},
            {"name": "己巳", "shi_shen": "比肩运", "start_year": 2011, "end_year": 2020},
        ]
        return f

    def test_大运窗口流年_年份段内(self):
        ctx = {"year": 2005, "target": "配偶星", "liunian": {"nian_zhi": "酉", "nian_gan": "乙", "shi_shen": "偏财"}}
        self.assertEqual(_liu_op("大运窗口流年", ["目标星"], self._fac(), "male", None, ctx), 1)

    def test_大运窗口流年_窗口外(self):
        ctx = {"year": 1989, "target": "配偶星", "liunian": {}}
        self.assertEqual(_liu_op("大运窗口流年", ["目标星"], self._fac(), "male", None, ctx), 0)

    def test_大运窗口流年_非目标星大运(self):
        # 丁卯=偏印运（非配偶星）——1995 窗口内但 shi_shen 不含目标星 → 0
        fac = self._fac()
        ctx = {"year": 1995, "target": "配偶星", "liunian": {}}
        self.assertEqual(_liu_op("大运窗口流年", ["目标星"], fac, "male", None, ctx), 0)

    def test_换运流年_首年(self):
        ctx = {"year": 2001, "target": "配偶星", "liunian": {}}
        self.assertEqual(_liu_op("换运流年", ["目标星"], self._fac(), "male", None, ctx), 1)

    def test_换运流年_非首年(self):
        ctx = {"year": 2005, "target": "配偶星", "liunian": {}}
        self.assertEqual(_liu_op("换运流年", ["目标星"], self._fac(), "male", None, ctx), 0)


if __name__ == "__main__":
    unittest.main()
