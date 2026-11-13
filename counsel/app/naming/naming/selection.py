"""Selected-yongshen driven character selection."""

from app.naming.naming.loader import _char_by_element, _ensure_loaded
from app.naming.naming.models import Character, SelectedYongshenPlan
from app.naming.naming.wuxing import wuxing_from_chinese
from app.natal.tools.errors import LikiToolError

PICK_ROLES = ("yong", "xi")


def _normalize_selected_yongshen(selected_yongshen: object) -> SelectedYongshenPlan:
    """Normalize the bazi decision into concrete yong/xi/ji lists."""
    if isinstance(selected_yongshen, dict):
        raw_selected_yongshen = selected_yongshen
    else:
        try:
            raw_selected_yongshen = selected_yongshen.model_dump()
        except AttributeError as error:
            raise LikiToolError("selected_yongshen must be an object") from error
    if not isinstance(raw_selected_yongshen, dict):
        raise ValueError("selected_yongshen must be an object")

    if raw_selected_yongshen.get("status") not in ("resolved", "partial"):
        raise ValueError("selected_yongshen.status 必须是 resolved 或 partial")
    raw = raw_selected_yongshen.get("selected_yongshen")
    if not isinstance(raw, dict):
        raise ValueError("selected_yongshen.selected_yongshen must be an object")

    yong = wuxing_from_chinese(raw.get("yong", ""))
    if not yong:
        raise ValueError(f"invalid selected_yongshen.selected_yongshen.yong {raw.get('yong')!r}")

    xi: list[str] = []
    ji: list[str] = []
    assigned = {yong: "yong"}
    for role, elements in (("xi", raw.get("xi")), ("ji", raw.get("ji"))):
        values: list[str] = []
        if not isinstance(elements, list):
            raise LikiToolError(f"selected_yongshen.selected_yongshen.{role} must be an array")
        for element in elements:
            normalized_element = wuxing_from_chinese(element)
            if not normalized_element:
                raise LikiToolError(f"invalid selected_yongshen.selected_yongshen.{role} {element!r}")
            if normalized_element in assigned:
                raise LikiToolError(f"五行 {normalized_element} 同时属于 {assigned[normalized_element]} 和 {role} 角色")
            values.append(normalized_element)
            assigned[normalized_element] = role
        if role == "xi":
            xi = values
        else:
            ji = values
    return SelectedYongshenPlan(yong=yong, xi=tuple(xi), ji=tuple(ji))


def _pick_role_elements(plan: SelectedYongshenPlan, role: str) -> list[str]:
    return list(plan.elements_for(role))


def pick_chars_by_selected_yongshen(
    selected_yongshen: object,
    first_role: str = "yong",
    second_role: str | None = None,
) -> dict:
    """按 用/喜 角色取字池；一个 role 对多个五行时返回多个 pool。"""
    _ensure_loaded()
    normalized_plan = _normalize_selected_yongshen(selected_yongshen)
    if first_role not in PICK_ROLES or (second_role is not None and second_role not in PICK_ROLES):
        raise ValueError("naming_pick 只支持 yong/xi 角色")
    if second_role is None:
        if first_role != "yong":
            raise LikiToolError("单名 first_role 只能是 yong")
    elif "yong" not in (first_role, second_role):
        raise ValueError("双名至少一个字位必须是 yong")

    def append_pools(char_position: str, role: str) -> None:
        for element in _pick_role_elements(normalized_plan, role):
            chars = _char_by_element.get(element, [])
            if not chars:
                raise LikiToolError(f"no characters available for wuxing {element!r}")
            result["pools"].append(
                {
                    "char_position": char_position,
                    "role": role,
                    "wuxing": element,
                    "chars": [_picked(character) for character in chars],
                }
            )

    result: dict = {"pools": []}
    append_pools("first", first_role)
    if second_role is not None:
        append_pools("second", second_role)
    return result


def _picked(c: Character) -> dict:
    return {"char": c.char, "frequency": c.frequency}


# -- 组名 -------------------------------------------------------------------
