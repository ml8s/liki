"""Payment order aggregate and lifecycle invariants."""

from __future__ import annotations

from dataclasses import dataclass, fields
from enum import StrEnum
from typing import Any


class OrderState(StrEnum):
    PENDING_PAYMENT = "PENDING_PAYMENT"
    PAID = "PAID"
    FULFILLED = "FULFILLED"


class FulfillmentState(StrEnum):
    UNFULFILLED = "UNFULFILLED"
    PENDING_CONFIRM = "PENDING_CONFIRM"
    FULFILLED = "FULFILLED"


class OrderTransitionError(ValueError):
    """The current order lifecycle does not allow the requested transition."""


@dataclass(slots=True)
class PaymentOrder:
    """支付订单聚合根；拥有生命周期不变量，仓库只负责持久化。"""

    out_trade_no: str
    trade_no: str | None
    state: OrderState
    tool_name: str
    amount: str
    currency: str
    goods_name: str
    pay_before: str
    fulfillment_state: FulfillmentState

    @classmethod
    def from_snapshot(cls, order: Any) -> PaymentOrder:
        snapshot_fields = fields(order)
        return cls(**{field.name: getattr(order, field.name) for field in snapshot_fields})

    def mark_paid(self) -> None:
        """Mark a verified pending payment; repeated verification is idempotent."""
        if self.state == OrderState.PAID:
            if self.fulfillment_state != FulfillmentState.UNFULFILLED:
                raise OrderTransitionError("已支付订单不允许重复支付")
            return
        if self.state != OrderState.PENDING_PAYMENT or self.fulfillment_state != FulfillmentState.UNFULFILLED:
            raise OrderTransitionError("订单状态不允许支付")
        self.state = OrderState.PAID

    def reserve_fulfillment(self) -> None:
        """Reserve one usable execution; repeated reservation stays idempotent."""
        if self.fulfillment_state == FulfillmentState.PENDING_CONFIRM:
            return
        if self.state == OrderState.PAID and self.fulfillment_state == FulfillmentState.UNFULFILLED:
            self.fulfillment_state = FulfillmentState.PENDING_CONFIRM
            return
        raise OrderTransitionError("订单状态不允许履约")

    def mark_fulfilled(self, trade_no: str) -> None:
        if self.fulfillment_state != FulfillmentState.PENDING_CONFIRM:
            raise OrderTransitionError("订单尚未进入履约状态")
        self.trade_no = trade_no
        self.state = OrderState.FULFILLED
        self.fulfillment_state = FulfillmentState.FULFILLED
