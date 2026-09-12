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
ATOMIC_STRING_FACTS = (
    "day_master_element", "day_master_stem", "day_branch",
    "month_longevity", "year_stem_ten_god", "month_main_ten_god",
    "hour_stem_ten_god", "spouse_palace_state", "day_branch_type",
)
ATOMIC_BOOLEAN_FACTS = (
    "pattern_god_transparent", "officer_killing_cleaned",
    "wealth_tomb_present", "wealth_star_in_tomb", "year_officer_killing",
)


def _require_nonempty_string(value) -> bool:
    return isinstance(value, str) and bool(value)


def _validate_engine_fact_items(action: str, full: dict) -> None:
    for key in ENGINE_LIST_FACTS:
        for index, item in enumerate(full[key]):
            if not isinstance(item, dict):
                raise PanSchemaError(f"{action} pan.full.{key}[{index}] 必须是 object。")

    for index, item in enumerate(full["relation_groups"]):
        if not _require_nonempty_string(item.get("field")) \
                or not _require_nonempty_string(item.get("group")):
            raise PanSchemaError(
                f"{action} pan.full.relation_groups[{index}] 缺少 field/group 字符串。"
            )

    for index, item in enumerate(full["lu_roots"]):
        if not all(_require_nonempty_string(item.get(key)) for key in ("shi_shen", "gan", "zhi")):
            raise PanSchemaError(
                f"{action} pan.full.lu_roots[{index}] 缺少 shi_shen/gan/zhi 字符串。"
            )

    for index, item in enumerate(full["ten_god_states"]):
        string_fields = ("shi_shen", "wuxing", "strength")
        bool_fields = ("transparent", "hidden", "rooted", "timely")
        if not all(_require_nonempty_string(item.get(key)) for key in string_fields):
            raise PanSchemaError(
                f"{action} pan.full.ten_god_states[{index}] 缺少字符串状态。"
            )
        if not all(isinstance(item.get(key), bool) for key in bool_fields):
            raise PanSchemaError(
                f"{action} pan.full.ten_god_states[{index}] 缺少布尔状态。"
            )
        count = item.get("count")
        if not isinstance(count, int) or isinstance(count, bool):
            raise PanSchemaError(
                f"{action} pan.full.ten_god_states[{index}].count 必须是整数。"
            )

    element_names = {item.get("wuxing") for item in full["element_states"]}
    if element_names != set(_CONSTANTS["五行"]):
        raise PanSchemaError(f"{action} pan.full.element_states 必须覆盖五行闭集。")
    for index, item in enumerate(full["element_states"]):
        string_fields = (
            "wuxing", "season_strength", "strength", "controls",
            "controlled_by", "controller_strength",
        )
        if not all(_require_nonempty_string(item.get(key)) for key in string_fields):
            raise PanSchemaError(
                f"{action} pan.full.element_states[{index}] 缺少字符串状态。"
            )
        if not all(isinstance(item.get(key), bool) for key in ("transparent", "rooted")):
            raise PanSchemaError(
                f"{action} pan.full.element_states[{index}] 缺少布尔状态。"
            )
        if not isinstance(item.get("reasons"), list):
            raise PanSchemaError(
                f"{action} pan.full.element_states[{index}].reasons 缺失或不是 array。"
            )


def _validate_yong_shen(action: str, full: dict) -> None:
    yong_shen = full.get("yong_shen")
    if not isinstance(yong_shen, dict):
        raise PanSchemaError(f"{action} pan.full.yong_shen 缺失或不是 object。")
    for school in ("fu_yi", "tiao_hou", "ge_ju"):
        if not isinstance(yong_shen.get(school), dict):
            raise PanSchemaError(f"{action} pan.full.yong_shen.{school} 缺失或不是 object。")
    fu_yi = yong_shen["fu_yi"]
    for key in ("wuxing_count", "wang_shuai"):
        if not isinstance(fu_yi.get(key), dict):
            raise PanSchemaError(f"{action} pan.full.yong_shen.fu_yi.{key} 缺失或不是 object。")
    for key in ("yong", "xi", "ji"):
        if not isinstance(fu_yi.get(key), str):
            raise PanSchemaError(
                f"{action} pan.full.yong_shen.fu_yi.{key} 缺失或不是字符串。"
            )
    if not _require_nonempty_string(fu_yi.get("qiangruo")):
        raise PanSchemaError(
            f"{action} pan.full.yong_shen.fu_yi.qiangruo 缺失或不是非空字符串。"
        )


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
    for key in ATOMIC_STRING_FACTS:
        if not _require_nonempty_string(atomic.get(key)):
            raise PanSchemaError(f"{action} pan.full.atomic_facts.{key} 缺失或不是非空字符串。")
    for key in ATOMIC_BOOLEAN_FACTS:
        if not isinstance(atomic.get(key), bool):
            raise PanSchemaError(f"{action} pan.full.atomic_facts.{key} 缺失或不是 boolean。")
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
    _validate_engine_fact_items(action, full)
    _validate_yong_shen(action, full)
    if not isinstance(pan["ziwei"].get("gong_wei"), list):
        raise PanSchemaError(f"{action} pan.ziwei.gong_wei 缺失或不是 array。")
    expected_palaces = tuple(_CONSTANTS["紫微宫位"])
    for index, palace in enumerate(pan["ziwei"]["gong_wei"]):
        if not isinstance(palace, dict) or not _require_nonempty_string(palace.get("name")):
            raise PanSchemaError(f"{action} pan.ziwei.gong_wei[{index}] 缺少 name 字符串。")
    actual_palaces = tuple(palace["name"] for palace in pan["ziwei"]["gong_wei"])
    if actual_palaces != expected_palaces:
        raise PanSchemaError(
            f"{action} pan.ziwei.gong_wei 必须按 engine 宫位闭集输出 12 宫。"
        )
    palace_facts = pan["ziwei"].get("palace_facts")
    if not isinstance(palace_facts, list):
        raise PanSchemaError(f"{action} pan.ziwei.palace_facts 缺失或不是 array。")
    if not palace_facts:
        raise PanSchemaError(f"{action} pan.ziwei.palace_facts 不能为空。")
    for index, fact in enumerate(palace_facts):
        if not isinstance(fact, dict) \
                or not _require_nonempty_string(fact.get("palace")) \
                or not _require_nonempty_string(fact.get("kind")) \
                or not _require_nonempty_string(fact.get("target")):
                raise PanSchemaError(
                    f"{action} pan.ziwei.palace_facts[{index}] 缺少 palace/kind/target 字符串。"
                )
        if fact["palace"] not in expected_palaces:
            raise PanSchemaError(
                f"{action} pan.ziwei.palace_facts[{index}].palace 不在 engine 宫位闭集。"
            )
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
