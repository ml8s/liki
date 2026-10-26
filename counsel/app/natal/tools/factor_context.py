"""因子求值上下文对象：显式生命周期，不污染调用方 natal_chart。"""

from collections.abc import Mapping
from dataclasses import dataclass

__all__ = ["FactorContext", "NatalContext", "PalaceIndex"]


class PalaceIndex:
    """紫微宫位事实索引：一次构建，O(1) 查询。

    索引存放在 FactorContext 上，不修改调用方 natal_chart。
    支持四种查询模式（palace 具体/任意 × star 具体/任意）。
    """

    __slots__ = ("_exact", "_any_palace", "_any_star", "_any_both", "_stars_by_pkt")

    def __init__(self, facts):
        self._exact: set = set()       # (palace, kind, target, star)
        self._any_palace: set = set()  # (kind, target, star)
        self._any_star: set = set()    # (palace, kind, target)
        self._any_both: set = set()    # (kind, target)
        self._stars_by_pkt: dict = {}  # (palace, kind, target) → set[star]
        for fact in facts:
            if not isinstance(fact, dict):
                continue
            p = fact.get("palace", "")
            k = fact.get("kind", "")
            t = fact.get("target", "")
            s = fact.get("star", "")
            self._exact.add((p, k, t, s))
            self._any_palace.add((k, t, s))
            self._any_star.add((p, k, t))
            self._any_both.add((k, t))
            pkt = (p, k, t)
            stars = self._stars_by_pkt.get(pkt)
            if stars is None:
                self._stars_by_pkt[pkt] = {s}
            else:
                stars.add(s)

    def hit(self, palace: str, kind: str, target: str, star: str = "") -> bool:
        """O(1) 查询。star="" 表示任意 star；palace=任意 表示跨宫。"""
        if palace == "任意":
            if star:
                return (kind, target, star) in self._any_palace
            return (kind, target) in self._any_both
        if star:
            return (palace, kind, target, star) in self._exact
        return (palace, kind, target) in self._any_star

    def stars_for(self, palace: str, kind: str, target: str) -> frozenset:
        """返回 (palace, kind, target) 匹配的 star 集合。"""
        if palace == "任意":
            return frozenset(
                s for (k, t, s) in self._any_palace if k == kind and t == target
            )
        return frozenset(self._stars_by_pkt.get((palace, kind, target)) or ())


class FactorContext(Mapping):
    """一次因子求值的只读 natal_chart 视图 + 已构建基础上下文。

    实现 Mapping 是为了让既有算子继续用 `ctx.get("full")` 直读 natal_chart；
    `base` 显式挂载聚合结果，避免写回调用方 natal_chart。
    """

    def __init__(self, natal_chart: dict, base: dict):
        self.natal_chart = natal_chart or {}
        self.base = base or {}
        self._palace_index = None

    def __getitem__(self, key):
        return self.natal_chart[key]

    def __iter__(self):
        return iter(self.natal_chart)

    def __len__(self) -> int:
        return len(self.natal_chart)

    @property
    def palace_index(self) -> PalaceIndex:
        """紫微宫位事实索引（lazy，首次访问构建，后续复用）。"""
        if self._palace_index is None:
            facts = ((self.natal_chart.get("ziwei") or {}).get("palace_facts") or [])
            self._palace_index = PalaceIndex(facts)
        return self._palace_index


@dataclass(frozen=True)
class NatalContext:
    """多年流年复用的本命求值上下文与八字快照。"""

    evaluation: FactorContext
    snapshot: dict
