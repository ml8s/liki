"""订单持久化仓库（SQLite）。

A2M 协议要求原子创建、唯一约束防重复核销、幂等履约标记。
SQLite 单文件满足当前规模；多实例部署时换 Redis/Postgres。
"""

from __future__ import annotations

import sqlite3
from dataclasses import dataclass
from enum import StrEnum
from pathlib import Path

__all__ = [
    "Order",
    "OrderTransitionError",
    "OrderState",
    "FulfillmentState",
    "FulfillmentResult",
    "PaymentOrder",
    "OrderStore",
    "PendingOrder",
]


class OrderState(StrEnum):
    PENDING_PAYMENT = "PENDING_PAYMENT"
    PAID = "PAID"
    FULFILLED = "FULFILLED"


class FulfillmentState(StrEnum):
    UNFULFILLED = "UNFULFILLED"
    PENDING_CONFIRM = "PENDING_CONFIRM"
    FULFILLED = "FULFILLED"


@dataclass(frozen=True)
class Order:
    out_trade_no: str
    trade_no: str | None
    state: OrderState
    tool_name: str
    amount: str
    currency: str
    goods_name: str
    pay_before: str
    fulfillment_state: FulfillmentState


@dataclass(frozen=True)
class PendingOrder:
    out_trade_no: str
    tool_name: str
    amount: str
    currency: str
    goods_name: str
    pay_before: str


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
    def from_snapshot(cls, order: Order) -> PaymentOrder:
        return cls(**order.__dict__)

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


@dataclass(frozen=True)
class FulfillmentResult:
    state: str
    order: Order | None = None


