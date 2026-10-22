"""natal_chart 契约层 — 只接受 build_natal_chart 的完整返回结构。"""

from app.natal.tools.chart_integrity import DIGEST_FIELD, validate_natal_chart_digest
from app.natal.tools.errors import ChartSchemaError
from app.natal.tools.factor_constants import load_constants

_CONSTANTS = load_constants()
REQUIRED_FIELDS = (
    "instant", "timezone", "chart", "full", "ziwei",
    "ziwei_chart", "ziwei_daxian", "gender", DIGEST_FIELD,
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


def _validate_optional_element(action: str, value: object, label: str) -> None:
    if value is None:
        return
    if not isinstance(value, str) or value not in _CONSTANTS["五行"]:
        raise ChartSchemaError(f"{action} {label} 必须是 null 或五行闭集字符串。")


def _validate_engine_fact_items(action: str, full: dict) -> None:
    for key in ENGINE_LIST_FACTS:
        for index, item in enumerate(full[key]):
            if not isinstance(item, dict):
                raise ChartSchemaError(f"{action} natal_chart.full.{key}[{index}] 必须是 object。")

    for index, item in enumerate(full["relation_groups"]):
        if not _require_nonempty_string(item.get("field")) \
                or not _require_nonempty_string(item.get("group")):
            raise ChartSchemaError(
                f"{action} natal_chart.full.relation_groups[{index}] 缺少 field/group 字符串。"
            )

    for index, item in enumerate(full["lu_roots"]):
        if not all(_require_nonempty_string(item.get(key)) for key in ("shi_shen", "gan", "zhi")):
            raise ChartSchemaError(
                f"{action} natal_chart.full.lu_roots[{index}] 缺少 shi_shen/gan/zhi 字符串。"
            )

    for index, item in enumerate(full["ten_god_states"]):
        string_fields = ("shi_shen", "wuxing", "strength")
        bool_fields = ("transparent", "hidden", "rooted", "timely")
        if not all(_require_nonempty_string(item.get(key)) for key in string_fields):
            raise ChartSchemaError(
                f"{action} natal_chart.full.ten_god_states[{index}] 缺少字符串状态。"
            )
        if not all(isinstance(item.get(key), bool) for key in bool_fields):
            raise ChartSchemaError(
                f"{action} natal_chart.full.ten_god_states[{index}] 缺少布尔状态。"
            )
        count = item.get("count")
        if not isinstance(count, int) or isinstance(count, bool):
            raise ChartSchemaError(
                f"{action} natal_chart.full.ten_god_states[{index}].count 必须是整数。"
            )

    element_names = {item.get("wuxing") for item in full["element_states"]}
    if element_names != set(_CONSTANTS["五行"]):
        raise ChartSchemaError(f"{action} natal_chart.full.element_states 必须覆盖五行闭集。")
    for index, item in enumerate(full["element_states"]):
        string_fields = (
            "wuxing", "season_strength", "strength", "controls",
            "controlled_by", "controller_strength",
        )
        if not all(_require_nonempty_string(item.get(key)) for key in string_fields):
            raise ChartSchemaError(
                f"{action} natal_chart.full.element_states[{index}] 缺少字符串状态。"
            )
        if not all(isinstance(item.get(key), bool) for key in ("transparent", "rooted")):
            raise ChartSchemaError(
                f"{action} natal_chart.full.element_states[{index}] 缺少布尔状态。"
            )
        if not isinstance(item.get("reasons"), list):
            raise ChartSchemaError(
                f"{action} natal_chart.full.element_states[{index}].reasons 缺失或不是 array。"
            )


def _validate_yongshen_schools(action: str, full: dict) -> None:
    for school in ("fu_yi", "tiao_hou", "ge_ju"):
        if not isinstance(full.get(school), dict):
            raise ChartSchemaError(f"{action} natal_chart.full.{school} 缺失或不是 object。")
    fu_yi = full.get("fu_yi")
    tiao_hou = full.get("tiao_hou")
    ge_ju = full.get("ge_ju")
    if not isinstance(fu_yi, dict):
            raise ChartSchemaError(f"{action} natal_chart.full.{school} 缺失或不是 object。")
    fu_yi = full.get("fu_yi")
    for key in ("wuxing_count", "wang_shuai"):
        if not isinstance(fu_yi.get(key), dict):
            raise ChartSchemaError(f"{action} natal_chart.full.fu_yi.{key} 缺失或不是 object。")
    for key in ("yong", "xi", "ji"):
        if key not in fu_yi:
            raise ChartSchemaError(
                f"{action} natal_chart.full.fu_yi.{key} 缺失。"
            )
        _validate_optional_element(
            action, fu_yi.get(key), f"natal_chart.full.fu_yi.{key}"
        )
    if not _require_nonempty_string(fu_yi.get("qiangruo")):
        raise ChartSchemaError(
            f"{action} natal_chart.full.fu_yi.qiangruo 缺失或不是非空字符串。"
        )
    _validate_fu_yi_basis(action, fu_yi.get("basis"))

    tiao_hou = full.get("tiao_hou")
    stale_tiao_hou = [key for key in ("yong", "xi") if key in tiao_hou]
    if stale_tiao_hou:
        raise ChartSchemaError(
            f"{action} natal_chart.full.tiao_hou 不应输出旧用神字段: {', '.join(stale_tiao_hou)}"
        )
    if not _require_nonempty_string(tiao_hou.get("primary_wuxing")):
        raise ChartSchemaError(
            f"{action} natal_chart.full.tiao_hou.primary_wuxing 缺失或不是非空字符串。"
        )
    if "secondary_wuxing" not in tiao_hou:
        raise ChartSchemaError(f"{action} natal_chart.full.tiao_hou.secondary_wuxing 缺失。")
    _validate_optional_element(
        action, tiao_hou.get("secondary_wuxing"),
        "natal_chart.full.tiao_hou.secondary_wuxing",
    )
    _validate_tiao_hou_availability(action, tiao_hou.get("primary"), "primary")
    secondary_condition = tiao_hou.get("secondary_condition")
    if secondary_condition is not None and not isinstance(secondary_condition, str):
        raise ChartSchemaError(
            f"{action} natal_chart.full.tiao_hou.secondary_condition 不是 string。"
        )
    if secondary_condition or tiao_hou.get("secondary_wuxing") is not None:
        _validate_tiao_hou_availability(action, tiao_hou.get("secondary"), "secondary")
    if secondary_condition and not tiao_hou.get("secondary"):
        raise ChartSchemaError(
            f"{action} natal_chart.full.tiao_hou.secondary_condition 缺少 secondary。"
        )

    ge_ju = full.get("ge_ju")
    stale = [key for key in ("yong", "xi", "ji") if key in ge_ju]
    if stale:
        raise ChartSchemaError(
            f"{action} natal_chart.full.ge_ju 不应输出旧用神字段: {', '.join(stale)}"
        )
    for key in ("ge_ju", "yong_fa", "pattern_god_source"):
        if not _require_nonempty_string(ge_ju.get(key)):
            raise ChartSchemaError(
                f"{action} natal_chart.full.ge_ju.{key} 缺失或不是非空字符串。"
            )
    structure = ge_ju.get("structure")
    if not isinstance(structure, dict):
        raise ChartSchemaError(f"{action} natal_chart.full.ge_ju.structure 缺失或不是 object。")
    for element in ("pattern", "controller", "generator"):
        _validate_element_structure(action, structure.get(element), element)
    _validate_relation_facts(action, structure.get("relation_facts"), "ge_ju.structure")


def _validate_fu_yi_basis(action: str, basis: object) -> None:
    if not isinstance(basis, dict):
        raise ChartSchemaError(f"{action} natal_chart.full.fu_yi.basis 缺失或不是 object。")
    for key in ("root_type", "season"):
        if not _require_nonempty_string(basis.get(key)):
            raise ChartSchemaError(f"{action} natal_chart.full.fu_yi.basis.{key} 缺失或不是非空字符串。")
    if not isinstance(basis.get("yin_bi_count"), int):
        raise ChartSchemaError(f"{action} natal_chart.full.fu_yi.basis.yin_bi_count 缺失或不是 integer。")
    _validate_stem_occurrences(action, basis.get("day_master_roots"), "fu_yi.basis.day_master_roots")
    _validate_relation_facts(action, basis.get("relation_facts"), "fu_yi.basis")


def _validate_tiao_hou_availability(action: str, value: object, name: str) -> None:
    if not isinstance(value, dict):
        raise ChartSchemaError(f"{action} natal_chart.full.tiao_hou.{name} 缺失或不是 object。")
    if not _require_nonempty_string(value.get("stem")):
        raise ChartSchemaError(f"{action} natal_chart.full.tiao_hou.{name}.stem 缺失或不是非空字符串。")
    for key in ("transparent", "hidden"):
        if not isinstance(value.get(key), bool):
            raise ChartSchemaError(f"{action} natal_chart.full.tiao_hou.{name}.{key} 缺失或不是 boolean。")
    _validate_stem_occurrences(action, value.get("occurrences"), f"tiao_hou.{name}.occurrences")
    _validate_relation_facts(action, value.get("relation_facts"), f"tiao_hou.{name}")


def _validate_element_structure(action: str, value: object, name: str) -> None:
    if not isinstance(value, dict):
        raise ChartSchemaError(f"{action} natal_chart.full.ge_ju.structure.{name} 缺失或不是 object。")
    for key in ("wuxing", "season", "strength"):
        if not _require_nonempty_string(value.get(key)):
            raise ChartSchemaError(
                f"{action} natal_chart.full.ge_ju.structure.{name}.{key} 缺失或不是非空字符串。"
            )
    if not isinstance(value.get("timely"), bool):
        raise ChartSchemaError(
            f"{action} natal_chart.full.ge_ju.structure.{name}.timely 缺失或不是 boolean。"
        )
    if not isinstance(value.get("transparent_pillars"), list):
        raise ChartSchemaError(
            f"{action} natal_chart.full.ge_ju.structure.{name}.transparent_pillars 缺失或不是 array。"
        )
    _validate_stem_occurrences(action, value.get("roots"), f"ge_ju.structure.{name}.roots")


def _validate_stem_occurrences(action: str, value: object, name: str) -> None:
    if not isinstance(value, list):
        raise ChartSchemaError(f"{action} natal_chart.full.{name} 缺失或不是 array。")
    for index, item in enumerate(value):
        if not isinstance(item, dict):
            raise ChartSchemaError(f"{action} natal_chart.full.{name}[{index}] 不是 object。")
        if not all(_require_nonempty_string(item.get(key)) for key in ("pillar", "branch", "stem", "source")):
            raise ChartSchemaError(f"{action} natal_chart.full.{name}[{index}] 缺少完整根字段。")


def _validate_relation_facts(action: str, value: object, name: str) -> None:
    if not isinstance(value, list):
        raise ChartSchemaError(f"{action} natal_chart.full.{name}.relation_facts 缺失或不是 array。")
    for index, item in enumerate(value):
        if not isinstance(item, dict):
            raise ChartSchemaError(f"{action} natal_chart.full.{name}.relation_facts[{index}] 不是 object。")
        if not all(_require_nonempty_string(item.get(key)) for key in ("field", "group")):
            raise ChartSchemaError(f"{action} natal_chart.full.{name}.relation_facts[{index}] 缺少 field/group。")
        for key in ("pillars", "branches", "targets"):
            if not isinstance(item.get(key), list):
                raise ChartSchemaError(
                    f"{action} natal_chart.full.{name}.relation_facts[{index}].{key} 缺失或不是 array。"
                )


def validate_natal_chart(natal_chart: object, action: str = "natal_chart") -> None:
    """拒绝快照、裁剪盘和半截盘；完整盘契约由本层统一维护。"""
    if not isinstance(natal_chart, dict):
        raise ChartSchemaError(
            f"{action} natal_chart 必须是 build_natal_chart 返回的完整本命盘 object。"
            "流年查询请走 yearly_range。"
        )
    missing = [key for key in REQUIRED_FIELDS if key not in natal_chart]
    if missing:
        raise ChartSchemaError(
            f"{action} natal_chart 必须是 build_natal_chart 返回的完整本命盘，缺少字段: "
            f"{', '.join(missing)}。禁止传快照或裁剪盘。"
        )
    wrong = [
        key for key in REQUIRED_FIELDS
        if key not in ("instant", "timezone", "gender", "ziwei_daxian", DIGEST_FIELD)
        and not isinstance(natal_chart[key], dict)
    ]
    if wrong:
        raise ChartSchemaError(
            f"{action} natal_chart 字段类型不完整: {', '.join(wrong)} 必须是 object。"
        )
    for key in ("instant", "timezone"):
        if not isinstance(natal_chart.get(key), str) or not natal_chart.get(key):
            raise ChartSchemaError(f"{action} natal_chart.{key} 必须是非空字符串。")
    if natal_chart["gender"] not in GENDERS:
        raise ChartSchemaError(
            f"{action} natal_chart.gender 必须是 male/female，收到: {natal_chart['gender']}"
        )

    chart, full = natal_chart["chart"], natal_chart["full"]
    missing_pillars = [
        key for key in PILLARS
        if not isinstance(chart.get(key), dict)
        or not isinstance(full.get(key), dict)
        or not full[key].get("gan")
        or not full[key].get("zhi")
    ]
    if missing_pillars:
        raise ChartSchemaError(
            f"{action} natal_chart 四柱结构不完整，缺少: {', '.join(missing_pillars)}。"
            "请重新执行 build_natal_chart，不要手工拼装半截盘。"
        )
    if not isinstance(chart.get("da_yun"), dict):
        raise ChartSchemaError(f"{action} natal_chart.chart.da_yun 缺失或不是 object。")
    for key in ENGINE_LIST_FACTS:
        if not isinstance(full.get(key), list):
            raise ChartSchemaError(f"{action} natal_chart.full.{key} 缺失或不是 array。")
    if not full["ten_god_states"]:
        raise ChartSchemaError(f"{action} natal_chart.full.ten_god_states 不能为空。")
    atomic = full.get("atomic_facts")
    if not isinstance(atomic, dict) or not atomic.get("day_master_element"):
        raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.day_master_element 缺失。")
    for key in ("day_master_stem", "day_branch"):
        if not isinstance(atomic.get(key), str):
            raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.{key} 缺失或不是 string。")
    for key in (
        "month_longevity", "year_stem_ten_god", "month_main_ten_god", "hour_stem_ten_god",
    ):
        if not isinstance(atomic.get(key), str):
            raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.{key} 缺失或不是 string。")
    for key in ATOMIC_STRING_FACTS:
        if not _require_nonempty_string(atomic.get(key)):
            raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.{key} 缺失或不是非空字符串。")
    for key in ATOMIC_BOOLEAN_FACTS:
        if not isinstance(atomic.get(key), bool):
            raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.{key} 缺失或不是 boolean。")
    if not isinstance(atomic.get("pattern_god_transparent"), bool):
        raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.pattern_god_transparent 缺失或不是 boolean。")
    if not isinstance(atomic.get("pillar_punishments"), dict):
        raise ChartSchemaError(f"{action} natal_chart.full.atomic_facts.pillar_punishments 缺失或不是 object。")
    full_dayun = full.get("da_yun")
    if not isinstance(full_dayun, dict) or not isinstance(full_dayun.get("steps"), list):
        raise ChartSchemaError(f"{action} natal_chart.full.da_yun.steps 缺失或不是 array。")
    for index, step in enumerate(full_dayun["steps"]):
        if not isinstance(step, dict) or not isinstance(step.get("rooted"), bool):
            raise ChartSchemaError(
                f"{action} natal_chart.full.da_yun.steps[{index}] 缺少 rooted engine 事实。"
            )
        if step["rooted"] and not isinstance(step.get("root_refs"), list):
            raise ChartSchemaError(
                f"{action} natal_chart.full.da_yun.steps[{index}].root_refs 缺失或不是 array。"
            )
    _validate_engine_fact_items(action, full)
    _validate_yongshen_schools(action, full)
    if not isinstance(natal_chart["ziwei"].get("gong_wei"), list):
        raise ChartSchemaError(f"{action} natal_chart.ziwei.gong_wei 缺失或不是 array。")
    expected_palaces = tuple(_CONSTANTS["紫微宫位"])
    for index, palace in enumerate(natal_chart["ziwei"]["gong_wei"]):
        if not isinstance(palace, dict) or not _require_nonempty_string(palace.get("name")):
            raise ChartSchemaError(f"{action} natal_chart.ziwei.gong_wei[{index}] 缺少 name 字符串。")
    actual_palaces = tuple(palace["name"] for palace in natal_chart["ziwei"]["gong_wei"])
    if actual_palaces != expected_palaces:
        raise ChartSchemaError(
            f"{action} natal_chart.ziwei.gong_wei 必须按 engine 宫位闭集输出 12 宫。"
        )
    palace_facts = natal_chart["ziwei"].get("palace_facts")
    if not isinstance(palace_facts, list):
        raise ChartSchemaError(f"{action} natal_chart.ziwei.palace_facts 缺失或不是 array。")
    if not palace_facts:
        raise ChartSchemaError(f"{action} natal_chart.ziwei.palace_facts 不能为空。")
    for index, fact in enumerate(palace_facts):
        if not isinstance(fact, dict) \
                or not _require_nonempty_string(fact.get("palace")) \
                or not _require_nonempty_string(fact.get("kind")) \
                or not _require_nonempty_string(fact.get("target")):
                raise ChartSchemaError(
                    f"{action} natal_chart.ziwei.palace_facts[{index}] 缺少 palace/kind/target 字符串。"
                )
        if fact["palace"] not in expected_palaces:
            raise ChartSchemaError(
                f"{action} natal_chart.ziwei.palace_facts[{index}].palace 不在 engine 宫位闭集。"
            )
    if not isinstance(natal_chart["ziwei_daxian"], list):
        raise ChartSchemaError(f"{action} natal_chart.ziwei_daxian 缺失或不是 array。")
    daxian_count = _CONSTANTS["大限段数"]
    if len(natal_chart["ziwei_daxian"]) != daxian_count:
        raise ChartSchemaError(
            f"{action} natal_chart.ziwei_daxian 必须包含 {daxian_count} 个大限段。"
        )
    daxian_palaces = []
    for index, step in enumerate(natal_chart["ziwei_daxian"]):
        if not isinstance(step, dict):
            raise ChartSchemaError(f"{action} natal_chart.ziwei_daxian[{index}] 必须是 object。")
        missing_step = [
            key for key in ("gong", "name", "start_year", "end_year", "qi_sui", "zhi_sui")
            if key not in step
        ]
        if missing_step:
            raise ChartSchemaError(
                f"{action} natal_chart.ziwei_daxian[{index}] 缺少字段: {', '.join(missing_step)}。"
            )
        if not isinstance(step["gong"], str) or not step["gong"].strip():
            raise ChartSchemaError(f"{action} natal_chart.ziwei_daxian[{index}].gong 必须是非空字符串。")
        if not isinstance(step["name"], str) or not step["name"].strip():
            raise ChartSchemaError(f"{action} natal_chart.ziwei_daxian[{index}].name 必须是非空字符串。")
        integer_fields = []
        for key in ("start_year", "end_year", "qi_sui", "zhi_sui"):
            if not isinstance(step[key], int) or isinstance(step[key], bool):
                integer_fields.append(key)
        if integer_fields:
            raise ChartSchemaError(
                f"{action} natal_chart.ziwei_daxian[{index}] 字段必须为整数: {', '.join(integer_fields)}。"
            )
        if step["start_year"] > step["end_year"] or step["qi_sui"] > step["zhi_sui"]:
            raise ChartSchemaError(f"{action} natal_chart.ziwei_daxian[{index}] 年段/虚岁区间倒置。")
        daxian_palaces.append(step["gong"])
    if len(set(daxian_palaces)) != daxian_count:
        raise ChartSchemaError(
            f"{action} natal_chart.ziwei_daxian 的 {daxian_count} 个大限宫位必须唯一。"
        )
    validate_natal_chart_digest(natal_chart, action=action)


def validate_side_chart(natal_chart: object, side: str, action: str = "compute_factors") -> None:
    """校验正交分侧盘（compute_factors / period_query 的输入）。

    与 validate_natal_chart（完整本命盘）不同：分侧盘只含单侧引擎数据
    （bazi: chart+full；ziwei: chart+ziwei+ziwei_daxian）。这里对分侧必需结构
    fail-closed，避免 engine 字段漂移时因子静默算成 0。
    """
    if not isinstance(natal_chart, dict):
        raise ChartSchemaError(f"{action} natal_chart 必须是 object。")
    gender = natal_chart.get("gender")
    if gender not in GENDERS:
        raise ChartSchemaError(
            f"{action} natal_chart.gender 必须是 male/female，收到: {gender!r}"
        )
    chart = natal_chart.get("chart")
    if not isinstance(chart, dict) or not chart:
        raise ChartSchemaError(f"{action} natal_chart.chart 必须是非空对象。")

    if side == "bazi":
        full = natal_chart.get("full")
        if not isinstance(full, dict) or not full:
            raise ChartSchemaError(f"{action} bazi natal_chart.full 必须是非空对象。")
        missing_pillars = [
            key for key in PILLARS
            if not isinstance(chart.get(key), dict)
            or not isinstance(full.get(key), dict)
            or not full[key].get("gan")
            or not full[key].get("zhi")
        ]
        if missing_pillars:
            raise ChartSchemaError(
                f"{action} bazi 四柱结构不完整，缺少: {', '.join(missing_pillars)}。"
            )
    elif side == "ziwei":
        ziwei = natal_chart.get("ziwei")
        if not isinstance(ziwei, dict) or not ziwei:
            raise ChartSchemaError(f"{action} ziwei natal_chart.ziwei 必须是非空对象。")
        gong_wei = ziwei.get("gong_wei")
        if not isinstance(gong_wei, list) or len(gong_wei) != 12:
            raise ChartSchemaError(
                f"{action} ziwei natal_chart.ziwei.gong_wei 必须是 12 宫数组。"
            )
        daxian = natal_chart.get("ziwei_daxian")
        if not isinstance(daxian, list) or not daxian:
            raise ChartSchemaError(f"{action} ziwei natal_chart.ziwei_daxian 必须是非空数组。")
    else:
        raise ChartSchemaError(f"{action} side 只支持 bazi/ziwei，收到: {side!r}")
