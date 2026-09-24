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

import collections
import json
import os
import pathlib
import sys
import time

from typing import Literal, Optional

from starlette.applications import Starlette
from starlette.routing import Mount

from mcp.server import MCPServer
from pydantic import BaseModel
from starlette.applications import Starlette
from starlette.responses import JSONResponse
from starlette.routing import Route

JUDGMENT_SCHEMA = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "counsel-tools.json"
NAMING_SCHEMA = pathlib.Path(__file__).resolve().parent / "naming" / "tools" / "skill-tools.json"
NAMING_DIR = pathlib.Path(__file__).resolve().parent / "naming"
DIVINATION_SCHEMA = pathlib.Path(__file__).resolve().parent / "divination" / "tools" / "skill-tools.json"
DIVINATION_DIR = pathlib.Path(__file__).resolve().parent / "divination" / "tools"

# natal 工具模块同目录 import（from duanyu / from paipan …），进程内调用需目录在 sys.path。
sys.path.insert(0, str(JUDGMENT_SCHEMA.parent))
sys.path.insert(0, str(NAMING_DIR))
sys.path.insert(0, str(DIVINATION_DIR))


def _signature_params(schema: dict) -> tuple[str, str]:
    """从参数 schema 生成 handler 签名与 args 组装代码（exec 动态构建工具 handler）。"""
    required = set(schema.get("required", []))
    params: list[str] = []
    body_items: list[str] = []
    for pname, ps in schema.get("properties", {}).items():
        ptype = ps.get("type")
        if "enum" in ps and ptype == "string":
            enum_items = ", ".join(repr(e) for e in ps["enum"])
            type_expr = f"Literal[{enum_items}]"
        elif ptype == "string":
            type_expr = "str"
        elif ptype == "integer":
            type_expr = "int"
        elif ptype == "number":
            type_expr = "float"
        elif ptype == "boolean":
            type_expr = "bool"
        elif ptype == "array":
            type_expr = "list"
        else:
            type_expr = "dict"
        if pname in required:
            params.append(f"{pname}: {type_expr}")
        else:
            params.append(f"{pname}: Optional[{type_expr}] = None")
        body_items.append(f"'{pname}': {pname}")
    sig = ", ".join(params)
    body = ", ".join(body_items)
    return sig, body


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
    if domain in ("liuyao", "qimen"):
        return _create_divination_server(domain)
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


def create_counsel_root_server() -> MCPServer:
    """Create the aggregate counsel surface used by the root Liki skill.

    Bazi and Ziwei are intentionally excluded: those are exposed through the
    expert skills and domain-specific endpoints.
    """
    server = MCPServer(
        name="counsel",
        title="Liki Counsel",
        description="命理判断层：起名、六爻与奇门能力聚合。",
        version="1.0.0",
    )
    _add_naming_tools(server)
    _add_liuyao_tools(server)
    _add_qimen_tools(server)
    return server


def _add_divination_tools(server: MCPServer, domain: str) -> None:
    """六爻/奇门（liuyao/qimen）：snapshot 创建（排盘+因子）+ query 追问（ask→query）。

    复用 analysis 模块（create/ask——起卦/排盘/断语），engine 完成排盘；
    counsel 是补集（判断），ask 做成 query（追问出断语）。
    """
    if domain == "liuyao":
        from liuyao_snapshot import create as snapshot_create
        from liuyao_ask import ask as query_ask
    else:
        from qimen_snapshot import create as snapshot_create
        from qimen_ask import ask as query_ask
    schema = json.loads(DIVINATION_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        if name == f"{domain}_snapshot":
            impl = snapshot_create
            out_name = name
        elif name == f"{domain}_ask":
            impl = query_ask
            out_name = f"{domain}_query"  # ask → query（追问出断语）
        else:
            continue
        sig, body = _signature_params(fn["parameters"])
        ns: dict = {"json": json, "_fn": impl, "Optional": Optional, "Literal": Literal}
        code = (
            f"async def _handler({sig}) -> str:\n"
            f"    kwargs = {{{body}}}\n"
            "    result = _fn(**{k: v for k, v in kwargs.items() if v is not None})\n"
            "    return json.dumps(result, ensure_ascii=False)\n"
        )
        exec(code, ns)  # noqa: S102
        server.add_tool(ns["_handler"], name=out_name, description=fn["description"])


def _create_divination_server(domain: str) -> MCPServer:
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 顾问（{domain}）",
        description=(
            f"{'六爻' if domain == 'liuyao' else '奇门'}顾问：snapshot 创建（起卦/排盘+因子），"
            f"query 追问（基于 snapshot 出断语，不重排）。排盘由 engine 完成。"
        ),
        version="1.0.0",
    )
    _add_divination_tools(server, domain)
    return server


