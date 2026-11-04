"""SkillPay A2M 按量付费领域服务。

公共 API：PaymentService, PaywallResponse, VerifyResult。
支付宝 SDK 类在方法内懒加载；未安装时 import 不报错。

协议严格按官方 A2MPaymentDemo.md：
1. 无 Payment-Proof → 建单返回 402 + Payment-Needed
2. 有 Payment-Proof → SDK 验付 → 履约确认 → 返回结果
签名/验签由 alipay-sdk-python DefaultAlipayClient 自动处理。
"""

from __future__ import annotations

import base64
import json
import time
import uuid
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any

# alipay SDK 类在 _init_client / verify_proof / confirm_fulfillment 中懒加载：
# SDK 未安装时 import 不报错，仅 paywall enabled 时才需要。

from app.config import AipayConfig
from app.order_store import OrderStore


@dataclass(frozen=True)
class VerifyResult:
    ok: bool
    trade_no: str | None
    out_trade_no: str | None
    error: str | None = None


@dataclass(frozen=True)
class PaywallResponse:
    """HTTP 响应（402 或 502），gateway 直接返回给调用方。"""
    status_code: int
    headers: dict[str, str]
    body: dict[str, Any]

    @classmethod
    def payment_required(cls, body: dict[str, Any]) -> "PaywallResponse":
        return cls(status_code=402, headers={"Content-Type": "application/json"}, body=body)

    @classmethod
    def processing_error(cls, code: str, message: str) -> "PaywallResponse":
        return cls(status_code=502, headers={"Content-Type": "application/json"}, body={"code": code, "message": message})


class PaymentService:
    """A2M 支付领域服务。委托 alipay-sdk-python 处理签名和通信。"""

    def __init__(self, config: AipayConfig, store: OrderStore) -> None:
        self._config = config
        self._store = store
        self._client = None  # lazy: SDK 加载延迟到首次调用

    def _get_client(self):
        if self._client is None:
            self._client = self._init_client()
        return self._client

    def _init_client(self):
        from alipay.aop.api.AlipayClientConfig import AlipayClientConfig
        from alipay.aop.api.DefaultAlipayClient import DefaultAlipayClient

        client_config = AlipayClientConfig()
        client_config.app_id = self._config.app_id
        client_config.app_private_key = self._config.private_key
        client_config.alipay_public_key = self._config.alipay_public_key
        client_config.server_url = self._config.gateway
        client_config.sign_type = "RSA2"
        return DefaultAlipayClient(client_config)

    def build_402_response(self, tool_name: str) -> PaywallResponse:
        """无凭证时：建单并返回 402 + Payment-Needed 响应。"""
        out_trade_no = (
            f"LIKI_{int(time.time() * 1000)}_{uuid.uuid4().hex[:12]}"
        )
        amount = self._config.amount
        goods_name = self._config.goods_name
        pay_before = (
            datetime.now(timezone.utc) + timedelta(minutes=30)
        ).isoformat()

        self._store.create_pending(out_trade_no, tool_name)

        payment_needed = base64.urlsafe_b64encode(json.dumps({
            "app_id": self._config.app_id,
            "amount": amount,
            "currency": "CNY",
            "goods_name": goods_name,
            "out_trade_no": out_trade_no,
            "seller_id": self._config.seller_id,
            "service_id": self._config.service_id,
            "pay_before": pay_before,
        }, ensure_ascii=False).encode("utf-8")).decode("utf-8")

        body = {
            "code": "PAYMENT_REQUIRED",
            "message": f"工具 {tool_name} 需要付费使用",
            "out_trade_no": out_trade_no,
            "amount": amount,
            "currency": "CNY",
            "goods_name": goods_name,
        }

        response = PaywallResponse.payment_required(body)
        return PaywallResponse(
            status_code=response.status_code,
            headers={**response.headers, "Payment-Needed": payment_needed},
            body=response.body,
        )

    def verify_proof(self, proof: str) -> VerifyResult:
        """有凭证时：用 SDK 验证 Payment-Proof。"""
        from alipay.aop.api.domain.AlipayAipayAgentPaymentVerifyModel import (
            AlipayAipayAgentPaymentVerifyModel,
        )
        from alipay.aop.api.request.AlipayAipayAgentPaymentVerifyRequest import (
            AlipayAipayAgentPaymentVerifyRequest,
        )

        try:
            request = AlipayAipayAgentPaymentVerifyRequest()
            model = AlipayAipayAgentPaymentVerifyModel()
            model.payment_proof = proof
            request.biz_model = model

            response_content = self._get_client().execute(request)
            response = json.loads(response_content)
            data = response.get(
                "alipay_aipay_agent_payment_verify_response", response
            )

            if data.get("code") != "10000":
                return VerifyResult(
                    ok=False, trade_no=None, out_trade_no=None,
                    error=f"verify failed: {data.get('sub_msg', data.get('code'))}",
                )

            trade_no = data.get("trade_no", "")
            out_trade_no = data.get("out_trade_no", "")
            return VerifyResult(
                ok=True, trade_no=trade_no, out_trade_no=out_trade_no,
            )
        except Exception as exc:
            return VerifyResult(
                ok=False, trade_no=None, out_trade_no=None,
                error=f"verify exception: {exc}",
            )

    def confirm_fulfillment(self, trade_no: str) -> bool:
        """履约确认：通知支付宝服务已交付。"""
        if not trade_no:
            return False
        from alipay.aop.api.domain.AlipayAipayAgentFulfillmentConfirmModel import (
            AlipayAipayAgentFulfillmentConfirmModel,
        )
        from alipay.aop.api.request.AlipayAipayAgentFulfillmentConfirmRequest import (
            AlipayAipayAgentFulfillmentConfirmRequest,
        )

        try:
            request = AlipayAipayAgentFulfillmentConfirmRequest()
            model = AlipayAipayAgentFulfillmentConfirmModel()
            model.trade_no = trade_no
            request.biz_model = model

            response_content = self._get_client().execute(request)
            response = json.loads(response_content)
            data = response.get(
                "alipay_aipay_agent_fulfillment_confirm_response", response
            )
            return data.get("code") == "10000"
        except Exception:
            return False

    def enforce_paywall(
        self, tool_name: str, proof: str | None
    ) -> PaywallResponse | None:
        """入口：检查付费工具的支付状态。

        返回 PaywallResponse 表示需要支付（调用方应返回 402）。
        返回 None 表示支付通过（或工具免费），可继续透传 upstream。
        """
        if not self._config.enabled:
            return None
        if tool_name not in self._config.paid_tools:
            return None

        if not proof or not proof.strip():
            return self.build_402_response(tool_name)

        verify = self.verify_proof(proof)
        if not verify.ok:
            return PaywallResponse.processing_error("VERIFY_FAILED", verify.error or "unknown")

        fulfillment = self._store.prepare_fulfillment(verify.out_trade_no or "")

        if fulfillment.state == "FULFILLED":
            return None  # 幂等：已履约，直接透传

        if not self.confirm_fulfillment(verify.trade_no or ""):
            return PaywallResponse.processing_error("CONFIRM_FAILED", "履约确认失败，请重试")

        self._store.mark_fulfilled(
            verify.out_trade_no or "", verify.trade_no or ""
        )
        return None  # 支付+履约完成，透传 upstream
