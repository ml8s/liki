"""skill 工具端到端测试（从 liki-web 迁移——测的是 liki 自身的功能）。

覆盖：
- skill-tools.json 解析 → 6 工具全部注册
- agent_cli.py 非法输入 → 错误透传（ValueError 非 crash）
- query(rule, pan) 传 mock pan → 返回 {八字:[], 紫微:[]}
"""
import json
import os
import subprocess
import sys
import unittest

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_TOOLS = os.path.join(_ROOT, "skills", "liki-bazi", "tools")
_SCHEMA = os.path.join(_TOOLS, "skill-tools.json")


class TestSkillToolsRegister(unittest.TestCase):
    """skill-tools.json 存在且 6 工具全部定义。"""

    def test_schema_parses_and_has_6_tools(self):
        with open(_SCHEMA, encoding="utf-8") as f:
            d = json.load(f)
        names = [t["function"]["name"] for t in d["tools"]]
        self.assertEqual(len(names), 6)
        expected = {"city_coords", "full_paipan", "query", "yearly_range", "calibrate", "bond"}
        self.assertEqual(set(names), expected)


class TestAgentCliErrorPropagation(unittest.TestCase):
    """agent_cli.py 非法输入 → 错误透传为 JSON（不 crash）。"""

    def _run(self, stdin_text):
        p = subprocess.run(
            [sys.executable, os.path.join(_TOOLS, "agent_cli.py")],
            input=stdin_text.encode(), capture_output=True, timeout=30)
        return json.loads(p.stdout)

    def test_unknown_tool(self):
        out = self._run('{"fn":"nonexistent","args":{}}')
        self.assertFalse(out["ok"])
        self.assertIn("unknown tool", out["error"])

    def test_missing_arg(self):
        out = self._run('{"fn":"query","args":{"rule":"十神"}}')
        self.assertFalse(out["ok"])
        self.assertIn("missing arg", out["error"])


class TestQueryWithMockPan(unittest.TestCase):
    """query(rule, pan) 传含 fac 的 mock pan → 返回 {八字:[], 紫微:[]}。"""

    def test_query_returns_bazi_ziwei(self):
        sys.path.insert(0, _TOOLS)
        from duanyu import query

        # 构造最小合法 pan
        mock_pan = {
            "solar": "1990-06-01T12:00:00+08:00",
            "lunar": {"year": 1990, "month": 5, "day": 9, "shichen": "午"},
            "chart": {
                "nian": {"gan": "庚", "zhi": "午"},
                "yue": {"gan": "壬", "zhi": "午"},
                "ri": {"gan": "己", "zhi": "亥"},
                "shi": {"gan": "庚", "zhi": "午"},
            },
            "full": {},
            "yongshen": {},
            "ziwei": {},
            "gender": "male",
            "fac": {},
        }
        r = query("十神", mock_pan)
        self.assertIn("八字", r)
        self.assertIn("紫微", r)

    def test_query_empty_pan_raises(self):
        sys.path.insert(0, _TOOLS)
        from duanyu import query
        with self.assertRaises(ValueError):
            query("十神", {})


if __name__ == "__main__":
    unittest.main()
