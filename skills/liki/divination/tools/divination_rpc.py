"""问卦 Skill 的统一 JSON-RPC 访问层。"""
from __future__ import annotations

import json
import os
import urllib.request
from urllib.error import HTTPError, URLError
from pathlib import Path


TIMEOUT = 30
RETRYABLE_HTTP_CODES = {408, 429}
VERSION_PATH = Path(__file__).resolve().parents[2] / "VERSION.txt"
DISCOVER_SCOPES = ("liuyao", "qimen", "huangli", "city", "tianwen", "time")
REQUIRED_METHODS = (
    "liuyao.qigua", "liuyao.chart", "qimen.chart", "huangli.days",
    "city.coords", "tianwen.time", "time.now",
)


class RPCError(RuntimeError):
    """统一 RPC 失败类型。"""


def call(method: str, params: dict, retries: int = 1) -> dict:
    endpoint = rpc_endpoint()
    body = json.dumps(
        {"jsonrpc": "2.0", "method": method, "params": params, "id": 1},
        ensure_ascii=False,
    ).encode("utf-8")
    last_error: Exception | None = None
    for _ in range(retries + 1):
        try:
            request = urllib.request.Request(
                endpoint,
                data=body,
                headers={"Content-Type": "application/json; charset=utf-8"},
            )
            with urllib.request.urlopen(request, timeout=TIMEOUT) as response:
                try:
                    result = json.loads(response.read().decode("utf-8"))
                except (UnicodeDecodeError, json.JSONDecodeError) as error:
                    raise RPCError(f"{method}: malformed JSON-RPC response: {error}") from error
            if not isinstance(result, dict):
                raise RPCError(f"{method}: malformed JSON-RPC response")
            if "error" in result:
                raise RPCError(f"{method}: {result['error']}")
            if "result" not in result:
                raise RPCError(f"{method}: JSON-RPC response missing result")
            return result["result"]
        except HTTPError as error:
            try:
                detail = error.read().decode("utf-8", errors="replace")[:500]
            except Exception:  # noqa: BLE001 - 保留原始 RPC 错误
                detail = ""
            if error.code in RETRYABLE_HTTP_CODES or error.code >= 500:
                last_error = error
                continue
            raise RPCError(f"{method}: HTTP {error.code}: {detail or error.reason}") from error
        except (URLError, ConnectionError, TimeoutError, OSError) as error:
            last_error = error
    raise RPCError(f"{method} 失败: {last_error}")


def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]


def engine_items(method: str, params: dict) -> list:
    """调用返回 data 为数组的 engine RPC。"""
    response = call(method, params)
    data = response.get("data") if isinstance(response, dict) else None
    if not isinstance(data, list):
        raise RPCError(f"{method}: engine response missing data array")
    return data


def server_time() -> str:
    data = engine_data("time.now", {})
    cst = data.get("cst")
    if not isinstance(cst, str) or not cst:
        raise RPCError("time.now: engine response missing cst")
    return cst


def rpc_endpoint() -> str:
    return os.environ.get("LIKI_RPC_URL", "https://liki.hk/jsonrpc")


def engine_version() -> str:
    payload = call("rpc.discover", {"methods": ",".join(DISCOVER_SCOPES)}, retries=0)
    info = payload.get("info") if isinstance(payload, dict) else None
    version = info.get("version") if isinstance(info, dict) else None
    if not isinstance(version, str) or not version:
        raise RPCError("engine rpc.discover response missing version")
    available = {
        method.get("name")
        for method in payload.get("methods", [])
        if isinstance(method, dict)
    }
    missing = [name for name in REQUIRED_METHODS if name not in available]
    if missing:
        raise RPCError(f"engine rpc.discover missing methods: {', '.join(missing)}")
    return version


def _version_key(version: str) -> tuple[int, ...]:
    try:
        key = tuple(int(part) for part in version.split("."))
        return key + (0,) * (4 - len(key))
    except ValueError as error:
        raise RPCError(f"engine version is invalid: {version}") from error


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


def required_engine_version() -> str:
    """Require the engine to meet the installed skill's CalVer, not a stale floor."""
    return skill_version()


def ensure_engine_compatible() -> None:
    """Fail closed when an updated skill points at an incompatible old engine."""
    version = engine_version()
    required = required_engine_version()
    if _version_key(version) < _version_key(required):
        raise RPCError(
            f"engine version {version} is incompatible; skill VERSION.txt requires engine >= {required}"
        )