class OrderStore:
    """订单聚合根仓库。持久连接复用，避免每次操作创建新连接。"""

    def __init__(self, db_path: str) -> None:
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)
        self._db_path = db_path
        self._conn: sqlite3.Connection | None = None
        self._init_schema()

    def _get_conn(self) -> sqlite3.Connection:
        if self._conn is None:
            conn = sqlite3.connect(self._db_path, check_same_thread=False)
            conn.execute("PRAGMA journal_mode=WAL")
            conn.execute("PRAGMA busy_timeout=5000")
            self._conn = conn
        return self._conn

    def _init_schema(self) -> None:
        conn = self._get_conn()
        conn.execute("""
            CREATE TABLE IF NOT EXISTS aipay_orders (
                out_trade_no TEXT PRIMARY KEY,
                trade_no TEXT UNIQUE,
                state TEXT NOT NULL,
                tool_name TEXT NOT NULL,
                amount TEXT NOT NULL DEFAULT '',
                currency TEXT NOT NULL DEFAULT '',
                goods_name TEXT NOT NULL DEFAULT '',
                pay_before TEXT NOT NULL DEFAULT '',
                fulfillment_state TEXT NOT NULL DEFAULT 'UNFULFILLED',
                created_at TEXT NOT NULL DEFAULT (datetime('now')),
                updated_at TEXT NOT NULL DEFAULT (datetime('now'))
            )
        """)
        columns = {row[1] for row in conn.execute("PRAGMA table_info(aipay_orders)")}
        migrations = {
            "amount": "ALTER TABLE aipay_orders ADD COLUMN amount TEXT NOT NULL DEFAULT ''",
            "currency": "ALTER TABLE aipay_orders ADD COLUMN currency TEXT NOT NULL DEFAULT ''",
            "goods_name": "ALTER TABLE aipay_orders ADD COLUMN goods_name TEXT NOT NULL DEFAULT ''",
            "pay_before": "ALTER TABLE aipay_orders ADD COLUMN pay_before TEXT NOT NULL DEFAULT ''",
            "fulfillment_state": (
                "ALTER TABLE aipay_orders ADD COLUMN fulfillment_state TEXT NOT NULL DEFAULT 'UNFULFILLED'"
            ),
        }
        for name, statement in migrations.items():
            if name not in columns:
                conn.execute(statement)
        conn.commit()

    def create_pending(self, order: PendingOrder) -> Order:
        conn = self._get_conn()
        conn.execute(
            """
            INSERT INTO aipay_orders
                (out_trade_no, state, tool_name, amount, currency, goods_name,
                 pay_before, fulfillment_state)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                order.out_trade_no,
                OrderState.PENDING_PAYMENT,
                order.tool_name,
                order.amount,
                order.currency,
                order.goods_name,
                order.pay_before,
                FulfillmentState.UNFULFILLED,
            ),
        )
        conn.commit()
        return self.find_by_out_trade_no(order.out_trade_no)

    def mark_paid(self, out_trade_no: str) -> None:
        """Persist a verified payment transition guarded by pending state."""
        order = self.find_by_out_trade_no(out_trade_no)
        if order is None:
            raise OrderTransitionError("订单不存在")
        aggregate = PaymentOrder.from_snapshot(order)
        aggregate.mark_paid()
        conn = self._get_conn()
        cursor = conn.execute(
            """
            UPDATE aipay_orders
               SET state = ?, updated_at = datetime('now')
             WHERE out_trade_no = ?
               AND state = ?
               AND fulfillment_state = ?
            """,
            (
                OrderState.PAID,
                out_trade_no,
                OrderState.PENDING_PAYMENT,
                FulfillmentState.UNFULFILLED,
            ),
        )
        if cursor.rowcount == 0:
            current = self.find_by_out_trade_no(out_trade_no)
            if current is None:
                raise OrderTransitionError("订单不存在")
            aggregate = PaymentOrder.from_snapshot(current)
            try:
                aggregate.mark_paid()
            except OrderTransitionError:
                raise OrderTransitionError("订单状态被并发修改") from None
            return
        conn.commit()

    def find_by_out_trade_no(self, out_trade_no: str) -> Order | None:
        row = (
            self._get_conn()
            .execute(
                """
            SELECT out_trade_no, trade_no, state, tool_name, amount, currency,
                   goods_name, pay_before, fulfillment_state
            FROM aipay_orders
            WHERE out_trade_no = ?
            """,
                (out_trade_no,),
            )
            .fetchone()
        )
        if row is None:
            return None
        return Order(
            out_trade_no=row[0],
            trade_no=row[1],
            state=OrderState(row[2]),
            tool_name=row[3],
            amount=row[4],
            currency=row[5],
            goods_name=row[6],
            pay_before=row[7],
            fulfillment_state=FulfillmentState(row[8]),
        )

    def reserve_fulfillment(self, out_trade_no: str) -> FulfillmentResult:
        order = self.find_by_out_trade_no(out_trade_no)
        if order is None:
            return FulfillmentResult(state="NOT_FOUND")
        if order.fulfillment_state == FulfillmentState.FULFILLED:
            return FulfillmentResult(state="FULFILLED", order=order)

        aggregate = PaymentOrder.from_snapshot(order)
        try:
            aggregate.reserve_fulfillment()
        except OrderTransitionError:
            return FulfillmentResult(state="INVALID_STATE", order=order)

        conn = self._get_conn()
        cursor = conn.execute(
            """
            UPDATE aipay_orders
               SET state = ?, fulfillment_state = ?, updated_at = datetime('now')
             WHERE out_trade_no = ?
               AND state = ?
               AND fulfillment_state = ?
            """,
            (
                OrderState.PAID,
                FulfillmentState.PENDING_CONFIRM,
                out_trade_no,
                OrderState.PAID,
                FulfillmentState.UNFULFILLED,
            ),
        )
        conn.commit()
        if cursor.rowcount == 1:
            reserved = self.find_by_out_trade_no(out_trade_no)
            return FulfillmentResult(
                state="PENDING_CONFIRM",
                order=reserved,
            )

        reread = self.find_by_out_trade_no(out_trade_no)
        if reread is None:
            return FulfillmentResult(state="NOT_FOUND")
        if reread.fulfillment_state == FulfillmentState.PENDING_CONFIRM:
            return FulfillmentResult(state="PENDING_CONFIRM", order=reread)
        if reread.fulfillment_state == FulfillmentState.FULFILLED:
            return FulfillmentResult(state="FULFILLED", order=reread)
        aggregate = PaymentOrder.from_snapshot(reread)
        try:
            aggregate.reserve_fulfillment()
        except OrderTransitionError:
            return FulfillmentResult(state="INVALID_STATE", order=reread)
        # A concurrent writer changed the row after the guarded update; the
        # aggregate invariant is the authoritative compatibility check.
        return FulfillmentResult(state="INVALID_STATE", order=reread)

    def mark_fulfilled(self, out_trade_no: str, trade_no: str) -> None:
        order = self.find_by_out_trade_no(out_trade_no)
        if order is None:
            raise OrderTransitionError("订单不存在")
        if order.fulfillment_state == FulfillmentState.FULFILLED:
            if order.trade_no != trade_no:
                raise OrderTransitionError("履约 trade_no 不一致")
            return
        aggregate = PaymentOrder.from_snapshot(order)
        aggregate.mark_fulfilled(trade_no)
        conn = self._get_conn()
        cursor = conn.execute(
            """
            UPDATE aipay_orders
               SET trade_no = ?, state = ?, fulfillment_state = ?,
                   updated_at = datetime('now')
             WHERE out_trade_no = ?
               AND fulfillment_state = ?
            """,
            (
                trade_no,
                OrderState.FULFILLED,
                FulfillmentState.FULFILLED,
                out_trade_no,
                FulfillmentState.PENDING_CONFIRM,
            ),
        )
        if cursor.rowcount == 0:
            raise OrderTransitionError("订单状态被并发修改")
        conn.commit()

    def close(self) -> None:
        if self._conn is not None:
            self._conn.close()
            self._conn = None
