"""Counsel domain MCP server construction and tool registration."""
from __future__ import annotations

import json
import pathlib
import sys

from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError

JUDGMENT_SCHEMA = pathlib.Path(__file__).resolve().parent / "natal" / "tools" / "counsel-tools.json"
NAMING_SCHEMA = pathlib.Path(__file__).resolve().parent / "naming" / "tools" / "skill-tools.json"
NAMING_DIR = pathlib.Path(__file__).resolve().parent / "naming"
DIVINATION_SCHEMA = pathlib.Path(__file__).resolve().parent / "divination" / "tools" / "skill-tools.json"
DIVINATION_DIR = pathlib.Path(__file__).resolve().parent / "divination" / "tools"
VERSION_PATH = pathlib.Path(__file__).resolve().parent / "VERSION.txt"
RUNTIME_VERSION = VERSION_PATH.read_text("utf-8").strip()

# natal 工具模块同目录 import（from duanyu / from paipan …），进程内调用需目录在 sys.path。
sys.path.insert(0, str(JUDGMENT_SCHEMA.parent))
sys.path.insert(0, str(NAMING_DIR))
sys.path.insert(0, str(DIVINATION_DIR))

from app.mcp_tool_adapter import add_schema_tool as _add_schema_tool  # noqa: E402
from app.natal.tools import counsel  # noqa: E402


def _require_domain(domain: str, chart: dict) -> None:
    """域校验：bazi 域只收八字盘，ziwei 域只收紫微盘。"""
    is_bazi = "ri" in chart
    if domain == "bazi" and not is_bazi:
        raise ToolError("invalid_arguments: counsel/bazi 只接收八字盘（chart 含四柱），收到紫微盘。")
    if domain == "ziwei" and is_bazi:
        raise ToolError("invalid_arguments: counsel/ziwei 只接收紫微盘（chart 为宫位结构），收到八字盘。")


def _chart_domain(chart: dict) -> str:
    return "bazi" if "ri" in chart else "ziwei"


def _factor_side(arguments: dict) -> str:
    factors = arguments.get("factors")
    if not isinstance(factors, dict):
        raise ToolError("invalid_arguments: factors 必须是 compute_factors 返回的对象")
    provenance = factors.get("_provenance")
    if not isinstance(provenance, dict):
        raise ToolError("invalid_arguments: factors 缺少可验证的 side provenance")
    side = provenance.get("side")
    if side not in ("bazi", "ziwei"):
        raise ToolError("invalid_arguments: factors 缺少可验证的 side provenance")
    return side


def _add_natal_tools(server: MCPServer, domain: str | None = None) -> None:
    """注册本命判断工具。

    domain endpoint 固定 bazi/ziwei；root aggregate 传入 None 时从 chart 和
    factors provenance 推断领域，避免在同一 MCP server 上重复注册同名工具。
    """
    schema = json.loads(JUDGMENT_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]

        def invoke_natal(arguments: dict, _name=name):
            if _name == "compute_factors":
                side = domain or _chart_domain(arguments["chart"])
                _require_domain(side, arguments["chart"])
                return counsel.compute_factors(arguments["chart"])
            if _name == "natal_query":
                side = domain or _factor_side(arguments)
                return counsel.natal_query(
                    arguments["factors"],
                    arguments["topics"],
                    arguments["factors_digest"],
                    context=arguments.get("context"),
                    side=side,
                )
            side = domain or _factor_side(arguments)
            _require_domain(side, arguments["chart"])
            return counsel.period_query(
                arguments["factors"],
                arguments["factors_digest"],
                arguments["time_scope"],
                arguments["topics"],
                arguments["chart"],
                side=side,
            )

        _add_schema_tool(
            server,
            name=name,
            description=fn["description"],
            schema=fn["parameters"],
            output_schema=fn.get("output_schema"),
            invoke=invoke_natal,
        )


