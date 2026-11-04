"""SkillPay paywall integration tests.

Tests the gateway paywall logic with mocked Alipay SDK.
The alipay-sdk-python is not required to run these tests.
"""

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from app.config import AipayConfig, GatewayConfig, UpstreamConfig

# Mock alipay SDK modules so tests run without the SDK installed.
# The SDK is only needed in production when paywall is enabled.
import sys
from unittest.mock import MagicMock as _MM

_alipay_modules = [
    "alipay", "alipay.aop", "alipay.aop.api",
    "alipay.aop.api.AlipayClientConfig",
    "alipay.aop.api.DefaultAlipayClient",
    "alipay.aop.api.domain",
    "alipay.aop.api.domain.AlipayAipayAgentPaymentVerifyModel",
    "alipay.aop.api.domain.AlipayAipayAgentFulfillmentConfirmModel",
    "alipay.aop.api.request",
    "alipay.aop.api.request.AlipayAipayAgentPaymentVerifyRequest",
    "alipay.aop.api.request.AlipayAipayAgentFulfillmentConfirmRequest",
]
for _mod in _alipay_modules:
    if _mod not in sys.modules:
        sys.modules[_mod] = _MM()
from app.order_store import OrderStore
from app.payment import PaymentService


def _aipay_config(enabled: bool = True) -> AipayConfig:
    return AipayConfig(
        enabled=enabled,
        app_id="test-app-id",
        private_key="test-key",
        alipay_public_key="test-pub-key",
        gateway="https://openapi-sandbox.dl.alipaydev.com/gateway.do",
        seller_id="2088test",
        service_id="test-service",
        amount="9.90",
        goods_name="Liki 命理服务",
        seller_name="Liki",
        paid_tools=frozenset({"subject_candidates"}),
    )


def _gateway_config(aipay: AipayConfig) -> GatewayConfig:
    return GatewayConfig(
        server_name="liki-mcp",
        server_title="LIKI product MCP",
        server_version="0.1.0",
        engine=UpstreamConfig(name="engine", url="http://engine:8081/mcp", timeout_seconds=35),
        counsel=UpstreamConfig(name="counsel", url="http://counsel:8086/mcp", timeout_seconds=35),
        exposed_tools={},
        forbidden_tools=frozenset(),
        aipay=aipay,
    )


class TestOrderStore:
    def test_create_and_find(self, tmp_path):
        store = OrderStore(str(tmp_path / "test.db"))
        store.create_pending("ORDER_1", "subject_candidates")
        order = store.find_by_out_trade_no("ORDER_1")
        assert order is not None
        assert order.out_trade_no == "ORDER_1"
        assert order.tool_name == "subject_candidates"

    def test_find_nonexistent(self, tmp_path):
        store = OrderStore(str(tmp_path / "test.db"))
        assert store.find_by_out_trade_no("NOPE") is None

    def test_fulfillment_idempotent(self, tmp_path):
        store = OrderStore(str(tmp_path / "test.db"))
        store.create_pending("ORDER_2", "subject_candidates")
        store.prepare_fulfillment("ORDER_2")  # transitions PENDING_PAYMENT → PAID
        result = store.prepare_fulfillment("ORDER_2")
        assert result.state in ("PENDING_CONFIRM", "FULFILLED")
        store.mark_fulfilled("ORDER_2", "TRADE_1")
        result2 = store.prepare_fulfillment("ORDER_2")
        assert result2.state == "FULFILLED"

    def test_unique_constraint(self, tmp_path):
        store = OrderStore(str(tmp_path / "test.db"))
        store.create_pending("DUP", "subject_candidates")
        store.create_pending("DUP", "subject_candidates")  # INSERT OR IGNORE
        order = store.find_by_out_trade_no("DUP")
        assert order is not None


class TestPaymentService:
    @pytest.fixture
    def service(self, tmp_path):
        config = _aipay_config()
        store = OrderStore(str(tmp_path / "test.db"))
        return PaymentService(config, store)

    @pytest.fixture
    def disabled_service(self, tmp_path):
        config = _aipay_config(enabled=False)
        store = OrderStore(str(tmp_path / "test.db"))
        return PaymentService(config, store)

    def test_free_tool_returns_none(self, service):
        result = service.enforce_paywall("time_now", None)
        assert result is None

    def test_disabled_paywall_returns_none(self, disabled_service):
        result = disabled_service.enforce_paywall("subject_candidates", None)
        assert result is None

    def test_paid_tool_no_proof_returns_402(self, service):
        result = service.enforce_paywall("subject_candidates", None)
        assert result is not None
        assert result.status_code == 402
        assert "Payment-Needed" in result.headers
        assert result.body["code"] == "PAYMENT_REQUIRED"

    def test_paid_tool_with_proof_verifies(self, service):
        mock_client = MagicMock()
        mock_client.execute.return_value = json.dumps({
            "alipay_aipay_agent_payment_verify_response": {
                "code": "10000",
                "trade_no": "T123",
                "out_trade_no": "ORDER_1",
            }
        })
        service._client = mock_client

        with patch.object(service, "_store") as mock_store, \
             patch.object(service, "confirm_fulfillment", return_value=True):
            from app.order_store import FulfillmentResult
            mock_store.prepare_fulfillment.return_value = FulfillmentResult(state="PENDING_CONFIRM")
            mock_store.mark_fulfilled.return_value = None

            result = service.enforce_paywall("subject_candidates", "proof-token")
            assert result is None  # Payment passed, should continue to upstream

    def test_402_body_structure(self, service):
        result = service.build_402_response("subject_candidates")
        assert result.status_code == 402
        assert "Payment-Needed" in result.headers
        body = result.body
        assert body["code"] == "PAYMENT_REQUIRED"
        assert "out_trade_no" in body
        assert body["amount"] == "9.90"
