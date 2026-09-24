"""问卦层的统一 MCP 引擎访问层。"""
from __future__ import annotations

import os
import sys
from pathlib import Path

# MCP 引擎客户端（counsel 统一走 MCP）
_LIKI_ANALYSIS = Path(__file__).resolve().parents[2]
if str(_LIKI_ANALYSIS) not in sys.path:
    sys.path.insert(0, str(_LIKI_ANALYSIS))
from engine_client import call, engine_version, MCPError  # noqa: E402

VERSION_PATH = Path(__file__).resolve().parents[2] / "VERSION.txt"


class RPCError(MCPError):
    """统一引擎失败类型（MCP 客户端）。"""

def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]


def engine_items(method: str, params: dict) -> list:
    """调用返回 data 为数组的 engine 工具。"""
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
