"""Bazi Python 工具的 LLM-facing result contract。

这里的 schema 只描述工具输出边界；pan 的深层 runtime 校验由 pan_schema.py 负责。
``scripts/generate_bazi_tool_contracts.py`` 用本文件生成 skill-tools.json，
避免 query / yearly_range / calibrate 手工维护多份重复断语契约。
"""
from __future__ import annotations

from typing import Any

from factor_constants import load_constants
from factor_tokens import FACTOR_WILDCARD

_SIDE_LABELS = load_constants()["命理侧"]["标签"]

FACTOR_VALUE_SCHEMA: dict[str, Any] = {
    "description": "因子值：0/1 命题、受控领域字符串；空字符串表示字符串因子不可用。",
    "not": {"const": FACTOR_WILDCARD},
    "oneOf": [
        {"type": "integer", "enum": [0, 1]},
        {"type": "string"},
    ],
}

TRACE_FACTOR_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "expected": FACTOR_VALUE_SCHEMA,
        "actual": FACTOR_VALUE_SCHEMA,
    },
    "required": ["expected", "actual"],
    "additionalProperties": False,
}

FACTOR_MAP_SCHEMA: dict[str, Any] = {
    "type": "object",
    "description": "因子名 -> 因子值。",
    "additionalProperties": FACTOR_VALUE_SCHEMA,
}

TRACE_FACTOR_MAP_SCHEMA: dict[str, Any] = {
    "type": "object",
    "description": "因子名 -> 命中证据。",
    "additionalProperties": TRACE_FACTOR_SCHEMA,
}

TRACE_ENTRY_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "condition_group": {"type": "integer", "minimum": 1},
        "factors": TRACE_FACTOR_MAP_SCHEMA,
    },
    "required": ["condition_group", "factors"],
    "additionalProperties": False,
}


CALIBRATION_HINT_SCHEMA: dict[str, Any] = {
    "type": "object",
    "description": "时辰交界提示；只表达窗口风险，不做换日或吉凶判断。",
    "properties": {
        "near_boundary": {"type": "boolean", "const": True},
        "minutes_to_boundary": {"type": "integer", "minimum": 0, "maximum": 30},
        "boundary_offset_minutes": {
            "type": "integer",
            "minimum": -30,
            "maximum": 30,
            "description": "负数=早于交界；正数=晚于交界。",
        },
        "boundary_solar": {"type": "string", "format": "date-time"},
        "threshold_minutes": {"type": "integer", "minimum": 0, "maximum": 30},
        "current_shichen": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "branch": {"type": "string"},
                "span": {"type": "string"},
            },
            "required": ["name", "branch", "span"],
            "additionalProperties": False,
        },
        "alternate_shichen": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "branch": {"type": "string"},
                "span": {"type": "string"},
            },
            "required": ["name", "branch", "span"],
            "additionalProperties": False,
        },
        "direction": {"type": "string", "enum": ["earlier", "later"]},
        "message": {"type": "string"},
    },
    "required": [
        "near_boundary",
        "minutes_to_boundary",
        "boundary_offset_minutes",
        "boundary_solar",
        "threshold_minutes",
        "current_shichen",
        "alternate_shichen",
        "direction",
        "message",
    ],
    "additionalProperties": False,
}


def _constraint_group_schema() -> dict[str, Any]:
    return {
        "type": "array",
        "description": "不同数组元素为 OR；同一对象内字段为 AND。",
        "items": {
            "type": "object",
            "additionalProperties": FACTOR_VALUE_SCHEMA,
        },
    }


def brief_assertion_hit_schema() -> dict[str, Any]:
    """detail=false 时的稳定精简断语契约。"""
    return {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "领域": {"type": "string"},
            "事件类型": {"type": "string"},
            "时间层": {"type": "string"},
            "事件": {"type": "string"},
            "结论": {"type": "string"},
        },
        "required": ["id", "领域", "事件类型", "时间层", "事件", "结论"],
        "additionalProperties": False,
    }


