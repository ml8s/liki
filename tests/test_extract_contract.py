"""提取层契约测试：extract(pan) 必须产出因子层实际消费的 fac 键。"""
import inspect

import _helpers  # noqa: F401 —— 注入 tools 路径
from extract import extract
from factors import _liu_op, _op


def _pan(ri_zhi: str = "子") -> dict:
    return {
        "full": {
            "nian": {"gan": "庚", "zhi": "午", "shi_shens": []},
            "yue": {"gan": "壬", "zhi": "午", "shi_shens": []},
            "ri": {"gan": "甲", "zhi": ri_zhi, "shi_shens": []},
            "shi": {"gan": "庚", "zhi": "午", "shi_shens": []},
        },
        "yongshen": {"fu_yi": {}},
        "chart": {"ri": {"gan": "甲", "zhi": ri_zhi}},
    }


def test_extract_outputs_day_palace_for_flow_operators() -> None:
    fac = extract(_pan("子"))

    assert fac["palace_ri"] == {"zhi": "子"}


def test_operator_parameters_use_fac_entity_name() -> None:
    assert list(inspect.signature(_op).parameters)[2] == "fac"
    assert list(inspect.signature(_liu_op).parameters)[2] == "fac"
