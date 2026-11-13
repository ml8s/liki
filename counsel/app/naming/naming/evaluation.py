"""Name evaluation pipeline."""

from app.naming.naming.loader import _char_by_char, _ensure_loaded, _negative_chars
from app.naming.naming.models import Character
from app.naming.naming.selection import _normalize_selected_yongshen
from app.natal.tools.errors import LikiToolError


def evaluate_names(
    given_names: list,
    selected_yongshen: object | None = None,
    constraints: object | None = None,
) -> list[dict]:
    """评估候选名的结构、用户约束、音韵与逐字五行角色。"""
    _ensure_loaded()
    if not given_names:
        raise ValueError("given_names must contain at least one name")

    normalized_plan = (
        _normalize_selected_yongshen(selected_yongshen)
        if selected_yongshen is not None
        else None
    )
    normalized_constraints = _normalize_constraints(constraints)
    return [
        _evaluate_name(name, normalized_plan, normalized_constraints)
        for name in given_names
    ]


def _normalize_constraints(constraints: object) -> dict:
    """Normalize optional user character constraints."""
    if constraints is None:
        return {"forbidden": set(), "required": set()}
    if isinstance(constraints, dict):
        raw = constraints
    else:
        try:
            raw = constraints.model_dump()
        except AttributeError as error:
            raise LikiToolError("constraints must be an object") from error
    if not isinstance(raw, dict):
        raise ValueError("constraints must be an object")

    def normalize_list(field: str) -> set[str]:
        values = raw.get(field) or []
        if not isinstance(values, list):
            raise LikiToolError(f"constraints.{field} must be an array")
        result: set[str] = set()
        for value in values:
            if not isinstance(value, str) or len(value) != 1:
                raise LikiToolError(f"invalid constraints.{field} {value!r}")
            result.add(value)
        return result

    forbidden = normalize_list("forbidden_chars")
    required = normalize_list("required_chars")
    overlap = forbidden & required
    if overlap:
        raise ValueError(
            f"同一字不能同时属于 forbidden_chars 和 required_chars: {sorted(overlap)}"
        )
    return {"forbidden": forbidden, "required": required}


def _named_characters(
    given_name: str,
    evaluation: dict,
) -> list[Character]:
    characters: list[Character] = []
    for char in given_name:
        if char not in _char_by_char:
            _append_error(evaluation, "character_not_found", char)
            continue
        if char in _negative_chars:
            _append_error(evaluation, "negative_character_forbidden", char)
            continue
        characters.append(_char_by_char[char])
    return characters


def _check_user_constraints(
    given_name: str,
    evaluation: dict,
    constraints: dict,
) -> None:
    name_chars = set(given_name)
    for char in sorted(name_chars & constraints["forbidden"]):
        _append_error(evaluation, "forbidden_character", char)
    for char in sorted(constraints["required"] - name_chars):
        _append_error(evaluation, "required_character_missing", char)


def _wuxing_role(plan: dict | None, element: str) -> str | None:
    if plan is None:
        return None
    if element == plan["yong"]:
        return "yong"
    if element in plan["xi"]:
        return "xi"
    if element in plan["ji"]:
        return "ji"
    return "neutral"


def _evaluate_name(
    given_name: str,
    plan: dict | None,
    constraints: dict,
) -> dict:
    evaluation: dict = {"given_name": given_name, "valid": False}
    if not 1 <= len(given_name) <= 2:
        evaluation["errors"] = [{"code": "invalid_name_length"}]
        return evaluation

    _check_user_constraints(given_name, evaluation, constraints)
    characters = _named_characters(given_name, evaluation)
    if evaluation.get("errors"):
        return evaluation

    evaluation["valid"] = True
    character_facts = []
    for character in characters:
        item = _character_json(character)
        role = _wuxing_role(plan, character.element)
        if role is not None:
            item["wuxing_role"] = role
        character_facts.append(item)
    evaluation["characters"] = character_facts
    evaluation["phonetic"] = {"tones": "-".join(str(c.tone) for c in characters)}
    return evaluation


def _character_json(c: Character) -> dict:
    out = {
        "char": c.char, "frequency": c.frequency, "wuxing": c.element,
        "stroke": c.stroke, "pinyin": c.pinyin, "tone": c.tone,
    }
    if c.radical:
        out["radical"] = c.radical
    return out


def _append_error(evaluation: dict, code: str, char: str) -> None:
    errors = evaluation.setdefault("errors", [])
    if {"code": code, "char": char} in errors:
        return
    errors.append({"code": code, "char": char})


# -- 配姓 -------------------------------------------------------------------

