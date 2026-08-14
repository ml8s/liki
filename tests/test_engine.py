"""单元测试：match（真值表匹配）+ 核心算子（现/透/藏/得令/克/旺）。

纯函数测试——mock 因子快照/排盘数据结构，不依赖真实 RPC。
"""
import sys, os, unittest
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'tools'))
from engine import match
from atoms import _op, _resolve_tens


def _mock_factors(**shishen):
    """构造最小因子快照（fac）。"""
    default = {"shishen": {}, "ri_gan": "甲", "qiangruo": "中和",
               "wuxing": {"wang_shuai": {}}}
    fac = default
    fac["shishen"] = shishen
    return fac


class TestMatch(unittest.TestCase):
    def test_命中(self):
        table = [{"id": "t1", "约束": {"配偶星透干": 1}, "结论": "婚可成"}]
        snapshot = {"配偶星透干": 1}
        self.assertEqual([h["id"] for h in match(table, snapshot)], ["t1"])

    def test_未命中(self):
        table = [{"id": "t1", "约束": {"配偶星透干": 1}, "结论": "婚可成"}]
        self.assertEqual(match(table, {"配偶星透干": 0}), [])

    def test_多约束全命中才命中(self):
        table = [{"id": "t1", "约束": {"配偶星透干": 1, "比劫重": 0}, "结论": "x"}]
        self.assertEqual(match(table, {"配偶星透干": 1, "比劫重": 0})[0]["id"], "t1")
        self.assertEqual(match(table, {"配偶星透干": 1, "比劫重": 1}), [])

    def test_exclusive取表序最前(self):
        table = [{"id": "a", "约束": {}, "结论": "A"}, {"id": "b", "约束": {}, "结论": "B"}]
        self.assertEqual(match(table, {}, exclusive=True)[0]["id"], "a")

    def test_约束值为0也须匹配(self):
        # 约束列值=0 表示"该因子必须为 0"（排他）
        table = [{"id": "t1", "约束": {"官杀透": 0}, "结论": "x"}]
        self.assertEqual(match(table, {"官杀透": 0})[0]["id"], "t1")
        self.assertEqual(match(table, {"官杀透": 1}), [])


class TestOperators(unittest.TestCase):
    def test_现(self):
        fac = _mock_factors(正财={"count": 1, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], fac, "male", None), 1)
        fac2 = _mock_factors(正财={"count": 0, "wuxing": "土"})
        self.assertEqual(_op("现", ["正财"], fac2, "male", None), 0)

    def test_透(self):
        fac = _mock_factors(正财={"tou_gan": True, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], fac, "male", None), 1)
        fac2 = _mock_factors(正财={"tou_gan": False, "wuxing": "土"})
        self.assertEqual(_op("透", ["正财"], fac2, "male", None), 0)

    def test_得令(self):
        fac = _mock_factors(正财={"de_ling": True, "wuxing": "土"})
        self.assertEqual(_op("得令", ["正财"], fac, "male", None), 1)

    def test_克_五行生克(self):
        # 财（土）克印（水）——土克水 → 1
        fac = _mock_factors(正财={"wuxing": "土", "count": 1}, 正印={"wuxing": "水", "count": 1})
        self.assertEqual(_op("克", ["财", "印"], fac, "male", None), 1)

    def test_克_不克(self):
        # 财（土）不克官（木）——土不克木 → 0
        fac = _mock_factors(正财={"wuxing": "土", "count": 1}, 正官={"wuxing": "木", "count": 1})
        self.assertEqual(_op("克", ["财", "官杀"], fac, "male", None), 0)


if __name__ == "__main__":
    unittest.main()


class TestOperatorsExtended(unittest.TestCase):
    """补测核心算子：藏/有根/旺/弱/缺（评审遗留 #2）。"""

    def test_藏(self):
        fac = _mock_factors(正印={"cang_zhi": True, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], fac, "male", None), 1)
        fac2 = _mock_factors(正印={"cang_zhi": False, "wuxing": "水"})
        self.assertEqual(_op("藏", ["正印"], fac2, "male", None), 0)

    def test_有根(self):
        fac = _mock_factors(正财={"has_root": True, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], fac, "male", None), 1)
        fac2 = _mock_factors(正财={"has_root": False, "wuxing": "土"})
        self.assertEqual(_op("有根", ["正财"], fac2, "male", None), 0)

    def test_旺_五行直读(self):
        fac = _mock_factors()
        fac["wuxing"] = {"wang_shuai": {"木": "旺", "水": "休"}}
        self.assertEqual(_op("旺", ["木"], fac, "male", None), 1)
        self.assertEqual(_op("旺", ["水"], fac, "male", None), 0)

    def test_弱_十神三条件(self):
        # 失令 + 不透 + 无根 → 弱
        fac = _mock_factors(正财={"wuxing": "土", "tou_gan": False, "has_root": False})
        fac["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], fac, "male", None), 1)
        # 透干则不算弱
        fac2 = _mock_factors(正财={"wuxing": "土", "tou_gan": True, "has_root": False})
        fac2["wuxing"] = {"wang_shuai": {"土": "死"}}
        self.assertEqual(_op("弱", ["正财"], fac2, "male", None), 0)

    def test_缺_五行(self):
        # 缺[土]：五行 counts 中土=0 → 1
        fac = _mock_factors()
        fac["wuxing"] = {"count": {"木": 2, "火": 1, "土": 0, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], fac, "male", None), 1)
        fac2 = _mock_factors()
        fac2["wuxing"] = {"count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 1}}
        self.assertEqual(_op("缺", ["土"], fac2, "male", None), 0)


class TestEvaluateFactors(unittest.TestCase):
    """因子快照（duanyu.evaluate_factors）——核心逻辑回归保护（评审遗留 #2）。"""

    @classmethod
    def setUpClass(cls):
        import sys
        sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'tools'))
        import duanyu
        cls.duanyu = duanyu

    def _fac(self, **shishen):
        f = _mock_factors(**shishen)
        f["wuxing"] = {"wang_shuai": {"木": "旺", "火": "相", "土": "休", "金": "囚", "水": "死"},
                       "count": {"木": 2, "火": 1, "土": 1, "金": 1, "水": 0}}
        f["yongshen"] = {"fu_yi": {"qiangruo": "身强", "yong": "金", "xi": "土", "ji": "木"}}
        return f

    def test_快照含关键因子(self):
        # 正印 2 个且得令 → 「印星现」(现[正印,偏印]) 与「正印旺」(得令[正印]) 命中
        fac = self._fac(正印={"count": 2, "tou_gan": True, "cang_zhi": True, "de_ling": True, "has_root": True, "wuxing": "水"})
        snap = self.duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 1)
        self.assertEqual(snap.get("正印旺"), 1)
        self.assertEqual(snap.get("gender"), "male")

    def test_未命中因子为0(self):
        # 无正印 → 印星现=0
        fac = self._fac(正财={"count": 0, "wuxing": "土"})
        snap = self.duanyu.evaluate_factors(fac, "female", {}, shushi="bazi")
        self.assertEqual(snap.get("印星现"), 0)

    def test_多遍稳定_确定性(self):
        fac = self._fac(正官={"count": 1, "tou_gan": True, "de_ling": False, "wuxing": "木"})
        a = self.duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        b = self.duanyu.evaluate_factors(fac, "male", {}, shushi="bazi")
        self.assertEqual(a, b)

    def test_shushi_过滤(self):
        fac = self._fac()
        ziwei_snap = self.duanyu.evaluate_factors(fac, "male", {}, shushi="ziwei")
        # 紫微快照不应含八字因子
        self.assertNotIn("印星现", ziwei_snap)