def _qiming_check(given_names, yongshen, xishen=None, jishen=None):
    """适配 schema 参数名（yongshen/xishen/jishen）→ evaluate_names（yong_shen 等）。"""
    from qiming import evaluate_names

    return evaluate_names(
        given_names, yongshen, xishen or [], jishen or []
    )


def _add_naming_tools(server: MCPServer) -> None:
    """起名（naming）：qiming 工具复用 qiming 逻辑（字库来自 engine）。"""
    from qiming import compose_names, lookup_char, match_surnames, pick_chars

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


def _add_liuyao_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "liuyao")


def _add_qimen_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "qimen")


def _create_naming_server() -> MCPServer:
    server = MCPServer(
        name="counsel-naming",
        title="Liki 起名顾问",
        description=(
            "起名顾问：姓氏匹配、按五行取字、单字查询、组名、候选名评估。"
            "字库来自 engine（继承），用神/喜忌由调用方传入（八字判断结果）。"
        ),
        version="1.0.0",
    )
    _add_naming_tools(server)
    return server


def _make_app():
    """统一 app：挂载全部域的 MCP server（/counsel/mcp/{domain}）。

    网关只路由 /counsel 到本服务（一个容器），服务内按域分发——
    与 engine 对称（/engine/mcp/{domain}）。MCP 服务根路由为 /mcp。
    """
    domains = ("bazi", "ziwei", "liuyao", "qimen", "naming")
    domain_apps = {
        domain: create_counsel_server(domain).streamable_http_app(
            streamable_http_path="/mcp",
            json_response=True,
            stateless_http=True,
        )
        for domain in domains
    }

    async def healthz(scope, receive, send):
        response = JSONResponse(
            {"status": "ok", "domains": list(domains)},
        )
        await response(scope, receive, send)

    class RateLimiter:
        """Small non-durable limiter; enough for one-process MCP service."""

        def __init__(self, limit: int, window_seconds: float):
            self.limit = limit
            self.window_seconds = window_seconds
            self.hits: dict[str, collections.deque[float]] = {}

        def allow(self, key: str) -> bool:
            now = time.monotonic()
            hits = self.hits.setdefault(key, collections.deque())
            while hits and hits[0] <= now - self.window_seconds:
                hits.popleft()
            if len(hits) >= self.limit:
                return False
            hits.append(now)
            return True

    def client_key(scope: dict) -> str:
        for name, value in scope.get("headers") or []:
            if name == b"x-forwarded-for":
                # With a single trusted proxy, the last address is the one
                # observed by that proxy; earlier entries can be spoofed.
                return value.decode("latin-1").split(",", 1)[-1].strip()
        client = scope.get("client") or ("unknown", 0)
        return str(client[0])

    class MultiDomainCounselApp:
        def __init__(self):
            self.rate_limiter = RateLimiter(
                limit=int(os.environ.get("LIKI_COUNSEL_RATE_LIMIT", "240")),
                window_seconds=float(os.environ.get("LIKI_COUNSEL_RATE_WINDOW_SECONDS", "60")),
            )

        async def __call__(self, scope, receive, send):
            if scope["type"] == "http" and scope.get("path") == "/healthz":
                await healthz(scope, receive, send)
                return

            if not self.rate_limiter.allow(client_key(scope)):
                response = JSONResponse(
                    {"error": {"code": "rate_limited", "message": "too many requests"}},
                    status_code=429,
                )
                await response(scope, receive, send)
                return

            path = scope.get("path", "").rstrip("/") or "/"
            for domain in domains:
                if path == f"/counsel/mcp/{domain}":
                    target = domain_apps[domain]
                    proxied_scope = dict(scope)
                    proxied_scope["path"] = "/mcp"
                    proxied_scope["raw_path"] = "/mcp".encode()
                    await target(proxied_scope, receive, send)
                    return

            response = JSONResponse(
                {"error": {"code": "not_found", "message": "unknown counsel endpoint"}},
                status_code=404,
            )
            await response(scope, receive, send)

    return MultiDomainCounselApp()


app = _make_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8091)
