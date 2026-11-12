"""Engine liunian series keys must match the factors layer without aliases."""

from app.natal.tools.charts import (
    _bazi_liunian_range,  # noqa: E402
    _ziwei_liunian_range,  # noqa: E402
)
from app.natal.tools.counsel import _build_natal_chart  # noqa: E402
from app.natal.tools.factors import (
    evaluate_liunian_snap_from_natal_chart,  # noqa: E402
    prepare_natal_context,  # noqa: E402
)


def _subject() -> dict:
    return {
        "standard_time": "1990-05-20T11:49:00+08:00",
        "longitude": 116.4,
        "gender": "male",
    }


def test_bazi_liunian_series_exposes_canonical_gan_zhi() -> None:
    items = _bazi_liunian_range(_subject(), 2026, 2)
    assert len(items) == 2
    assert items[0]["gan"] and items[0]["zhi"]
    assert items[1]["gan"] != items[0]["gan"] or items[1]["zhi"] != items[0]["zhi"]


def test_ziwei_liunian_series_exposes_canonical_gan_zhi() -> None:
    items = _ziwei_liunian_range(_subject(), 2026, 2)
    assert len(items) == 2
    assert items[0]["zhi"]
    assert items[1]["zhi"] != items[0]["zhi"]


def test_spouse_palace_value_has_annual_branch_cadence() -> None:
    natal_chart = _build_natal_chart(_subject())
    context = prepare_natal_context(
        natal_chart, factor_names={"流年配偶星值宫"}
    )
    assert context.evaluation.base["palace_ri"]["zhi"]
    hits = []
    for year, liunian_bazi in enumerate(
        _bazi_liunian_range(_subject(), 2026, 12), start=2026
    ):
        liunian = {"bazi": liunian_bazi, "ziwei": {}}
        snapshot = evaluate_liunian_snap_from_natal_chart(
            natal_chart, liunian, year=year,
            natal_context=context, factor_names={"流年配偶星值宫"},
        )
        if snapshot["八字"]["流年配偶星值宫"] == "命中":
            hits.append(liunian["bazi"]["zhi"])
    assert len(hits) == 1
