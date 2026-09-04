"""单元测试：duanyu.match（断语真值表匹配）。"""
import unittest

import _helpers  # noqa: F401 —— 导入即注入 tools 路径
from duanyu import match


class TestMatch(unittest.TestCase):
    def test_命中(self):
        table = [{"id": "t1", "约束组": [{"配偶星透": 1}], "结论": "婚可成"}]
        snapshot = {"配偶星透": 1}
        self.assertEqual([h["id"] for h in match(table, snapshot)], ["t1"])

    def test_未命中(self):
        table = [{"id": "t1", "约束组": [{"配偶星透": 1}], "结论": "婚可成"}]
        self.assertEqual(match(table, {"配偶星透": 0}), [])

    def test_多约束全命中才命中(self):
        table = [{"id": "t1", "约束组": [{"配偶星透": 1, "比劫重": 0}], "结论": "x"}]
        self.assertEqual(match(table, {"配偶星透": 1, "比劫重": 0})[0]["id"], "t1")
        self.assertEqual(match(table, {"配偶星透": 1, "比劫重": 1}), [])

    def test_exclusive取表序最前(self):
        table = [{"id": "a", "约束组": [{"x": 1}], "结论": "A"},
                 {"id": "b", "约束组": [{"x": 1}], "结论": "B"}]
        self.assertEqual(match(table, {"x": 1}, exclusive=True)[0]["id"], "a")

    def test_约束值为0也须匹配(self):
        # 约束列值=0 表示"该因子必须为 0"（排他）
        table = [{"id": "t1", "约束组": [{"官杀透": 0}], "结论": "x"}]
        self.assertEqual(match(table, {"官杀透": 0})[0]["id"], "t1")
        self.assertEqual(match(table, {"官杀透": 1}), [])

    def test_空表返回空(self):
        self.assertEqual(match([], {"x": 1}), [])

    def test_条件组之间为或(self):
        table = [{
            "id": "t1",
            "约束组": [{"印为用": 1, "大运印星运": 1}, {"官杀为用": 1, "大运官杀运": 1}],
            "结论": "x",
        }]
        assert [h["id"] for h in match(table, {"印为用": 1, "大运印星运": 1})] == ["t1"]
        assert [h["id"] for h in match(table, {"官杀为用": 1, "大运官杀运": 1})] == ["t1"]
        assert match(table, {"印为用": 1, "大运官杀运": 1}) == []


if __name__ == "__main__":
    unittest.main()
