"""Counsel domain MCP server construction and tool registration."""

import pathlib
from typing import Any

from mcp.server import MCPServer
from mcp.server.mcpserver.exceptions import ToolError

APP_DIR = pathlib.Path(__file__).resolve().parent

# 工具层为 app.* 绝对导入包（无 sys.path 注入）。

from app.contracts import COUNSEL_TOOLS  # noqa: E402
from app.contracts import VERSION as RUNTIME_VERSION  # noqa: E402
from app.contracts.yongshen import SelectedYongShenCore, YongShenSource  # noqa: E402
from app.mcp_tool_adapter import add_contract_tool as _add_contract_tool  # noqa: E402
from app.natal.tools import counsel  # noqa: E402

_CONTRACTS = {contract.name: contract for contract in COUNSEL_TOOLS}


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
    names = ["compute_factors", "natal_query", "period_query"]
    if domain in (None, "bazi"):
        names.insert(0, "bazi_yongshen")
    for name in names:
        contract = _CONTRACTS[name]

        def invoke_natal(_name=name, **a):
            """关键字直通：领域路由是唯一附加逻辑，不做 dict 二次拆包。"""
            if _name == "bazi_yongshen":
                _require_domain("bazi", a["chart"])
                return counsel.bazi_yongshen(chart=a["chart"])
            if _name == "compute_factors":
                side = domain or _chart_domain(a["chart"])
                _require_domain(side, a["chart"])
                return counsel.compute_factors(chart=a["chart"])
            if _name == "natal_query":
                side = domain or _factor_side(a)
                return counsel.natal_query(
                    factors=a["factors"],
                    topics=a["topics"],
                    factors_digest=a["factors_digest"],
                    side=side,
                )
            side = domain or _factor_side(a)
            _require_domain(side, a["chart"])
            return counsel.period_query(
                factors=a["factors"],
                factors_digest=a["factors_digest"],
                time_scope=a["time_scope"],
                topics=a["topics"],
                chart=a["chart"],
                side=side,
            )

        _add_contract_tool(server, contract, invoke_natal)


def create_counsel_mcp(domain: str) -> MCPServer:
    if domain == "naming":
        return _create_naming_server()
    if domain in ("liuyao", "qimen"):
        return _create_divination_server(domain)

    if domain not in ("bazi", "ziwei"):
        raise ValueError(f"counsel domain 无效: {domain!r}")
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 断语引擎（{domain}）",
        description=(
            f"断语引擎（{domain}）："
            + (
                "用神裁决 + "
                if domain == "bazi" else ""
            )
            + "因子快照、本命断语、应期断语；排盘由排盘引擎完成（盘面为输入）。"
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
        title="Liki Reading Engine",
        description="断语引擎：八字/紫微本命、起名、六爻与奇门能力聚合。",
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
    counsel 是补集（判断），MCP 暴露 {domain}_snapshot / {domain}_query，
    query 基于 snapshot 追问出断语。
    """
    if domain == "liuyao":
        from app.divination.tools.liuyao_ask import ask as query_ask
        from app.divination.tools.liuyao_snapshot import create as snapshot_create
    else:
        from app.divination.tools.qimen_ask import ask as query_ask
        from app.divination.tools.qimen_snapshot import create as snapshot_create
    for suffix, impl in (("snapshot", snapshot_create), ("query", query_ask)):
        contract = _CONTRACTS[f"{domain}_{suffix}"]
        # create / ask 本身就是 (**kwargs) 类型化纯函数，直通注册。
        _add_contract_tool(server, contract, impl)


def _create_divination_server(domain: str) -> MCPServer:
    server = MCPServer(
        name=f"counsel-{domain}",
        title=f"Liki 断语引擎（{domain}）",
        description=(
            f"{'六爻' if domain == 'liuyao' else '奇门'}：快照（起卦/排盘 + 因子）与追问断语，"
            f"基于快照不重排；排盘由排盘引擎完成。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_divination_tools(server, domain)
    return server


def _naming_check(given_names, yongshen=None, xishen=None, jishen=None,
                  selected_yongshen=None):
    """评估候选名；selected_yongshen 是唯一八字用神来源。"""
    from app.naming.naming import evaluate_names

    provenance: dict[str, Any] | None = None
    if selected_yongshen is not None:
        if isinstance(selected_yongshen, SelectedYongShenCore):
            selected_yongshen = selected_yongshen.model_dump()
        if not isinstance(selected_yongshen, dict):
            raise ToolError("invalid_arguments: selected_yongshen 必须是用神裁决返回的对象")
        status = selected_yongshen.get("status")
        if status in ("needs_review", "unresolved"):
            raise ToolError(
                "invalid_arguments: selected_yongshen 未形成可消费结论；"
                "请先复核用神，不得静默降级"
            )
        selected = selected_yongshen.get("selected")
        if not isinstance(selected, dict):
            raise ToolError("invalid_arguments: selected_yongshen.selected 缺失")
        yongshen = selected.get("yong")
        xishen = selected.get("xi") or []
        jishen = selected.get("ji") or []
        if not yongshen:
            raise ToolError("invalid_arguments: selected_yongshen.selected.yong 为空")
        missing = [
            field for field in ("policy_id", "yongshen_digest", "chart_digest")
            if not selected_yongshen.get(field)
        ]
        if missing:
            raise ToolError(
                "invalid_arguments: selected_yongshen 缺少来源字段: "
                + ", ".join(missing)
            )
        provenance = YongShenSource(
            **{
                field: selected_yongshen.get(field)
                for field in (
                    "policy_id", "rule_id", "status",
                    "chart_digest", "yongshen_digest",
                )
            }
        ).model_dump()
    evaluations = evaluate_names(
        given_names, yongshen or "", xishen or [], jishen or []
    )
    if provenance is not None:
        for item in evaluations:
            item["yongshen_source"] = dict(provenance)
    return evaluations


def _add_naming_tools(server: MCPServer) -> None:
    """起名（naming）：姓氏匹配、取字、单字查询、组名与候选评估（字库与音韵在 counsel 本地）。"""
    from app.naming.naming import compose_names, lookup_char, match_surnames, pick_chars

    def _pick(first, second=None):
        # 契约 {first, second?} → 实现 pick_chars(wuxing1, wuxing2, count)。
        return pick_chars(first, second or "", count=2 if second else 1)

    def _compose(first, second=None, max_names=100):
        return compose_names(first, second or [], max_names)

    naming_fns = {
        "naming_surname": match_surnames,
        "naming_pick": _pick,
        "naming_char": lookup_char,
        "naming_compose": _compose,
        "naming_check": _naming_check,
    }
    for name, impl in naming_fns.items():
        _add_contract_tool(server, _CONTRACTS[name], impl)


def _add_liuyao_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "liuyao")


def _add_qimen_tools(server: MCPServer) -> None:
    _add_divination_tools(server, "qimen")


def _create_naming_server() -> MCPServer:
    server = MCPServer(
        name="counsel-naming",
        title="Liki 断语引擎（naming）",
        description=(
            "起名评估：姓氏匹配、按五行取字、单字查询、组名、候选名评估。"
            "字库与音韵在 counsel 本地，用神/喜忌由调用方传入（八字判断结果）。"
        ),
        version=RUNTIME_VERSION,
    )
    _add_naming_tools(server)
    return server