def create_counsel_mcp(domain: str) -> MCPServer:
    if domain == "naming":
        return _create_naming_server()
    if domain in ("liuyao", "qimen"):
        return _create_divination_server(domain)

    if domain not in ("bazi", "ziwei"):
        raise ValueError(f"counsel domain 无效: {domain!r}")
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 判断层（{domain}）",
        description=(
            f"命理判断层（{domain}）：compute_factors 因子快照 + natal_query 本命断语 + "
            f"period_query 应期断语。排盘由 engine 完成（chart 为输入）。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_natal_tools(server, domain)
    return server


def create_counsel_root_mcp() -> MCPServer:
    """Create the aggregate counsel surface used by the root Liki skill.

    The root exposes one namespaced-free union used by the Liki Chat expert
    deployment. Natal tools infer bazi/ziwei from chart shape and verified
    factor provenance, so both domains can share the standard MCP tool names.
    """
    server = MCPServer(
        name="counsel",
        title="Liki Counsel",
        description="命理判断层：八字/紫微本命、起名、六爻与奇门能力聚合。",
        version=RUNTIME_VERSION,
    )
    _add_natal_tools(server)
    _add_naming_tools(server)
    _add_liuyao_tools(server)
    _add_qimen_tools(server)
    return server


def _add_divination_tools(server: MCPServer, domain: str) -> None:
    """六爻/奇门（liuyao/qimen）：snapshot 创建（排盘+因子）+ query 追问。

    复用 counsel 模块（create/ask——起卦/排盘/断语），engine 完成排盘；
    counsel 是补集（判断），MCP 暴露名与 skill-tools.json 一致（{domain}_snapshot / {domain}_query），
    query 基于 snapshot 追问出断语。
    """
    if domain == "liuyao":
        from liuyao_ask import ask as query_ask
        from liuyao_snapshot import create as snapshot_create
    else:
        from qimen_ask import ask as query_ask
        from qimen_snapshot import create as snapshot_create
    schema = json.loads(DIVINATION_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        if name == f"{domain}_snapshot":
            impl = snapshot_create
        elif name == f"{domain}_query":
            impl = query_ask
        else:
            continue
        _add_schema_tool(
            server,
            name=name,
            description=fn["description"],
            schema=fn["parameters"],
            output_schema=fn.get("output_schema"),
            invoke=lambda arguments, _impl=impl: _impl(**arguments),
        )


def _create_divination_server(domain: str) -> MCPServer:
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 顾问（{domain}）",
        description=(
            f"{'六爻' if domain == 'liuyao' else '奇门'}顾问：snapshot 创建（起卦/排盘+因子），"
            f"query 追问（基于 snapshot 出断语，不重排）。排盘由 engine 完成。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_divination_tools(server, domain)
    return server


def _naming_check(given_names, yongshen, xishen=None, jishen=None):
    """适配 schema 参数名（yongshen/xishen/jishen）→ evaluate_names（yong_shen 等）。"""
    from naming import evaluate_names

    return evaluate_names(
        given_names, yongshen, xishen or [], jishen or []
    )


def _add_naming_tools(server: MCPServer) -> None:
    """起名（naming）：qiming 工具复用 qiming 逻辑（字库来自 engine）。"""
    from naming import compose_names, lookup_char, match_surnames, pick_chars

    naming_fns = {
        "naming_surname": match_surnames,
        "naming_pick": pick_chars,
        "naming_char": lookup_char,
        "naming_compose": compose_names,
        "naming_check": _naming_check,
    }
    schema = json.loads(NAMING_SCHEMA.read_text("utf-8"))
    for tool in schema["tools"]:
        fn = tool["function"]
        name = fn["name"]
        _add_schema_tool(
            server,
            name=name,
            description=fn["description"],
            schema=fn["parameters"],
            output_schema=fn.get("output_schema"),
            invoke=lambda arguments, _name=name: naming_fns[_name](**arguments),
        )


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
        version=RUNTIME_VERSION,
    )
    _add_naming_tools(server)
    return server
