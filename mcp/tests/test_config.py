import base64
import json
from functools import lru_cache

import rsa
from app.config import EXPOSE_CONFIG_PATH, FORBIDDEN_TOOLS, load_config


@lru_cache(maxsize=1)
def python_pkcs1_key() -> str:
    key = rsa.newkeys(2048)[1]
    return base64.b64encode(key.save_pkcs1(format="DER")).decode("ascii")


def test_expose_contract_is_the_single_source():
    contract = json.loads(EXPOSE_CONFIG_PATH.read_text(encoding="utf-8"))["expose"]
    exposed = load_config().exposed_tools

    assert set(contract) == {"engine", "counsel"}
    assert exposed == {
        name: server
        for server, tools in contract.items()
        for name in tools
    }
    assert FORBIDDEN_TOOLS.isdisjoint(exposed)


def test_sandbox_profile_is_mapped_at_config_boundary(monkeypatch, tmp_path):
    profile = {
        "appIds": [{
            "appId": "sandbox-app-id",
            "appPrivatePkcsKey": python_pkcs1_key(),
            "alipayPublicKey": "sandbox-alipay-public-key",
        }],
        "sandboxAccounts": {"partner": {"userId": "2088sandbox-seller"}},
    }
    (tmp_path / ".alipay-sandbox.json").write_text(
        json.dumps(profile),
        encoding="utf-8",
    )
    monkeypatch.chdir(tmp_path)
    for name in (
        "LIKI_ALIPAY_SANDBOX_CONFIG",
        "LIKI_PAYWALL_ENABLED",
        "ALIPAY_APP_ID",
        "ALIPAY_PRIVATE_KEY",
        "ALIPAY_PUBLIC_KEY",
        "ALIPAY_SELLER_ID",
        "ALIPAY_SERVICE_ID",
        "ALIPAY_GATEWAY",
    ):
        monkeypatch.delenv(name, raising=False)
    monkeypatch.setenv("LIKI_PAYWALL_ENABLED", "true")
    monkeypatch.setenv(
        "LIKI_ALIPAY_SANDBOX_CONFIG",
        str(tmp_path / ".alipay-sandbox.json"),
    )

    config = load_config().aipay
    assert config.app_id == "sandbox-app-id"
    assert config.private_key == python_pkcs1_key()
    assert config.alipay_public_key == "sandbox-alipay-public-key"
    assert config.seller_id == "2088sandbox-seller"
    assert config.gateway.endswith("openapi-sandbox.dl.alipaydev.com/gateway.do")
    assert config.service_id == "api_mock_service_id"
    assert config.amount == "9.90"
    assert config.db_path == "/data/aipay.db"


def test_explicit_pkcs1_key_is_preferred_for_python_gateway(monkeypatch):
    for name in (
        "LIKI_ALIPAY_SANDBOX_CONFIG",
        "LIKI_PAYWALL_ENABLED",
        "ALIPAY_APP_ID",
        "ALIPAY_PRIVATE_KEY",
        "ALIPAY_PRIVATE_KEY_PKCS1",
        "ALIPAY_PUBLIC_KEY",
        "ALIPAY_SELLER_ID",
        "ALIPAY_SERVICE_ID",
        "ALIPAY_GATEWAY",
    ):
        monkeypatch.delenv(name, raising=False)
    monkeypatch.setenv("LIKI_PAYWALL_ENABLED", "true")
    monkeypatch.setenv("ALIPAY_APP_ID", "app")
    monkeypatch.setenv("ALIPAY_PRIVATE_KEY", "go-compatible-key")
    monkeypatch.setenv("ALIPAY_PRIVATE_KEY_PKCS1", python_pkcs1_key())
    monkeypatch.setenv("ALIPAY_PUBLIC_KEY", "public")
    monkeypatch.setenv("ALIPAY_SELLER_ID", "2088test")
    monkeypatch.setenv("ALIPAY_SERVICE_ID", "service")
    monkeypatch.setenv("ALIPAY_GATEWAY", "https://openapi.alipay.com/gateway.do")

    assert load_config().aipay.private_key == python_pkcs1_key()


def test_production_pkcs8_key_is_not_a_python_fallback(monkeypatch):
    for name in (
        "LIKI_ALIPAY_SANDBOX_CONFIG",
        "LIKI_PAYWALL_ENABLED",
        "ALIPAY_APP_ID",
        "ALIPAY_PRIVATE_KEY",
        "ALIPAY_PRIVATE_KEY_PKCS1",
        "ALIPAY_PUBLIC_KEY",
        "ALIPAY_SELLER_ID",
        "ALIPAY_SERVICE_ID",
        "ALIPAY_GATEWAY",
    ):
        monkeypatch.delenv(name, raising=False)
    monkeypatch.setenv("LIKI_PAYWALL_ENABLED", "true")
    monkeypatch.setenv("ALIPAY_APP_ID", "app")
    monkeypatch.setenv("ALIPAY_PRIVATE_KEY", "go-compatible-pkcs8-key")
    monkeypatch.setenv("ALIPAY_PUBLIC_KEY", "public")
    monkeypatch.setenv("ALIPAY_SELLER_ID", "2088test")
    monkeypatch.setenv("ALIPAY_SERVICE_ID", "service")
    monkeypatch.setenv("ALIPAY_GATEWAY", "https://openapi.alipay.com/gateway.do")

    try:
        load_config()
        raise AssertionError("PKCS#8 fallback was accepted")
    except ValueError as error:
        assert "ALIPAY_PRIVATE_KEY_PKCS1" in str(error)
