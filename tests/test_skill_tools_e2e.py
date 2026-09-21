"""skill 工具端到端契约测试。

覆盖：
- skill-tools.json 解析 → 5 工具全部注册
- agent_cli.py 非法输入 → 错误透传（ValueError 非 crash）
- query(rule, pan) 传 mock pan → 返回 {八字:[], 紫微:[], 合参:[]}
"""
import os
import json
import subprocess
import sys
import unittest
from unittest import mock

import _helpers  # noqa: F401 —— 提供完整 daxian mock
from pan_integrity import with_natal_digest
_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_TOOLS = os.path.join(_ROOT, "skills", "liki", "natal", "tools")
_SCHEMA = os.path.join(_TOOLS, "skill-tools.json")


class TestSkillToolsRegister(unittest.TestCase):
    """skill-tools.json 存在且 5 工具全部定义。"""

    def test_schema_parses_and_has_5_tools(self):
        with open(_SCHEMA, encoding="utf-8") as f:
            d = json.load(f)
        names = [t["function"]["name"] for t in d["tools"]]
        self.assertEqual(len(names), 5)
        expected = {
            "create_birth_chart", "analyze_natal", "analyze_periods",
            "compare_birth_charts", "calibrate_birth_time",
        }
        self.assertEqual(set(names), expected)


class TestAgentCliErrorPropagation(unittest.TestCase):
    """agent_cli.py 非法输入 → 错误透传为 JSON（不 crash）。"""

    def _run(self, stdin_text):
        sys.path.insert(0, _TOOLS)
        import agent_cli
        output = {}
        with mock.patch.object(agent_cli, "ensure_engine_compatible"), \
             mock.patch("sys.stdin") as stdin, \
             mock.patch("builtins.print") as printed:
            stdin.read.return_value = stdin_text
            agent_cli.main()
            output = json.loads(printed.call_args.args[0])
        return output

    def test_unknown_tool(self):
        out = self._run('{"fn":"nonexistent","args":{}}')
        self.assertFalse(out["ok"])
        self.assertEqual(out["error"]["code"], "INVALID_INPUT")
        self.assertIn("unknown tool", out["error"]["message"])

    def test_missing_arg(self):
        out = self._run('{"fn":"analyze_natal","args":{"chart_ref":{"token":"x"}}}')
        self.assertFalse(out["ok"])
        self.assertEqual(out["error"]["code"], "INVALID_INPUT")
        self.assertIn("missing arg", out["error"]["message"])


class TestQueryWithMockPan(unittest.TestCase):
    """query(rule, pan) 传含 base 的 mock pan → 返回 {八字:[], 紫微:[], 合参:[]}。"""

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
                "da_yun": {"steps": [], "current_step_index": -1},
            },
            "full": {
                "nian": {"gan": "庚", "zhi": "午"},
                "yue": {"gan": "壬", "zhi": "午"},
                "ri": {"gan": "己", "zhi": "亥"},
                "shi": {"gan": "庚", "zhi": "午"},
                **_helpers.mock_engine_facts(),
            **_helpers.mock_yongshen_fields(),
            },
            "fu_yi": {}, "tiao_hou": {}, "ge_ju": {},
            "ziwei": _helpers.mock_ziwei(),
            "ziwei_daxian": _helpers.valid_daxian(),
            "gender": "male",
        }
        r = query("十神", with_natal_digest(mock_pan))
        self.assertIn("八字", r)
        self.assertIn("紫微", r)

    def test_query_empty_pan_raises(self):
        sys.path.insert(0, _TOOLS)
        from duanyu import query
        with self.assertRaises(ValueError):
            query("十神", {})


if __name__ == "__main__":
    unittest.main()
