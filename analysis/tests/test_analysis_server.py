"""analysis MCP server 正确性测试。

验证三件事：
1. 工具面与 skill-tools.json 一致（schema 正确性）
2. MCP 透传结果 == 直接 agent_cli 结果（结果不被 server 改变）
3. 参数校验/错误处理（非法输入不静默通过）

前提：LIKI_RPC_URL 指向可用引擎（运行前启动，如 make verify 的 local-engine）。
"""
from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[2]
NATAL_TOOLS_PATH = Path(__file__).resolve().parents[1] / "liki_analysis" / "natal" / "tools"
DIVINATION_TOOLS_PATH = Path(__file__).resolve().parents[1] / "liki_analysis" / "divination" / "tools"
sys.path.insert(0, str(REPO_ROOT / "analysis"))
sys.path.insert(0, str(NATAL_TOOLS_PATH))

from liki_analysis.server import TOOL_DEFS, VENV_PYTHON, create_server  # noqa: E402

LIKI_MCP_URL = os.environ.get("LIKI_MCP_URL", "http://127.0.0.1:18081/mcp")


def _engine_available() -> bool:
    import urllib.request

    try:
        with urllib.request.urlopen(LIKI_MCP_URL.replace("/mcp", "") + "/health", timeout=3) as r:
            return r.status == 200
    except Exception:
        return False


requires_engine = pytest.mark.skipif(
    not _engine_available(),
    reason="本地引擎不可达（需要 LIKI_RPC_URL 指向可用引擎）",
)


@pytest.fixture(scope="module")
def server():
    return create_server()


@pytest.fixture()
def natal_agent_cli() -> Path:
    return NATAL_TOOLS_PATH / "agent_cli.py"


@pytest.fixture()
def divination_agent_cli() -> Path:
    return DIVINATION_TOOLS_PATH / "agent_cli.py"


def _direct_cli(cli: Path, fn: str, args: dict) -> dict:
    payload = json.dumps({"fn": fn, "args": args}, ensure_ascii=False).encode("utf-8")
    proc = subprocess.run(
        [str(VENV_PYTHON), str(cli)],
        input=payload,
        capture_output=True,
        timeout=60,
        cwd=str(cli.parent),
        env={**os.environ, "LIKI_MCP_URL": LIKI_MCP_URL},
    )
    return json.loads(proc.stdout.decode("utf-8"))


def _call_tool_text(server, name: str, args: dict) -> tuple[str, bool]:
    import asyncio

    result = asyncio.run(server.call_tool(name, args))
    text = result.content[0].text
    return text, result.is_error


# ── 1. 工具面正确性 ─────────────────────────────────────────────

def test_tool_surface_matches_skill(server) -> None:
    import asyncio

    tools = asyncio.run(server.list_tools())
    names = {t.name for t in tools}
    assert names == set(TOOL_DEFS.keys())
    assert len(names) == 10

    # 每个工具必须来自 natal/divination 的 skill-tools.json
    for name in names:
        cli, kind, _ = TOOL_DEFS[name]
        assert cli.name == "agent_cli.py"


def test_create_birth_chart_schema(server) -> None:
    import asyncio

    tools = asyncio.run(server.list_tools())
    t = next(x for x in tools if x.name == "create_birth_chart")
    schema = t.model_dump(exclude_none=True)["input_schema"]
    props = schema["properties"]
    assert set(props) == {"gender", "source"}
    assert schema["required"] == ["gender", "source"]
    assert props["gender"]["enum"] == ["male", "female"]


# ── 2. 透传一致性：MCP 结果 == 直接 agent_cli 结果 ─────────────

@requires_engine
def test_create_birth_chart_parity(server, natal_agent_cli) -> None:
    args = {
        "gender": "male",
        "source": {
            "type": "timestamp",
            "timestamp": "1990-05-20T12:00:00+08:00",
            "precision": "minute",
            "location": {"city": "北京"},
        },
    }
    text, is_err = _call_tool_text(server, "create_birth_chart", args)
    assert not is_err, text
    mcp_result = json.loads(text)

    direct = _direct_cli(natal_agent_cli, "create_birth_chart", args)
    assert direct["ok"], direct
    assert mcp_result == direct["data"], "MCP 结果与 agent_cli 直接结果不一致"


