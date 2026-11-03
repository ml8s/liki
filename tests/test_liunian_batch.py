"""yearly_range 流年预取的单次 RPC 与 engine count 边界契约。"""
import threading

import _helpers  # noqa: F401 —— 注入 tools 路径
import pytest
from app.natal.tools.charts import (
    MAX_LIUNIAN_BATCH,
    RPCError,
    _bazi_liunian_range,
    _ziwei_liunian_range,
)
from app.natal.tools.duanyu import _batch_liunian_data


def _pan(*, bazi=True, ziwei=True) -> dict:
    return {
        "subject": {},
        "bazi": {"ri": {}} if bazi else {},
        "ziwei": {"palaces": []} if ziwei else {},
    }


def test_batch_liunian_uses_one_call_per_side(monkeypatch) -> None:
    calls = []

    def bazi_range(_chart, start, count):
        calls.append(("bazi", start, count))
        return [{"year": start + i} for i in range(count)]

    def ziwei_range(_chart, start, count):
        calls.append(("ziwei", start, count))
        return [{"lunar_year": start + i} for i in range(count)]

    monkeypatch.setattr("app.natal.tools.charts._bazi_liunian_range", bazi_range)
    monkeypatch.setattr("app.natal.tools.charts._ziwei_liunian_range", ziwei_range)

    cache = _batch_liunian_data(_pan(), 2026, MAX_LIUNIAN_BATCH)

    assert calls == [
        ("bazi", 2026, MAX_LIUNIAN_BATCH),
        ("ziwei", 2026, MAX_LIUNIAN_BATCH),
    ]
    assert list(cache) == list(range(2026, 2026 + MAX_LIUNIAN_BATCH))
    assert cache[2075]["bazi"] == {"year": 2075}
    assert cache[2075]["ziwei"] == {"lunar_year": 2075}


def test_batch_liunian_runs_side_calls_concurrently(monkeypatch) -> None:
    barrier = threading.Barrier(2, timeout=1)

    def wait_for_side(*_):
        barrier.wait()
        return []

    monkeypatch.setattr(
        "app.natal.tools.charts._bazi_liunian_range", wait_for_side
    )
    monkeypatch.setattr(
        "app.natal.tools.charts._ziwei_liunian_range", wait_for_side
    )
    cache = _batch_liunian_data(_pan(), 2026, 1)
    assert cache == {2026: {"bazi": {}, "ziwei": {}}}


def test_batch_liunian_fails_closed_for_the_requested_range(monkeypatch) -> None:
    def bazi_range(_chart, start, count):
        return [{"year": year} for year in range(start, start + count)]

    def ziwei_range(_chart, start, count):
        if start == 2026:
            raise ConnectionError("engine unavailable")
        return [{} for _ in range(count)]

    monkeypatch.setattr("app.natal.tools.charts._bazi_liunian_range", bazi_range)
    monkeypatch.setattr("app.natal.tools.charts._ziwei_liunian_range", ziwei_range)

    cache = _batch_liunian_data(_pan(), 2026, 2)

    assert cache[2026]["error"] == "ConnectionError: engine unavailable"
    assert cache[2027]["error"] == "ConnectionError: engine unavailable"
    assert all(item["bazi"] == {} and item["ziwei"] == {} for item in cache.values())


def test_batch_liunian_keeps_bazi_only_analysis_pan_on_one_side(monkeypatch) -> None:
    calls = []

    def bazi_range(_chart, start, count):
        calls.append("bazi")
        return [{"year": year} for year in range(start, start + count)]

    def ziwei_range(*_):
        calls.append("ziwei")
        return []

    monkeypatch.setattr("app.natal.tools.charts._bazi_liunian_range", bazi_range)
    monkeypatch.setattr("app.natal.tools.charts._ziwei_liunian_range", ziwei_range)

    cache = _batch_liunian_data(_pan(ziwei=False), 2026, 1)

    assert calls == ["bazi"]
    assert cache == {2026: {"bazi": {"year": 2026}, "ziwei": {}}}


def test_batch_liunian_keeps_ziwei_only_analysis_pan_on_one_side(monkeypatch) -> None:
    calls = []

    def bazi_range(*_):
        calls.append("bazi")
        return []

    def ziwei_range(_chart, start, count):
        calls.append("ziwei")
        return [{"lunar_year": year} for year in range(start, start + count)]

    monkeypatch.setattr("app.natal.tools.charts._bazi_liunian_range", bazi_range)
    monkeypatch.setattr("app.natal.tools.charts._ziwei_liunian_range", ziwei_range)

    cache = _batch_liunian_data(_pan(bazi=False), 2026, 1)

    assert calls == ["ziwei"]
    assert cache == {2026: {"bazi": {}, "ziwei": {"lunar_year": 2026}}}


@pytest.mark.parametrize("range_call", [_bazi_liunian_range, _ziwei_liunian_range])
def test_liunian_range_rejects_short_engine_items(monkeypatch, range_call) -> None:
    monkeypatch.setattr(
        "app.natal.tools.charts.call",
        lambda *_: {"data": {"items": [{"year": 2026}]}},
    )
    with pytest.raises(RPCError, match="engine response expected 2 items"):
        range_call({"chart": {}}, 2026, 2)


@pytest.mark.parametrize("range_call", [_bazi_liunian_range, _ziwei_liunian_range])
def test_liunian_range_rejects_missing_batch_items(monkeypatch, range_call) -> None:
    monkeypatch.setattr(
        "app.natal.tools.charts.call",
        lambda *_: {"data": {"year": 2026}},
    )
    with pytest.raises(RPCError, match="engine response expected 2 items"):
        range_call({"chart": {}}, 2026, 2)
