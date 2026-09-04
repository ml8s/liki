"""流年宫位算子必须消费 operators_natal._base_ctx_from_pan(pan) 产出的上下文。"""
import _helpers  # noqa: F401 —— 注入 tools 路径
from operators_liunian import _liu_op
from operators_natal import _base_ctx_from_pan


def _pan() -> dict:
    return {
        "full": {
            "nian": {"gan": "庚", "zhi": "午", "shi_shens": []},
            "yue": {"gan": "壬", "zhi": "午", "shi_shens": []},
            "ri": {"gan": "甲", "zhi": "子", "shi_shens": []},
            "shi": {"gan": "庚", "zhi": "午", "shi_shens": []},
        },
        "yongshen": {"fu_yi": {}},
        "chart": {
            "nian": {"gan": "庚", "zhi": "午"},
            "yue": {"gan": "壬", "zhi": "午"},
            "ri": {"gan": "甲", "zhi": "子"},
            "shi": {"gan": "庚", "zhi": "午"},
        },
    }


def _ctx(nian_zhi: str, relation_type: str | None = None) -> dict:
    relation = (
        [{"zhi_a": nian_zhi, "zhi_b": "子", "type": relation_type}]
        if relation_type
        else []
    )
    return {
        "year": 2006,
        "chart": _pan(),
        "liunian": {
            "nian_zhi": nian_zhi,
            "natal_interactions": [{"zhi_rels": relation}],
        },
    }


def test_flow_value_uses_extracted_day_palace() -> None:
    pan = _pan()
    base = _base_ctx_from_pan(pan)

    assert _liu_op("流年值", ["配偶星"], "male", pan, _ctx("子")) == 1
    assert _liu_op("流年值", ["配偶星"], "male", pan, _ctx("丑")) == 0


def test_flow_clash_uses_extracted_day_palace() -> None:
    pan = _pan()
    base = _base_ctx_from_pan(pan)

    assert _liu_op("流年冲", ["配偶星"], "male", pan, _ctx("午", "六冲")) == 1
    assert _liu_op("流年合", ["配偶星"], "male", pan, _ctx("午", "六冲")) == 0


def test_flow_combination_uses_extracted_day_palace() -> None:
    pan = _pan()
    base = _base_ctx_from_pan(pan)

    assert _liu_op("流年合", ["配偶星"], "male", pan, _ctx("丑", "六合")) == 1
    assert _liu_op("流年冲", ["配偶星"], "male", pan, _ctx("丑", "六合")) == 0
