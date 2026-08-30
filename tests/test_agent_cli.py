"""agent_cli 分派单元测试（不连引擎——mock 工具链函数）。

覆盖：6 函数分派映射、参数透传、非法 fn、stdin 协议错误包装。
"""
import json
import os
import sys
import unittest
from unittest import mock

sys.path.insert(0, __import__('os').path.join(
    __import__('os').path.dirname(__import__('os').path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools'))

import agent_cli


class TestDispatch(unittest.TestCase):
    def setUp(self):
        # mock 工具链函数——验证分派映射与参数透传，不触发网络/真值表
        self.patchers = [
            mock.patch('agent_cli.full_paipan', return_value={"pan": "full"}),
            mock.patch('agent_cli.city_coords', return_value={"longitude": 116.4}),
            mock.patch('agent_cli.query', return_value={"八字": [], "紫微": []}),
            mock.patch('agent_cli.yearly_range', return_value={"years": {}}),
            mock.patch('agent_cli.calibrate', return_value={}),
            mock.patch('agent_cli.bond', return_value={"bazi": {}, "ziwei": {}}),
        ]
        for p in self.patchers:
            p.start()
        self.addCleanup(lambda: [p.stop() for p in self.patchers])

    def test_full_paipan_分派(self):
        data = agent_cli._dispatch("full_paipan", {"gregorian": "1990-06-01T12:00:00+08:00", "gender": "male"})
        self.assertEqual(data, {"pan": "full"})
        agent_cli.full_paipan.assert_called_once_with(
            "1990-06-01T12:00:00+08:00", "male", longitude=None, correct=True)

    def test_full_paipan_默认值(self):
        agent_cli._dispatch("full_paipan", {"gregorian": "t", "gender": "female", "longitude": 116.4})
        agent_cli.full_paipan.assert_called_once_with("t", "female", longitude=116.4, correct=True)

    def test_city_coords_分派(self):
        agent_cli._dispatch("city_coords", {"city": "北京"})
        agent_cli.city_coords.assert_called_once_with("北京")

    def test_query_分派(self):
        agent_cli._dispatch("query", {"rule": "marriage", "pan": {}})
        agent_cli.query.assert_called_once_with("marriage", {})

    def test_yearly_range_分派(self):
        agent_cli._dispatch("yearly_range", {"pan": {}, "start": 2025, "end": 2026})
        agent_cli.yearly_range.assert_called_once_with({}, 2025, 2026, rules=None, detail=False)

    def test_calibrate_分派(self):
        agent_cli._dispatch("calibrate", {"candidates": [], "events": []})
        agent_cli.calibrate.assert_called_once_with([], [], detail=False)

    def test_bond_分派(self):
        agent_cli._dispatch("bond", {"pan_a": {}, "pan_b": {}})
        agent_cli.bond.assert_called_once_with({}, {})

    def test_非法fn(self):
        with self.assertRaises(ValueError):
            agent_cli._dispatch("evil", {})

    def test_缺参(self):
        with self.assertRaises(KeyError):
            agent_cli._dispatch("full_paipan", {})  # 缺 time


class TestMainProtocol(unittest.TestCase):
    """stdin → stdout 协议：{ok,data} / {ok:false,error}。"""

    def _run_main(self, stdin_text):
        out = {}
        with mock.patch('sys.stdin') as stdin, \
             mock.patch('builtins.print') as pr:
            stdin.read.return_value = stdin_text
            agent_cli.main()
            # 捕获 print 的 JSON
            for call in pr.call_args_list:
                out = json.loads(call.args[0])
        return out

    def test_成功(self):
        with mock.patch('agent_cli._dispatch', return_value={"ok_data": 1}):
            out = self._run_main('{"fn":"query","args":{}}')
        self.assertTrue(out["ok"])
        self.assertEqual(out["data"], {"ok_data": 1})

    def test_失败_错误包装(self):
        with mock.patch('agent_cli._dispatch', side_effect=ValueError("boom")):
            out = self._run_main('{"fn":"query","args":{}}')
        self.assertFalse(out["ok"])
        self.assertIn("boom", out["error"])

    def test_空输入(self):
        out = self._run_main('')
        self.assertFalse(out["ok"])
        self.assertIn("empty", out["error"])

    def test_非法JSON(self):
        out = self._run_main('not json')
        self.assertFalse(out["ok"])

class TestSchemaConsistency(unittest.TestCase):
    """skill-tools.json（schema 单一来源）与 agent_cli 分派实现一致（R6）。"""

    def test_schema工具名全部分派支持(self):
        import os
        p = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                         "skills", "liki-bazi", "tools", "skill-tools.json")
        with open(p, encoding="utf-8") as f:
            doc = json.load(f)
        names = [t["function"]["name"] for t in doc["tools"]]
        self.assertEqual(len(names), 6)
        # 实际分派验证：白名单 6 名全部分派成功
        for n in names:
            with mock.patch("agent_cli.full_paipan"), \
                 mock.patch("agent_cli.city_coords"), \
                 mock.patch("agent_cli.query"), \
                 mock.patch("agent_cli.yearly_range"), \
                 mock.patch("agent_cli.calibrate"), \
                 mock.patch("agent_cli.bond"):
                if n == "full_paipan":
                    agent_cli._dispatch(n, {"gregorian": "t", "gender": "male"})
                elif n == "city_coords":
                    agent_cli._dispatch(n, {"city": "北京"})
                elif n == "yearly_range":
                    agent_cli._dispatch(n, {"pan": {}, "start": 2025, "end": 2026})
                elif n == "calibrate":
                    agent_cli._dispatch(n, {"candidates": [], "events": []})
                elif n == "bond":
                    agent_cli._dispatch(n, {"pan_a": {}, "pan_b": {}})
                else:
                    agent_cli._dispatch(n, {"rule": "marriage", "pan": {}})

    def test_schema_rule_enums_match_runtime_whitelists(self):
        from duanyu import _NATAL_RULES, _YEARLY_RULES

        p = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                         "skills", "liki-bazi", "tools", "skill-tools.json")
        with open(p, encoding="utf-8") as f:
            tools = {
                item["function"]["name"]: item["function"]
                for item in json.load(f)["tools"]
            }

        assert set(tools["query"]["parameters"]["properties"]["rule"]["enum"]) == _NATAL_RULES
        assert set(tools["yearly_range"]["parameters"]["properties"]["rules"]["items"]["enum"]) == _YEARLY_RULES
        assert set(tools["calibrate"]["parameters"]["properties"]["events"]["items"]
                   ["properties"]["rule"]["enum"]) == _YEARLY_RULES


if __name__ == "__main__":
    unittest.main()
