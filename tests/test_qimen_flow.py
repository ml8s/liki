from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_report  # noqa: E402
import qimen_session  # noqa: E402


def _read_result():
    return {
        "schema_version": "qimen-read-v1",
        "question": "该往哪个方向推进？",
        "input": {
            "city": "上海", "longitude": 121.47,
            "local_time": "2026-09-08T12:00:00+08:00",
            "solar_time": "2026-09-08T11:57:00+08:00",
        },
        "snapshot": {
            "method": {"scope": "hour", "school": "zhuanpan"},
            "ri_gan_gong": "震", "shi_gan_gong": "兑", "ying_qi": [],
        },
        "special": None,
    }


def test_read_report_session_flow():
    read_result = _read_result()
    report = qimen_report.template(read_result)
    report.update({
        "headline": "当前路径有支持，方向仍需现实核验",
        "verdict": "日时宫位可作推进候选，但不能替代合同与成本核查。",
        "action": "先做一次现实条件核查，再按候选方向推进。",
    })
    audit = qimen_report.validate_report(report, read_result)
    assert audit["accepted"] is True
    session = qimen_session.create_session(
        read_result=read_result, headline=report["headline"],
        verdict=report["verdict"], confidence="medium", report=report,
    )
    assert session["snapshot_digest"]
    assert session["policy"]["immutable_input"] is True
