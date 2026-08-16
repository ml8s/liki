"""单元测试：engine.match（真值表匹配）。"""
import unittest

from _helpers import mock_factors  # noqa: F401 —— 导入即注入 tools 路径
from engine import match


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

    def test_空表返回空(self):
        self.assertEqual(match([], {"x": 1}), [])


if __name__ == "__main__":
    unittest.main()
