"""Official Alipay SDK adapter; this module owns no payment policy."""

from __future__ import annotations

import json
from typing import Any

from app.config import AipayConfig
from app.payment.proof import ProofComponents, VerifyResult, response_data, returned


def seller_signature(params: dict[str, str], private_key: str) -> str:
    """Generate the A2M seller bill signature with the official SDK."""
    from alipay.aop.api.util.SignatureUtils import sign_with_rsa2

    sign_string = "&".join(f"{key}={params[key]}" for key in sorted(params) if params[key] != "")
    return sign_with_rsa2(private_key.strip(), sign_string, "utf-8")


def create_client(config: AipayConfig) -> Any:
    from alipay.aop.api.AlipayClientConfig import AlipayClientConfig
    from alipay.aop.api.DefaultAlipayClient import DefaultAlipayClient

    client_config = AlipayClientConfig()
    client_config.app_id = config.app_id
    client_config.app_private_key = config.private_key
    client_config.alipay_public_key = config.alipay_public_key
    client_config.server_url = config.gateway
    client_config.sign_type = "RSA2"
    return DefaultAlipayClient(alipay_client_config=client_config)


def verify_proof(client: Any, components: ProofComponents) -> VerifyResult:
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

        response = json.loads(client.execute(request))
        data = response_data(response, "alipay_aipay_agent_payment_verify_response")
        if data.get("code") != "10000":
            return VerifyResult(
                ok=False,
                error=f"verify failed: {data.get('sub_msg', data.get('code'))}",
            )
        if data.get("active") is not True:
            return VerifyResult(ok=False, error="payment proof is not active")

        return VerifyResult(
            ok=True,
            trade_no=returned(data, "trade_no", "tradeNo") or components.trade_no,
            out_trade_no=returned(data, "out_trade_no", "outTradeNo"),
            amount=returned(data, "amount", "amount"),
            resource_id=returned(data, "resource_id", "resourceId"),
        )
    except Exception as exc:
        return VerifyResult(ok=False, error=f"verify exception: {exc}", retryable=True)


def confirm_fulfillment(client: Any, trade_no: str) -> bool:
    """Confirm fulfillment through the official Alipay API."""
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
        response = json.loads(client.execute(request))
        data = response_data(response, "alipay_aipay_agent_fulfillment_confirm_response")
        return data.get("code") == "10000"
    except Exception:
        return False
