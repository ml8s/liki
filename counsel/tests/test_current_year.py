"""WP6: current-year boundaries (立春 for bazi, 春节 for ziwei)."""

import pytest
from app.domain.natal.duanyu.cache import (  # noqa: E402
    reset_current_year_cache,
    resolve_current_year,
    resolve_current_years,
)
from app.natal.tools import charts  # noqa: E402

GZ = {2024: ("甲", "辰"), 2025: ("乙", "巳"), 2026: ("丙", "午")}
# 测试日期避开设节当日；立春 2025-02-03 / 2026-02-04，春节 2025-01-29 / 2026-02-17。
BAZI = {
    "2025-01-15": 2024,
    "2025-02-10": 2025,
    "2026-01-15": 2025,
    "2026-02-10": 2026,
    "2026-02-20": 2026,
    "2026-10-21": 2026,
}
LUNAR = {
    "2025-01-15": 2024,
    "2025-02-10": 2025,
    "2026-01-15": 2025,
    "2026-02-10": 2025,
    "2026-02-20": 2026,
    "2026-10-21": 2026,
}


def _date(gt):
    return gt[:10]


def _fake_call(cst):
    def call(name, params):
        if name == "time_now":
            return {
                "data": {
                    "local_standard_time": cst + "T12:00:00+08:00",
                    "offset": 28800,
                }
            }
        gt = params["standard_time"]
        if name == "bazi_chart":
            year = BAZI[_date(gt)]
            gan, zhi = GZ[year]
            return {"data": {"nian": {"gan": gan, "zhi": zhi}, "birth_year": int(gt[:4])}}
        if name == "ziwei_chart":
            return {"data": {"birth_year": LUNAR[_date(gt)]}}
        if name == "bazi_liupan":
            year = BAZI[_date(gt)]
            gan, zhi = GZ[year]
            return {
                "data": {
                    "liu_nian": [
                        {
                            "year": year,
                            "nian": {"gan": gan, "zhi": zhi},
                            "nian_jie": {"lichun_year": year},
                        }
                    ]
                }
            }
        if name == "ziwei_liupan":
            year = LUNAR[_date(gt)]
            return {
                "data": {
                    "liu_nian": [
                        {
                            "nian_jie": {"lunar_year": year},
                        }
                    ]
                }
            }
        raise AssertionError(f"unexpected tool {name}")

    return call


@pytest.mark.parametrize("cst", sorted(BAZI))
def test_current_year_boundaries(monkeypatch, cst):
    reset_current_year_cache()
    monkeypatch.setattr(charts, "call", _fake_call(cst + "T12:00:00+08:00"))
    bazi_year, lunar_year, source = resolve_current_years("Asia/Shanghai")
    assert (bazi_year, lunar_year, source) == (BAZI[cst], LUNAR[cst], "server")
    assert resolve_current_year("Asia/Shanghai") == (BAZI[cst], "server")
    reset_current_year_cache()
