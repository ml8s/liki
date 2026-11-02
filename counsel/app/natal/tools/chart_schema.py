"""natal_chart canonical contract and fail-closed structural validation."""

from app.natal.tools.chart_integrity import DIGEST_FIELD, validate_natal_chart_digest
from app.natal.tools.errors import ChartSchemaError
from app.natal.tools.factor_constants import load_constants

_CONSTANTS = load_constants()
REQUIRED_FIELDS = ("subject", "bazi", "ziwei", "ziwei_daxian", DIGEST_FIELD)
PILLARS = tuple(_CONSTANTS["四柱"])
GENDERS = tuple(_CONSTANTS["性别闭集"])
PALACES = tuple(_CONSTANTS["紫微宫位"])


def _require_object(value: object, action: str, label: str) -> dict:
    if not isinstance(value, dict) or not value:
        raise ChartSchemaError(f"{action} {label} 必须是非空 object。")
    return value


def _require_list(value: object, action: str, label: str, length: int | None = None) -> list:
    if not isinstance(value, list):
        raise ChartSchemaError(f"{action} {label} 必须是 array。")
    if length is not None and len(value) != length:
        raise ChartSchemaError(f"{action} {label} 必须包含 {length} 项。")
    return value


def _validate_subject(subject: object, action: str) -> dict:
    subject = _require_object(subject, action, "subject")
    missing = [key for key in ("birth_time", "birth_longitude", "gender") if key not in subject]
    if missing:
        raise ChartSchemaError(f"{action} subject 缺少字段: {', '.join(missing)}。")
    if not isinstance(subject["birth_time"], str) or not subject["birth_time"]:
        raise ChartSchemaError(f"{action} subject.birth_time 必须是非空字符串。")
    longitude = subject["birth_longitude"]
    if not isinstance(longitude, (int, float)) or isinstance(longitude, bool) \
            or not -180 <= float(longitude) <= 180:
        raise ChartSchemaError(f"{action} subject.birth_longitude 必须在 [-180,180]。")
    if subject["gender"] not in GENDERS:
        raise ChartSchemaError(f"{action} subject.gender 必须是 male/female。")
    return subject


def _validate_bazi(bazi: object, action: str) -> None:
    bazi = _require_object(bazi, action, "bazi")
    missing = [pillar for pillar in PILLARS if not isinstance(bazi.get(pillar), dict)]
    if missing:
        raise ChartSchemaError(f"{action} bazi 缺少四柱: {', '.join(missing)}。")
    for pillar in PILLARS:
        pillar_data = bazi[pillar]
        for key in ("gan", "zhi"):
            if not isinstance(pillar_data.get(key), str) or not pillar_data[key]:
                raise ChartSchemaError(f"{action} bazi.{pillar}.{key} 必须是非空字符串。")


def _validate_ziwei(ziwei: object, action: str) -> None:
    ziwei = _require_object(ziwei, action, "ziwei")
    palaces = _require_list(ziwei.get("shi_er_gong"), action, "ziwei.shi_er_gong", 12)
    actual = tuple(item.get("ming") for item in palaces if isinstance(item, dict))
    if actual != PALACES:
        raise ChartSchemaError(
            f"{action} ziwei.shi_er_gong 必须按 engine 宫位闭集从命宫顺布: {PALACES}。"
        )
    for index, palace in enumerate(palaces):
        stars = palace.get("xing_yao")
        if not isinstance(stars, list):
            raise ChartSchemaError(f"{action} ziwei.shi_er_gong[{index}].xing_yao 必须是 array。")
        for star_index, star in enumerate(stars):
            if not isinstance(star, dict) or not isinstance(star.get("ming"), str) \
                    or not star["ming"]:
                raise ChartSchemaError(
                    f"{action} ziwei.shi_er_gong[{index}].xing_yao[{star_index}].ming 必须是非空字符串。"
                )


def _validate_daxian(daxian: object, action: str) -> None:
    count = _CONSTANTS["大限段数"]
    daxian = _require_list(daxian, action, "ziwei_daxian", count)
    palaces: list[str] = []
    for index, step in enumerate(daxian):
        if not isinstance(step, dict):
            raise ChartSchemaError(f"{action} ziwei_daxian[{index}] 必须是 object。")
        for key in ("ming_gong", "start_year", "end_year", "qi_sui", "zhi_sui"):
            if key not in step:
                raise ChartSchemaError(f"{action} ziwei_daxian[{index}] 缺少 {key}。")
        if not isinstance(step["ming_gong"], str) or not step["ming_gong"]:
            raise ChartSchemaError(f"{action} ziwei_daxian[{index}].ming_gong 必须是非空字符串。")
        integer_fields = ("start_year", "end_year", "qi_sui", "zhi_sui")
        if any(not isinstance(step[key], int) or isinstance(step[key], bool) for key in integer_fields):
            raise ChartSchemaError(f"{action} ziwei_daxian[{index}] 年段/虚岁必须是整数。")
        palaces.append(step["ming_gong"])
    if len(set(palaces)) != count:
        raise ChartSchemaError(f"{action} ziwei_daxian 的 {count} 个宫位必须唯一。")


def validate_natal_chart(natal_chart: object, action: str = "natal_chart") -> None:
    """只接受 build_natal_chart 的 canonical subject + bazi + ziwei 结构。"""
    if not isinstance(natal_chart, dict):
        raise ChartSchemaError(f"{action} natal_chart 必须是 build_natal_chart 返回的 object。")
    missing = [key for key in REQUIRED_FIELDS if key not in natal_chart]
    if missing:
        raise ChartSchemaError(f"{action} natal_chart 缺少字段: {', '.join(missing)}。")
    _validate_subject(natal_chart["subject"], action)
    _validate_bazi(natal_chart["bazi"], action)
    _validate_ziwei(natal_chart["ziwei"], action)
    _validate_daxian(natal_chart["ziwei_daxian"], action)
    validate_natal_chart_digest(natal_chart, action=action)


def validate_side_chart(natal_chart: object, side: str, action: str = "natal_chart") -> None:
    """校验单侧求值输入；输入仍是 canonical natal_chart。"""
    if not isinstance(natal_chart, dict):
        raise ChartSchemaError(f"{action} natal_chart 必须是 object。")
    validate_natal_chart(natal_chart, action=action)
    if technique not in ("bazi", "ziwei"):
        raise ChartSchemaError(f"{action} side 只支持 bazi/ziwei，收到: {side!r}。")
