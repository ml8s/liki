"""counsel MCP server（正交化判断层）。

暴露判断工具，按 domain 分域（bazi/ziwei/liuyao/qimen/naming）：
- bazi/ziwei：compute_factors / natal_query / period_query（因子快照 + 断语）
- naming：起名（surname/pick/char/compose/check，复用 qiming 逻辑 + engine 字库）
内部调 engine（排盘/字库）取判断所需字段，再在进程内求因子/断语/评估；
无状态、无存储依赖。counsel 是 engine 的补集（判断），排盘/历法在 engine。

Run:
    LIKI_COUNSEL_SERVICE_DOMAIN=bazi .venv/bin/python -m uvicorn app.counsel_server:app \
        --port 8091
"""
from __future__ import annotations

import json
import os
import pathlib
import sys

from typing import Literal, Optional

from mcp.server import MCPServer
from pydantic import BaseModel

from app.server import _build_model, _signature_params

JUDGMENT_SCHEMA = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "counsel-tools.json"
NAMING_SCHEMA = pathlib.Path(__file__).resolve().parent / "naming" / "tools" / "skill-tools.json"
NAMING_DIR = pathlib.Path(__file__).resolve().parent / "naming"

# natal 工具模块同目录 import（from duanyu / from paipan …），进程内调用需目录在 sys.path。
sys.path.insert(0, str(JUDGMENT_SCHEMA.parent))
sys.path.insert(0, str(NAMING_DIR))


def _require_domain(domain: str, chart: dict) -> None:
    """域校验：bazi 域只收八字盘，ziwei 域只收紫微盘。"""
    is_bazi = "ri" in chart
    if domain == "bazi" and not is_bazi:
        raise ValueError(f"counsel/bazi 只接收八字盘（chart 含四柱），收到紫微盘。")
    if domain == "ziwei" and is_bazi:
        raise ValueError(f"counsel/ziwei 只接收紫微盘（chart 为宫位结构），收到八字盘。")


def create_counsel_server(domain: str) -> MCPServer:
    if domain == "naming":
        return _create_naming_server()
    from app.natal.tools import counsel

    if domain not in ("bazi", "ziwei"):
        raise ValueError(f"counsel domain 无效: {domain!r}")
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 判断层（{domain}）",
        description=(
            f"命理判断层（{domain}）：compute_factors 因子快照 + natal_query 本命断语 + "
            f"period_query 应期断语。排盘由 engine 完成（chart 为输入）。"
        ),
        version="1.0.0",
    )
    schema = json.loads(JUDGMENT_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        params_schema = fn["parameters"]
        sig, body = _signature_params(params_schema)
        ns: dict = {
            "json": json,
            "_domain": domain,
            "_require": _require_domain,
            "_counsel": counsel,
            "Optional": Optional,
            "Literal": Literal,
        }
        if name == "compute_factors":
            call = (
                "_require(_domain, chart)\n"
                "result = _counsel.compute_factors(chart)"
            )
        elif name == "natal_query":
            call = (
                "kwargs = {k: v for k, v in "
                "({'factors': factors, 'topics': topics, 'context': context, "
                "'side': _domain}).items() if v is not None}\n"
                "result = _counsel.natal_query(**kwargs)"
            )
        else:  # period_query
            call = (
                "_require(_domain, chart)\n"
                "result = _counsel.period_query(factors, time_scope, topics, chart, "
                "side=_domain)"
            )
        code = (
            f"async def _handler({sig}) -> str:\n"
            f"    {call.replace(chr(10), chr(10) + '    ')}\n"
            "    return json.dumps(result, ensure_ascii=False)\n"
        )
        exec(code, ns)  # noqa: S102
        server.add_tool(ns["_handler"], name=name, description=fn["description"])
    return server


def _qiming_check(given_names, yongshen, xishen=None, jishen=None):
    """适配 schema 参数名（yongshen/xishen/jishen）→ evaluate_names（yong_shen 等）。"""
    from qiming import evaluate_names

    return evaluate_names(
        given_names, yongshen, xishen or [], jishen or []
    )


def _create_naming_server() -> MCPServer:
    """起名（naming）：qiming 工具复用 qiming 逻辑（字库来自 engine）。"""
    from qiming import compose_names, lookup_char, match_surnames, pick_chars

    server = MCPServer(
        name="counsel-naming",
        title="Liki 起名顾问",
        description=(
            "起名顾问：姓氏匹配、按五行取字、单字查询、组名、候选名评估。"
            "字库来自 engine（继承），用神/喜忌由调用方传入（八字判断结果）。"
        ),
        version="1.0.0",
    )
    qiming_fns = {
        "qiming_surname": match_surnames,
        "qiming_pick": pick_chars,
        "qiming_char": lookup_char,
        "qiming_compose": compose_names,
        "qiming_check": _qiming_check,
    }
    schema = json.loads(NAMING_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        sig, body = _signature_params(fn["parameters"])
        ns: dict = {"json": json, "_fn": qiming_fns[name], "Optional": Optional, "Literal": Literal}
        compose_default = (
            "    if 'second' not in kwargs or kwargs['second'] is None:\n"
            "        kwargs['second'] = []\n"
        )
        code = (
            f"async def _handler({sig}) -> str:\n"
            f"    kwargs = {{{body}}}\n"
            f"{compose_default if name == 'qiming_compose' else ''}"
            "    result = _fn(**{k: v for k, v in kwargs.items() if v is not None})\n"
            "    return json.dumps(result, ensure_ascii=False)\n"
        )
        exec(code, ns)  # noqa: S102
        server.add_tool(ns["_handler"], name=name, description=fn["description"])
    return server


def _make_app():
    domain = os.environ.get("LIKI_COUNSEL_SERVICE_DOMAIN", "").strip() or None
    if domain is None:
        raise RuntimeError("LIKI_COUNSEL_SERVICE_DOMAIN 必须设置（bazi/ziwei/naming 等）")
    return create_counsel_server(domain).streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
    )


app = _make_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8091)