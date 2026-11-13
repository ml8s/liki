"""Candidate composition and character lookup."""

from app.naming.naming.evaluation import _character_json
from app.naming.naming.loader import _char_by_char, _ensure_loaded, _negative_chars
from app.naming.naming.models import Character
from app.natal.tools.errors import LikiToolError

COMPOSE_DEFAULT_NAMES = 100
COMPOSE_MAX_CANDIDATES = 256
COMPOSE_MAX_NAMES = 1000


def compose_names(first: list, second: list, max_names: int = 0) -> dict:
    """组合候选名（对齐 Go ComposeNames）。second 空=单名。"""
    _ensure_loaded()
    if not first or len(first) > COMPOSE_MAX_CANDIDATES:
        raise ValueError(f"first must contain 1 to {COMPOSE_MAX_CANDIDATES} characters")
    double_name = bool(second)
    if double_name and len(second) > COMPOSE_MAX_CANDIDATES:
        raise ValueError(f"second must contain 1 to {COMPOSE_MAX_CANDIDATES} characters")
    _reject_duplicates("first", first)
    _reject_duplicates("second", second)
    if max_names == 0:
        max_names = COMPOSE_DEFAULT_NAMES
    if not 1 <= max_names <= COMPOSE_MAX_NAMES:
        raise ValueError(f"max_names must be between 1 and {COMPOSE_MAX_NAMES}")

    first_chars = _lookup_characters(first)
    second_chars = _lookup_characters(second)
    names: list[str] = []
    for fc in first_chars:
        if len(names) >= max_names:
            break
        if not double_name:
            names.append(fc.char)
            continue
        for sc in second_chars:
            if len(names) >= max_names:
                break
            names.append(fc.char + sc.char)
    return {"names": names}


def _lookup_characters(chars: list) -> list[Character]:
    out: list[Character] = []
    for char in chars:
        out.append(_lookup_character(char))
    return out


def _lookup_character(char: str) -> Character:
    _ensure_loaded()
    if len(char) != 1:
        raise ValueError(f"character {char!r} must be a single character")
    if char not in _char_by_char:
        raise ValueError(f"character {char!r} not found in database")
    if char in _negative_chars:
        raise ValueError(f"character {char!r} is excluded from naming")
    return _char_by_char[char]


def lookup_char(char: str) -> dict | None:
    """查单个字（对齐 Go LookupChar，未收录返回 None）。"""
    _ensure_loaded()
    if len(char) != 1 or char not in _char_by_char:
        return None
    return _character_json(_char_by_char[char])


def _reject_duplicates(char_position: str, chars: list) -> None:
    seen: set[str] = set()
    for char in chars:
        if char in seen:
            raise LikiToolError(f"{char_position} contains duplicate character {char!r}")
        seen.add(char)


# -- 评估 -------------------------------------------------------------------

