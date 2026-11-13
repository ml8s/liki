"""Small TTL cache for server-derived current-year facts."""

import time


class TtlCache:
    """有界 TTL 缓存：显式生命周期，不暴露裸全局变量。"""

    __slots__ = ("_data", "_at", "_ttl")

    def __init__(self, ttl: float):
        self._data = None
        self._at: float | None = None
        self._ttl = ttl

    def get(self, key, now: float):
        """命中返回 value，过期/不匹配返回 None。"""
        if self._at is None or self._data is None:
            return None
        if now - self._at >= self._ttl:
            return None
        if self._data[-1] != key:
            return None
        return self._data[:-1]

    def set(self, key, value: tuple, now: float):
        self._data = (*value, key)
        self._at = now

    def clear(self):
        self._data = None
        self._at = None


# RPC 失败仍抛错（fail-fast，绝不回退本地时钟）；仅成功结果被缓存。
year_cache = TtlCache(ttl=60)


def resolve_current_years(timezone: str):
    """返回 (八字年号[立春界], 农历年号[春节界], source)。timezone 必填。"""
    from app.natal.tools.charts import current_years

    if not timezone:
        raise ValueError("timezone is required")
    now = time.monotonic()
    cached = year_cache.get(timezone, now)
    if cached is not None:
        return cached
    bazi_year, lunar_year = current_years(timezone)
    year_cache.set(timezone, (bazi_year, lunar_year, "server"), now)
    return bazi_year, lunar_year, "server"


def resolve_current_year(timezone: str):
    """兼容入口：返回 (八字年号, source)。timezone 必填。"""
    bazi_year, _lunar_year, source = resolve_current_years(timezone)
    return bazi_year, source


def reset_current_year_cache():
    year_cache.clear()
