"""Canonical natal chart contract guardrails."""

import pytest
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart
from app.natal.tools.factor_constants import load_constants

CONSTANTS = load_constants()
PILLARS = CONSTANTS["四柱"]
PALACES = CONSTANTS["紫微宫位"]


def _pan():
    chart = {
        "subject": {
            "birth_time": "1990-05-20T12:00:00+08:00",
            "birth_longitude": 116.4,
            "gender": "male",
        },
        "bazi": {pillar: {"gan": "甲", "zhi": "子"} for pillar in PILLARS},
        "ziwei": {
            "shi_er_gong": [
                {"ming": palace, "xing_yao": []} for palace in PALACES
            ],
        },
        "ziwei_daxian": [
            {
                "ming_gong": palace,
                "start_year": index * 10 + 1,
                "end_year": index * 10 + 10,
                "qi_sui": index * 10 + 1,
                "zhi_sui": index * 10 + 10,
            }
            for index, palace in enumerate(PALACES)
        ],
    }
    return with_natal_chart_digest(chart)


def test_valid_natal_chart_passes():
    validate_natal_chart(_pan(), action="test")


def test_digest_rejects_tampering():
    chart = _pan()
    chart["subject"]["birth_longitude"] = 117.0
    with pytest.raises(ValueError, match="digest mismatch"):
        validate_natal_chart(chart, action="test")


def test_required_envelope_fields_are_required():
    for key in ("subject", "bazi", "ziwei", "ziwei_daxian", "natal_chart_digest"):
        chart = _pan()
        chart.pop(key)
        with pytest.raises(ValueError, match=key):
            validate_natal_chart(chart, action="test")


def test_subject_and_four_pillars_are_required():
    chart = _pan()
    chart["subject"].pop("gender")
    with pytest.raises(ValueError, match="gender"):
        validate_natal_chart(chart, action="test")

    for pillar in PILLARS:
        chart = _pan()
        chart["bazi"].pop(pillar)
        with pytest.raises(ValueError, match=pillar):
            validate_natal_chart(chart, action="test")


def test_ziwei_palaces_and_daxian_are_complete():
    chart = _pan()
    chart["ziwei"]["shi_er_gong"] = chart["ziwei"]["shi_er_gong"][:11]
    with pytest.raises(ValueError, match="12"):
        validate_natal_chart(chart, action="test")

    chart = _pan()
    chart["ziwei_daxian"] = chart["ziwei_daxian"][:-1]
    with pytest.raises(ValueError, match="12"):
        validate_natal_chart(chart, action="test")


def test_redigest_does_not_hide_missing_chart_side():
    chart = _pan()
    chart.pop("ziwei")
    with pytest.raises(ValueError, match="ziwei"):
        validate_natal_chart(with_natal_chart_digest(chart), action="test")
