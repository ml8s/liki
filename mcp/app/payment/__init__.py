"""AI 按量付费 A2M 支付领域。"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .authorization import amounts_equal, is_future, order_usable
    from .proof import ProofComponents, VerifyResult, decode_proof
    from .responses import Authorization, PaymentValidation, PaywallResponse
    from .service import PaymentService

__all__ = [
    "Authorization",
    "PaymentService",
    "PaymentValidation",
    "PaywallResponse",
    "ProofComponents",
    "VerifyResult",
    "amounts_equal",
    "decode_proof",
    "is_future",
    "order_usable",
]

_LAZY_IMPORTS = {
    "Authorization": (".responses", "Authorization"),
    "PaymentService": (".service", "PaymentService"),
    "PaymentValidation": (".responses", "PaymentValidation"),
    "PaywallResponse": (".responses", "PaywallResponse"),
    "ProofComponents": (".proof", "ProofComponents"),
    "VerifyResult": (".proof", "VerifyResult"),
    "amounts_equal": (".authorization", "amounts_equal"),
    "decode_proof": (".proof", "decode_proof"),
    "is_future": (".authorization", "is_future"),
    "order_usable": (".authorization", "order_usable"),
}


def __getattr__(name: str):
    if name not in _LAZY_IMPORTS:
        raise AttributeError(name)
    from importlib import import_module

    module_name, attribute = _LAZY_IMPORTS[name]
    return getattr(import_module(module_name, __name__), attribute)
