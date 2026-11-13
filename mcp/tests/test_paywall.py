"""AI 按量付费 A2M paywall tests with mocked Alipay SDK calls."""

from __future__ import annotations

import asyncio
import base64
import json
from contextlib import asynccontextmanager
from unittest.mock import MagicMock, patch

import pytest
from app.config import AipayConfig, GatewayConfig, UpstreamConfig
from app.gateway import create_app
from app.order_store import OrderStore, PendingOrder
from app.payment import PaymentService
from httpx2 import ASGITransport, AsyncClient
from mcp.types import CallToolResult


def _aipay_config(enabled: bool = True) -> AipayConfig:
    return AipayConfig(
        enabled=enabled,
        app_id="test-app-id",
        private_key="test-key",
        alipay_public_key="test-pub-key",
        gateway="https://openapi-sandbox.dl.alipaydev.com/gateway.do",
        seller_id="2088test",
        service_id="api_mock_service_id",
        amount="0.01",
        goods_name="Liki 命理服务",
        seller_name="Liki",
        paid_tools=frozenset({"subject_candidates"}),
        db_path="/data/aipay.db",
    )


def _gateway_config(aipay: AipayConfig) -> GatewayConfig:
    return GatewayConfig(
        server_name="liki-mcp",
        server_title="LIKI product MCP",
        server_version="0.1.0",
        engine=UpstreamConfig(
            name="engine",
            url="http://engine:8081/mcp",
            timeout_seconds=35,
        ),
        counsel=UpstreamConfig(
            name="counsel",
            url="http://counsel:8086/mcp",
            timeout_seconds=35,
        ),
        exposed_tools={
            "time_now": "engine",
            "subject_candidates": "engine",
        },
        forbidden_tools=frozenset(),
        aipay=aipay,
    )


def _pending(service: PaymentService) -> str:
    with patch("app.payment.service._seller_signature", return_value="seller-signature"):
        response = service.build_402_response("subject_candidates")
    assert response.status_code == 402
    encoded = response.headers["Payment-Needed"]
    bill = json.loads(base64.urlsafe_b64decode(encoded + "=="))
    assert set(bill) == {"protocol", "method"}
    assert bill["protocol"]["resource_id"] == "subject_candidates"
    assert bill["method"]["service_id"] == "api_mock_service_id"
    assert response.body["code"] == "Payment-Needed"
    return encoded


def _proof(out_trade_no: str, trade_no: str = "T123") -> str:
    value = {
        "protocol": {
            "payment_proof": "inner-proof",
            "trade_no": trade_no,
        },
        "method": {
            "client_session": "session",
        },
    }
    raw = base64.urlsafe_b64encode(json.dumps(value).encode()).rstrip(b"=")
    return raw.decode()


def _verify_response(
    *,
    out_trade_no: str,
    trade_no: str = "T123",
    amount: str = "0.01",
    resource_id: str = "subject_candidates",
    active: bool = True,
) -> str:
    return json.dumps(
        {
            "alipay_aipay_agent_payment_verify_response": {
                "code": "10000",
                "active": active,
                "trade_no": trade_no,
                "out_trade_no": out_trade_no,
                "amount": amount,
                "resource_id": resource_id,
            }
        }
    )


