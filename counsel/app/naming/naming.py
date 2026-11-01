"""起名（qiming）——函数式实现。

起名本质是"命理之上的应用"（字库查找 + 五行匹配 + 组名 + 评估），不依赖
命理排盘，天然属于 counsel（判断/应用）层。本模块为纯函数 + 模块级只读数据。

数据：同目录 data/（naming_characters.csv / negative_chars.txt / surnames.csv），
字库与音韵随本模块维护，本层是唯一来源。
"""

import csv
import threading
import unicodedata
from dataclasses import dataclass, field
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent / "data"

# -- 数据结构 ---------------------------------------------------------------


@dataclass(frozen=True)
class Character:
    char: str
    frequency: str
    element: str
    stroke: int
    radical: str = ""
    pinyin: str = ""
    tone: int = 0


@dataclass(frozen=True)
class SurnameRecord:
    surname: str
    pinyin: str
    plain_pinyin: str
    baijiaxing_index: int
    aliases: list = field(default_factory=list)


# -- 数据加载（模块级只读，进程内一次） -------------------------------------

_char_by_char: dict[str, Character] = {}
_negative_chars: set[str] = set()
_char_by_element: dict[str, list[Character]] = {}
_surname_records: list[SurnameRecord] = []
_surnames_loaded = False
_LOAD_LOCK = threading.Lock()


def wuxing_from_chinese(name: str) -> str:
    """五行中文 → 五行中文；无效返回 ''（对齐 Go ParseWuxing→0）。

    线上契约只接受中文五行闭集；英文名别名已随契约收窄删除。
    """
    return {
        "木": "木",
        "火": "火",
        "土": "土",
        "金": "金",
        "水": "水",
    }.get(name, "")


def _load_naming() -> None:
    columns = ("char", "frequency", "pinyin", "radical", "stroke", "wuxing", "tone")
    with (DATA_DIR / "naming_characters.csv").open(encoding="utf-8", newline="") as fh:
        reader = csv.DictReader(fh)
        if reader.fieldnames is None or set(reader.fieldnames) != set(columns):
            raise ValueError(
                "naming_characters.csv columns mismatch; "
                f"expected {sorted(columns)}, got {reader.fieldnames}"
            )
        for i, row in enumerate(reader, start=2):
            word = row["char"]
            if len(word) != 1:
                raise ValueError(f"naming_characters.csv row {i}: invalid character {word!r}")
            radical = row["radical"] if row["radical"] != "NULL" else ""
            stroke = int(row["stroke"])
            if stroke <= 0:
                raise ValueError(f"naming_characters.csv row {i}: invalid stroke for {word!r}")
            tone = int(row["tone"])
            if not 1 <= tone <= 5:
                raise ValueError(f"naming_characters.csv row {i}: invalid tone for {word!r}")
            pinyin = row["pinyin"].strip()
            if not pinyin:
                raise ValueError(f"naming_characters.csv row {i}: missing pinyin for {word!r}")
            frequency = row["frequency"]
            if frequency not in ("common", "standard", "rare"):
                raise ValueError(f"naming_characters.csv row {i}: invalid frequency {frequency!r}")
            element = wuxing_from_chinese(row["wuxing"])
            if not element:
                raise ValueError(f"naming_characters.csv row {i}: missing naming element for {word!r}")
            if word in _char_by_char:
                raise ValueError(f"naming_characters.csv row {i}: duplicate character {word!r}")
            _char_by_char[word] = Character(
                char=word, frequency=frequency, element=element,
                stroke=stroke, radical=radical, pinyin=pinyin, tone=tone,
            )

    for line_no, excluded in enumerate(
        (DATA_DIR / "negative_chars.txt").read_text(encoding="utf-8").splitlines(), start=1
    ):
        excluded = excluded.strip()
        if not excluded:
            continue
        if len(excluded) != 1 or excluded not in _char_by_char:
            raise ValueError(f"negative_chars.txt line {line_no}: unknown character {excluded!r}")
        if excluded in _negative_chars:
            raise ValueError(f"negative_chars.txt line {line_no}: duplicate character {excluded!r}")
        _negative_chars.add(excluded)

    naming_chars = [c for c in _char_by_char.values() if c.char not in _negative_chars]
    naming_chars.sort(key=lambda c: c.char)
    for character in naming_chars:
        _char_by_element.setdefault(character.element, []).append(character)


