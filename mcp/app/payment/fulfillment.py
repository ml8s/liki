"""Fulfillment decision mapping for the payment service."""

from __future__ import annotations

from app.order_store import FulfillmentResult, Order
from app.payment.responses import Authorization, PaywallResponse


def map_reservation(
    reservation: FulfillmentResult,
    authorization: Authorization,
    build_402,
) -> PaywallResponse | None:
    if reservation.state == "FULFILLED":
        return None
    if reservation.state == "PENDING_CONFIRM":
        return None
    if reservation.state == "NOT_FOUND":
        return build_402(authorization.resource_id)
    return PaywallResponse.processing_error("FULFILLMENT_ERROR", "订单状态不允许履约")


def completable(order: Order) -> bool:
    return (order.state == "PAID" and order.fulfillment_state == "PENDING_CONFIRM") or (
        order.state == "FULFILLED" and order.fulfillment_state == "FULFILLED"
    )


__all__ = [
    "completable",
    "map_reservation",
]