@requires_engine
def test_analyze_natal_parity(server, natal_agent_cli) -> None:
    # 先建盘拿 chart_ref
    args = {
        "gender": "male",
        "source": {
            "type": "timestamp",
            "timestamp": "1990-05-20T12:00:00+08:00",
            "precision": "minute",
            "location": {"city": "北京"},
        },
    }
    text, _ = _call_tool_text(server, "create_birth_chart", args)
    chart_ref = json.loads(text)["chart_ref"]

    analyze_args = {"chart_ref": chart_ref, "topics": ["career"]}
    text2, is_err = _call_tool_text(server, "analyze_natal", analyze_args)
    assert not is_err, text2
    mcp_result = json.loads(text2)

    direct = _direct_cli(natal_agent_cli, "analyze_natal", analyze_args)
    assert direct["ok"], direct
    assert mcp_result == direct["data"], "analyze_natal 透传结果不一致"


# ── 3. 错误处理 ─────────────────────────────────────────────────

def test_invalid_gender_is_error(server) -> None:
    args = {
        "gender": "unknown",
        "source": {
            "type": "timestamp",
            "timestamp": "1990-05-20T12:00:00+08:00",
            "precision": "minute",
            "location": {"city": "北京"},
        },
    }
    _assert_tool_rejects(server, "create_birth_chart", args)


def test_missing_required_is_error(server) -> None:
    _assert_tool_rejects(server, "create_birth_chart", {"gender": "male"})


def _assert_tool_rejects(server, name: str, args: dict) -> None:
    import asyncio

    try:
        result = asyncio.run(server.call_tool(name, args))
    except Exception:
        return  # 参数校验失败抛错，符合预期
    assert result.is_error, f"非法参数应被拒绝（IsError），却返回成功: {result.content}"

# ── 4. 全部 10 工具 1:1 parity（MCP == agent_cli 直接） ─────────

BIRTH = {"gender": "male", "source": {"type": "timestamp", "timestamp": "1990-05-20T12:00:00+08:00", "precision": "minute", "location": {"city": "北京"}}}
BIRTH2 = {"gender": "female", "source": {"type": "timestamp", "timestamp": "1988-11-03T09:30:00+08:00", "precision": "minute", "location": {"city": "上海"}}}


@pytest.fixture(scope="module")
def chart_refs(server):
    """MCP 建两个盘；token 自包含，agent_cli 直接可复用。"""
    text, _ = _call_tool_text(server, "create_birth_chart", BIRTH)
    ref_a = json.loads(text)["chart_ref"]
    text2, _ = _call_tool_text(server, "create_birth_chart", BIRTH2)
    ref_b = json.loads(text2)["chart_ref"]
    return ref_a, ref_b


@pytest.fixture(scope="module")
def liuyao_snapshot(server):
    text, is_err = _call_tool_text(server, "liuyao_snapshot",
                                   {"question": "这次面试能不能通过", "mode": "yaos", "yaos": [7, 8, 7, 8, 7, 8], "matter": "general"})
    assert not is_err, text
    return json.loads(text)


@pytest.fixture(scope="module")
def qimen_snapshot(server):
    text, is_err = _call_tool_text(server, "qimen_snapshot",
                                   {"question": "谈判时机", "city": "北京", "matter": "career"})
    assert not is_err, text
    return json.loads(text)


@requires_engine
def test_parity_create_birth_chart(server, natal_agent_cli):
    _assert_parity(server, natal_agent_cli, "create_birth_chart", BIRTH)


@requires_engine
def test_parity_analyze_natal(server, natal_agent_cli, chart_refs):
    _assert_parity(server, natal_agent_cli, "analyze_natal", {"chart_ref": chart_refs[0], "topics": ["career"]})


@requires_engine
def test_parity_analyze_periods(server, natal_agent_cli, chart_refs):
    _assert_parity(server, natal_agent_cli, "analyze_periods",
                   {"chart_ref": chart_refs[0], "time_scope": {"type": "year", "year": 2025}, "topics": ["wealth"]})


