"""judgment MCP server（正交化判断层）。

暴露 3 个判断工具（compute_factors / natal_query / period_query），按 domain
分域（bazi/ziwei）：compute_factors 内部调 engine（fullchart/流年/大限）取
判断所需字段，再在进程内求因子快照与断语；无状态、无存储依赖。

Run:
    LIKI_JUDGMENT_DOMAIN=bazi .venv/bin/python -m uvicorn app.judgment_server:app \
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

JUDGMENT_SCHEMA = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "judgment-tools.json"

# natal 工具模块同目录 import（from duanyu / from paipan …），进程内调用需目录在 sys.path。
sys.path.insert(0, str(JUDGMENT_SCHEMA.parent))


def _require_domain(domain: str, chart: dict) -> None:
    """域校验：bazi 域只收八字盘，ziwei 域只收紫微盘。"""
    is_bazi = "ri" in chart
    if domain == "bazi" and not is_bazi:
        raise ValueError(f"judgment/bazi 只接收八字盘（chart 含四柱），收到紫微盘。")
    if domain == "ziwei" and is_bazi:
        raise ValueError(f"judgment/ziwei 只接收紫微盘（chart 为宫位结构），收到八字盘。")


def create_judgment_server(domain: str) -> MCPServer:
    from app.natal.tools import judgment

    if domain not in ("bazi", "ziwei"):
        raise ValueError(f"judgment domain 无效: {domain!r}")
    server = MCPServer(
        name=f"judgment-{domain}",
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
            "_judgment": judgment,
            "Optional": Optional,
            "Literal": Literal,
        }
        if name == "compute_factors":
            call = (
                "_require(_domain, chart)\n"
                "result = _judgment.compute_factors(chart)"
            )
        elif name == "natal_query":
            call = (
                "kwargs = {k: v for k, v in "
                "({'factors': factors, 'topics': topics, 'context': context, "
                "'side': _domain}).items() if v is not None}\n"
                "result = _judgment.natal_query(**kwargs)"
            )
        else:  # period_query
            call = (
                "_require(_domain, chart)\n"
                "result = _judgment.period_query(factors, time_scope, topics, chart, "
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


def _make_app():
    domain = os.environ.get("LIKI_JUDGMENT_DOMAIN", "").strip() or None
    if domain is None:
        raise RuntimeError("LIKI_JUDGMENT_DOMAIN 必须设置（bazi/ziwei）")
    return create_judgment_server(domain).streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
    )


app = _make_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8091)