"""AI 按量付费 A2M 支付领域服务。

公开平台 API 和账单签名都委托官方 alipay-sdk-python。
本模块只负责协议适配与状态校验，不实现自己的 RSA/SHA256 密码学算法。
"""

from __future__ import annotations

import base64
import json
import time
import uuid
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any

from app.config import AipayConfig, normalize_amount
from app.order_store import OrderStore, PendingOrder


def _b64url_encode(value: str) -> str:
    return base64.urlsafe_b64encode(value.encode("utf-8")).rstrip(b"=").decode("ascii")


def _b64url_decode(value: str) -> str:
    compact = value.strip()
    padded = compact + ("=" * ((4 - len(compact) % 4) % 4))
    return base64.urlsafe_b64decode(padded.encode("ascii")).decode("utf-8")


def _amounts_equal(left: Any, right: Any) -> bool:
    normalized_left = normalize_amount(str(left or ""))
    normalized_right = normalize_amount(str(right or ""))
    return normalized_left is not None and normalized_left == normalized_right


def _is_future(value: str) -> bool:
    try:
        return datetime.fromisoformat(value) > datetime.now(timezone.utc)
    except (TypeError, ValueError):
        return False


def _seller_signature(params: dict[str, str], private_key: str) -> str:
    """Generate the A2M seller bill signature with the official SDK."""
    from alipay.aop.api.util.SignatureUtils import sign_with_rsa2

    sign_string = "&".join(
        f"{key}={params[key]}" for key in sorted(params) if params[key] != ""
    )
    return sign_with_rsa2(private_key.strip(), sign_string, "utf-8")


def _response_data(response: dict[str, Any], key: str) -> dict[str, Any]:
    nested = response.get(key)
    return nested if isinstance(nested, dict) else response


def _returned(response: dict[str, Any], snake: str, camel: str) -> str:
    value = response.get(snake, response.get(camel, ""))
    return str(value) if value is not None else ""


@dataclass(frozen=True)
class VerifyResult:
    ok: bool
    trade_no: str | None = None
    out_trade_no: str | None = None
    amount: str | None = None
    resource_id: str | None = None
    error: str | None = None
    retryable: bool = False


@dataclass(frozen=True)
class ProofComponents:
    payment_proof: str
    trade_no: str
    client_session: str | None = None


@dataclass(frozen=True)
class Authorization:
    trade_no: str
    out_trade_no: str
    resource_id: str


@dataclass(frozen=True)
class PaymentValidation:
    trade_no: str
    out_trade_no: str
    resource_id: str

    def header(self) -> str:
        return _b64url_encode(json.dumps({
            "trade_no": self.trade_no,
            "out_trade_no": self.out_trade_no,
            "resource_id": self.resource_id,
            "validated": True,
        }, ensure_ascii=False, separators=(",", ":")))


@dataclass(frozen=True)
class PaywallResponse:
    status_code: int
    headers: dict[str, str]
    body: dict[str, Any]

    @classmethod
    def processing_error(cls, code: str, message: str) -> "PaywallResponse":
        return cls(
            status_code=502,
            headers={"Content-Type": "application/json"},
            body={"code": code, "message": message},
        )


