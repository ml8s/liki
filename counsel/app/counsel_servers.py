"""Counsel aggregate MCP server construction and tool registration."""

from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError

# 工具层为 app.* 绝对导入包（无 sys.path 注入）。
from app.contracts import COUNSEL_TOOLS  # noqa: E402
from app.contracts import VERSION as RUNTIME_VERSION  # noqa: E402
from app.mcp_tool_adapter import add_contract_tool as _add_contract_tool  # noqa: E402
from app.natal.tools import counsel  # noqa: E402

_CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}


def _require_side(side: str, chart: dict) -> None:
    """Validate that a chart belongs to the natal side inferred from factors."""
    chart_side = "bazi" if "ri" in chart else "ziwei"
    if chart_side != side:
        raise ToolError(
            f"invalid_arguments: factors 属于 {side}，收到 {chart_side} 盘；"
            "必须原样传递同一 engine 排盘和 natal_compute_factors 结果。"
        )


def _chart_side(chart: dict) -> str:
    return "bazi" if "ri" in chart else "ziwei"


def _factor_side(arguments: dict) -> str:
    factors = arguments.get("factors")
    if not isinstance(factors, dict):
        raise ToolError("invalid_arguments: factors 必须是 natal_compute_factors 返回的对象")
    provenance = factors.get("_provenance")
    if not isinstance(provenance, dict):
        raise ToolError("invalid_arguments: factors 缺少可验证的 side provenance")
    side = provenance.get("side")
    if side not in ("bazi", "ziwei"):
        raise ToolError("invalid_arguments: factors 缺少可验证的 side provenance")
    return side


def _add_natal_tools(server: MCPServer) -> None:
    """Register the aggregate natal surface.

    Bazi/ziwei is an internal technique side, not a separate MCP surface. The
    chart shape and factor provenance keep the two sides fail-closed.
    """
    for name in (
        "bazi_yongshen",
        "natal_compute_factors",
        "natal_query",
        "natal_period_query",
    ):
        contract = _CONTRACTS[name]

        def invoke_natal(_name=name, **arguments):
            if _name == "bazi_yongshen":
                if _chart_side(arguments["chart"]) != "bazi":
                    raise ToolError("invalid_arguments: bazi_yongshen 只接受八字盘。")
                return counsel.bazi_yongshen(chart=arguments["chart"])
            if _name == "natal_compute_factors":
                return counsel.natal_compute_factors(chart=arguments["chart"])
            if _name == "natal_query":
                side = _factor_side(arguments)
                return counsel.natal_query(
                    factors=arguments["factors"],
                    topics=arguments["topics"],
                    factors_digest=arguments["factors_digest"],
                    side=side,
                )
            side = _factor_side(arguments)
            _require_side(side, arguments["chart"])
            return counsel.natal_period_query(
                factors=arguments["factors"],
                factors_digest=arguments["factors_digest"],
                time_scope=arguments["time_scope"],
                topics=arguments["topics"],
                chart=arguments["chart"],
                side=side,
            )

        _add_contract_tool(server, contract, invoke_natal)


def create_counsel_root_mcp() -> MCPServer:
    """Create the single aggregate counsel surface used by the Liki expert."""
    server = MCPServer(
        name="counsel",
        title="Liki Reading Engine",
        description="断语引擎：八字/紫微本命、起名、六爻与奇门能力聚合。",
        version=RUNTIME_VERSION,
    )
    _add_natal_tools(server)
    _add_naming_tools(server)
    _add_divination_tools(server, "liuyao")
    _add_divination_tools(server, "qimen")
    return server


def _add_divination_tools(server: MCPServer, domain: str) -> None:
    """Register snapshot/query for one divination technique."""
    if domain == "liuyao":
        from app.divination.tools.liuyao_ask import ask as query_ask
        from app.divination.tools.liuyao_snapshot import create as snapshot_create
    else:
        from app.divination.tools.qimen_ask import ask as query_ask
        from app.divination.tools.qimen_snapshot import create as snapshot_create
    for suffix, implementation in (
        ("snapshot", snapshot_create),
        ("query", query_ask),
    ):
        _add_contract_tool(
            server,
            _CONTRACTS[f"{domain}_{suffix}"],
            implementation,
        )


def _naming_check(given_names, selected_yongshen=None):
    """Evaluate names against the selected bazi yong/xi/ji plan."""
    from app.naming.naming import evaluate_names

    return {"evaluations": evaluate_names(given_names, selected_yongshen)}


def _add_naming_tools(server: MCPServer) -> None:
    """Register surname matching, character pools, composition, and checks."""
    from app.naming.naming import (
        compose_names,
        lookup_char,
        match_surnames,
        pick_chars_by_selected_yongshen,
    )

    def _pick(selected_yongshen, first_role="yong", second_role=None):
        return pick_chars_by_selected_yongshen(
            selected_yongshen,
            first_role,
            second_role,
        )

    def _compose(first_chars, second_chars=None, max_names=100):
        return compose_names(first_chars, second_chars or [], max_names)

    def _char(char):
        return {"character": lookup_char(char)}

    implementations = {
        "naming_surname": match_surnames,
        "naming_pick": _pick,
        "naming_char": _char,
        "naming_compose": _compose,
        "naming_check": _naming_check,
    }
    for name, implementation in implementations.items():
        _add_contract_tool(server, _CONTRACTS[name], implementation)
