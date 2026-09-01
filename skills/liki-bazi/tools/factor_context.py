"""因子求值上下文对象：显式生命周期，不污染调用方 pan。"""
from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass

__all__ = ["FactorContext", "NatalContext"]


class FactorContext(Mapping):
    """一次因子求值的只读 pan 视图 + 已构建基础上下文。

    实现 Mapping 是为了让既有算子继续用 `ctx.get("full")` 直读 pan；
    `base` 显式挂载聚合结果，避免写回调用方 pan。
    """

    def __init__(self, pan: dict, base: dict):
        self.pan = pan or {}
        self.base = base or {}

    def __getitem__(self, key):
        return self.pan[key]

    def __iter__(self):
        return iter(self.pan)

    def __len__(self) -> int:
        return len(self.pan)


@dataclass(frozen=True)
class NatalContext:
    """多年流年复用的本命求值上下文与八字快照。"""

    evaluation: FactorContext
    snapshot: dict
