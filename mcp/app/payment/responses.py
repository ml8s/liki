"""HTTP-facing value objects for the paywall flow."""

from __future__ import annotations

import base64
import json
from dataclasses import dataclass
from typing import Any


def b64url_encode(value: str) -> str:
    return base64.urlsafe_b64encode(value.encode("utf-8")).rstrip(b"=").decode("ascii")


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
        return b64url_encode(
            json.dumps(
                {
                    "trade_no": self.trade_no,
                    "out_trade_no": self.out_trade_no,
                    "resource_id": self.resource_id,
                    "validated": True,
                },
                ensure_ascii=False,
                separators=(",", ":"),
            )
        )


@dataclass(frozen=True)
class PaywallResponse:
    status_code: int
    headers: dict[str, str]
    body: dict[str, Any]

    @classmethod
    def processing_error(cls, code: str, message: str) -> PaywallResponse:
        return cls(
            status_code=502,
            headers={"Content-Type": "application/json"},
            body={"code": code, "message": message},
        )