class PaymentService:
    """A2M payment flow backed by the official Alipay Python SDK."""

    def __init__(self, config: AipayConfig, store: OrderStore) -> None:
        self._config = config
        self._store = store
        self._client: Any | None = None

    def _get_client(self) -> Any:
        if self._client is None:
            from alipay.aop.api.AlipayClientConfig import AlipayClientConfig
            from alipay.aop.api.DefaultAlipayClient import DefaultAlipayClient

            client_config = AlipayClientConfig()
            client_config.app_id = self._config.app_id
            client_config.app_private_key = self._config.private_key
            client_config.alipay_public_key = self._config.alipay_public_key
            client_config.server_url = self._config.gateway
            client_config.sign_type = "RSA2"
            self._client = DefaultAlipayClient(alipay_client_config=client_config)
        return self._client

    def build_402_response(self, tool_name: str) -> PaywallResponse:
        """Create, persist, and encode a protocol-compliant Payment-Needed bill."""
        try:
            out_trade_no = f"LIKI_{int(time.time() * 1000)}_{uuid.uuid4().hex[:12]}"
            amount = self._config.amount
            if normalize_amount(amount) is None:
                raise ValueError("ALIPAY_AIPAY_AMOUNT must be a positive decimal amount")
            currency = "CNY"
            resource_id = tool_name
            goods_name = self._config.goods_name
            pay_before = (
                datetime.now(timezone.utc).astimezone() + timedelta(minutes=30)
            ).isoformat()

            signature = _seller_signature({
                "amount": amount,
                "currency": currency,
                "goods_name": goods_name,
                "out_trade_no": out_trade_no,
                "pay_before": pay_before,
                "resource_id": resource_id,
                "seller_id": self._config.seller_id,
                "service_id": self._config.service_id,
            }, self._config.private_key)

            self._store.create_pending(PendingOrder(
                out_trade_no=out_trade_no,
                tool_name=tool_name,
                amount=amount,
                currency=currency,
                goods_name=goods_name,
                pay_before=pay_before,
            ))

            payment_needed = {
                "protocol": {
                    "out_trade_no": out_trade_no,
                    "amount": amount,
                    "currency": currency,
                    "resource_id": resource_id,
                    "pay_before": pay_before,
                    "seller_signature": signature,
                    "seller_sign_type": "RSA2",
                    "seller_unique_id": self._config.seller_id,
                },
                "method": {
                    "seller_name": self._config.seller_name,
                    "seller_id": self._config.seller_id,
                    "seller_app_id": self._config.app_id,
                    "goods_name": goods_name,
                    "seller_unique_id_key": "seller_id",
                    "service_id": self._config.service_id,
                },
            }
            encoded_bill = _b64url_encode(json.dumps(
                payment_needed,
                ensure_ascii=False,
                separators=(",", ":"),
            ))
            body = {
                "code": "Payment-Needed",
                "message": f"工具 {tool_name} 需要付费使用",
                "out_trade_no": out_trade_no,
                "amount": amount,
                "currency": currency,
                "goods_name": goods_name,
            }
            return PaywallResponse(
                status_code=402,
                headers={
                    "Content-Type": "application/json",
                    "Payment-Needed": encoded_bill,
                },
                body=body,
            )
        except Exception as exc:
            return PaywallResponse.processing_error(
                "CREATE_ORDER_ERROR",
                f"创建支付订单失败: {exc}",
            )

    def _decode_proof(self, proof: str) -> ProofComponents | None:
        try:
            decoded = json.loads(_b64url_decode(proof))
            if not isinstance(decoded, dict):
                return None
            protocol = decoded.get("protocol")
            method = decoded.get("method")
            if not isinstance(protocol, dict):
                return None
            payment_proof = str(protocol.get("payment_proof") or "").strip()
            trade_no = str(protocol.get("trade_no") or "").strip()
            if not payment_proof or not trade_no:
                return None
            client_session = (
                str(method.get("client_session")).strip()
                if isinstance(method, dict) and method.get("client_session")
                else None
            )
            return ProofComponents(
                payment_proof=payment_proof,
                trade_no=trade_no,
                client_session=client_session,
            )
        except Exception:
            return None

    def _verify_proof(self, components: ProofComponents) -> VerifyResult:
        """Verify an A2M Payment-Proof with the official Alipay API."""
        try:
            from alipay.aop.api.domain.AlipayAipayAgentPaymentVerifyModel import (
                AlipayAipayAgentPaymentVerifyModel,
            )
            from alipay.aop.api.request.AlipayAipayAgentPaymentVerifyRequest import (
                AlipayAipayAgentPaymentVerifyRequest,
            )

            request = AlipayAipayAgentPaymentVerifyRequest()
            model = AlipayAipayAgentPaymentVerifyModel()
            model.payment_proof = components.payment_proof
            model.trade_no = components.trade_no
            if components.client_session:
                model.client_session = components.client_session
            request.biz_model = model

            response = json.loads(self._get_client().execute(request))
            data = _response_data(response, "alipay_aipay_agent_payment_verify_response")
            if data.get("code") != "10000":
                return VerifyResult(
                    ok=False,
                    error=f"verify failed: {data.get('sub_msg', data.get('code'))}",
                )
            if data.get("active") is not True:
                return VerifyResult(ok=False, error="payment proof is not active")

            trade_no = _returned(data, "trade_no", "tradeNo") or components.trade_no
            out_trade_no = _returned(data, "out_trade_no", "outTradeNo")
            amount = _returned(data, "amount", "amount")
            resource_id = _returned(data, "resource_id", "resourceId")

            return VerifyResult(
                ok=True,
                trade_no=trade_no,
                out_trade_no=out_trade_no,
                amount=amount,
                resource_id=resource_id,
            )
        except Exception as exc:
            return VerifyResult(
                ok=False,
                error=f"verify exception: {exc}",
                retryable=True,
            )

    def confirm_fulfillment(self, trade_no: str) -> bool:
        """Confirm fulfillment through the official Alipay API."""
        if not trade_no:
            return False

        try:
            from alipay.aop.api.domain.AlipayAipayAgentFulfillmentConfirmModel import (
                AlipayAipayAgentFulfillmentConfirmModel,
            )
            from alipay.aop.api.request.AlipayAipayAgentFulfillmentConfirmRequest import (
                AlipayAipayAgentFulfillmentConfirmRequest,
            )

            request = AlipayAipayAgentFulfillmentConfirmRequest()
            model = AlipayAipayAgentFulfillmentConfirmModel()
            model.trade_no = trade_no
            request.biz_model = model
            response = json.loads(self._get_client().execute(request))
            data = _response_data(
                response,
                "alipay_aipay_agent_fulfillment_confirm_response",
            )
            return data.get("code") == "10000"
        except Exception:
            return False

    def close(self) -> None:
        """Release the owned SQLite connection when cached config is replaced."""
        self._store.close()

    def authorize_paid_call(
        self,
        tool_name: str,
        proof: str | None,
    ) -> PaywallResponse | Authorization | None:
        """Return 402/502, a validated authorization, or None for a free call."""
        if not self._config.enabled or tool_name not in self._config.paid_tools:
            return None

        if not proof or not proof.strip():
            return self.build_402_response(tool_name)

        components = self._decode_proof(proof)
        if components is None:
            return self.build_402_response(tool_name)

        verification = self._verify_proof(components)
        if not verification.ok:
            if verification.retryable:
                return PaywallResponse.processing_error(
                    "VERIFY_FAILED",
                    verification.error or "unknown",
                )
            return self.build_402_response(tool_name)

        if not verification.trade_no or verification.trade_no != components.trade_no:
            return self.build_402_response(tool_name)

        order = self._store.find_by_out_trade_no(verification.out_trade_no or "")
        if order is None:
            return self.build_402_response(tool_name)

        verify_amount = verification.amount
        verify_resource = verification.resource_id
        if self._config.is_exact_sandbox:
            verify_amount = verify_amount or order.amount
            verify_resource = verify_resource or order.tool_name

        order_usable = (
            order.currency == "CNY"
            and (
                (
                    order.state == "PENDING_PAYMENT"
                    and order.fulfillment_state == "UNFULFILLED"
                    and _is_future(order.pay_before)
                )
                or (
                    order.state == "PAID"
                    and order.fulfillment_state == "UNFULFILLED"
                )
                or (
                    order.state == "PAID"
                    and order.fulfillment_state == "PENDING_CONFIRM"
                )
                or (
                    order.state == "FULFILLED"
                    and order.fulfillment_state == "FULFILLED"
                )
            )
        )
        if (
            not order_usable
            or not _amounts_equal(order.amount, verify_amount)
            or order.tool_name != tool_name
            or verify_resource != tool_name
        ):
            return self.build_402_response(tool_name)

        if order.state == "PENDING_PAYMENT":
            self._store.mark_paid(order.out_trade_no)

        return Authorization(
            trade_no=verification.trade_no,
            out_trade_no=order.out_trade_no,
            resource_id=order.tool_name,
        )

    def reserve_fulfillment(
        self,
        authorization: Authorization,
    ) -> PaywallResponse | None:
        reservation = self._store.reserve_fulfillment(authorization.out_trade_no)
        if reservation.state == "FULFILLED":
            return None
        if reservation.state == "PENDING_CONFIRM":
            return None
        if reservation.state == "NOT_FOUND":
            return self.build_402_response(authorization.resource_id)
        return PaywallResponse.processing_error(
            "FULFILLMENT_ERROR",
            "订单状态不允许履约",
        )

    def complete_fulfillment(
        self,
        authorization: Authorization,
    ) -> PaywallResponse | PaymentValidation:
        order = self._store.find_by_out_trade_no(authorization.out_trade_no)
        if order is None or not (
            (
                order.state == "PAID"
                and order.fulfillment_state == "PENDING_CONFIRM"
            )
            or (
                order.state == "FULFILLED"
                and order.fulfillment_state == "FULFILLED"
            )
        ):
            return PaywallResponse.processing_error(
                "FULFILLMENT_ERROR",
                "订单尚未进入履约状态",
            )
        if order.fulfillment_state == "PENDING_CONFIRM":
            if not self.confirm_fulfillment(authorization.trade_no):
                return PaywallResponse.processing_error(
                    "FULFILLMENT_CONFIRM_FAILED",
                    "履约确认失败，请使用同一 Payment-Proof 重试",
                )
            self._store.mark_fulfilled(authorization.out_trade_no, authorization.trade_no)
        return PaymentValidation(
            trade_no=authorization.trade_no,
            out_trade_no=authorization.out_trade_no,
            resource_id=authorization.resource_id,
        )
