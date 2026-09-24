"""正交化因子解析一致性测试。

组合盘（analysis 输入：engine 排盘 chart）产出的因子必须与完整盘
（full_paipan，旧分析层）同侧因子完全一致——正交化只改输入来源，
不改因子求值逻辑。覆盖本命因子与流年因子、八字与紫微、多样本。
"""
from __future__ import annotations

import os

os.environ.setdefault("LIKI_COUNSEL_SERVICE_DOMAIN", "bazi")

import json  # noqa: E402
import sys  # noqa: E402
import urllib.request  # noqa: E402
from pathlib import Path  # noqa: E402

import pytest  # noqa: E402

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "natal" / "tools"))

SAMPLES = [
    ("1990-05-20T11:49:00+08:00", "male", 116.4),
    ("1985-11-03T08:15:00+08:00", "female", 104.07),
    ("2000-01-01T00:30:00+08:00", "male", 121.47),
]


def _engine_call(domain, name, args):
    url = os.environ["LIKI_MCP_URL"] + "/" + domain
    body = json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": "tools/call",
        "params": {"name": name, "arguments": args, "_meta": {
            "io.modelcontextprotocol/protocolVersion": "2026-07-28",
            "io.modelcontextprotocol/clientCapabilities": {},
        }},
    }).encode()
    req = urllib.request.Request(
        url, data=body,
        headers={
            "Content-Type": "application/json",
            "MCP-Protocol-Version": "2026-07-28",
            "Mcp-Method": "tools/call", "Mcp-Name": name,
            "Accept": "application/json, text/event-stream",
        },
    )
    with urllib.request.urlopen(req, timeout=8) as r:
        return json.loads(json.loads(r.read())["result"]["content"][0]["text"])


def _full_pan(solar_time, gender, longitude):
    from paipan import full_paipan

    return full_paipan(solar_time, gender, longitude=longitude)


def _bazi_combo_chart(solar_time, gender, longitude):
    return _engine_call("bazi", "bazi_chart", {
        "solar_time": solar_time, "gender": gender, "longitude": longitude,
    })


def _ziwei_combo_chart(solar_time, gender, longitude):
    t = _engine_call("aux", "tianwen_time", {"time": solar_time, "longitude": longitude})
    return _engine_call("ziwei", "ziwei_chart", {"lunar": t["lunar"], "gender": gender})


def _diff_fields(a: dict, b: dict) -> tuple[list[str], list[str]]:
    diff = [k for k in a if a.get(k) != b.get(k)]
    missing = [k for k in a if k not in b]
    return diff, missing


@pytest.mark.parametrize("sample", SAMPLES)
def test_natal_factors_match_full_paipan(sample):
    """本命因子：组合盘（compute_factors）== 完整盘（full_paipan）同侧。"""
    from factors import evaluate_factors
    from counsel import compute_factors

    solar_time, gender, longitude = sample
    pan = _full_pan(solar_time, gender, longitude)
    for side, shushi, chart in (
        ("bazi", "bazi", _bazi_combo_chart(solar_time, gender, longitude)),
        ("ziwei", "ziwei", _ziwei_combo_chart(solar_time, gender, longitude)),
    ):
        full = evaluate_factors(gender, pan, shushi=shushi)
        combo = compute_factors(chart)["factors"]
        diff, missing = _diff_fields(full, combo)
        assert not diff and not missing, (
            f"{side} 本命因子组合盘与完整盘不一致: diff={diff[:5]} missing={missing[:5]}"
        )


@pytest.mark.parametrize("sample", SAMPLES)
def test_liunian_factors_match_full_paipan(sample):
    """流年因子：组合盘 == 完整盘同侧（八字与紫微）。"""
    from duanyu import prepare_natal_context
    from factors import evaluate_liunian_snap_from_pan
    from paipan import _bazi_liunian, _ziwei_liunian, call

    solar_time, gender, longitude = sample
    pan = _full_pan(solar_time, gender, longitude)
    nc = prepare_natal_context(pan)
    full_snap = evaluate_liunian_snap_from_pan(
        pan,
        {"bazi": _bazi_liunian(pan["chart"], 2030), "ziwei": _ziwei_liunian(pan["ziwei"], 2030)},
        year=2030, natal_context=nc,
    )
    # 八字组合盘
    bc = _bazi_combo_chart(solar_time, gender, longitude)
    full = call("bazi.fullchart", {"chart": bc})["data"]
    comp_bazi = {"chart": bc, "full": full, "gender": gender}
    nc_b = prepare_natal_context(comp_bazi)
    snap_b = evaluate_liunian_snap_from_pan(
        comp_bazi, {"bazi": _bazi_liunian(comp_bazi["chart"], 2030), "ziwei": {}},
        year=2030, natal_context=nc_b,
    )
    diff, missing = _diff_fields(full_snap["八字"], snap_b["八字"])
    assert not diff and not missing, (
        f"八字流年因子组合盘与完整盘不一致: diff={diff[:5]} missing={missing[:5]}"
    )
    # 紫微组合盘
    zwc = _ziwei_combo_chart(solar_time, gender, longitude)
    zw = call("ziwei.fullchart", {"chart": zwc})["data"]
    daxian = call("ziwei.daxian", {"chart": zw})["data"]
    comp_zw = {"chart": zwc, "full": zw, "ziwei": zw, "ziwei_daxian": daxian, "gender": gender}
    nc_z = prepare_natal_context(comp_zw)
    snap_z = evaluate_liunian_snap_from_pan(
        comp_zw, {"bazi": {}, "ziwei": _ziwei_liunian(comp_zw["ziwei"], 2030)},
        year=2030, natal_context=nc_z,
    )
    diff, missing = _diff_fields(full_snap["紫微"], snap_z["紫微"])
    assert not diff and not missing, (
        f"紫微流年因子组合盘与完整盘不一致: diff={diff[:5]} missing={missing[:5]}"
    )