class TestOrderStore:
    def test_create_and_find_full_order_snapshot(self, tmp_path):
        service = PaymentService(
            _aipay_config(),
            OrderStore(str(tmp_path / "aipay.db")),
        )
        _pending(service)
        store = service._store
        row = store._get_conn().execute("SELECT out_trade_no FROM aipay_orders").fetchone()
        order = store.find_by_out_trade_no(row[0])
        assert order is not None
        assert order.amount == "0.01"
        assert order.currency == "CNY"
        assert order.fulfillment_state == "UNFULFILLED"

    def test_fulfillment_is_atomic_and_idempotent(self, tmp_path):
        store = OrderStore(str(tmp_path / "test.db"))
        store.create_pending(
            PendingOrder(
                out_trade_no="ORDER_1",
                tool_name="subject_candidates",
                amount="0.01",
                currency="CNY",
                goods_name="Liki",
                pay_before="2999-01-01T00:00:00+00:00",
            )
        )
        store.mark_paid("ORDER_1")
        first = store.reserve_fulfillment("ORDER_1")
        assert first.state == "PENDING_CONFIRM"
        assert first.order is not None
        assert first.order.state == "PAID"
        assert first.order.fulfillment_state == "PENDING_CONFIRM"
        again = store.reserve_fulfillment("ORDER_1")
        assert again.state == "PENDING_CONFIRM"
        store.mark_fulfilled("ORDER_1", "T123")
        final = store.reserve_fulfillment("ORDER_1")
        assert final.state == "FULFILLED"

    def test_pending_order_cannot_reserve_before_payment(self, tmp_path):
        store = OrderStore(str(tmp_path / "pending.db"))
        store.create_pending(
            PendingOrder(
                out_trade_no="ORDER_1",
                tool_name="subject_candidates",
                amount="0.01",
                currency="CNY",
                goods_name="Liki",
                pay_before="2999-01-01T00:00:00+00:00",
            )
        )
        result = store.reserve_fulfillment("ORDER_1")
        assert result.state == "INVALID_STATE"
        assert result.order is not None
        assert result.order.state == "PENDING_PAYMENT"

    def test_payment_and_fulfillment_marks_are_guarded_and_idempotent(self, tmp_path):
        store = OrderStore(str(tmp_path / "guarded.db"))
        store.create_pending(
            PendingOrder(
                out_trade_no="ORDER_1",
                tool_name="subject_candidates",
                amount="0.01",
                currency="CNY",
                goods_name="Liki",
                pay_before="2999-01-01T00:00:00+00:00",
            )
        )
        store.mark_paid("ORDER_1")
        store.mark_paid("ORDER_1")
        store.reserve_fulfillment("ORDER_1")
        with pytest.raises(ValueError, match="重复支付"):
            store.mark_paid("ORDER_1")

        store.mark_fulfilled("ORDER_1", "T123")
        store.mark_fulfilled("ORDER_1", "T123")
        with pytest.raises(ValueError, match="trade_no 不一致"):
            store.mark_fulfilled("ORDER_1", "T456")


class TestPaymentService:
    @pytest.fixture
    def service(self, tmp_path):
        return PaymentService(
            _aipay_config(),
            OrderStore(str(tmp_path / "test.db")),
        )

    def test_disabled_paywall_allows_paid_tool(self, tmp_path):
        service = PaymentService(
            _aipay_config(enabled=False),
            OrderStore(str(tmp_path / "test.db")),
        )
        assert service.authorize_paid_call("subject_candidates", None) is None

    def test_paid_tool_no_proof_returns_official_402(self, service):
        with patch("app.payment.service._seller_signature", return_value="seller-signature"):
            response = service.authorize_paid_call("subject_candidates", None)
        assert response is not None
        assert response.status_code == 402
        encoded = response.headers["Payment-Needed"]
        bill = json.loads(base64.urlsafe_b64decode(encoded + "=="))
        required_protocol = {
            "out_trade_no",
            "amount",
            "currency",
            "resource_id",
            "pay_before",
            "seller_signature",
            "seller_sign_type",
            "seller_unique_id",
        }
        required_method = {
            "seller_name",
            "seller_id",
            "seller_app_id",
            "goods_name",
            "seller_unique_id_key",
            "service_id",
        }
        assert set(bill["protocol"]) == required_protocol
        assert set(bill["method"]) == required_method
        assert bill["method"]["service_id"] == "api_mock_service_id"

    def test_verify_and_fulfillment_strictly_bind_local_order(self, service):
        _pending(service)
        order = service._store._get_conn().execute("SELECT out_trade_no FROM aipay_orders").fetchone()[0]
        service._client = MagicMock()
        service._client.execute.return_value = _verify_response(
            out_trade_no=order,
        )
        proof = _proof(order)

        authorization = service.authorize_paid_call("subject_candidates", proof)
        assert authorization.out_trade_no == order
        assert authorization.resource_id == "subject_candidates"

        assert service.reserve_fulfillment(authorization) is None
        with patch.object(service, "confirm_fulfillment", return_value=True):
            validation = service.complete_fulfillment(authorization)
        assert validation.header()
        final = service._store.find_by_out_trade_no(order)
        assert final is not None
        assert final.fulfillment_state == "FULFILLED"

    def test_mismatched_amount_is_rejected_with_402(self, service):
        bill = _pending(service)
        decoded = json.loads(base64.urlsafe_b64decode(bill + "=="))
        order = decoded["protocol"]["out_trade_no"]
        service._client = MagicMock()
        service._client.execute.return_value = _verify_response(
            out_trade_no=order,
            amount="9.90",
        )
        with patch("app.payment.service._seller_signature", return_value="seller-signature"):
            response = service.authorize_paid_call(
                "subject_candidates",
                _proof(order),
            )
        assert response is not None
        assert response.status_code == 402


