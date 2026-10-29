"""统一 MCP 引擎客户端。

counsel-mcp 通过标准 Streamable HTTP MCP 调用 engine-mcp。公开部署经 Caddy
剥离 /engine 前缀；容器内直接使用 engine-mcp 的 MCP endpoint。

暴露统一工具调用接口：`call(method, params)` 返回 `{"data": <MCP 工具结果>}`，
方法名自动做点号 → MCP 工具名映射（bazi.chart → bazi_chart）。
"""

import http.client
import json
import os
import threading
import time
import urllib.parse
import urllib.request
from collections.abc import Callable
from dataclasses import dataclass
from typing import Any
from urllib.error import HTTPError

from app.metrics import observe_engine_request
from app.version import runtime_version

MCP_TIMEOUT = int(os.environ.get("LIKI_ENGINE_MCP_TIMEOUT_SECONDS", "30"))
MAX_RETRIES = int(os.environ.get("LIKI_ENGINE_MCP_MAX_RETRIES", "2"))
PROTOCOL_VERSION = "2026-07-28"
RETRYABLE_HTTP_CODES = {408, 429, 500, 502, 503, 504}
HTTP_IDLE_TIMEOUT = float(os.environ.get("LIKI_HTTP_IDLE_TIMEOUT_SECONDS", "55"))

CLIENT_INFO = {"name": "counsel", "version": runtime_version()}
_HTTP_TRANSPORT: Callable[[urllib.request.Request, int], Any] | None = None

_thread_local = threading.local()


@dataclass(frozen=True)
class _Endpoint:
    origin: tuple[str, str, int]
    path: str

    @property
    def base_url(self) -> str:
        scheme, host, port = self.origin
        default = 443 if scheme == "https" else 80
        authority = host if port == default else f"{host}:{port}"
        return f"{scheme}://{authority}"


class MCPError(Exception):
    pass


def _endpoint() -> str:
    url = os.environ.get("LIKI_ENGINE_MCP_URL", "").strip()
    if not url:
        raise MCPError(
            "LIKI_ENGINE_MCP_URL is required (engine-mcp endpoint, "
            "e.g. http://engine-mcp:8081/mcp)"
        )
    return url.rstrip("/")


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


def _endpoint_parts() -> _Endpoint:
    raw = _endpoint()
    parsed = urllib.parse.urlsplit(raw)
    if parsed.scheme not in ("http", "https") or not parsed.hostname:
        raise MCPError(f"LIKI_ENGINE_MCP_URL must be an http(s) URL: {raw}")
    default_port = 443 if parsed.scheme == "https" else 80
    port = parsed.port or default_port
    path = parsed.path.rstrip("/") or ""
    if parsed.query:
        path = f"{path}?{parsed.query}"
    return _Endpoint((parsed.scheme, parsed.hostname.lower(), port), path)


def _new_connection(
    endpoint: _Endpoint, timeout: float = MCP_TIMEOUT
) -> http.client.HTTPConnection:
    scheme, host, port = endpoint.origin
    if scheme == "https":
        return http.client.HTTPSConnection(host, port, timeout=timeout)
    return http.client.HTTPConnection(host, port, timeout=timeout)


def _borrow_connection(endpoint: _Endpoint) -> http.client.HTTPConnection:
    """Return this worker thread's reusable connection, opening one when stale."""
    connections = getattr(_thread_local, "connections", None)
    if connections is None:
        connections = {}
        _thread_local.connections = connections

    connection = connections.get(endpoint.origin)
    stale = connection is None or getattr(connection, "sock", None) is None
    if not stale:
        last_used = getattr(connection, "last_used", 0.0)
        stale = HTTP_IDLE_TIMEOUT > 0 and time.monotonic() - last_used > HTTP_IDLE_TIMEOUT
    if stale:
        if connection is not None:
            connection.close()
        connection = _new_connection(endpoint)
        connection.last_used = time.monotonic()
        connections[endpoint.origin] = connection
    return connection


def _release_connection(
    endpoint: _Endpoint, connection: http.client.HTTPConnection, response: http.client.HTTPResponse
) -> None:
    if response.will_close:
        connection.close()
        connections = getattr(_thread_local, "connections", None)
        if connections and connections.get(endpoint.origin) is connection:
            del connections[endpoint.origin]
        return
    setattr(connection, "last_used", time.monotonic())


def _discard_connection(endpoint: _Endpoint, connection: http.client.HTTPConnection) -> None:
    """Drop a pooled connection after a transport failure; it may be unsynchronized."""
    connection.close()
    connections = getattr(_thread_local, "connections", None)
    if connections and connections.get(endpoint.origin) is connection:
        del connections[endpoint.origin]


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
    try:
        text = raw.decode("utf-8")
    except UnicodeDecodeError as error:
        raise MCPError("无法解析 MCP 响应: body is not UTF-8") from error
    if text.lstrip().startswith("{"):
        return _parse_json(text, text[:200])
    for line in text.splitlines():
        if line.startswith("data:"):
            payload = line[5:].lstrip()
            return _parse_json(payload, text[:200])
    raise MCPError(f"无法解析 MCP 响应: {text[:200]}")


