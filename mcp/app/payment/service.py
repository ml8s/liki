"""A2M payment flow orchestration."""

from __future__ import annotations

import json
import time
import uuid
from datetime import datetime, timedelta, timezone
from typing import Any

import app.payment.alipay_client as alipay_client
from app.config import AipayConfig, normalize_amount
from app.order_store import OrderStore, PendingOrder
from app.payment.alipay_client import create_client
from app.payment.alipay_client import seller_signature as sdk_seller_signature
from app.payment.authorization import amounts_equal, order_usable
from app.payment.fulfillment import completable, map_reservation
from app.payment.proof import ProofComponents, VerifyResult, decode_proof
from app.payment.responses import Authorization, PaymentValidation, PaywallResponse, b64url_encode


class PaymentService:
    """A2M payment flow backed by the official Alipay Python SDK."""

    def __init__(self, config: AipayConfig, store: OrderStore) -> None:
        self._config = config
        self._store = store
        self._client: Any | None = None

    def _get_client(self) -> Any:
        if self._client is None:
            self._client = create_client(self._config)
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
            pay_before = (datetime.now(timezone.utc).astimezone() + timedelta(minutes=30)).isoformat()

            signature = _seller_signature(
                {
                    "amount": amount,
                    "currency": currency,
                    "goods_name": goods_name,
                    "out_trade_no": out_trade_no,
                    "pay_before": pay_before,
                    "resource_id": resource_id,
                    "seller_id": self._config.seller_id,
                    "service_id": self._config.service_id,
                },
                self._config.private_key,
            )

            self._store.create_pending(
                PendingOrder(
                    out_trade_no=out_trade_no,
                    tool_name=tool_name,
                    amount=amount,
                    currency=currency,
                    goods_name=goods_name,
                    pay_before=pay_before,
                )
            )

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
            encoded_bill = b64url_encode(
                json.dumps(
                    payment_needed,
                    ensure_ascii=False,
                    separators=(",", ":"),
                )
            )
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
        return decode_proof(proof)

    def _verify_proof(self, components: ProofComponents) -> VerifyResult:
        return alipay_client.verify_proof(self._get_client(), components)

    def confirm_fulfillment(self, trade_no: str) -> bool:
        """Confirm fulfillment through the official Alipay API."""
        if not trade_no:
            return False
        return alipay_client.confirm_fulfillment(self._get_client(), trade_no)

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

        if (
            not order_usable(order)
            or not amounts_equal(order.amount, verify_amount)
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
        return map_reservation(
            reservation,
            authorization,
            self.build_402_response,
        )

    def complete_fulfillment(
        self,
        authorization: Authorization,
    ) -> PaywallResponse | PaymentValidation:
        order = self._store.find_by_out_trade_no(authorization.out_trade_no)
        if order is None or not completable(order):
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


def _seller_signature(params: dict[str, str], private_key: str) -> str:
    """Seam retained for tests; production calls delegate to the SDK adapter."""
    return sdk_seller_signature(params, private_key)
