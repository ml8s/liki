"""Pure payment-order aggregate lifecycle contracts."""

import pytest
from app.payment.order import FulfillmentState, OrderState, OrderTransitionError, PaymentOrder, PendingOrder


def _pending() -> PaymentOrder:
    pending = PendingOrder(
        out_trade_no="ORDER_1",
        tool_name="subject_candidates",
        amount="0.01",
        currency="CNY",
        goods_name="Liki",
        pay_before="2999-01-01T00:00:00+00:00",
    )
    return PaymentOrder(
        out_trade_no=pending.out_trade_no,
        trade_no=None,
        state=OrderState.PENDING_PAYMENT,
        tool_name=pending.tool_name,
        amount=pending.amount,
        currency=pending.currency,
        goods_name=pending.goods_name,
        pay_before=pending.pay_before,
        fulfillment_state=FulfillmentState.UNFULFILLED,
    )


def test_pending_order_can_be_marked_paid():
    order = _pending()
    order.mark_paid()
    assert order.state == OrderState.PAID
    assert order.fulfillment_state == FulfillmentState.UNFULFILLED


def test_paid_unfulfilled_order_is_idempotent_for_payment():
    order = _pending()
    order.mark_paid()
    order.mark_paid()
    assert order.state == OrderState.PAID


def test_paid_pending_confirm_order_rejects_repeat_payment():
    order = _pending()
    order.mark_paid()
    order.reserve_fulfillment()
    with pytest.raises(OrderTransitionError, match="重复支付"):
        order.mark_paid()


def test_pending_order_cannot_reserve_fulfillment():
    order = _pending()
    with pytest.raises(OrderTransitionError, match="订单状态不允许履约"):
        order.reserve_fulfillment()


def test_paid_order_reserves_fulfillment():
    order = _pending()
    order.mark_paid()
    order.reserve_fulfillment()
    assert order.state == OrderState.PAID
    assert order.fulfillment_state == FulfillmentState.PENDING_CONFIRM


def test_pending_confirm_order_reserves_idempotently():
    order = _pending()
    order.mark_paid()
    order.reserve_fulfillment()
    order.reserve_fulfillment()
    assert order.fulfillment_state == FulfillmentState.PENDING_CONFIRM


def test_pending_confirm_order_can_be_fulfilled():
    order = _pending()
    order.mark_paid()
    order.reserve_fulfillment()
    order.mark_fulfilled("T123")
    assert order.trade_no == "T123"
    assert order.state == OrderState.FULFILLED
    assert order.fulfillment_state == FulfillmentState.FULFILLED


def test_fulfilled_order_rejects_different_trade_no():
    order = _pending()
    order.mark_paid()
    order.reserve_fulfillment()
    order.mark_fulfilled("T123")
    with pytest.raises(OrderTransitionError, match="订单尚未进入履约状态"):
        order.mark_fulfilled("T456")
