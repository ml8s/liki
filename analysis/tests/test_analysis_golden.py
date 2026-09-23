"""重构回归测试：与 golden 基线对照（逻辑一致）+ 专家 domain 过滤。"""
from __future__ import annotations

import asyncio
import json
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from app.server import create_server  # noqa: E402

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "analysis" / "baseline.json"

BIRTH = {
    "gender": "male",
    "source": {
        "type": "timestamp",
        "timestamp": "1990-05-20T12:00:00+08:00",
        "precision": "minute",
        "location": {"city": "北京"},
    },
}
BIRTH2 = {
    "gender": "female",
    "source": {
        "type": "timestamp",
        "timestamp": "1988-11-03T09:30:00+08:00",
        "precision": "minute",
        "location": {"city": "上海"},
    },
}

# 归一化时忽略的动态/时间字段（digest/当前时刻等每次运行不同）
IGNORE_KEYS = {"solar_time", "local_time", "digest", "snapshot_digest", "chart_ref", "pan_digest"}


def norm(d):
    if isinstance(d, dict):
        return {k: norm(v) for k, v in d.items() if k not in IGNORE_KEYS}
    if isinstance(d, list):
        return [norm(v) for v in d]
    return d


def call(srv, name, args):
    r = asyncio.run(srv.call_tool(name, args))
    if r.is_error:
        raise RuntimeError(f"{name}: {r.content[0].text}")
    return json.loads(r.content[0].text)


@pytest.fixture(scope="module")
def charts():
    full = create_server()
    a = call(full, "create_birth_chart", BIRTH)
    b = call(full, "create_birth_chart", BIRTH2)
    return full, a, b


def keys_only(d):
    """只保留键结构（值置空），用于时间敏感工具的 schema 对照。"""
    if isinstance(d, dict):
        return {k: keys_only(v) for k, v in d.items()}
    if isinstance(d, list):
        return [keys_only(v) for v in d] if d else []
    return type(d).__name__


def top_keys(d):
    """时间敏感工具只对照顶层字段集合（局内容随时刻变化）。"""
    return sorted(d) if isinstance(d, dict) else sorted(range(len(d)))


def test_golden_baseline_consistent(charts):
    """全量输出与 golden 基线一致（时间字段归一化后）。

    由当前时刻定局的工具（liuyao/qimen snapshot 及依赖它的 ask）只对照顶层
    字段集合；确定性工具（排盘/本命/流年/合盘/黄历）做逐字段精确对照。
    """
    if not GOLDEN.exists():
        pytest.skip("golden 基线不存在")
    full, chart_a, chart_b = charts
    ref_a, ref_b = chart_a["chart_ref"], chart_b["chart_ref"]
    liuyao = call(full, "liuyao_snapshot", {
        "question": "这次面试能不能通过", "mode": "yaos",
        "yaos": [7, 8, 7, 8, 7, 8], "matter": "general",
    })
    qimen = call(full, "qimen_snapshot", {
        "question": "谈判时机", "city": "北京", "matter": "career",
    })
    cur = {
        "create_birth_chart": chart_a,
        "analyze_natal": call(full, "analyze_natal", {"chart_ref": ref_a, "topics": ["career"]}),
        "analyze_periods": call(full, "analyze_periods", {
            "chart_ref": ref_a, "time_scope": {"type": "year", "year": 2025}, "topics": ["wealth"],
        }),
        "compare_birth_charts": call(full, "compare_birth_charts", {
            "chart_ref_a": ref_a, "chart_ref_b": ref_b,
        }),
        "liuyao_snapshot": liuyao,
        "liuyao_ask": call(full, "liuyao_ask", {"snapshot": liuyao, "message": "什么时候有结果"}),
        "qimen_snapshot": qimen,
        "qimen_ask": call(full, "qimen_ask", {"snapshot": qimen, "message": "哪个方向有利"}),
        "huangli_days": call(full, "huangli_days", {
            "question": "下个月哪天适合搬家", "start_date": "2026-10-01", "event": "move", "days": 5,
        }),
    }
    base = json.loads(GOLDEN.read_text("utf-8"))
    assert set(cur) == set(base)
    for key in base:
        if key.startswith(("liuyao_", "qimen_")):
            assert top_keys(cur[key]) == top_keys(base[key]), f"{key} 顶层字段不一致"
        else:
            assert norm(cur[key]) == norm(base[key]), f"{key} 与 golden 不一致"


def test_domain_filter_pure_side(charts):
    """bazi/ziwei 专家 analyze_natal 只出自己 side，且与全量对应子集一致。"""
    full, chart_a, _ = charts
    ref = chart_a["chart_ref"]
    base = call(full, "analyze_natal", {"chart_ref": ref, "topics": ["career"]})
    base_by_side: dict[str, list] = {}
    for a in base["assertions"]:
        base_by_side.setdefault(a["side"], []).append(a)
    for domain in ("bazi", "ziwei"):
        srv = create_server(domain)
        r = call(srv, "analyze_natal", {"chart_ref": ref, "topics": ["career"]})
        sides: dict[str, list] = {}
        for a in r["assertions"]:
            sides.setdefault(a["side"], []).append(a)
        assert set(sides) == {domain}, f"{domain} 专家混入了 {set(sides) - {domain}}"
        got = sorted(a["assertion_id"] for a in sides[domain])
        want = sorted(a["assertion_id"] for a in base_by_side.get(domain, []))
        assert got == want, f"{domain} 专家断语与全量子集不一致"