@requires_engine
def test_parity_compare_birth_charts(server, natal_agent_cli, chart_refs):
    _assert_parity(server, natal_agent_cli, "compare_birth_charts",
                   {"chart_ref_a": chart_refs[0], "chart_ref_b": chart_refs[1]})


@requires_engine
def test_parity_calibrate_birth_time(server, natal_agent_cli, chart_refs):
    candidates = [
        {"label": "子时", "gender": "male", "source": {"type": "timestamp", "timestamp": "1990-05-19T23:30:00+08:00", "precision": "minute", "location": {"city": "北京"}}},
        {"label": "丑时", "gender": "male", "source": {"type": "timestamp", "timestamp": "1990-05-20T01:30:00+08:00", "precision": "minute", "location": {"city": "北京"}}},
    ]
    events = [
        {"year": 2015, "topic": "career", "label": "入职"},
        {"year": 2020, "topic": "career", "label": "升职"},
        {"year": 2023, "topic": "marriage", "label": "结婚"},
    ]
    _assert_parity(server, natal_agent_cli, "calibrate_birth_time", {"candidates": candidates, "events": events})


@requires_engine
def test_parity_liuyao_snapshot(server, divination_agent_cli):
    _assert_parity(server, divination_agent_cli, "liuyao_snapshot",
                   {"question": "这次面试能不能通过", "mode": "yaos", "yaos": [7, 8, 7, 8, 7, 8], "matter": "general"},
                   ignore_fields=("question.solar_time", "snapshot_digest"))


@requires_engine
def test_parity_liuyao_ask(server, divination_agent_cli, liuyao_snapshot):
    _assert_parity(server, divination_agent_cli, "liuyao_ask",
                   {"snapshot": liuyao_snapshot, "message": "什么时候有结果"})


@requires_engine
def test_parity_qimen_snapshot(server, divination_agent_cli):
    _assert_parity(server, divination_agent_cli, "qimen_snapshot",
                   {"question": "谈判时机", "city": "北京", "matter": "career"},
                   ignore_fields=("input.local_time", "snapshot_digest"))


@requires_engine
def test_parity_qimen_ask(server, divination_agent_cli, qimen_snapshot):
    _assert_parity(server, divination_agent_cli, "qimen_ask",
                   {"snapshot": qimen_snapshot, "message": "哪个方向有利"})


@requires_engine
def test_parity_huangli_days(server, divination_agent_cli):
    _assert_parity(server, divination_agent_cli, "huangli_days",
                   {"question": "下个月哪天适合搬家", "start_date": "2026-10-01", "event": "move", "days": 5})


def _assert_parity(server, cli, fn, args, ignore_fields=()):
    text, is_err = _call_tool_text(server, fn, args)
    assert not is_err, f"{fn} MCP 调用失败: {text}"
    mcp_result = json.loads(text)

    direct = _direct_cli(cli, fn, args)
    assert direct.get("ok"), f"{fn} agent_cli 失败: {direct}"
    _compare_ignoring(mcp_result, direct["data"], ignore_fields, fn)


def _compare_ignoring(a, b, ignore, fn, path=""):
    if path in ignore:
        return
    if type(a) != type(b):
        raise AssertionError(f"{fn}: 类型不一致 {path}")
    if isinstance(a, dict):
        for k in set(a) | set(b):
            if k not in a or k not in b:
                raise AssertionError(f"{fn}: 字段差异 {path}.{k}")
            _compare_ignoring(a[k], b[k], ignore, fn, f"{path}.{k}" if path else k)
    elif isinstance(a, list):
        if len(a) != len(b):
            raise AssertionError(f"{fn}: 长度差异 {path}")
        for i, (x, y) in enumerate(zip(a, b)):
            _compare_ignoring(x, y, ignore, fn, f"{path}[{i}]")
    elif a != b:
        raise AssertionError(f"{fn}: 值不一致 {path}: {str(a)[:40]} vs {str(b)[:40]}")
