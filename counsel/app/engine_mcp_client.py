"""统一 MCP 引擎客户端。

counsel-mcp 通过标准 Streamable HTTP MCP 调用 engine-mcp。公开部署经 Caddy
剥离 /engine 前缀；容器内直接使用 engine-mcp 的 MCP endpoint。

暴露统一工具调用接口：`call(method, params)` 返回 `{"data": <MCP 工具结果>}`，
方法名自动做点号 → MCP 工具名映射（bazi.chart → bazi_chart）。
"""

import json
import os
import time
import urllib.request
from urllib.error import HTTPError, URLError

from app.version import runtime_version

MCP_TIMEOUT = int(os.environ.get("LIKI_MCP_TIMEOUT", "30"))
MAX_RETRIES = int(os.environ.get("LIKI_MCP_MAX_RETRIES", "2"))
PROTOCOL_VERSION = "2026-07-28"
RETRYABLE_HTTP_CODES = {408, 429, 500, 502, 503, 504}

CLIENT_INFO = {"name": "counsel", "version": runtime_version()}


class MCPError(Exception):
    pass


def _endpoint() -> str:
    url = os.environ.get("LIKI_MCP_URL", "").strip()
    if not url:
        raise MCPError(
            "LIKI_MCP_URL is required (engine-mcp endpoint, "
            "e.g. http://engine-mcp:8081/mcp)"
        )
    return url.rstrip("/")


def _engine_token() -> str:
    """Return the outbound engine token without coupling inbound/outbound auth."""
    return os.environ.get("LIKI_ENGINE_MCP_TOKEN", "") or os.environ.get(
        "LIKI_MCP_TOKEN", ""
    )


# MCP 工具名前缀 → 引擎分域端点后缀（各术数排盘 + 共享 aux）
_DOMAIN_BY_PREFIX = (
    ("bazi", "/bazi"),
    ("ziwei", "/ziwei"),
    ("qimen", "/qimen"),
    ("liuyao", "/liuyao"),
    ("huangli", "/huangli"),
    ("bazhai", "/fengshui"),
    ("xuankong", "/fengshui"),
    ("time", "/aux"),
    ("tianwen", "/aux"),
    ("city", "/aux"),
)


def _domain_suffix(method: str | None) -> str:
    if method is None:
        return ""
    for prefix, suffix in _DOMAIN_BY_PREFIX:
        if method.startswith(prefix):
            return suffix
    return ""


def _version_key(version: str) -> tuple[int, ...]:
    try:
        key = tuple(int(part) for part in version.split("."))
        return key + (0,) * (4 - len(key))
    except ValueError as error:
        raise MCPError(f"engine version is invalid: {version}") from error


def _meta() -> dict:
    return {
        "io.modelcontextprotocol/protocolVersion": PROTOCOL_VERSION,
        "io.modelcontextprotocol/clientInfo": CLIENT_INFO,
        "io.modelcontextprotocol/clientCapabilities": {},
    }


def _parse_body(raw: bytes) -> dict:
    """解析 MCP 响应体：可能是纯 JSON 或 SSE（event: message\\ndata: {...}）。"""
    text = raw.decode("utf-8")
    if text.lstrip().startswith("{"):
        return json.loads(text)
    for line in text.splitlines():
        if line.startswith("data: "):
            return json.loads(line[6:])
    raise MCPError(f"无法解析 MCP 响应: {text[:200]}")


def _post(method: str, name: str | None, params: dict, retries: int = 0) -> dict:
    body = json.dumps(
        {"jsonrpc": "2.0", "id": 1, "method": method, "params": {**params, "_meta": _meta()}}
    ).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "MCP-Protocol-Version": PROTOCOL_VERSION,
        "Mcp-Method": method,
        "Accept": "application/json, text/event-stream",
    }
    if name:
        headers["Mcp-Name"] = name
    if token := _engine_token():
        headers["Authorization"] = f"Bearer {token}"

    last_err: Exception | None = None
    for attempt in range(retries + 1):
        try:
            req = urllib.request.Request(_endpoint() + _domain_suffix(name), data=body, headers=headers)
            with urllib.request.urlopen(req, timeout=MCP_TIMEOUT) as resp:
                doc = _parse_body(resp.read())
            if "error" in doc:
                raise MCPError(f"{method}: {doc['error']}")
            return doc.get("result", {})
        except HTTPError as e:
            last_err = e
            if e.code not in RETRYABLE_HTTP_CODES:
                raise MCPError(f"{method}: HTTP {e.code}: {e.reason}") from e
        except (URLError, ConnectionError, TimeoutError, OSError) as e:
            last_err = e
        if attempt < retries:
            time.sleep(min(0.25 * (2**attempt), 2.0))
    raise MCPError(f"{method} 失败: {last_err}")


def call(method: str, params: dict, retries: int = MAX_RETRIES) -> dict:
    """调用 engine MCP 工具；返回 {"data": <工具结果>}。"""
    tool = method.replace(".", "_")
    result = _post("tools/call", tool, {"name": tool, "arguments": params}, retries)
    content = result.get("content") or []
    text = content[0].get("text", "") if content else ""
    if result.get("isError") or not text:
        raise MCPError(f"{method}: {text or result}")
    return {"data": json.loads(text)}


def server_info(retries: int = 0) -> dict:
    """获取引擎 MCP 的 serverInfo；stateless 下位于 _meta。仅用于诊断。"""
    result = _post("server/discover", None, {}, retries)
    meta = result.get("_meta", {}) or {}
    return meta.get("io.modelcontextprotocol/serverInfo", {}) or {}