def _load_surnames() -> None:
    with (DATA_DIR / "surnames.csv").open(encoding="utf-8", newline="") as fh:
        reader = csv.DictReader(fh)
        seen_indexes: set[int] = set()
        seen_surnames: set[str] = set()
        for line_no, row in enumerate(reader, start=2):
            surname = row["surname"].strip()
            pinyin = row["pinyin"].strip()
            plain = fold_latin(row["plain_pinyin"])
            if not surname or not pinyin or not plain:
                raise ValueError(f"surnames.csv row {line_no}: empty surname or pinyin")
            if not 1 <= len(surname) <= 2:
                raise ValueError(f"surnames.csv row {line_no}: invalid surname {surname!r}")
            if " " in plain or not is_latin_lower(plain):
                raise ValueError(f"surnames.csv row {line_no}: invalid plain pinyin {plain!r}")
            if tone_from_pinyin(pinyin) == 5:
                raise ValueError(f"surnames.csv row {line_no}: pinyin {pinyin!r} must contain a tone mark")
            idx = int(row["baijiaxing_index"])
            if idx != line_no - 1 or idx > 504:
                raise ValueError(f"surnames.csv row {line_no}: invalid baijiaxing index")
            if idx in seen_indexes:
                raise ValueError(f"surnames.csv row {line_no}: duplicate baijiaxing index {idx}")
            seen_indexes.add(idx)
            aliases: list[str] = []
            for alias in row["aliases"].split(","):
                alias = fold_latin(alias)
                if not alias:
                    continue
                if not is_latin_lower(alias) or " " in alias or alias == plain or alias in aliases:
                    raise ValueError(f"surnames.csv row {line_no}: invalid alias {alias!r}")
                aliases.append(alias)
            if surname in seen_surnames:
                continue
            seen_surnames.add(surname)
            _surname_records.append(SurnameRecord(
                surname=surname, pinyin=pinyin, plain_pinyin=plain,
                baijiaxing_index=idx, aliases=aliases,
            ))


def _ensure_loaded() -> None:
    global _surnames_loaded
    if _char_by_char and _surnames_loaded:
        return
    with _LOAD_LOCK:
        if not _char_by_char:
            _load_naming()
        if not _surnames_loaded:
            _load_surnames()
            _surnames_loaded = True


# -- 取字池 -----------------------------------------------------------------

PICK_ROLES = ("yong", "xi")


def _normalize_selected_yongshen(selected_yongshen: object) -> dict:
    """Normalize the bazi decision into concrete yong/xi/ji lists."""
    if isinstance(selected_yongshen, dict):
        raw_selected_yongshen = selected_yongshen
    else:
        try:
            raw_selected_yongshen = selected_yongshen.model_dump()
        except AttributeError as error:
            raise ValueError("selected_yongshen must be an object") from error
    if not isinstance(raw_selected_yongshen, dict):
        raise ValueError("selected_yongshen must be an object")

    if raw_selected_yongshen.get("status") not in ("resolved", "partial"):
        raise ValueError("selected_yongshen.status 必须是 resolved 或 partial")
    raw = raw_selected_yongshen.get("selected")
    if not isinstance(raw, dict):
        raise ValueError("selected_yongshen.selected must be an object")

    yong = wuxing_from_chinese(raw.get("yong", ""))
    if not yong:
        raise ValueError(f"invalid selected_yongshen.selected.yong {raw.get('yong')!r}")

    normalized: dict = {"yong": yong, "xi": [], "ji": []}
    assigned = {yong: "yong"}
    for role in ("xi", "ji"):
        elements = raw.get(role) or []
        if not isinstance(elements, list):
            raise ValueError(f"selected_yongshen.selected.{role} must be an array")
        for element in elements:
            normalized_element = wuxing_from_chinese(element)
            if not normalized_element:
                raise ValueError(
                    f"invalid selected_yongshen.selected.{role} {element!r}"
                )
            if normalized_element in assigned:
                raise ValueError(
                    f"五行 {normalized_element} 同时属于 "
                    f"{assigned[normalized_element]} 和 {role} 角色"
                )
            normalized[role].append(normalized_element)
            assigned[normalized_element] = role
    return normalized


