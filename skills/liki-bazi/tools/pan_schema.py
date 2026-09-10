"""pan 契约层 — 只接受 full_paipan 的完整返回结构。"""
from __future__ import annotations

from errors import PanSchemaError
from factor_constants import load_constants
from pan_integrity import DIGEST_FIELD, validate_natal_digest

_CONSTANTS = load_constants()
REQUIRED_FIELDS = (
    "solar", "lunar", "chart", "full", "ziwei",
    "ziwei_daxian", "gender", DIGEST_FIELD,
)
PILLARS = tuple(_CONSTANTS["四柱"])
GENDERS = tuple(_CONSTANTS["性别闭集"])
ENGINE_LIST_FACTS = ("lu_roots", "relation_groups", "ten_god_states", "element_states")


def validate_natal_pan(pan: object, action: str = "pan") -> None:
    """拒绝快照、裁剪盘和半截盘；完整盘契约由本层统一维护。"""
    if not isinstance(pan, dict):
        raise PanSchemaError(
            f"{action} pan 必须是 full_paipan 返回的完整本命盘 object。"
            "流年查询请走 yearly_range。"
        )
    missing = [key for key in REQUIRED_FIELDS if key not in pan]
    if missing:
        raise PanSchemaError(
            f"{action} pan 必须是 full_paipan 返回的完整本命盘，缺少字段: "
            f"{', '.join(missing)}。禁止传快照或裁剪盘。"
        )
    wrong = [
        key for key in REQUIRED_FIELDS
        if key not in ("solar", "gender", "ziwei_daxian", DIGEST_FIELD)
        and not isinstance(pan[key], dict)
    ]
    if wrong:
        raise PanSchemaError(
            f"{action} pan 字段类型不完整: {', '.join(wrong)} 必须是 object。"
        )
    if not isinstance(pan["solar"], str) or not pan["solar"]:
        raise PanSchemaError(f"{action} pan.solar 必须是非空字符串。")
    if pan["gender"] not in GENDERS:
        raise PanSchemaError(
            f"{action} pan.gender 必须是 male/female，收到: {pan['gender']}"
        )

    chart, full = pan["chart"], pan["full"]
    missing_pillars = [
        key for key in PILLARS
        if not isinstance(chart.get(key), dict)
        or not isinstance(full.get(key), dict)
        or not full[key].get("gan")
        or not full[key].get("zhi")
    ]
    if missing_pillars:
        raise PanSchemaError(
            f"{action} pan 四柱结构不完整，缺少: {', '.join(missing_pillars)}。"
            "请重新执行 full_paipan，不要手工拼装半截盘。"
        )
    if not isinstance(chart.get("da_yun"), dict):
        raise PanSchemaError(f"{action} pan.chart.da_yun 缺失或不是 object。")
    for key in ENGINE_LIST_FACTS:
        if not isinstance(full.get(key), list):
            raise PanSchemaError(f"{action} pan.full.{key} 缺失或不是 array。")
    if not full["ten_god_states"]:
        raise PanSchemaError(f"{action} pan.full.ten_god_states 不能为空。")
    if len(full["element_states"]) != len(_CONSTANTS["五行"]):
        raise PanSchemaError(f"{action} pan.full.element_states 必须覆盖五行闭集。")
    atomic = full.get("atomic_facts")
    if not isinstance(atomic, dict) or not atomic.get("day_master_element"):
        raise PanSchemaError(f"{action} pan.full.atomic_facts.day_master_element 缺失。")
    for key in ("day_master_stem", "day_branch"):
        if not isinstance(atomic.get(key), str):
            raise PanSchemaError(f"{action} pan.full.atomic_facts.{key} 缺失或不是 string。")
    for key in (
        "month_longevity", "year_stem_ten_god", "month_main_ten_god", "hour_stem_ten_god",
    ):
        if not isinstance(atomic.get(key), str):
            raise PanSchemaError(f"{action} pan.full.atomic_facts.{key} 缺失或不是 string。")
    if not isinstance(atomic.get("pattern_god_transparent"), bool):
        raise PanSchemaError(f"{action} pan.full.atomic_facts.pattern_god_transparent 缺失或不是 boolean。")
    if not isinstance(atomic.get("pillar_punishments"), dict):
        raise PanSchemaError(f"{action} pan.full.atomic_facts.pillar_punishments 缺失或不是 object。")
    full_dayun = full.get("da_yun")
    if not isinstance(full_dayun, dict) or not isinstance(full_dayun.get("steps"), list):
        raise PanSchemaError(f"{action} pan.full.da_yun.steps 缺失或不是 array。")
    for index, step in enumerate(full_dayun["steps"]):
        if not isinstance(step, dict) or not isinstance(step.get("rooted"), bool):
            raise PanSchemaError(
                f"{action} pan.full.da_yun.steps[{index}] 缺少 rooted engine 事实。"
            )
        if step["rooted"] and not isinstance(step.get("root_refs"), list):
            raise PanSchemaError(
                f"{action} pan.full.da_yun.steps[{index}].root_refs 缺失或不是 array。"
            )
    if not isinstance(pan["ziwei"].get("gong_wei"), list):
        raise PanSchemaError(f"{action} pan.ziwei.gong_wei 缺失或不是 array。")
    if not isinstance(pan["ziwei_daxian"], list):
        raise PanSchemaError(f"{action} pan.ziwei_daxian 缺失或不是 array。")
    daxian_count = _CONSTANTS["大限段数"]
    if len(pan["ziwei_daxian"]) != daxian_count:
        raise PanSchemaError(
            f"{action} pan.ziwei_daxian 必须包含 {daxian_count} 个大限段。"
        )
    daxian_palaces = []
    for index, step in enumerate(pan["ziwei_daxian"]):
        if not isinstance(step, dict):
            raise PanSchemaError(f"{action} pan.ziwei_daxian[{index}] 必须是 object。")
        missing_step = [
            key for key in ("gong", "name", "start_year", "end_year", "qi_sui", "zhi_sui")
            if key not in step
        ]
        if missing_step:
            raise PanSchemaError(
                f"{action} pan.ziwei_daxian[{index}] 缺少字段: {', '.join(missing_step)}。"
            )
        if not isinstance(step["gong"], str) or not step["gong"].strip():
            raise PanSchemaError(f"{action} pan.ziwei_daxian[{index}].gong 必须是非空字符串。")
        if not isinstance(step["name"], str) or not step["name"].strip():
            raise PanSchemaError(f"{action} pan.ziwei_daxian[{index}].name 必须是非空字符串。")
        integer_fields = []
        for key in ("start_year", "end_year", "qi_sui", "zhi_sui"):
            if not isinstance(step[key], int) or isinstance(step[key], bool):
                integer_fields.append(key)
        if integer_fields:
            raise PanSchemaError(
                f"{action} pan.ziwei_daxian[{index}] 字段必须为整数: {', '.join(integer_fields)}。"
            )
        if step["start_year"] > step["end_year"] or step["qi_sui"] > step["zhi_sui"]:
            raise PanSchemaError(f"{action} pan.ziwei_daxian[{index}] 年段/虚岁区间倒置。")
        daxian_palaces.append(step["gong"])
    if len(set(daxian_palaces)) != daxian_count:
        raise PanSchemaError(
            f"{action} pan.ziwei_daxian 的 {daxian_count} 个大限宫位必须唯一。"
        )
    validate_natal_digest(pan, action=action)
