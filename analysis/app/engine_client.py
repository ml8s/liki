"""统一 MCP 引擎客户端。

analysis 工具层通过标准 MCP（Streamable HTTP，stateless）调用引擎，替代旧的
JSON-RPC（/jsonrpc）。RPC 仍由引擎保留在线，仅用于线上旧 skill 兼容；dev 全部走 MCP。

暴露与旧 call 兼容的接口：`call(method, params)` 返回 `{"data": <MCP 工具结果>}`，
方法名自动做 RPC 点号 → MCP 下划线映射（bazi.chart → bazi_chart）。
"""
from __future__ import annotations

import json
import os
import time
import urllib.request
from urllib.error import HTTPError, URLError

MCP_TIMEOUT = int(os.environ.get("LIKI_MCP_TIMEOUT", "30"))
MAX_RETRIES = int(os.environ.get("LIKI_MCP_MAX_RETRIES", "2"))
PROTOCOL_VERSION = "2026-07-28"
RETRYABLE_HTTP_CODES = {408, 429, 500, 502, 503, 504}

CLIENT_INFO = {"name": "liki-analysis", "version": "0.1.0"}


class MCPError(Exception):
    pass


def _endpoint() -> str:
    return os.environ.get("LIKI_MCP_URL", "https://liki.hk/mcp")


# RPC 方法前缀 → 引擎分域端点后缀（各术数排盘 + 共享 aux）
_DOMAIN_BY_PREFIX = (
    ("bazi.", "/bazi"),
    ("ziwei.", "/ziwei"),
    ("qimen.", "/qimen"),
    ("liuyao.", "/liuyao"),
    ("time.", "/aux"),
    ("tianwen.", "/aux"),
    ("city.", "/aux"),
)


def _domain_suffix(method: str) -> str:
    for prefix, suffix in _DOMAIN_BY_PREFIX:
        if method.startswith(prefix):
            return suffix
    return ""


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

    last_err: Exception | None = None
    for attempt in range(retries + 1):
        try:
            req = urllib.request.Request(_endpoint() + _domain_suffix(method), data=body, headers=headers)
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
    """调引擎 MCP 工具；返回 {"data": <工具结果>}，兼容旧 RPC 调用点。"""
    tool = method.replace(".", "_")
    result = _post("tools/call", tool, {"name": tool, "arguments": params}, retries)
    content = result.get("content") or []
    text = content[0].get("text", "") if content else ""
    if result.get("isError") or not text:
        raise MCPError(f"{method}: {text or result}")
    return {"data": json.loads(text)}


def server_info(retries: int = 0) -> dict:
    """获取引擎 MCP 的 serverInfo（含版本）；stateless 下位于 _meta。"""
    result = _post("server/discover", None, {}, retries)
    meta = result.get("_meta", {}) or {}
    return meta.get("io.modelcontextprotocol/serverInfo", {}) or {}


def engine_version() -> str:
    info = server_info()
    version = info.get("version")
    if not isinstance(version, str) or not version:
        raise MCPError("engine server.discover response missing version")
    return version