def _pick_role_elements(plan: dict, role: str) -> list[str]:
    if role == "yong":
        return [plan["yong"]]
    if role == "xi":
        return list(plan["xi"])
    raise ValueError(f"naming_pick 不支持 {role} 字池")


def pick_chars_by_selected_yongshen(
    selected_yongshen: object,
    first_role: str = "yong",
    second_role: str | None = None,
) -> dict:
    """按 用/喜 角色取字池；一个 role 对多个五行时返回多个 pool。"""
    _ensure_loaded()
    normalized_plan = _normalize_selected_yongshen(selected_yongshen)
    if first_role not in PICK_ROLES or (
        second_role is not None and second_role not in PICK_ROLES
    ):
        raise ValueError("naming_pick 只支持 yong/xi 角色")
    if second_role is None:
        if first_role != "yong":
            raise ValueError("单名 first_role 只能是 yong")
    elif "yong" not in (first_role, second_role):
        raise ValueError("双名至少一个字位必须是 yong")

    def append_pools(slot: str, role: str) -> None:
        for element in _pick_role_elements(normalized_plan, role):
            chars = _char_by_element.get(element, [])
            if not chars:
                raise ValueError(f"no characters available for wuxing {element!r}")
            result["pools"].append({
                "slot": slot,
                "role": role,
                "wuxing": element,
                "chars": [_picked(character) for character in chars],
            })

    result: dict = {"pools": []}
    append_pools("first", first_role)
    if second_role is not None:
        append_pools("second", second_role)
    return result


def _picked(c: Character) -> dict:
    return {"char": c.char, "frequency": c.frequency}


# -- 组名 -------------------------------------------------------------------

COMPOSE_MAX_CANDIDATES = 256
COMPOSE_DEFAULT_NAMES = 100
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
    total_possible = len(first_chars) * len(second_chars) if double_name else len(first_chars)
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
    return {"total_possible": total_possible, "names": names}


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


def _reject_duplicates(slot: str, chars: list) -> None:
    seen: set[str] = set()
    for char in chars:
        if char in seen:
            raise ValueError(f"{slot} contains duplicate character {char!r}")
        seen.add(char)


# -- 评估 -------------------------------------------------------------------

def evaluate_names(
    given_names: list,
    selected_yongshen: object | None = None,
) -> list[dict]:
    """评估候选名的字库、禁用字、音韵与逐字五行角色。"""
    _ensure_loaded()
    if not given_names:
        raise ValueError("given_names must contain at least one name")

    normalized_plan = (
        _normalize_selected_yongshen(selected_yongshen)
        if selected_yongshen is not None
        else None
    )
    return [_evaluate_name(name, normalized_plan) for name in given_names]


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


def _evaluate_name(given_name: str, plan: dict | None) -> dict:
    evaluation: dict = {"given_name": given_name, "valid": False}
    if not 1 <= len(given_name) <= 2:
        evaluation["errors"] = [{"code": "invalid_name_length"}]
        return evaluation

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

MATCH_PINYIN_EXACT = "pinyin_exact"
MATCH_ROMANIZATION_EXACT = "romanization_exact"
MATCH_PHONETIC_CLOSE = "phonetic_close"
MATCH_FALLBACK_BAIJIAXING = "fallback_baijiaxing"
STRATEGY_PHONETIC = "phonetic"
STRATEGY_BAIJIAXING_FALLBACK = "baijiaxing_fallback"

_NON_DECOMPOSING_LATIN = {
    "æ": "ae", "œ": "oe", "ß": "ss", "ø": "o", "đ": "d", "ð": "d",
    "ł": "l", "þ": "th", "ı": "i", "ħ": "h", "ŋ": "n", "ſ": "s",
}


def match_surnames(source_surname: str, max_candidates: int = 0) -> dict:
    """按发音匹配候选姓（对齐 Go MatchSurnames）。"""
    _ensure_loaded()
    if max_candidates == 0:
        max_candidates = 6
    if not 1 <= max_candidates <= 12:
        raise ValueError("max_candidates must be between 1 and 12")
    source = source_surname.strip()
    if not source:
        raise ValueError("source_surname is required")
    if len(source) > 64:
        raise ValueError("source_surname must contain at most 64 characters")
    tokens = _surname_source_tokens(source)
    if not tokens:
        raise ValueError("source_surname must contain Latin letters")

    matches: list[tuple[dict, int]] = []
    for record in _surname_records:
        hit = _match_surname(record, tokens)
        if hit:
            level, basis, token_size = hit
            matches.append((_new_surname_candidate(record, level, basis), token_size))
    matches.sort(key=lambda m: (_match_order(m[0]["match_kind"]), -m[1], m[0]["baijiaxing_index"]))

    result: dict = {"source_surname": source, "strategy": STRATEGY_PHONETIC}
    if matches:
        result["candidates"] = [m[0] for m in matches[:max_candidates]]
        return result

    result["strategy"] = STRATEGY_BAIJIAXING_FALLBACK
    result["candidates"] = [
        _new_surname_candidate(r, MATCH_FALLBACK_BAIJIAXING, ["classic_baijiaxing_order"])
        for r in _surname_records[:max_candidates]
    ]
    return result


