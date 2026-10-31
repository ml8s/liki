"""问卦层的统一 MCP 引擎访问层。"""

# MCP 引擎客户端（counsel 统一走 MCP）
from app.engine_mcp_client import MCPError, call  # noqa: E402

RPCError = MCPError


def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]
