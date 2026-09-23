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
                                'analysis', 'app', 'natal', 'tools'))


@pytest.mark.integration
class TestIntegration_NatalTools(unittest.TestCase):
    """全链路集成测试：create_birth_chart → analyze / periods / compare / calibrate。"""

    def call_tool(self, fn: str, args: dict):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                           "analysis", "app", "natal", "tools", "agent_cli.py")
        env = dict(os.environ, LIKI_RPC_URL=url)
        p = subprocess.run(
            ["python3", cli], input=json.dumps({"fn": fn, "args": args}).encode(),
            capture_output=True, env=env, timeout=60,
        )
        return json.loads(p.stdout)

    def test_natal_full_chain(self):
        created = self.call_tool("create_birth_chart", {
            "gender": "male",
            "source": {
                "type": "timestamp",
                "timestamp": "1990-06-01T12:00:00+08:00",
                "location": {"longitude": 116.4},
                "solar_time_correction": "auto",
            },
        })
        self.assertTrue(created.get("ok"), created.get("error"))
        chart = created["data"]["chart"]
        ref = created["data"]["chart_ref"]
        self.assertEqual(chart["completeness"], "full")
        self.assertIn("pillars", chart["bazi"])
        self.assertIsInstance(chart["ziwei"]["ming_gong"], str)

        natal = self.call_tool("analyze_natal", {
            "chart_ref": ref,
            "topics": ["personality"],
        })
        self.assertTrue(natal.get("ok"), natal.get("error"))
        self.assertEqual(natal["data"]["query"]["scope"], "natal")
        self.assertEqual(natal["data"]["query"]["topics"], ["personality"])
        for row in natal["data"]["assertions"]:
            self.assertEqual(row["topic"], "personality")
            self.assertEqual(row["time_scope"], "natal")
            self.assertIn(row["side"], {"bazi", "ziwei", "combined"})

        periods = self.call_tool("analyze_periods", {
            "chart_ref": ref,
            "time_scope": {"type": "year", "year": 2006},
            "topics": ["marriage"],
        })
        self.assertTrue(periods.get("ok"), periods.get("error"))
        self.assertEqual(periods["data"]["query"]["scope"], "periods")
        self.assertEqual(periods["data"]["periods"][0]["time_scope"], {"type": "year", "year": 2006})
        for row in periods["data"]["periods"][0]["assertions"]:
            self.assertEqual(row["topic"], "marriage")
            self.assertEqual(row["year"], 2006)

    def test_compare_and_calibrate_full_chain(self):
        common_source = {
            "type": "timestamp",
            "timestamp": "1990-06-01T12:00:00+08:00",
            "location": {"longitude": 116.4},
            "solar_time_correction": "auto",
        }
        a = self.call_tool("create_birth_chart", {"gender": "male", "source": common_source})
        b = self.call_tool("create_birth_chart", {"gender": "female", "source": common_source})
        self.assertTrue(a.get("ok"), a.get("error"))
        self.assertTrue(b.get("ok"), b.get("error"))
        compared = self.call_tool("compare_birth_charts", {
            "chart_ref_a": a["data"]["chart_ref"],
            "chart_ref_b": b["data"]["chart_ref"],
        })
        self.assertTrue(compared.get("ok"), compared.get("error"))
        self.assertIn("bazi", compared["data"]["comparison"])
        self.assertIn("ziwei", compared["data"]["comparison"])

        calibrated = self.call_tool("calibrate_birth_time", {
            "candidates": [
                {"label": "11时", "gender": "male", "source": common_source},
                {"label": "12时", "gender": "male", "source": common_source},
            ],
            "events": [
                {"year": 2006, "topic": "marriage", "label": "婚恋"},
                {"year": 2006, "topic": "career", "label": "事业"},
                {"year": 2006, "topic": "relocation", "label": "搬迁"},
            ],
            "detail": True,
        })
        self.assertTrue(calibrated.get("ok"), calibrated.get("error"))
        self.assertEqual(
            [item["label"] for item in calibrated["data"]["candidates"]],
            ["11时", "12时"],
        )
        for candidate in calibrated["data"]["candidates"]:
            self.assertEqual(len(candidate["events"]), 3)


