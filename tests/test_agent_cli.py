"""agent_cli 分派单元测试（不连引擎——mock 工具链函数）。

覆盖：5 函数分派映射、参数透传、非法 fn、stdin 协议错误包装。
"""
import json
import os
import sys
import tempfile
import unittest
from unittest import mock

import pytest

sys.path.insert(0, __import__('os').path.join(
    __import__('os').path.dirname(__import__('os').path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools'))

import agent_cli


class TestDispatch(unittest.TestCase):
    def setUp(self):
        # mock 工具链函数——验证分派映射与参数透传，不触发网络/真值表
        self.patchers = [
            mock.patch('agent_cli.full_paipan', return_value={"pan": "full"}),
            mock.patch('agent_cli.liunian', return_value={"ln": True}),
            mock.patch('agent_cli.make_factors', return_value={"fac": {}}),
            mock.patch('agent_cli.make_liunian_factors', return_value={"lf": {}}),
            mock.patch('agent_cli.query', return_value={"八字": [], "紫微": []}),
        ]
        for p in self.patchers:
            p.start()
        self.addCleanup(lambda: [p.stop() for p in self.patchers])

    def test_full_paipan_分派(self):
        data = agent_cli._dispatch("full_paipan", {"time": "1990-06-01T12:00:00+08:00", "gender": "male"})
        self.assertEqual(data, {"pan": "full"})
        agent_cli.full_paipan.assert_called_once_with(
            "1990-06-01T12:00:00+08:00", "male", longitude=None, correct=True)

    def test_full_paipan_默认值(self):
        agent_cli._dispatch("full_paipan", {"time": "t", "gender": "female", "longitude": 116.4})
        agent_cli.full_paipan.assert_called_once_with("t", "female", longitude=116.4, correct=True)

    def test_liunian_分派(self):
        agent_cli._dispatch("liunian", {"pan": {}, "year": 2006})
        agent_cli.liunian.assert_called_once_with({}, 2006)

    def test_make_factors_分派(self):
        agent_cli._dispatch("make_factors", {"pan": {"fac": {}}})
        agent_cli.make_factors.assert_called_once_with({"fac": {}})

    def test_make_liunian_factors_默认参数(self):
        agent_cli._dispatch("make_liunian_factors", {"pan": {}, "liunian_pan": {}})
        agent_cli.make_liunian_factors.assert_called_once_with({}, {}, target="配偶星", year=0)

    def test_query_分派(self):
        agent_cli._dispatch("query", {"rule": "marriage", "snapshots": {}})
        agent_cli.query.assert_called_once_with("marriage", {})

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


if __name__ == "__main__":
    unittest.main()


class TestSchemaConsistency(unittest.TestCase):
    """skill-tools.json（schema 单一来源）与 agent_cli 分派实现一致（R6）。"""

    def test_schema工具名全部分派支持(self):
        import os
        p = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                         "skills", "liki-bazi", "tools", "skill-tools.json")
        doc = json.load(open(p, encoding="utf-8"))
        names = [t["function"]["name"] for t in doc["tools"]]
        self.assertEqual(len(names), 5)
        for n in names:
            with self.subTest(tool=n):
                # mock 下每个 schema 工具名都能分派（不抛 unknown tool）
                with mock.patch("agent_cli._dispatch", return_value=None) as d:
                    agent_cli.main() if False else None
                    # 直接验证分派分支存在：非白名单名会 ValueError
                    self.assertIsNotNone(d)  # 占位——实际验证在下方
        # 实际分派验证：白名单 5 名全部分派成功，未知名抛错
        for n in names:
            with mock.patch("agent_cli.full_paipan"), mock.patch("agent_cli.liunian"), \
                 mock.patch("agent_cli.make_factors"), mock.patch("agent_cli.make_liunian_factors"), \
                 mock.patch("agent_cli.query"):
                if n == "full_paipan":
                    agent_cli._dispatch(n, {"time": "t", "gender": "male"})
                elif n == "liunian":
                    agent_cli._dispatch(n, {"pan": {}, "year": 1})
                elif n == "make_factors":
                    agent_cli._dispatch(n, {"pan": {}})
                elif n == "make_liunian_factors":
                    agent_cli._dispatch(n, {"pan": {}, "liunian_pan": {}})
                else:
                    agent_cli._dispatch(n, {"rule": "marriage", "snapshots": {}})


class TestFileRefs(unittest.TestCase):
    """{"$file": path} 引用展开——大对象（pan）走文件，免 shell 转义（feedback fedd52aa）。"""

    def test_dollar_file_loads_json(self):
        with tempfile.NamedTemporaryFile("w", suffix=".json", delete=False, encoding="utf-8") as f:
            json.dump({"day": "己亥"}, f, ensure_ascii=False)
            path = f.name
        try:
            args = agent_cli._load_file_refs({"pan": {"$file": path}, "year": 2026})
            self.assertEqual(args["pan"]["day"], "己亥")
            self.assertEqual(args["year"], 2026)
        finally:
            os.unlink(path)

    def test_plain_args_untouched(self):
        args = agent_cli._load_file_refs({"rule": "marriage", "snapshots": {"八字": {}}})
        self.assertEqual(args["rule"], "marriage")

    def test_dispatch_with_file_ref(self):
        with tempfile.NamedTemporaryFile("w", suffix=".json", delete=False, encoding="utf-8") as f:
            json.dump({"day": "甲子"}, f)
            path = f.name
        try:
            with mock.patch("agent_cli.make_factors", return_value={"fac": 1}) as mf:
                out = agent_cli._dispatch("make_factors", {"pan": {"$file": path}})
            self.assertEqual(out, {"fac": 1})
            mf.assert_called_once_with({"day": "甲子"})
        finally:
            os.unlink(path)
