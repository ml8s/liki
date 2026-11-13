"""Paid-call authorization invariants."""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from app.config import normalize_amount


def amounts_equal(left: Any, right: Any) -> bool:
    normalized_left = normalize_amount(str(left or ""))
    normalized_right = normalize_amount(str(right or ""))
    return normalized_left is not None and normalized_left == normalized_right


def is_future(value: str) -> bool:
    try:
        return datetime.fromisoformat(value) > datetime.now(timezone.utc)
    except (TypeError, ValueError):
        return False


def order_usable(order: Any) -> bool:
    if order.currency != "CNY":
        return False
    state = order.state
    fulfillment = order.fulfillment_state
    if state == "PENDING_PAYMENT":
        return fulfillment == "UNFULFILLED" and is_future(order.pay_before)
    if state == "PAID":
        return fulfillment in {"UNFULFILLED", "PENDING_CONFIRM"}
    return state == "FULFILLED" and fulfillment == "FULFILLED"