def _parse_json(text: str, preview: str) -> dict:
    try:
        value = json.loads(text)
    except json.JSONDecodeError as error:
        raise MCPError(f"无法解析 MCP 响应: {preview}") from error
    if not isinstance(value, dict):
        raise MCPError(f"无法解析 MCP 响应: {preview}")
    return value


def _effective_retries(method: str, params: dict, retries: int) -> int:
    """Constrain generic retries by the semantic idempotency of a tool.

    Auto qigua is a server-side random event. A retryable transport response
    cannot tell whether the engine already cast a hexagram, so counsel must not
    silently cast another one. Explicit coins/yaos inputs remain retryable.
    """
    tool = method.replace(".", "_")
    if tool == "liuyao_qigua" and params.get("mode") == "auto":
        return 0
    return retries


def _post_once(
    method: str,
    name: str | None,
    params: dict,
    retries: int,
    timeout: float | None,
) -> dict:
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
        connection: http.client.HTTPConnection | None = None
        try:
            endpoint = _endpoint_parts()
            url = endpoint.path or "/"
            suffix = _domain_suffix(name)
            if suffix:
                if "?" in url:
                    url = url.replace("?", f"{suffix}?", 1)
                else:
                    url += suffix
            request = urllib.request.Request(endpoint.base_url + url, data=body, headers=headers)
            if _HTTP_TRANSPORT is not None:
                effective_timeout = MCP_TIMEOUT if timeout is None else timeout
                with _HTTP_TRANSPORT(request, effective_timeout) as response:
                    raw = response.read()
            else:
                if timeout is None:
                    connection = _borrow_connection(endpoint)
                else:
                    connection = _new_connection(endpoint, timeout)
                connection.request(
                    "POST", url, body=body,
                    headers={**headers, "Connection": "keep-alive"},
                )
                response = connection.getresponse()
                raw = response.read()
                if timeout is None:
                    _release_connection(endpoint, connection, response)
                else:
                    connection.close()
            if not 200 <= response.status < 300:
                error = HTTPError(request.full_url, response.status, response.reason, response.headers, None)
                last_err = error
                if response.status not in RETRYABLE_HTTP_CODES:
                    raise MCPError(f"{method}: HTTP {response.status}: {response.reason}") from error
                continue
            doc = _parse_body(raw)
            if "error" in doc:
                raise MCPError(f"{method}: {doc['error']}")
            return doc.get("result", {})
        except HTTPError as e:
            last_err = e
            if connection is not None:
                _discard_connection(endpoint, connection)
            if e.code not in RETRYABLE_HTTP_CODES:
                raise MCPError(f"{method}: HTTP {e.code}: {e.reason}") from e
        except (http.client.HTTPException, ConnectionError, TimeoutError, OSError) as e:
            last_err = e
            if connection is not None:
                _discard_connection(endpoint, connection)
        if attempt < retries:
            time.sleep(min(0.25 * (2**attempt), 2.0))
    raise MCPError(f"{method} 失败: {last_err}")


def _post(
    method: str,
    name: str | None,
    params: dict,
    retries: int = 0,
    timeout: float | None = None,
) -> dict:
    started = time.monotonic()
    status = "success"
    try:
        return _post_once(method, name, params, retries, timeout)
    except MCPError as error:
        message = str(error)
        if "HTTP 408" in message or "HTTP 429" in message or "HTTP 5" in message:
            status = "server_error"
        elif "HTTP 4" in message:
            status = "client_error"
        else:
            status = "transport_or_protocol_error"
        raise
    finally:
        observe_engine_request(method, status, time.monotonic() - started)


def call(method: str, params: dict, retries: int = MAX_RETRIES) -> dict:
    """调用 engine MCP 工具；返回 {"data": <工具结果>}。"""
    tool = method.replace(".", "_")
    retries = _effective_retries(method, params, retries)
    result = _post("tools/call", tool, {"name": tool, "arguments": params}, retries)
    content = result.get("content") or []
    text = content[0].get("text", "") if content else ""
    if result.get("isError"):
        raise MCPError(f"{method}: {text or result}")
    # Structured content is the canonical and only tool result; the engine
    # answers structured-only and errors arrive as text tool errors above.
    data = result.get("structuredContent")
    if data is None:
        raise MCPError(f"{method}: engine result missing structuredContent")
    return {"data": data}


def server_info(retries: int = 0, timeout: float | None = None) -> dict:
    """获取引擎 MCP 的 serverInfo；stateless 下位于 _meta。仅用于诊断。"""
    result = _post("server/discover", None, {}, retries, timeout)
    meta = result.get("_meta", {}) or {}
    return meta.get("io.modelcontextprotocol/serverInfo", {}) or {}