def detailed_assertion_hit_schema() -> dict[str, Any]:
    """query / detail=true 时的全量断语契约。"""
    return {
        "type": "object",
        "properties": {
            "id": {"type": "string"},
            "领域": {"type": "string"},
            "事件类型": {"type": "string"},
            "时间层": {"type": "string"},
            "事件": {"type": "string"},
            "约束组": _constraint_group_schema(),
            "结论": {"type": "string"},
            "依据": {"type": "string"},
            "经典依据": {"type": "string"},
            "trace": {"type": "array", "items": TRACE_ENTRY_SCHEMA},
        },
        "required": [
            "id", "领域", "事件类型", "时间层", "事件", "约束组",
            "结论", "依据", "经典依据", "trace",
        ],
        "additionalProperties": False,
    }


def assertion_result_schema(*, detailed: bool) -> dict[str, Any]:
    """固定三侧断语结果：八字 / 紫微 / 合参。"""
    item = (
        detailed_assertion_hit_schema() if detailed
        else brief_assertion_hit_schema()
    )
    side = {"type": "array", "items": item}
    return {
        "type": "object",
        "properties": {
            _SIDE_LABELS["bazi"]: side,
            _SIDE_LABELS["ziwei"]: side,
            _SIDE_LABELS["common"]: side,
        },
        "required": [_SIDE_LABELS["bazi"], _SIDE_LABELS["ziwei"], _SIDE_LABELS["common"]],
        "additionalProperties": False,
    }


def query_result_schema() -> dict[str, Any]:
    """query.data：三侧全量断语，及用神 / 限运特有上下文。"""
    result = assertion_result_schema(detailed=True)
    return {
        "type": "object",
        "properties": {
            **result["properties"],
            "yong_shen_context": {
                "description": "仅在 rule=用神 时由 runtime 返回。",
                "type": "object",
                "properties": {
                    "yong_shen": {"type": "object"},
                    "element_states": {"type": "array"},
                    "ten_god_states": {"type": "array"},
                },
                "required": ["yong_shen", "element_states", "ten_god_states"],
                "additionalProperties": False,
            },
            "current_year": {"type": "integer", "minimum": 1},
            "current_year_source": {
                "type": "string",
                "enum": ["specified", "server"],
                "description": "仅在 rule=大运 / 大限 时由 runtime 返回。",
            },
        },
        "required": result["required"],
        "additionalProperties": False,
    }


def yearly_result_schema() -> dict[str, Any]:
    """yearly_range.data：年份 key -> 规则结果 map；单年失败用 error 对象。"""
    brief = assertion_result_schema(detailed=False)
    detailed = assertion_result_schema(detailed=True)
    rule_results = {
        "type": "object",
        "description": "key=命理域或展开后的流年域；value=该域三侧断语结果。",
        "additionalProperties": {"anyOf": [brief, detailed]},
        "minProperties": 1,
    }
    year_error = {
        "type": "object",
        "properties": {"error": {"type": "string", "minLength": 1}},
        "required": ["error"],
        "additionalProperties": False,
    }
    return {
        "type": "object",
        "properties": {
            "current_year": {"type": "integer", "minimum": 1},
            "current_year_source": {"type": "string", "enum": ["specified", "server"]},
            "year_basis": {
                "type": "object",
                "properties": {
                    _SIDE_LABELS["bazi"]: {"type": "string"},
                    _SIDE_LABELS["ziwei"]: {"type": "string"},
                    "usage": {"type": "string"},
                },
                "required": [_SIDE_LABELS["bazi"], _SIDE_LABELS["ziwei"], "usage"],
                "additionalProperties": False,
            },
            "years": {
                "type": "object",
                "propertyNames": {"pattern": "^[0-9]{4}$"},
                "additionalProperties": {
                    "if": {"required": ["error"]},
                    "then": year_error,
                    "else": rule_results,
                },
            },
        },
        "required": ["current_year", "current_year_source", "year_basis", "years"],
        "additionalProperties": False,
    }


