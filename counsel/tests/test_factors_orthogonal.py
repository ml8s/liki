"""Canonical-chart factor consistency across counsel and direct evaluation."""

import json
import os
import urllib.request

import pytest

SAMPLES = [
    ("1990-05-20T03:49:00Z", "Asia/Shanghai", "male", 116.4),
    ("1985-11-03T00:15:00Z", "Asia/Shanghai", "female", 104.07),
    ("1999-12-31T16:30:00Z", "Asia/Shanghai", "male", 121.47),
]


def _engine_call(name: str, args: dict):
    url = os.environ["LIKI_ENGINE_MCP_URL"]
    body = json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": "tools/call",
        "params": {"name": name, "arguments": args, "_meta": {
            "io.modelcontextprotocol/protocolVersion": "2026-07-28",
            "io.modelcontextprotocol/clientCapabilities": {},
        }},
    }).encode()
    request = urllib.request.Request(url, data=body, headers={
        "Content-Type": "application/json",
        "Accept": "application/json, text/event-stream",
        "MCP-Protocol-Version": "2026-07-28",
        "Mcp-Method": "tools/call", "Mcp-Name": name,
    })
    with urllib.request.urlopen(request, timeout=8) as response:
        result = json.loads(response.read())["result"]
    payload = result.get("structuredContent")
    if payload is None:
        payload = json.loads(result["content"][0]["text"])
    return payload


def _subject(instant: str, timezone: str, gender: str, longitude: float) -> dict:
    from tests.time_helpers import standard_time_from_instant
    return {
        "birth_time": standard_time_from_instant(instant, timezone),
        "birth_longitude": longitude,
        "gender": gender,
    }


def _full_pan(instant, timezone, gender, longitude):
    from tests.time_helpers import civil_time_from_instant
    from app.natal.tools.charts import build_natal_chart
    return build_natal_chart(
        civil_time_from_instant(instant, timezone), timezone, gender,
        longitude=longitude,
    )


def _diff_fields(a: dict, b: dict) -> tuple[list[str], list[str]]:
    return [k for k in a if a.get(k) != b.get(k)], [k for k in a if k not in b]


@pytest.mark.parametrize("sample", SAMPLES)
def test_natal_compute_factors_matches_direct_evaluation(sample):
    from app.natal.tools.counsel import natal_compute_factors
    from app.natal.tools.factors import evaluate_factors

    instant, timezone, gender, longitude = sample
    subject = _subject(instant, timezone, gender, longitude)
    natal_chart = _full_pan(instant, timezone, gender, longitude)
    for technique, shushi in (("bazi", "bazi"), ("ziwei", "ziwei")):
        direct = evaluate_factors(gender, natal_chart, shushi=shushi)
        remote = natal_compute_factors(subject, technique)["factors"]
        if technique == "bazi":
            remote.pop("selected_yongshen", None)
        diff, missing = _diff_fields(direct, remote)
        assert not diff and not missing, (
            f"{technique} factors differ: diff={diff[:5]} missing={missing[:5]}"
        )


@pytest.mark.parametrize("sample", SAMPLES)
def test_liunian_snapshot_uses_subject_only(sample):
    from app.natal.tools.charts import _bazi_liunian, _ziwei_liunian
    from app.natal.tools.factors import evaluate_liunian_snap_from_natal_chart

    instant, timezone, gender, longitude = sample
    subject = _subject(instant, timezone, gender, longitude)
    natal_chart = _full_pan(instant, timezone, gender, longitude)
    first = evaluate_liunian_snap_from_natal_chart(
        natal_chart,
        {"bazi": _bazi_liunian(subject, 2030), "ziwei": _ziwei_liunian(subject, 2030)},
        year=2030,
    )
    second = evaluate_liunian_snap_from_natal_chart(
        natal_chart,
        {"bazi": _bazi_liunian(subject, 2030), "ziwei": _ziwei_liunian(subject, 2030)},
        year=2030,
    )
    assert first == second
    assert first["_snapshot_type"] == "liunian"
