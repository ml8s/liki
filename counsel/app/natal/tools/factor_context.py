"""因子求值上下文对象：显式生命周期，不污染调用方 natal_chart。"""

from collections.abc import Mapping
from dataclasses import dataclass

__all__ = ["FactorContext", "NatalContext"]


class FactorContext(Mapping):
    """一次因子求值的只读 natal_chart 视图 + 已构建基础上下文。

    实现 Mapping 是为了让既有算子继续用 `ctx.get("full")` 直读 natal_chart；
    `base` 显式挂载聚合结果，避免写回调用方 natal_chart。
    """

    def __init__(self, natal_chart: dict, base: dict):
        self.natal_chart = natal_chart or {}
        self.base = base or {}

    def __getitem__(self, key):
        return self.natal_chart[key]

    def __iter__(self):
        return iter(self.natal_chart)

    def __len__(self) -> int:
        return len(self.natal_chart)


@dataclass(frozen=True)
class NatalContext:
    """多年流年复用的本命求值上下文与八字快照。"""

    evaluation: FactorContext
    snapshot: dict
