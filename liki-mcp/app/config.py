"""LIKI-MCP gateway configuration."""

from __future__ import annotations

import json
import os
from decimal import Decimal, InvalidOperation
from dataclasses import dataclass
from pathlib import Path
from typing import Any

LATEST_PROTOCOL_VERSION = "2026-07-28"
EXPOSE_CONFIG_PATH = Path(__file__).with_name("expose.json")
SANDBOX_GATEWAY = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
PRODUCTION_GATEWAY = "https://openapi.alipay.com/gateway.do"
SANDBOX_SERVICE_ID = "api_mock_service_id"
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
        raise ValueError(f"expose config must contain a JSON object: {path}")
    return value


def _load_sandbox_profile() -> dict[str, Any]:
    """Load an explicitly configured sandbox profile at the server boundary."""
    configured = _env("LIKI_ALIPAY_SANDBOX_CONFIG", "")
    if not configured:
        return {}
    path = Path(configured)
    if not path.is_absolute():
        path = Path.cwd() / path
    if not path.is_file():
        raise ValueError(f"LIKI_ALIPAY_SANDBOX_CONFIG 文件不存在: {path}")
    profile = _load_json(str(path))
    apps = profile.get("appIds")
    app = apps[0] if isinstance(apps, list) and apps and isinstance(apps[0], dict) else {}
    accounts = profile.get("sandboxAccounts")
    partner = accounts.get("partner") if isinstance(accounts, dict) else {}
    return {
        "app_id": app.get("appId"),
        "private_key": app.get("appPrivatePkcsKey"),
        "alipay_public_key": app.get("alipayPublicKey"),
        "seller_id": partner.get("userId") if isinstance(partner, dict) else None,
    }


# 原始六爻 / 奇门盘面由 counsel 的 ask 方法封装，不从 LIKI 暴露。
FORBIDDEN_TOOLS = {
    "qimen_chart",
    "liuyao_chart",
    "liuyao_qigua",
}

@dataclass(frozen=True)
class UpstreamConfig:
    name: str
    url: str
    timeout_seconds: float


@dataclass(frozen=True)
class AipayConfig:
    """AI A2M 按量付费配置（领域值对象）。"""

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
    db_path: str

    @property
    def is_exact_sandbox(self) -> bool:
        return self.gateway == SANDBOX_GATEWAY and self.service_id == SANDBOX_SERVICE_ID


def normalize_amount(value: str) -> str | None:
    """Normalize an AI A2M amount to a positive two-decimal CNY string."""
    text = str(value or "").strip()
    try:
        amount = Decimal(text)
    except (InvalidOperation, ValueError):
        return None
    if amount <= 0 or amount.as_tuple().exponent < -2:
        return None
    return f"{amount:.2f}"


def _validate_pkcs1_private_key(private_key: str) -> None:
    """Fail fast when the Python SDK cannot use the configured key."""
    from alipay.aop.api.util.SignatureUtils import sign_with_rsa2

    try:
        sign_with_rsa2(private_key, "liki-config-probe", "utf-8")
    except Exception as exc:
        raise ValueError(
            "ALIPAY_PRIVATE_KEY_PKCS1 不是 Python SDK 可用的 PKCS#1 私钥"
        ) from exc