def calibrate_result_schema() -> dict[str, Any]:
    """calibrate.data：候选 label -> 事件报告数组。"""
    hit = {
        "oneOf": [
            brief_assertion_hit_schema(),
            detailed_assertion_hit_schema(),
        ]
    }
    side = {"type": "array", "items": hit}
    event = {
        "type": "object",
        "properties": {
            "year": {"type": "integer"},
            "label": {"type": "string"},
            "rule": {"type": "string"},
            _SIDE_LABELS["bazi"]: side,
            _SIDE_LABELS["ziwei"]: side,
            _SIDE_LABELS["common"]: side,
        },
        "required": [
            "year", "label", "rule",
            _SIDE_LABELS["bazi"], _SIDE_LABELS["ziwei"], _SIDE_LABELS["common"],
        ],
        "additionalProperties": False,
    }
    return {
        "type": "object",
        "description": "key=候选 label；value=按事件顺序排列的校准报告。",
        "minProperties": 2,
        "maxProperties": 3,
        "additionalProperties": {
            "type": "array",
            "items": event,
        },
    }


def full_paipan_surface_result_schema() -> dict[str, Any]:
    """full_paipan 的 LLM surface contract；深层 runtime 校验由 pan_schema.py 执行。"""
    return {
        "type": "object",
        "description": "完整本命盘。必须原样传回 query/yearly_range/bond/calibrate，不得裁剪或修改。",
        "properties": {
            "solar": {"type": "string", "format": "date-time"},
            "lunar": {"type": "object"},
            "chart": {"type": "object"},
            "full": {"type": "object"},
            "ziwei": {"type": "object"},
            "ziwei_daxian": {"type": "array"},
            "gender": {"type": "string", "enum": ["male", "female"]},
            "pan_digest": {"type": "string", "minLength": 64, "maxLength": 64},
            "calibration_hint": CALIBRATION_HINT_SCHEMA,
        },
        "required": [
            "solar", "lunar", "chart", "full", "ziwei", "ziwei_daxian",
            "gender", "pan_digest",
        ],
        "additionalProperties": False,
    }


def city_result_schema() -> dict[str, Any]:
    return {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "longitude": {"type": "number"},
            "latitude": {"type": "number"},
            "country": {"type": "string"},
        },
        "required": ["name", "longitude", "latitude", "country"],
        "additionalProperties": False,
    }


def bond_result_schema() -> dict[str, Any]:
    return {
        "type": "object",
        "properties": {
            "bazi": {
                "type": "object",
                "properties": {
                    "zhu_cross": {"type": "object"},
                    "shi_shen_cross": {"type": "object"},
                    "structure": {"type": "object"},
                    "nayin_cross": {"type": "object"},
                    "shensha_cross": {"type": "object"},
                },
                "required": [
                    "zhu_cross", "shi_shen_cross", "structure",
                    "nayin_cross", "shensha_cross",
                ],
                "additionalProperties": False,
            },
            "ziwei": {
                "type": "object",
                "properties": {
                    "jia_gong_name": {"type": "string"},
                    "yi_gong_name": {"type": "string"},
                    "ming_gong_hu_ru": {"type": "object"},
                    "fu_qi_gong": {"type": "object"},
                    "zi_nv_gong": {"type": "object"},
                    "ji_xing": {"type": "array"},
                    "sha_xing": {"type": "array"},
                    "lu_ma_ru": {"type": "array"},
                    "kong_wang": {"type": "array"},
                    "si_hua_ru": {"type": "array"},
                    "wu_xing_sheng_ke": {"type": "string"},
                    "summary": {"type": "string"},
                },
                "required": [
                    "jia_gong_name", "yi_gong_name", "ming_gong_hu_ru",
                    "fu_qi_gong", "zi_nv_gong", "ji_xing", "sha_xing",
                    "lu_ma_ru", "kong_wang", "si_hua_ru",
                    "wu_xing_sheng_ke", "summary",
                ],
                "additionalProperties": False,
            },
        },
        "required": ["bazi", "ziwei"],
        "additionalProperties": False,
    }
