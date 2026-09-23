"""liki-analysis MCP server.

Reuses the existing skill Python tool layer (skills/liki/{natal,divination}/tools
agent_cli.py) via subprocess, exposing the same 10 tools over a stateless
Streamable HTTP MCP endpoint. The engine (bazi/ziwei/liuyao/qimen/huangli
computation) is called by the tool layer internally; WorkBuddy clients only see
this server.

Run:
    .venv/bin/python -m uvicorn app.server:app --port 8090
"""
from __future__ import annotations

import json
import pathlib
import subprocess
from typing import Literal, Optional

from mcp.server import MCPServer
from mcp.server.streamable_http_manager import StreamableHTTPSessionManager
from mcp.server.streamable_http_manager import StreamableHTTPASGIApp
from mcp.types import Tool
from pydantic import BaseModel, Field, create_model

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]
NATAL_CLI = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "agent_cli.py"
DIVINATION_CLI = pathlib.Path(__file__).resolve().parent / "divination" / "tools" / "agent_cli.py"
NAMING_CLI = pathlib.Path(__file__).resolve().parent / "naming" / "agent_cli.py"
VENV_PYTHON = pathlib.Path(__file__).resolve().parents[1] / ".venv" / "bin" / "python"

ANALYSIS_DIR = pathlib.Path(__file__).resolve().parent
NATAL_TOOLS = ANALYSIS_DIR / "natal" / "tools"
DIVINATION_TOOLS = ANALYSIS_DIR / "divination" / "tools"
NAMING_TOOLS = ANALYSIS_DIR / "naming" / "tools"

# 每个工具的 {fn 名: (CLI 路径, 参数 schema 文件)}
TOOL_DEFS: dict[str, tuple[pathlib.Path, str, str]] = {}

_natal_schema = json.loads(
    (NATAL_TOOLS / "skill-tools.json").read_text("utf-8")
)
_div_schema = json.loads(
    (DIVINATION_TOOLS / "skill-tools.json").read_text("utf-8")
)
_naming_schema = json.loads(
    (NAMING_TOOLS / "skill-tools.json").read_text("utf-8")
)
for _schema, _cli, _kind in (
    (_natal_schema, NATAL_CLI, "natal"),
    (_div_schema, DIVINATION_CLI, "divination"),
    (_naming_schema, NAMING_CLI, "naming"),
):
    for _t in _schema["tools"]:
        _fn = _t["function"]
        TOOL_DEFS[_fn["name"]] = (_cli, _kind, _fn["name"])


def _build_model(name: str, schema: dict) -> type[BaseModel]:
    """从 skill-tools.json 的参数 schema 生成 pydantic 参数模型（顶层字段）。"""
    fields: dict[str, tuple] = {}
    required = set(schema.get("required", []))
    for pname, ps in schema.get("properties", {}).items():
        ptype = ps.get("type")
        desc = ps.get("description", "")
        if "enum" in ps and ptype == "string":
            field_type: type = Literal[tuple(ps["enum"])]  # type: ignore[valid-type]
        elif ptype == "string":
            field_type = str
        elif ptype == "integer":
            field_type = int
        elif ptype == "number":
            field_type = float
        elif ptype == "boolean":
            field_type = bool
        elif ptype == "array":
            field_type = list
        else:
            field_type = dict
        if pname in required:
            fields[pname] = (field_type, Field(description=desc))
        else:
            fields[pname] = (Optional[field_type], Field(default=None, description=desc))
    return create_model(name, __base__=BaseModel, **fields)


def _signature_params(schema: dict) -> tuple[str, str]:
    """从参数 schema 生成 handler 签名与 args 组装代码。

    returns (signature, body) 用于 exec 动态构建工具 handler。
    """
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


def _run_cli(cli: pathlib.Path, fn: str, args: dict) -> dict:
    """以 subprocess 调 agent_cli.py（stateless：每次独立进程，天然无状态）。"""
    payload = json.dumps({"fn": fn, "args": args}, ensure_ascii=False).encode("utf-8")
    proc = subprocess.run(
        [str(VENV_PYTHON), str(cli)],
        input=payload,
        capture_output=True,
        timeout=60,
        cwd=str(cli.parent),
    )
    if proc.returncode != 0:
        raise RuntimeError(f"tool {fn} crashed: {proc.stderr.decode('utf-8', 'replace')[:500]}")
    try:
        return json.loads(proc.stdout.decode("utf-8"))
    except json.JSONDecodeError as e:
        raise RuntimeError(f"tool {fn} bad output: {e}") from e


