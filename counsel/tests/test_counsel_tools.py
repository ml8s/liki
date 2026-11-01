"""Subject-bound counsel tool contract tests."""

import pytest


@pytest.fixture(scope="module")
def subject():
    return {
        "birth_time": "1990-05-20T11:49:00+08:00",
        "birth_longitude": 116.4,
        "gender": "male",
    }


def test_natal_compute_factors_returns_factors_and_digest(subject):
    from app.natal.tools.counsel import natal_compute_factors

    out = natal_compute_factors(subject, "bazi")
    assert set(out) == {"factors", "factors_digest"}


def test_natal_query_returns_bazi_assertions(subject):
    from app.natal.tools.counsel import natal_compute_factors, natal_query

    out = natal_compute_factors(subject, "bazi")
    result = natal_query(out["factors"], ["chart_structure"], out["factors_digest"])
    assert result["assertions"]
    assert all(item.get("side") == "bazi" for item in result["assertions"])


def test_natal_period_query_returns_bazi_assertions(subject):
    from app.natal.tools.counsel import natal_compute_factors, natal_period_query

    out = natal_compute_factors(subject, "bazi")
    got = natal_period_query(
        out["factors"], out["factors_digest"],
        {"type": "year_range", "start_year": 2020, "end_year": 2040, "timezone": "Asia/Shanghai"},
        ["marriage"], subject, "bazi",
    )
    assert got["periods"]
    assert any(period.get("assertions") for period in got["periods"])
    assert all(
        item.get("side") == "bazi"
        for period in got["periods"] for item in period["assertions"]
    )


def test_natal_query_rejects_modified_factors(subject):
    from app.natal.tools.counsel import natal_compute_factors, natal_query

    out = natal_compute_factors(subject, "bazi")
    factors = dict(out["factors"])
    factors[next(iter(factors))] = "tampered"
    with pytest.raises(ValueError, match="factors_digest mismatch"):
        natal_query(factors, ["chart_structure"], out["factors_digest"])


def test_natal_period_query_rejects_subject_from_another_snapshot(subject):
    from app.natal.tools.counsel import natal_compute_factors, natal_period_query

    out = natal_compute_factors(subject, "bazi")
    other = {**subject, "birth_longitude": 117.0}
    with pytest.raises(ValueError, match="subject 与 natal_compute_factors"):
        natal_period_query(
            out["factors"], out["factors_digest"],
            {"type": "year_range", "start_year": 2020, "end_year": 2040, "timezone": "Asia/Shanghai"},
            ["marriage"], other, "bazi",
        )