@pytest.mark.integration
class TestIntegration_DivinationSnapshotAsk(unittest.TestCase):
    """六爻 / 奇门 snapshot → ask 的真实引擎链路。"""

    def call_tool(self, fn: str, args: dict):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            "analysis", "app", "divination", "tools", "agent_cli.py",
        )
        env = dict(os.environ, LIKI_RPC_URL=url)
        process = subprocess.run(
            ["python3", cli],
            input=json.dumps({"fn": fn, "args": args}).encode(),
            capture_output=True,
            env=env,
            timeout=60,
        )
        result = json.loads(process.stdout)
        self.assertTrue(result.get("ok"), result.get("error"))
        return result["data"]

    def test_liuyao_snapshot_ask_and_tamper_rejection(self):
        snapshot = self.call_tool(**{
            "fn": "liuyao_snapshot",
            "args": {
                "question": "这次面试能不能通过？",
                "mode": "yaos",
                "yaos": [7, 7, 7, 7, 7, 7],
                "matter": "career",
            },
        })
        self.assertEqual(snapshot["method"], "liuyao")
        self.assertEqual(snapshot["schema_version"], "liuyao-snapshot-v7")
        self.assertTrue(snapshot["snapshot_digest"])
        self.assertIn("focus", snapshot)

        answer = self.call_tool(**{
            "fn": "liuyao_ask",
            "args": {
                "snapshot": snapshot,
                "message": "现在应该注意什么？",
            },
        })
        self.assertEqual(answer["method"], "liuyao")
        self.assertEqual(answer["snapshot_digest"], snapshot["snapshot_digest"])
        self.assertTrue(answer["audit"]["accepted"])

        tampered = dict(snapshot)
        tampered["question"] = dict(snapshot["question"], text="篡改后的问题")
        with self.assertRaises(AssertionError):
            self.call_tool(**{
                "fn": "liuyao_ask",
                "args": {"snapshot": tampered, "message": "现在应该注意什么？"},
            })

    def test_huangli_days_contract(self):
        result = self.call_tool(**{
            "fn": "huangli_days",
            "args": {"question": "哪天适合签约？", "event": "sign", "days": 3},
        })
        self.assertEqual(result["schema_version"], "huangli-days-v2")
        self.assertEqual(result["range"]["days"], 3)
        self.assertLessEqual(len(result["candidates"]), 3)

    def test_huangli_medical_topic_returns_result_with_safety_advisory(self):
        result = self.call_tool(**{
            "fn": "huangli_days",
            "args": {
                "question": "手术后的复诊日期哪天适合？",
                "event": "medical",
                "days": 3,
            },
        })
        self.assertEqual(result["schema_version"], "huangli-days-v2")
        self.assertTrue(result["candidates"])
        advisory = result["safety_advisory"]
        self.assertEqual(advisory["status"], "advisory")
        self.assertFalse(advisory["blocking"])
        self.assertEqual(advisory["category"], "serious_medical")
        self.assertIn("医生", advisory["guidance"])

    def test_liuyao_other_matter_routes_to_response_line(self):
        snapshot = self.call_tool(**{
            "fn": "liuyao_snapshot",
            "args": {
                "question": "对方现在怎么看这件事？",
                "matter": "other",
                "mode": "yaos",
                "yaos": [7, 7, 7, 7, 7, 7],
            },
        })
        self.assertEqual(snapshot["focus"]["yong_shen"]["name"], "应爻")
        self.assertEqual(snapshot["focus"]["yong_shen"]["position"], 3)
        self.assertEqual(snapshot["focus"]["yong_line"]["shi_ying"], "应")

    def test_liuyao_hidden_yong_shen_contract(self):
        snapshot = self.call_tool(**{
            "fn": "liuyao_snapshot",
            "args": {
                "question": "这笔财能不能拿到？",
                "mode": "yaos",
                "yaos": [8, 7, 9, 9, 7, 7],
                "yong_shen": "妻财",
            },
        })
        focus = snapshot["focus"]
        self.assertEqual(focus["yong_shen"]["position"], 0)
        self.assertTrue(focus["yong_shen"]["is_hidden"])
        self.assertEqual(focus["yong_shen"]["fu_shen"]["zhi"], "寅")
        self.assertIsNone(focus["yong_line"])
        primary_ids = {item["id"] for item in snapshot["evidence"]["primary"]}
        self.assertIn("yong-shen-hidden-state", primary_ids)
        timing_mechanisms = {item["mechanism"] for item in snapshot["timing_candidates"]}
        self.assertIn("冲飞出伏", timing_mechanisms)

        answer = self.call_tool(**{
            "fn": "liuyao_ask",
            "args": {"snapshot": snapshot, "message": "现在应该注意什么？"},
        })
        self.assertTrue(answer["audit"]["accepted"])
        self.assertIn("yong-shen-hidden-state", answer["primary_evidence_refs"])

    def test_qimen_jinhan_snapshot_ask_chain(self):
        snapshot = self.call_tool(**{
            "fn": "qimen_snapshot",
            "args": {
                "question": "今天整体态势如何？",
                "longitude": 121.47,
                "time": "2026-09-08T12:00:00+08:00",
                "scope": "day",
                "school": "jinhan_yujing",
            },
        })
        self.assertEqual(snapshot["snapshot_kind"], "jinhan")
        self.assertEqual(snapshot["factors"]["kind"], "jinhan")
        self.assertEqual(len(snapshot["factors"]["palaces"]), 9)
        self.assertEqual(len(snapshot["factors"]["day_spirits"]), 12)

        answer = self.call_tool(**{
            "fn": "qimen_ask",
            "args": {"snapshot": snapshot, "message": "今天整体态势如何？"},
        })
        self.assertEqual(answer["method"], "qimen")
        self.assertEqual(answer["snapshot_digest"], snapshot["snapshot_digest"])
        self.assertEqual(answer["assertion_refs"], [])
        self.assertEqual(answer["timing_refs"], [])

    def test_qimen_snapshot_ask_chain(self):
        snapshot = self.call_tool(**{
            "fn": "qimen_snapshot",
            "args": {
                "question": "当前应该往哪个方向推进？",
                "time": "2026-06-28T12:00:00+08:00",
                "longitude": 120.0,
                "matter": "career",
            },
        })
        self.assertEqual(snapshot["method"], "qimen")
        self.assertEqual(snapshot["schema_version"], "qimen-snapshot-v5")
        self.assertIn("method_context", snapshot)

        answer = self.call_tool(**{
            "fn": "qimen_ask",
            "args": {
                "snapshot": snapshot,
                "message": "现在适合推进吗？",
            },
        })
        self.assertEqual(answer["method"], "qimen")
        self.assertEqual(answer["snapshot_digest"], snapshot["snapshot_digest"])
        self.assertTrue(answer["audit"]["accepted"])

@pytest.mark.integration
class TestIntegration_QimenRules(unittest.TestCase):
    """奇门 RPC 排盘 + Python 表驱动解释链路。"""

    def test_qimen_specialized_rules_full_chain(self):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            "analysis", "app", "divination", "tools", "agent_cli.py",
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
        base = call("qimen_snapshot", {
            "question": "当前事业态势",
            "time": solar_time,
            "longitude": longitude,
            "matter": "career",
        })
        self.assertTrue(base["ok"], base.get("error"))
        self.assertEqual(base["data"]["matter"]["matter"], "career")
        self.assertIn("method_context", base["data"])
        self.assertIn("snapshot_digest", base["data"])

        specialized_rules = (
            "lost_property", "thief_capture", "thief_profile", "capture_escape",
        )
        for rule in specialized_rules:
            result = call("qimen_snapshot", {
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

        missing = call("qimen_snapshot", {
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
