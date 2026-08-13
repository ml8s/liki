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
