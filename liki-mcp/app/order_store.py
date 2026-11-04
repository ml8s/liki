"""订单持久化仓库（SQLite）。

A2M 协议要求原子创建、唯一约束防重复核销、幂等履约标记。
SQLite 单文件满足当前规模；多实例部署时换 Redis/Postgres。
"""

from __future__ import annotations

import sqlite3
from dataclasses import dataclass
from enum import StrEnum
from pathlib import Path


class OrderState(StrEnum):
    PENDING_PAYMENT = "PENDING_PAYMENT"
    PAID = "PAID"
    FULFILLED = "FULFILLED"


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
    """订单聚合根仓库。A2M demo 要求的四个端口方法。"""

    def __init__(self, db_path: str) -> None:
        self._db_path = db_path
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)
        self._init_schema()

    def _connect(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self._db_path)
        conn.execute("PRAGMA journal_mode=WAL")
        conn.execute("PRAGMA busy_timeout=5000")
        return conn

    def _init_schema(self) -> None:
        with self._connect() as conn:
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

    def create_pending(self, out_trade_no: str, tool_name: str) -> Order:
        """原子创建待支付订单；out_trade_no 唯一约束防重复。"""
        with self._connect() as conn:
            conn.execute(
                "INSERT OR IGNORE INTO aipay_orders (out_trade_no, state, tool_name) VALUES (?, ?, ?)",
                (out_trade_no, OrderState.PENDING_PAYMENT, tool_name),
            )
        return Order(out_trade_no=out_trade_no, trade_no=None,
                     state=OrderState.PENDING_PAYMENT, tool_name=tool_name)

    def find_by_out_trade_no(self, out_trade_no: str) -> Order | None:
        with self._connect() as conn:
            row = conn.execute(
                "SELECT out_trade_no, trade_no, state, tool_name FROM aipay_orders WHERE out_trade_no = ?",
                (out_trade_no,),
            ).fetchone()
        if row is None:
            return None
        return Order(out_trade_no=row[0], trade_no=row[1],
                     state=OrderState(row[2]), tool_name=row[3])

    def mark_paid(self, out_trade_no: str, trade_no: str) -> None:
        with self._connect() as conn:
            conn.execute(
                "UPDATE aipay_orders SET trade_no = ?, state = ?, updated_at = datetime('now') WHERE out_trade_no = ?",
                (trade_no, OrderState.PAID, out_trade_no),
            )

    def prepare_fulfillment(self, out_trade_no: str) -> FulfillmentResult:
        """原子校验订单并标记 PENDING_CONFIRM；返回履约上下文。"""
        order = self.find_by_out_trade_no(out_trade_no)
        if order is None:
            return FulfillmentResult(state="NOT_FOUND")
        if order.state == OrderState.FULFILLED:
            return FulfillmentResult(state="FULFILLED", order=order)
        with self._connect() as conn:
            conn.execute(
                "UPDATE aipay_orders SET state = ?, updated_at = datetime('now') WHERE out_trade_no = ? AND state = ?",
                (OrderState.PAID, out_trade_no, OrderState.PENDING_PAYMENT),
            )
        return FulfillmentResult(state="PENDING_CONFIRM", order=order)

    def mark_fulfilled(self, out_trade_no: str, trade_no: str) -> None:
        with self._connect() as conn:
            conn.execute(
                "UPDATE aipay_orders SET trade_no = ?, state = ?, updated_at = datetime('now') WHERE out_trade_no = ?",
                (trade_no, OrderState.FULFILLED, out_trade_no),
            )
