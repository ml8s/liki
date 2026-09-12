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
class TestIntegration_DivinationSnapshotAsk(unittest.TestCase):
    """六爻 / 奇门 snapshot → ask 的真实引擎链路。"""

    def call_tool(self, fn: str, args: dict):
        url = os.environ.get("LIKI_RPC_URL", "")
        if not url:
            self.skipTest("LIKI_RPC_URL 未设置，跳过全链路集成测试")
        cli = os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            "skills", "liki-divination", "tools", "agent_cli.py",
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
        self.assertEqual(snapshot["schema_version"], "liuyao-snapshot-v5")
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
        self.assertEqual(result["schema_version"], "huangli-days-v1")
        self.assertEqual(result["range"]["days"], 3)
        self.assertLessEqual(len(result["candidates"]), 3)

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
        self.assertEqual(snapshot["schema_version"], "qimen-snapshot-v4")
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