class TestGatewayPaywall:
    def test_paid_call_retries_with_proof_and_payment_validation(
        self,
        tmp_path,
        monkeypatch,
    ):
        config = _gateway_config(_aipay_config())
        service = PaymentService(
            config.aipay,
            OrderStore(str(tmp_path / "gateway.db")),
        )
        _pending(service)
        order = service._store._get_conn().execute("SELECT out_trade_no FROM aipay_orders").fetchone()[0]
        service._client = MagicMock()
        service._client.execute.return_value = _verify_response(
            out_trade_no=order,
        )
        proof = _proof(order)

        import app.gateway as gateway_module
        import app.tool_router as tool_router_module

        monkeypatch.setattr(gateway_module, "load_config", lambda: config)
        monkeypatch.setattr(gateway_module, "_payment_service", service)
        monkeypatch.setattr(gateway_module, "_payment_config", config.aipay)

        class FakeSession:
            async def call_tool(self, name, arguments=None):
                return CallToolResult(
                    content=[],
                    structuredContent={"candidates": []},
                )

        @asynccontextmanager
        async def fake_upstream_session(upstream_config, **kwargs):
            yield FakeSession()

        monkeypatch.setattr(tool_router_module, "upstream_session", fake_upstream_session)
        with patch.object(service, "confirm_fulfillment", return_value=True):

            async def run():
                app = create_app()
                async with AsyncClient(
                    transport=ASGITransport(app=app),
                    base_url="http://test",
                ) as client:
                    return await client.post(
                        "/mcp",
                        json={
                            "jsonrpc": "2.0",
                            "id": 1,
                            "method": "tools/call",
                            "params": {
                                "name": "subject_candidates",
                                "arguments": {},
                            },
                        },
                        headers={"Payment-Proof": proof},
                    )

            response = asyncio.run(run())

        assert response.status_code == 200
        assert "Payment-Validation" in response.headers
        assert response.json()["result"]["structuredContent"] == {"candidates": []}
        assert response.json()["result"]["content"] == []
        assert "mcpContent" not in response.json()["result"]
        assert response.json()["resource_id"] == "subject_candidates"
        evidence = response.json()["content"]
        assert evidence["resource_id"] == "subject_candidates"
        assert evidence["content"] == {"candidates": []}
        final = service._store.find_by_out_trade_no(order)
        assert final is not None
        assert final.fulfillment_state == "FULFILLED"

    def test_batch_with_paid_tool_is_fail_closed(self, monkeypatch):
        config = _gateway_config(_aipay_config())
        import app.gateway as gateway_module

        monkeypatch.setattr(gateway_module, "load_config", lambda: config)
        monkeypatch.setattr(gateway_module, "_payment_service", None)
        monkeypatch.setattr(gateway_module, "_payment_config", config.aipay)

        async def run():
            app = create_app()
            async with AsyncClient(
                transport=ASGITransport(app=app),
                base_url="http://test",
            ) as client:
                return await client.post(
                    "/mcp",
                    json=[
                        {
                            "jsonrpc": "2.0",
                            "id": 1,
                            "method": "tools/call",
                            "params": {"name": "subject_candidates", "arguments": {}},
                        }
                    ],
                )

        response = asyncio.run(run())
        assert response.status_code == 400
