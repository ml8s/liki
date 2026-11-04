"""订单持久化仓库（SQLite）。

A2M 协议要求原子创建、唯一约束防重复核销、幂等履约标记。
SQLite 单文件满足当前规模；多实例部署时换 Redis/Postgres。
"""

from __future__ import annotations

import sqlite3
from dataclasses import dataclass
from enum import StrEnum
from pathlib import Path

__all__ = ["Order", "OrderState", "FulfillmentResult", "OrderStore"]


_TRANSITIONS: dict[str, set[str]] = {
    "PENDING_PAYMENT": {"PAID", "FULFILLED"},
    "PAID": {"FULFILLED"},
    "FULFILLED": set(),
}


class OrderState(StrEnum):
    PENDING_PAYMENT = "PENDING_PAYMENT"
    PAID = "PAID"
    FULFILLED = "FULFILLED"

    def can_transition_to(self, target: "OrderState") -> bool:
        return target.value in _TRANSITIONS.get(self.value, set())


@dataclass(frozen=True)
class Order:
    out_trade_no: str
    trade_no: str | None
    state: OrderState
    tool_name: str


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
                created_at TEXT NOT NULL DEFAULT (datetime('now')),
                updated_at TEXT NOT NULL DEFAULT (datetime('now'))
            )
        """)
        conn.commit()

    def create_pending(self, out_trade_no: str, tool_name: str) -> Order:
        conn = self._get_conn()
        conn.execute(
            "INSERT OR IGNORE INTO aipay_orders (out_trade_no, state, tool_name) VALUES (?, ?, ?)",
            (out_trade_no, OrderState.PENDING_PAYMENT, tool_name),
        )
        conn.commit()
        return Order(out_trade_no=out_trade_no, trade_no=None,
                     state=OrderState.PENDING_PAYMENT, tool_name=tool_name)

    def find_by_out_trade_no(self, out_trade_no: str) -> Order | None:
        row = self._get_conn().execute(
            "SELECT out_trade_no, trade_no, state, tool_name FROM aipay_orders WHERE out_trade_no = ?",
            (out_trade_no,),
        ).fetchone()
        if row is None:
            return None
        return Order(out_trade_no=row[0], trade_no=row[1],
                     state=OrderState(row[2]), tool_name=row[3])

    def prepare_fulfillment(self, out_trade_no: str) -> FulfillmentResult:
        order = self.find_by_out_trade_no(out_trade_no)
        if order is None:
            return FulfillmentResult(state="NOT_FOUND")
        if order.state == OrderState.FULFILLED:
            return FulfillmentResult(state="FULFILLED", order=order)
        conn = self._get_conn()
        conn.execute(
            "UPDATE aipay_orders SET state = ?, updated_at = datetime('now') WHERE out_trade_no = ? AND state = ?",
            (OrderState.PAID, out_trade_no, OrderState.PENDING_PAYMENT),
        )
        conn.commit()
        return FulfillmentResult(state="PENDING_CONFIRM", order=order)

    def mark_fulfilled(self, out_trade_no: str, trade_no: str) -> None:
        conn = self._get_conn()
        conn.execute(
            "UPDATE aipay_orders SET trade_no = ?, state = ?, updated_at = datetime('now') WHERE out_trade_no = ?",
            (trade_no, OrderState.FULFILLED, out_trade_no),
        )
        conn.commit()

    def close(self) -> None:
        if self._conn is not None:
            self._conn.close()
            self._conn = None
