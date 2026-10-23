"""问卦层的统一 MCP 引擎访问层。"""

from pathlib import Path

# MCP 引擎客户端（counsel 统一走 MCP）
from app.engine_mcp_client import (  # noqa: E402
    MCPError,
    _version_key,
    call,
)

VERSION_PATH = Path(__file__).resolve().parents[2] / "VERSION.txt"


class RPCError(MCPError):
    """统一引擎失败类型（MCP 客户端）。"""

def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]


def server_time(timezone: str) -> dict:
    """返回 {instant, timezone, local_time}；timezone 必填（IANA）。"""
    if not timezone:
        raise RPCError("time.now: timezone is required")
    data = engine_data("time.now", {"timezone": timezone})
    instant = data.get("instant")
    local = data.get("local_time")
    if not isinstance(instant, str) or not instant:
        raise RPCError("time.now: engine response missing instant")
    return {"instant": instant, "timezone": timezone, "local_time": local}


def skill_version() -> str:
    """Read the distributed skill version; fail closed when missing or invalid."""
    try:
        version = VERSION_PATH.read_text(encoding="utf-8").strip()
    except OSError as error:
        raise RPCError(f"skill VERSION.txt is unavailable: {error}") from error
    if not version:
        raise RPCError("skill VERSION.txt is empty")
    _version_key(version)
    return version
