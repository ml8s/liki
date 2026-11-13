"""Name evaluation pipeline."""

from app.naming.naming.loader import _char_by_char, _ensure_loaded, _negative_chars
from app.naming.naming.models import (
    Character,
    NamingCharacterFact,
    NamingConstraints,
    NamingError,
    NamingEvaluation,
    SelectedYongshenPlan,
)
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

    normalized_plan = _normalize_selected_yongshen(selected_yongshen) if selected_yongshen is not None else None
    normalized_constraints = _normalize_constraints(constraints)
    return [_evaluate_name(name, normalized_plan, normalized_constraints).to_dict() for name in given_names]


def _normalize_constraints(constraints: object) -> NamingConstraints:
    """Normalize optional user character constraints."""
    if constraints is None:
        return NamingConstraints(frozenset(), frozenset())
    if isinstance(constraints, dict):
        raw = constraints
    else:
        try:
            raw = constraints.model_dump()
        except AttributeError as error:
            raise LikiToolError("constraints must be an object") from error
    if not isinstance(raw, dict):
        raise ValueError("constraints must be an object")

    forbidden = _normalize_constraint_list("forbidden_chars", raw.get("forbidden_chars"))
    required = _normalize_constraint_list("required_chars", raw.get("required_chars"))
    overlap = forbidden & required
    if overlap:
        raise ValueError(f"同一字不能同时属于 forbidden_chars 和 required_chars: {sorted(overlap)}")
    return NamingConstraints(frozenset(forbidden), frozenset(required))


def _normalize_constraint_list(field: str, values: object) -> set[str]:
    if not values:
        return set()
    if not isinstance(values, list):
        raise LikiToolError(f"constraints.{field} must be an array")
    result: set[str] = set()
    for value in values:
        if not isinstance(value, str) or len(value) != 1:
            raise LikiToolError(f"invalid constraints.{field} {value!r}")
        result.add(value)
    return result


def _named_characters(given_name: str) -> tuple[tuple[Character, ...], tuple[NamingError, ...]]:
    characters: list[Character] = []
    errors: list[NamingError] = []
    for char in given_name:
        if char not in _char_by_char:
            error = NamingError("character_not_found", char)
        elif char in _negative_chars:
            error = NamingError("negative_character_forbidden", char)
        else:
            characters.append(_char_by_char[char])
            continue
        # Repeated invalid characters keep the old single structured error.
        if error not in errors:
            errors.append(error)
    return tuple(characters), tuple(errors)


def _constraint_errors(given_name: str, constraints: NamingConstraints) -> list[NamingError]:
    name_chars = set(given_name)
    errors = [NamingError("forbidden_character", char) for char in sorted(name_chars & constraints.forbidden)]
    errors.extend(NamingError("required_character_missing", char) for char in sorted(constraints.required - name_chars))
    return errors


def _wuxing_role(plan: SelectedYongshenPlan | None, element: str) -> str | None:
    return plan.role_of(element) if plan is not None else None


def _evaluate_name(
    given_name: str,
    plan: SelectedYongshenPlan | None,
    constraints: NamingConstraints,
) -> NamingEvaluation:
    if not 1 <= len(given_name) <= 2:
        return NamingEvaluation(
            given_name=given_name,
            valid=False,
            errors=(NamingError("invalid_name_length", None),),
        )

    errors = tuple(_constraint_errors(given_name, constraints))
    characters, lookup_errors = _named_characters(given_name)
    all_errors = errors + lookup_errors
    if all_errors:
        return NamingEvaluation(given_name=given_name, valid=False, errors=all_errors)

    facts = tuple(
        NamingCharacterFact(
            char=character.char,
            frequency=character.frequency,
            wuxing=character.element,
            stroke=character.stroke,
            pinyin=character.pinyin,
            tone=character.tone,
            radical=character.radical,
            wuxing_role=_wuxing_role(plan, character.element),
        )
        for character in characters
    )
    return NamingEvaluation(
        given_name=given_name,
        valid=True,
        characters=facts,
        tones=tuple(character.tone for character in characters),
    )