def _new_surname_candidate(record: SurnameRecord, level: str, basis: list) -> dict:
    return {
        "surname": record.surname,
        "pinyin": record.pinyin,
        "tone": tone_from_pinyin(record.pinyin),
        "baijiaxing_index": record.baijiaxing_index,
        "match_kind": level,
        "basis": basis,
    }


def _match_surname(record: SurnameRecord, tokens: list[str]) -> tuple[str, list, int] | None:
    for token in tokens:
        if token == record.plain_pinyin:
            return MATCH_PINYIN_EXACT, [f"mandarin_pinyin={token}"], len(token)
    for token in tokens:
        for alias in record.aliases:
            if token == alias:
                return MATCH_ROMANIZATION_EXACT, [f"romanization={alias}"], len(token)
    for token in tokens:
        if latin_phonetic_close(token, record.plain_pinyin):
            return MATCH_PHONETIC_CLOSE, [f"latin={token}", f"pinyin={record.plain_pinyin}"], len(token)
    return None


def _match_order(level: str) -> int:
    return {
        MATCH_PINYIN_EXACT: 0,
        MATCH_ROMANIZATION_EXACT: 1,
        MATCH_PHONETIC_CLOSE: 2,
    }.get(level, 3)


def _surname_source_tokens(source: str) -> list[str]:
    raw = source.strip().lower()
    parts = [p for p in _split_non_alpha(raw) if p]
    folded: list[str] = []
    for part in parts:
        token = fold_latin(part)
        if _has_combining_diaeresis(part):
            if token == "lu":
                token = "lv"
            elif token == "nu":
                token = "nv"
        if not is_latin_lower(token):
            return []
        folded.append(token)
    tokens = folded
    if len(tokens) > 1:
        tokens.append("".join(tokens))
    return tokens


def _split_non_alpha(value: str) -> list[str]:
    import re
    return re.split(r"[^a-zA-Z]+", value)


def is_latin_lower(value: str) -> bool:
    return bool(value) and all("a" <= ch <= "z" for ch in value)


def latin_phonetic_close(source: str, plain: str) -> bool:
    if len(source) < 2 or len(plain) < 2:
        return False
    if source.startswith(plain) or plain.startswith(source):
        return True
    if source[0] != plain[0]:
        return False
    return first_vowel(source) == first_vowel(plain)


def first_vowel(s: str) -> str:
    for ch in s:
        if ch in "aeiouv":
            return ch
    return ""


def fold_latin(value: str) -> str:
    lowered = value.strip().lower()
    out: list[str] = []
    for ch in unicodedata.normalize("NFD", lowered):
        if unicodedata.combining(ch):
            continue
        out.append(_NON_DECOMPOSING_LATIN.get(ch, ch))
    return "".join(out)


def _has_combining_diaeresis(value: str) -> bool:
    return any(ch == "\u0308" for ch in unicodedata.normalize("NFD", value))


def tone_from_pinyin(pinyin: str) -> int:
    tone_by_vowel = {
        "ā": 1, "á": 2, "ǎ": 3, "à": 4,
        "ē": 1, "é": 2, "ě": 3, "è": 4,
        "ī": 1, "í": 2, "ǐ": 3, "ì": 4,
        "ō": 1, "ó": 2, "ǒ": 3, "ò": 4,
        "ū": 1, "ú": 2, "ǔ": 3, "ù": 4,
        "ǖ": 1, "ǘ": 2, "ǚ": 3, "ǜ": 4,
    }
    for ch in pinyin:
        if ch in tone_by_vowel:
            return tone_by_vowel[ch]
    return 5
