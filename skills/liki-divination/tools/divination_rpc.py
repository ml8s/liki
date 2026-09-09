"""问卦 Skill 的统一 JSON-RPC 访问层。"""
from __future__ import annotations

import json
import os
import urllib.request
from urllib.error import HTTPError, URLError


TIMEOUT = 30
RETRYABLE_HTTP_CODES = {408, 429}


class RPCError(RuntimeError):
    """统一 RPC 失败类型。"""


def call(method: str, params: dict, retries: int = 1) -> dict:
    endpoint = os.environ.get("LIKI_RPC_URL", "https://liki.hk/jsonrpc")
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