def create_server(domain: str | None = None) -> MCPServer:
    """创建 MCP server；domain 指定时只注册该专家工具（natal 工具注入 domain）。

    domain=None → 全量（双术数 + 占卜 + 起名）。
    domain="bazi"/"ziwei" → 5 个 natal 工具，analyze_* 自动注入 domain。
    domain="qimen"/"liuyao" → 对应 2 个占卜工具。
    domain="naming" → 5 个起名工具。
    """
    natal_domains = ("bazi", "ziwei")
    if domain is not None and domain not in (*natal_domains, "qimen", "liuyao", "naming"):
        raise ValueError(f"domain 无效: {domain!r}")
    server = MCPServer(
        name=f"engine-pro-{domain}" if domain else "engine-pro",
        title=f"Liki {domain} 专家" if domain else "Liki Engine Pro",
        description=(
            f"命理判断层（{domain} 专家）：因子/断语/应期/考时，复用 Liki 规则引擎。"
            if domain else "命理判断层：因子/断语/应期/考时，复用 Liki 规则引擎。"
        ),
        version="0.1.0",
    )
    for name, (cli, kind, fn) in TOOL_DEFS.items():
        if domain == "qimen" and fn not in ("qimen_snapshot", "qimen_ask"):
            continue
        if domain == "liuyao" and fn not in ("liuyao_snapshot", "liuyao_ask"):
            continue
        if domain in natal_domains and kind != "natal":
            continue
        if domain == "naming" and kind != "naming":
            continue
        schema_file = {
            "natal": NATAL_TOOLS / "skill-tools.json",
            "divination": DIVINATION_TOOLS / "skill-tools.json",
            "naming": NAMING_TOOLS / "skill-tools.json",
        }[kind]
        params_schema = json.loads(schema_file.read_text("utf-8"))
        tool_schema = next(
            t["function"]["parameters"]
            for t in params_schema["tools"]
            if t["function"]["name"] == fn
        )
        desc = next(
            t["function"]["description"]
            for t in params_schema["tools"]
            if t["function"]["name"] == fn
        )
        sig, body = _signature_params(tool_schema)
        ns: dict = {
            "_run_cli": _run_cli,
            "_cli": cli,
            "_fn": fn,
            "_domain": domain if domain in natal_domains else None,
            "json": json,
            "Optional": Optional,
            "Literal": Literal,
        }
        code = (
            f"async def _handler({sig}) -> str:\n"
            f"    args = {{{body}}}\n"
            "    if _domain is not None:\n"
            "        args['domain'] = _domain\n"
            "    filtered = {k: v for k, v in args.items() if v is not None}\n"
            "    result = _run_cli(_cli, _fn, filtered)\n"
            "    if not result.get('ok'):\n"
            "        err = result.get('error', {})\n"
            "        raise ValueError(json.dumps(err, ensure_ascii=False))\n"
            "    return json.dumps(result.get('data'), ensure_ascii=False)\n"
        )
        exec(code, ns)  # noqa: S102
        server.add_tool(ns["_handler"], name=name, description=desc)
    return server


def _make_app():
    """单 server 端点：LIKI_ANALYSIS_DOMAIN 未设 → 全量 10 工具；设置 → 该专家子集。

    同一进程内挂多个 session_manager 会触发 mcp 库的 anyio 嵌套上限（实测
    >4 层卡死），故每个专家用独立进程（LIKI_ANALYSIS_DOMAIN）部署，
    对外由网关按路径路由 /mcp/<专家>。
    """
    import os

    domain = os.environ.get("LIKI_ANALYSIS_DOMAIN", "").strip() or None
    return create_server(domain).streamable_http_app(
        streamable_http_path="/mcp",
        json_response=True,
        stateless_http=True,
    )

    return _Router(children)


app = _make_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8090)