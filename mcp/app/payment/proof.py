"""Payment-Proof decoding and provider verification results."""

from __future__ import annotations

import base64
import json
from dataclasses import dataclass
from typing import Any


def b64url_decode(value: str) -> str:
    compact = value.strip()
    padded = compact + ("=" * ((4 - len(compact) % 4) % 4))
    return base64.urlsafe_b64decode(padded.encode("ascii")).decode("utf-8")


def response_data(response: dict[str, Any], key: str) -> dict[str, Any]:
    nested = response.get(key)
    return nested if isinstance(nested, dict) else response


def returned(response: dict[str, Any], snake: str, camel: str) -> str:
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


def decode_proof(proof: str) -> ProofComponents | None:
    try:
        decoded = json.loads(b64url_decode(proof))
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
