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
    """全链路集成测试：full_paipan→liunian→make_liunian_factors→query。"""

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
            pan = call("full_paipan", {"time": "1990-06-01T12:00:00+08:00",
                                       "gender": "male", "longitude": 116.4, "correct": True})
        except Exception as e:  # noqa: BLE001
            self.fail(f"LIKI_RPC_URL={url} 已设置但引擎不可达: {e}")
        if not pan.get("ok"):
            self.fail(f"full_paipan 失败: {pan.get('error')}")

        ln = call("liunian", {"pan": pan["data"], "year": 2006})
        self.assertTrue(ln["ok"], ln.get("error"))
        lf = call("make_liunian_factors", {"pan": pan["data"], "liunian_pan": ln["data"],
                                           "target": "配偶星", "year": 2006})
        self.assertTrue(lf["ok"], lf.get("error"))
        q = call("query", {"rule": "yearly_marriage", "snapshots": lf["data"]})
        self.assertTrue(q["ok"], q.get("error"))
        self.assertGreater(len(lf["data"]["八字"]), 0)
        self.assertIn("八字", q["data"])


if __name__ == "__main__":
    unittest.main()