def _load_aipay_config() -> AipayConfig:
    """从环境变量读取 AI A2M 配置；enabled=true 时必填字段缺失即 fail-fast。"""
    enabled = _env("LIKI_PAYWALL_ENABLED", "false").lower() == "true"
    paid_tools_raw = _env("ALIPAY_PAID_TOOLS", "subject_candidates")
    paid_tools = frozenset(t.strip() for t in paid_tools_raw.split(",") if t.strip())
    amount = _env("ALIPAY_AIPAY_AMOUNT", "9.90")
    gateway = _env("ALIPAY_GATEWAY", SANDBOX_GATEWAY)
    profile = _load_sandbox_profile() if gateway == SANDBOX_GATEWAY else {}
    app_id = _env("ALIPAY_APP_ID", str(profile.get("app_id") or ""))
    # Python needs PKCS#1. The Go web deployment reads ALIPAY_PRIVATE_KEY
    # separately; its PKCS#8 value must not leak into this gateway.
    private_key = _env(
        "ALIPAY_PRIVATE_KEY_PKCS1",
        str(profile.get("private_key") or ""),
    )
    alipay_public_key = _env(
        "ALIPAY_PUBLIC_KEY",
        str(profile.get("alipay_public_key") or ""),
    )
    seller_id = _env("ALIPAY_SELLER_ID", str(profile.get("seller_id") or ""))
    service_id = _env("ALIPAY_SERVICE_ID", SANDBOX_SERVICE_ID)
    normalized_amount = normalize_amount(amount)

    if enabled:
        missing = [
            var for var, val in (
                ("ALIPAY_APP_ID", app_id),
                ("ALIPAY_PRIVATE_KEY_PKCS1", private_key),
                ("ALIPAY_PUBLIC_KEY", alipay_public_key),
                ("ALIPAY_SELLER_ID", seller_id),
                ("ALIPAY_SERVICE_ID", service_id),
            ) if not val.strip()
        ]
        if not normalized_amount:
            missing.append("ALIPAY_AIPAY_AMOUNT")
        if missing:
            raise ValueError(
                f"LIKI_PAYWALL_ENABLED=true 但缺少必需配置: {', '.join(missing)}"
            )
        _validate_pkcs1_private_key(private_key)
        if gateway == SANDBOX_GATEWAY and service_id != SANDBOX_SERVICE_ID:
            raise ValueError(
                f"支付宝沙箱网关必须使用 ALIPAY_SERVICE_ID={SANDBOX_SERVICE_ID}"
            )
        if gateway == PRODUCTION_GATEWAY and service_id == SANDBOX_SERVICE_ID:
            raise ValueError("生产网关禁止使用 AI 按量付费沙箱 service_id")
        if gateway not in {SANDBOX_GATEWAY, PRODUCTION_GATEWAY}:
            raise ValueError("ALIPAY_GATEWAY 只支持支付宝沙箱或生产网关")

    return AipayConfig(
        enabled=enabled,
        app_id=app_id,
        private_key=private_key,
        alipay_public_key=alipay_public_key,
        gateway=gateway,
        seller_id=seller_id,
        service_id=service_id,
        amount=normalized_amount or amount,
        goods_name=_env("ALIPAY_AIPAY_GOODS_NAME", "Liki 命理服务"),
        seller_name=_env("ALIPAY_SELLER_NAME", "Liki"),
        paid_tools=paid_tools,
        db_path=_env("LIKI_AIPAY_DB_PATH", "/data/aipay.db"),
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
    """Load upstream URLs and the LIKI product expose contract."""
    configured = _load_json(str(EXPOSE_CONFIG_PATH)).get("expose", {})
    if not isinstance(configured, dict):
        raise ValueError("expose must be an object")
    exposed: dict[str, str] = {}
    for server, tools in configured.items():
        if server not in {"engine", "counsel"} or not isinstance(tools, list):
            raise ValueError("expose server must be engine or counsel")
        for name in tools:
            exposed[str(name)] = server

    aipay = _load_aipay_config()
    unknown_paid_tools = aipay.paid_tools.difference(exposed)
    if aipay.enabled and unknown_paid_tools:
        raise ValueError(
            f"ALIPAY_PAID_TOOLS 包含未暴露工具: {', '.join(sorted(unknown_paid_tools))}"
        )

    unknown_forbidden = FORBIDDEN_TOOLS.intersection(exposed)
    if unknown_forbidden:
        raise ValueError(
            "LIKI may not expose raw divination charts: "
            + ", ".join(sorted(unknown_forbidden))
        )

    return GatewayConfig(
        server_name=_env("LIKI_MCP_SERVER_NAME", "hk.liki/liki"),
        server_title=_env("LIKI_MCP_SERVER_TITLE", "Liki — Professional AI Astrologer"),
        server_version=_env("LIKI_VERSION", "2026.11.09.1"),
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
        aipay=aipay,
    )
