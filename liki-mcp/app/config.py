"""LIKI-MCP gateway configuration."""

from __future__ import annotations

import json
import os
from dataclasses import dataclass
from pathlib import Path
from typing import Any

DEFAULT_PROTOCOL_VERSION = "2026-07-28"
SUPPORTED_PROTOCOL_VERSIONS = {
    "2025-03-26",
    "2025-06-18",
    "2025-11-25",
    "2026-07-28",
}


def _env(name: str, default: str) -> str:
    return os.environ.get(name, default).strip()


def _load_json(path: str) -> dict[str, Any]:
    with Path(path).open(encoding="utf-8") as stream:
        value = json.load(stream)
    if not isinstance(value, dict):
        raise ValueError(f"LIKI_EXPOSE_CONFIG must contain a JSON object: {path}")
    return value


# LIKI 暴露面是产品契约。新增工具必须显式加入这里。
ENGINE_EXPOSED_TOOLS_NAMES = (
    "time_now",
    "subject_candidates",
    "city_info",
    "huangli_days",
    "bazi_chart",
    "ziwei_chart",
    "bazi_liupan",
    "ziwei_liupan",
    "bazi_bond",
    "ziwei_bond",
    "bazhai_chart",
    "bazhai_layout",
    "xuankong_chart",
    "xuankong_liunian",
)

ENGINE_EXPOSED_TOOLS = {
    "time_now",
    "subject_candidates",
    "city_info",
    "huangli_days",
    "bazi_chart",
    "ziwei_chart",
    "bazi_liupan",
    "ziwei_liupan",
    "bazi_bond",
    "ziwei_bond",
    "bazhai_chart",
    "bazhai_layout",
    "xuankong_chart",
    "xuankong_liunian",
}

# 原始六爻 / 奇门盘面由 counsel 的 ask 方法封装，不从 LIKI 暴露。
FORBIDDEN_TOOLS = {
    "qimen_chart",
    "liuyao_chart",
    "liuyao_qigua",
}

COUNSEL_EXPOSED_TOOLS_NAMES = (
    "bazi_yongshen",
    "natal_query",
    "natal_liunian_query",
    "naming_surname",
    "naming_pick",
    "naming_character",
    "naming_compose",
    "naming_check",
    "liuyao_ask",
    "qimen_ask",
)

COUNSEL_EXPOSED_TOOLS = set(COUNSEL_EXPOSED_TOOLS_NAMES)


@dataclass(frozen=True)
class UpstreamConfig:
    name: str
    url: str
    timeout_seconds: float


@dataclass(frozen=True)
class AipayConfig:
    """SkillPay A2M 按量付费配置（领域值对象）。"""

    enabled: bool
    app_id: str
    private_key: str
    alipay_public_key: str
    gateway: str
    seller_id: str
    service_id: str
    amount: str
    goods_name: str
    seller_name: str
    paid_tools: frozenset[str]


def _load_aipay_config() -> AipayConfig:
    """从环境变量读取 SkillPay 配置；enabled=true 时必填字段缺失即 fail-fast。"""
    enabled = _env("LIKI_PAYWALL_ENABLED", "false").lower() == "true"
    paid_tools_raw = _env("ALIPAY_PAID_TOOLS", "subject_candidates")
    paid_tools = frozenset(t.strip() for t in paid_tools_raw.split(",") if t.strip())

    if enabled:
        missing = [
            var for var, val in (
                ("ALIPAY_APP_ID", _env("ALIPAY_APP_ID", "")),
                ("ALIPAY_PRIVATE_KEY", _env("ALIPAY_PRIVATE_KEY", "")),
                ("ALIPAY_PUBLIC_KEY", _env("ALIPAY_PUBLIC_KEY", "")),
                ("ALIPAY_SELLER_ID", _env("ALIPAY_SELLER_ID", "")),
                ("ALIPAY_SERVICE_ID", _env("ALIPAY_SERVICE_ID", "")),
            ) if not val.strip()
        ]
        if missing:
            raise ValueError(
                f"LIKI_PAYWALL_ENABLED=true 但缺少必需配置: {', '.join(missing)}"
            )

    return AipayConfig(
        enabled=enabled,
        app_id=_env("ALIPAY_APP_ID", ""),
        private_key=_env("ALIPAY_PRIVATE_KEY", ""),
        alipay_public_key=_env("ALIPAY_PUBLIC_KEY", ""),
        gateway=_env("ALIPAY_GATEWAY", "https://openapi-sandbox.dl.alipaydev.com/gateway.do"),
        seller_id=_env("ALIPAY_SELLER_ID", ""),
        service_id=_env("ALIPAY_SERVICE_ID", ""),
        amount=_env("ALIPAY_AIPAY_AMOUNT", "9.90"),
        goods_name=_env("ALIPAY_AIPAY_GOODS_NAME", "Liki 命理服务"),
        seller_name=_env("ALIPAY_SELLER_NAME", "Liki"),
        paid_tools=paid_tools,
    )


@dataclass(frozen=True)
class GatewayConfig:
    server_name: str
    server_title: str
    server_version: str
    engine: UpstreamConfig
    counsel: UpstreamConfig
    exposed_tools: dict[str, str]
    forbidden_tools: frozenset[str]
    aipay: AipayConfig


def load_config() -> GatewayConfig:
    """Load upstream URLs and the fixed LIKI expose allowlist."""
    config_path = _env("LIKI_EXPOSE_CONFIG", "")
    exposed: dict[str, str] = {
        name: "engine" for name in ENGINE_EXPOSED_TOOLS_NAMES
    }
    exposed.update({name: "counsel" for name in COUNSEL_EXPOSED_TOOLS_NAMES})

    if config_path:
        configured = _load_json(config_path).get("expose", {})
        if not isinstance(configured, dict):
            raise ValueError("expose must be an object")
        for server, tools in configured.items():
            if server not in {"engine", "counsel"} or not isinstance(tools, list):
                raise ValueError("expose server must be engine or counsel")
            for name in tools:
                exposed[str(name)] = server

    unknown_forbidden = FORBIDDEN_TOOLS.intersection(exposed)
    if unknown_forbidden:
        raise ValueError(
            "LIKI may not expose raw divination charts: "
            + ", ".join(sorted(unknown_forbidden))
        )

    return GatewayConfig(
        server_name=_env("LIKI_MCP_SERVER_NAME", "liki-mcp"),
        server_title=_env("LIKI_MCP_SERVER_TITLE", "LIKI product MCP"),
        server_version=_env("LIKI_VERSION", "0.1.0"),
        engine=UpstreamConfig(
            name="engine",
            url=_env("LIKI_UPSTREAM_ENGINE_URL", "http://engine-mcp:8081/mcp"),
            timeout_seconds=float(
                _env("LIKI_UPSTREAM_ENGINE_TIMEOUT_SECONDS", "35")
            ),
        ),
        counsel=UpstreamConfig(
            name="counsel",
            url=_env("LIKI_UPSTREAM_COUNSEL_URL", "http://counsel-mcp:8086/mcp"),
            timeout_seconds=float(
                _env("LIKI_UPSTREAM_COUNSEL_TIMEOUT_SECONDS", "35")
            ),
        ),
        exposed_tools=exposed,
        forbidden_tools=frozenset(FORBIDDEN_TOOLS),
        aipay=_load_aipay_config(),
    )
