"""统一 MCP 引擎客户端。

counsel-mcp 通过标准 Streamable HTTP MCP 调用 engine-mcp。公开部署经 Caddy
剥离 /engine 前缀；容器内直接使用 http://engine-mcp:8081/mcp。

暴露与旧 call 兼容的接口：`call(method, params)` 返回 `{"data": <MCP 工具结果>}`，
方法名自动做 RPC 点号 → MCP 下划线映射（bazi.chart → bazi_chart）。
"""
from __future__ import annotations

import json
import os
import pathlib
import threading
import time
import urllib.request
from urllib.error import HTTPError, URLError

MCP_TIMEOUT = int(os.environ.get("LIKI_MCP_TIMEOUT", "30"))
MAX_RETRIES = int(os.environ.get("LIKI_MCP_MAX_RETRIES", "2"))
PROTOCOL_VERSION = "2026-07-28"
RETRYABLE_HTTP_CODES = {408, 429, 500, 502, 503, 504}

VERSION_PATH = pathlib.Path(__file__).resolve().parent / "VERSION.txt"
CLIENT_INFO = {"name": "counsel", "version": VERSION_PATH.read_text(encoding="utf-8").strip()}
_COMPATIBILITY_LOCK = threading.Lock()
_COMPATIBILITY_CHECKED = False


class MCPError(Exception):
    pass


def _endpoint() -> str:
    return os.environ.get("LIKI_MCP_URL", "https://liki.hk/engine/mcp").rstrip("/")


def _engine_token() -> str:
    """Return the outbound engine token without coupling inbound/outbound auth."""
    return os.environ.get("LIKI_ENGINE_MCP_TOKEN", "") or os.environ.get(
        "LIKI_MCP_TOKEN", ""
    )


# RPC 方法前缀 → 引擎分域端点后缀（各术数排盘 + 共享 aux）
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


def required_engine_version() -> str:
    try:
        version = VERSION_PATH.read_text(encoding="utf-8").strip()
    except OSError as error:
        raise MCPError(f"counsel VERSION.txt is unavailable: {error}") from error
    if not version:
        raise MCPError("counsel VERSION.txt is empty")
    _version_key(version)
    return version


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
    """调引擎 MCP 工具；返回 {"data": <工具结果>}，兼容旧 RPC 调用点。"""
    ensure_engine_compatible()
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


def ensure_engine_compatible(
    version: str | None = None, required: str | None = None
) -> None:
    """每个进程做一次 engine/counsel CalVer 兼容检查，失败即 fail closed。"""
    global _COMPATIBILITY_CHECKED
    with _COMPATIBILITY_LOCK:
        if version is None or required is None:
            if _COMPATIBILITY_CHECKED:
                return
            version = engine_version()
            required = required_engine_version()
            version_key = _version_key(version)
            required_key = _version_key(required)
        else:
            version_key = _version_key(version)
            required_key = _version_key(required)
        if version_key < required_key:
            raise MCPError(
                f"engine version {version} is incompatible; "
                f"counsel VERSION.txt requires engine >= {required}"
            )
        _COMPATIBILITY_CHECKED = True
