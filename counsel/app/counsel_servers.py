"""Counsel aggregate MCP server construction and tool registration."""

from mcp.server import MCPServer

from app.build import build_id  # noqa: E402

# 工具层为 app.* 绝对导入包（无 sys.path 注入）。
from app.contracts import COUNSEL_TOOLS  # noqa: E402
from app.mcp_tool_adapter import add_contract_tool as _add_contract_tool  # noqa: E402
from app.natal.tools import counsel  # noqa: E402

_CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}




def _add_natal_tools(server: MCPServer) -> None:
    """Register the aggregate natal surface.

    Bazi/ziwei is an internal technique side, not a separate MCP surface. The
    chart shape and factor provenance keep the two sides fail-closed.
    """
    for name in (
        "bazi_yongshen",
        "natal_query",
        "natal_liunian_query",
    ):
        contract = _CONTRACTS[name]

        def invoke_natal(_name=name, **arguments):
            if _name == "bazi_yongshen":
                return counsel.bazi_yongshen(subject=arguments["subject"])
            if _name == "natal_query":
                return counsel.natal_query(
                    subject=arguments["subject"],
                    life_topics=arguments["life_topics"],
                )
            return counsel.natal_liunian_query(
                subject=arguments["subject"],
                start_year=arguments["start_year"],
                year_count=arguments["year_count"],
                life_topics=arguments["life_topics"],
            )

        _add_contract_tool(server, contract, invoke_natal)


def create_counsel_root_mcp() -> MCPServer:
    """Create the single aggregate counsel surface used by the Liki expert."""
    server = MCPServer(
        name="hk.liki/counsel",
        title="Liki Counsel",
        description="断语引擎：八字/紫微本命、起名、六爻与奇门能力聚合。",
        version=build_id(),
        instructions=(
            "Clients may compose tool arguments from explicit user facts and prior "
            "structured tool outputs; do not invent missing facts. For natal subjects, "
            "use subject_candidates when birth time is uncertain, ambiguous, near an "
            "hour boundary, or has multiple candidates; present candidates and wait for "
            "user confirmation. Once facts are explicit or confirmed, compose a valid "
            "subject without requiring a separate provenance token or digest."
        ),
        # Counsel tools are immutable per image and this stateless process does
        # not publish change events. A false-valued subscription bus keeps the
        # SDK from advertising listChanged or accepting listen streams.
        subscriptions=False,
    )
    _add_natal_tools(server)
    _add_naming_tools(server)
    _add_divination_tools(server, "liuyao")
    _add_divination_tools(server, "qimen")
    return server


def _add_divination_tools(server: MCPServer, domain: str) -> None:
    """Register the one-shot ask capability for a divination technique."""
    if domain == "liuyao":
        from app.divination.tools.liuyao_ask import ask
    else:
        from app.divination.tools.qimen_ask import ask

    _add_contract_tool(
        server,
        _CONTRACTS[f"{domain}_ask"],
        ask,
    )


def _naming_check(**arguments):
    """Evaluate names against the selected bazi yong/xi/ji plan."""
    from app.naming.naming import evaluate_names

    return {
        "evaluations": evaluate_names(
            arguments["given_names"],
            arguments.get("selected_yongshen"),
            {
                "forbidden_chars": arguments.get("forbidden_chars") or [],
                "required_chars": arguments.get("required_chars") or [],
            },
        )
    }


def _add_naming_tools(server: MCPServer) -> None:
    """Register surname matching, character pools, composition, and checks."""
    from app.naming.naming import (
        compose_names,
        lookup_char,
        match_surnames,
        pick_chars_by_selected_yongshen,
    )

    def _pick(**arguments):
        return pick_chars_by_selected_yongshen(
            arguments["selected_yongshen"],
            arguments.get("first_role", "yong"),
            arguments.get("second_role"),
        )

    def _compose(first_chars, second_chars=None, max_names=100):
        return compose_names(first_chars, second_chars or [], max_names)

    def _char(char):
        return {"character": lookup_char(char)}

    implementations = {
        "naming_surname": match_surnames,
        "naming_pick": _pick,
        "naming_character": _char,
        "naming_compose": _compose,
        "naming_check": _naming_check,
    }
    for name, implementation in implementations.items():
        _add_contract_tool(server, _CONTRACTS[name], implementation)
