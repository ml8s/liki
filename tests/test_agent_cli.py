"""agent_cli 分派与协议测试；领域服务全部 mock，不触发引擎。"""
import json
import os
import sys
import unittest
from unittest import mock

sys.path.insert(0, os.path.join(
    os.path.dirname(os.path.abspath(__file__)), '..', 'analysis', 'app', 'natal', 'tools'))

import agent_cli


class TestDispatch(unittest.TestCase):
    def setUp(self):
        self.patchers = [
            mock.patch('agent_cli.create_birth_chart', return_value={"chart": {}}),
            mock.patch('agent_cli.analyze_natal', return_value={"assertions": []}),
            mock.patch('agent_cli.analyze_periods', return_value={"periods": []}),
            mock.patch('agent_cli.compare_birth_charts', return_value={"comparison": {}}),
            mock.patch('agent_cli.calibrate_birth_time', return_value={"candidates": []}),
        ]
        for patcher in self.patchers:
            patcher.start()
        self.addCleanup(lambda: [patcher.stop() for patcher in self.patchers])

    def test_create_birth_chart_dispatch(self):
        args = {"gender": "male", "source": {"type": "hour", "timestamp": "1990-06-01T12:00:00+08:00", "solar_time_correction": "off"}}
        self.assertEqual(agent_cli._dispatch("create_birth_chart", args), {"chart": {}})
        agent_cli.create_birth_chart.assert_called_once_with(args)

    def test_analyze_natal_dispatch(self):
        args = {"chart_ref": {"token": "t", "digest": "d"}, "topics": ["marriage"]}
        self.assertEqual(agent_cli._dispatch("analyze_natal", args), {"assertions": []})
        agent_cli.analyze_natal.assert_called_once_with(args)

    def test_analyze_periods_dispatch(self):
        args = {
            "chart_ref": {"token": "t", "digest": "d"},
            "time_scope": {"type": "year", "year": 2026},
            "topics": ["marriage"],
        }
        self.assertEqual(agent_cli._dispatch("analyze_periods", args), {"periods": []})
        agent_cli.analyze_periods.assert_called_once_with(args)

    def test_compare_birth_charts_dispatch(self):
        args = {
            "chart_ref_a": {"token": "a", "digest": "a"},
            "chart_ref_b": {"token": "b", "digest": "b"},
        }
        agent_cli._dispatch("compare_birth_charts", args)
        agent_cli.compare_birth_charts.assert_called_once_with(args)

    def test_calibrate_birth_time_dispatch(self):
        args = {"candidates": [{"label": "A", "gender": "male", "source": {"type": "hour", "timestamp": "t", "solar_time_correction": "off"}}], "events": []}
        agent_cli._dispatch("calibrate_birth_time", args)
        agent_cli.calibrate_birth_time.assert_called_once_with(args)

    def test_unknown_tool(self):
        with self.assertRaises(ValueError):
            agent_cli._dispatch("query", {})


class TestMainProtocol(unittest.TestCase):
    def _run_main(self, stdin_text):
        out = {}
        with mock.patch('agent_cli.ensure_engine_compatible'), \
             mock.patch('sys.stdin') as stdin, \
             mock.patch('builtins.print') as printed:
            stdin.read.return_value = stdin_text
            agent_cli.main()
            out = json.loads(printed.call_args.args[0])
        return out

    def test_unknown_and_missing_args_skip_version_check(self):
        with mock.patch("agent_cli.ensure_engine_compatible") as version_check:
            out = self._run_main('{"fn":"query","args":{}}')
            self.assertFalse(out["ok"])
            self.assertEqual(out["error"]["code"], "INVALID_INPUT")
            self.assertIn("unknown tool", out["error"]["message"])

            out = self._run_main('{"fn":"analyze_natal","args":{}}')
            self.assertIn("missing arg", out["error"]["message"])
        version_check.assert_not_called()

    def test_success(self):
        with mock.patch('agent_cli._dispatch', return_value={"ok_data": 1}):
            out = self._run_main('{"fn":"analyze_natal","args":{"chart_ref":{},"topics":["marriage"]}}')
        self.assertTrue(out["ok"])
        self.assertEqual(out["data"], {"ok_data": 1})

    def test_error_contract(self):
        with mock.patch('agent_cli._dispatch', side_effect=ValueError("boom")):
            out = self._run_main('{"fn":"analyze_natal","args":{"chart_ref":{},"topics":["marriage"]}}')
        self.assertFalse(out["ok"])
        self.assertEqual(out["error"], {"code": "INVALID_INPUT", "message": "boom"})

    def test_empty_input(self):
        out = self._run_main('')
        self.assertEqual(out["error"]["code"], "INVALID_INPUT")

    def test_invalid_json(self):
        out = self._run_main('no-json')
        self.assertEqual(out["error"]["code"], "INVALID_INPUT")


class TestSchemaConsistency(unittest.TestCase):
    def test_schema_tools_match_dispatch(self):
        p = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                         "skills", "liki", "natal", "tools", "skill-tools.json")
        with open(p, encoding="utf-8") as f:
            names = [item["function"]["name"] for item in json.load(f)["tools"]]
        self.assertEqual(set(names), set(agent_cli._DISPATCH))
        self.assertEqual(len(names), 5)

    def test_schema_required_args_match_cli_precheck(self):
        p = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                         "analysis/app/natal/tools/skill-tools.json")
        with open(p, encoding="utf-8") as f:
            schema_required = {
                item["function"]["name"]: set(item["function"]["parameters"]["required"])
                for item in json.load(f)["tools"]
            }
        self.assertEqual({name: set(args) for name, args in agent_cli._REQUIRED_ARGS.items()}, schema_required)


if __name__ == "__main__":
    unittest.main()
