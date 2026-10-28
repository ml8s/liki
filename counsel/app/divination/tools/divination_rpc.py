"""问卦层的统一 MCP 引擎访问层。"""


# MCP 引擎客户端（counsel 统一走 MCP）
from app.engine_mcp_client import (  # noqa: E402
    MCPError,
    _version_key,
    call,
)
from app.version import runtime_version

RPCError = MCPError

def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]


def skill_version() -> str:
    """Read the distributed skill version; fail closed when missing or invalid."""
    try:
        version = runtime_version()
    except OSError as error:
        raise RPCError(f"skill VERSION.txt is unavailable: {error}") from error
    if not version:
        raise RPCError("skill VERSION.txt is empty")
    _version_key(version)
    return version
