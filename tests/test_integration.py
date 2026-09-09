"""全链路集成测试（需真实引擎服务，LIKI_RPC_URL）。

独立文件——无服务阶段（make test / CI / test-skills）用 --ignore 文件级排除，
不显示 deselected（与"全程无排除显示"一致）；服务已起阶段（make test-all Docker 段）全量运行。

失败语义：LIKI_RPC_URL 已显式设置（make test-all / CI e2e）时，引擎不可达或链路
失败一律算 FAIL——防止本地引擎没起来时静默 skip 造成"假绿"；未设置时（裸跑
pytest）保留 skip。
"""
import json
import os
import subprocess
import sys
import unittest

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                                'skills', 'liki-bazi', 'tools'))


@pytest.mark.integration
class TestIntegration_FullChain(unittest.TestCase):
    """全链路集成测试：full_paipan → query（本命） + yearly_range（流年）。"""

    def test_full_chain(self):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                           "skills", "liki-bazi", "tools", "agent_cli.py")
        env = dict(os.environ, LIKI_RPC_URL=url)

        def call(fn, args):
            p = subprocess.run(["python3", cli], input=json.dumps({"fn": fn, "args": args}).encode(),
                               capture_output=True, env=env, timeout=60)
            return json.loads(p.stdout)

        try:
            pan = call("full_paipan", {"gregorian": "1990-06-01T12:00:00+08:00",
                                       "gender": "male", "longitude": 116.4, "correct": True})
        except Exception as e:  # noqa: BLE001
            self.fail(f"LIKI_RPC_URL={url} 已设置但引擎不可达: {e}")
        if not pan.get("ok"):
            self.fail(f"full_paipan 失败: {pan.get('error')}")
        self.assertIsInstance(pan["data"]["ziwei_daxian"], list)

        q = call("query", {"rule": "十神", "pan": pan["data"]})
        self.assertTrue(q["ok"], q.get("error"))
        self.assertIn("八字", q["data"])

        dx = call("query", {"rule": "大限", "pan": pan["data"], "year": 2000})
        self.assertTrue(dx["ok"], dx.get("error"))
        self.assertIn("合参", dx["data"])
        self.assertEqual(dx["data"]["current_year"], 2000)
        self.assertEqual(dx["data"]["current_year_source"], "specified")
        self.assertTrue(any(row["id"].startswith("dx_") for row in dx["data"]["紫微"]))

        yr = call("yearly_range", {"pan": pan["data"], "start": 2006, "end": 2006,
                                   "rules": ["yearly_marriage", "yingqi"]})
        self.assertTrue(yr["ok"], yr.get("error"))
        self.assertIn("current_year", yr["data"])
        self.assertIn("2006", yr["data"]["years"])
        self.assertTrue(all("合参" in result for result in yr["data"]["years"]["2006"].values()))

    def test_calibrate_full_chain(self):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                           "skills/liki-bazi/tools/agent_cli.py")
        env = dict(os.environ, LIKI_RPC_URL=url)

        def call(fn, args):
            p = subprocess.run(["python3", cli], input=json.dumps({"fn": fn, "args": args}).encode(),
                               capture_output=True, env=env, timeout=60)
            return json.loads(p.stdout)

        candidates = [
            {"label": "11时", "gregorian": "1990-06-01T11:00:00+08:00",
             "gender": "male", "longitude": 116.4, "correct": True},
            {"label": "12时", "gregorian": "1990-06-01T12:00:00+08:00",
             "gender": "male", "longitude": 116.4, "correct": True},
        ]
        events = [
            {"year": 2006, "rule": "yearly_marriage", "label": "婚恋"},
            {"year": 2006, "rule": "yingqi", "label": "应期"},
            {"year": 2006, "rule": "yearly_career", "label": "事业"},
        ]
        result = call("calibrate", {"candidates": candidates, "events": events, "detail": True})
        self.assertTrue(result["ok"], result.get("error"))
        self.assertEqual(set(result["data"]), {"11时", "12时"})
        self.assertTrue(
            all(set(event) >= {"八字", "紫微", "合参"}
                for events in result["data"].values() for event in events)
        )
        self.assertTrue(
            all("evidence" in event
                for events in result["data"].values() for event in events)
        )


@pytest.mark.integration
class TestIntegration_QimenRules(unittest.TestCase):
    """奇门 RPC 排盘 + Python 表驱动解释链路。"""

    def test_qimen_specialized_rules_full_chain(self):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            "skills", "liki-divination", "tools", "agent_cli.py",
        )
        env = dict(os.environ, LIKI_RPC_URL=url)

        def call(fn, args):
            p = subprocess.run(
                ["python3", cli],
                input=json.dumps({"fn": fn, "args": args}).encode(),
                capture_output=True,
                env=env,
                timeout=60,
            )
            return json.loads(p.stdout)

        solar_time = "2026-06-28T12:00:00+08:00"
        longitude = 120.0
        base = call("qimen_read", {
            "question": "当前态势",
            "time": solar_time,
            "longitude": longitude,
        })
        self.assertTrue(base["ok"], base.get("error"))
        self.assertIsNone(base["data"]["matter"])
        self.assertIn("method", base["data"]["chart"])
        self.assertIn("snapshot", base["data"])

        specialized_rules = (
            "lost_property", "thief_capture", "thief_profile", "capture_escape",
        )
        for rule in specialized_rules:
            result = call("qimen_read", {
                "question": f"专占：{rule}",
                "time": solar_time,
                "longitude": longitude,
                "rule": rule,
            })
            self.assertTrue(result["ok"], result.get("error"))
            self.assertEqual(result["data"]["special"]["rule"], rule)
            assertions = result["data"]["special"]["assertions"]
            self.assertIsInstance(assertions, list)
            self.assertTrue(all(item["basis"] for item in assertions))

        missing = call("qimen_read", {
            "question": "家人走失了",
            "time": solar_time,
            "longitude": longitude,
            "matter": "missing_person",
            "rule": "missing_person",
        })
        self.assertTrue(missing["ok"], missing.get("error"))
        self.assertEqual(missing["data"]["matter"]["matter"], "missing_person")
        self.assertEqual(
            missing["data"]["special"]["rule"], "missing_person"
        )
        self.assertTrue(
            all(item["basis"] for item in missing["data"]["special"]["assertions"])
        )


if __name__ == "__main__":
    unittest.main()
