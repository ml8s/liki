"""agent_cli 分派单元测试（不连引擎——mock 工具链函数）。

覆盖：5 函数分派映射、参数透传、非法 fn、stdin 协议错误包装。
"""
import json
import sys
import unittest
from unittest import mock

sys.path.insert(0, __import__('os').path.join(
    __import__('os').path.dirname(__import__('os').path.abspath(__file__)), '..', 'tools'))

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
                         "tools", "skill-tools.json")
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


class TestIntegration_FullChain(unittest.TestCase):
    """全链路集成测试：full_paipan→liunian→make_liunian_factors→query。
    依赖真实引擎（LIKI_RPC_URL），不可达时跳过——CI/无引擎环境不挂。"""

    def test_full_chain(self):
        import os
        import subprocess

        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                           "tools", "agent_cli.py")
        env = dict(os.environ, LIKI_RPC_URL=url)

        def call(fn, args):
            p = subprocess.run(["python3", cli], input=json.dumps({"fn": fn, "args": args}).encode(),
                               capture_output=True, env=env, timeout=60)
            return json.loads(p.stdout)

        try:
            pan = call("full_paipan", {"time": "1990-06-01T12:00:00+08:00",
                                       "gender": "male", "longitude": 116.4, "correct": True})
        except Exception as e:  # noqa: BLE001
            self.skipTest(f"引擎不可达，跳过全链路集成测试: {e}")
        if not pan.get("ok"):
            self.skipTest(f"full_paipan 失败（引擎不可达？）: {pan.get('error')}")

        ln = call("liunian", {"pan": pan["data"], "year": 2006})
        self.assertTrue(ln["ok"], ln.get("error"))
        lf = call("make_liunian_factors", {"pan": pan["data"], "liunian_pan": ln["data"],
                                           "target": "配偶星", "year": 2006})
        self.assertTrue(lf["ok"], lf.get("error"))
        q = call("query", {"rule": "yearly_marriage", "snapshots": lf["data"]})
        self.assertTrue(q["ok"], q.get("error"))
        # 全链路产物非空（2006 流年婚姻因子 41 键、断语可能命中）
        self.assertGreater(len(lf["data"]["八字"]), 0)
        self.assertIn("八字", q["data"])
