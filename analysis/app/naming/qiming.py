"""起名（qiming）——函数式实现，移植自 engine Go 的 qiming 域。

起名本质是"命理之上的应用"（字库查找 + 五行匹配 + 组名 + 评估），不依赖
命理排盘，天然属于 analysis（Python）层。本模块为纯函数 + 模块级只读数据，
逻辑与 Go 实现逐字段对齐（对照测试保证一致）。

数据来源：engine/internal/engine/qiming/data/（naming_characters.csv /
negative_chars.txt / surnames.csv），与本文件同步维护。
"""
from __future__ import annotations

import csv
import unicodedata
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

DATA_DIR = Path(__file__).resolve().parent / "data"

WUXING_NAMES = ("木", "火", "土", "金", "水")

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


@dataclass(frozen=True)
class PickedCharacter:
    char: str
    frequency: str


# -- 数据加载（模块级只读，进程内一次） -------------------------------------

_char_by_char: dict[str, Character] = {}
_negative_chars: set[str] = set()
_char_by_element: dict[str, list[Character]] = {}
_surname_records: list[SurnameRecord] = []
_surnames_loaded = False


def wuxing_from_chinese(name: str) -> str:
    """五行名（中/英）→ 五行中文；无效返回 ''（对齐 Go ParseWuxing→0）。"""
    return {
        "木": "木", "wood": "木",
        "火": "火", "fire": "火",
        "土": "土", "earth": "土",
        "金": "金", "metal": "金",
        "水": "水", "water": "水",
    }.get(name, "")


def _load_naming() -> None:
    columns = ("char", "frequency", "pinyin", "radical", "stroke", "wuxing", "tone")
    with (DATA_DIR / "naming_characters.csv").open(encoding="utf-8", newline="") as fh:
        reader = csv.DictReader(fh)
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
    if not _char_by_char:
        _load_naming()
    if not _surnames_loaded:
        _load_surnames()
        _surnames_loaded = True


# -- 取字池 -----------------------------------------------------------------

def pick_chars(wuxing1: str, wuxing2: str, count: int) -> dict:
    """按五行取字池（对齐 Go PickChars）。count=1 单名，2 双名。"""
    _ensure_loaded()
    elem1 = wuxing_from_chinese(wuxing1)
    if not elem1:
        raise ValueError(f"invalid wuxing1 {wuxing1!r}")
    if count not in (1, 2):
        raise ValueError("count must be 1 or 2")
    if count == 1 and wuxing2:
        raise ValueError("wuxing2 is not allowed when count is 1")
    if not wuxing2:
        wuxing2 = wuxing1
    elem2 = wuxing_from_chinese(wuxing2)
    if not elem2:
        raise ValueError(f"invalid wuxing2 {wuxing2!r}")

    first = _char_by_element.get(elem1, [])
    if not first:
        raise ValueError(f"no characters available for wuxing {wuxing1!r}")
    result: dict = {"wuxing1": elem1, "pools": [{
        "slot": "first", "chars": [_picked(c) for c in first],
    }]}
    if count == 1:
        return result

    second = _char_by_element.get(elem2, [])
    if not second:
        raise ValueError(f"no characters available for wuxing {wuxing2!r}")
    if elem2 != elem1:
        result["wuxing2"] = elem2
    result["pools"].append({"slot": "second", "chars": [_picked(c) for c in second]})
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

def evaluate_names(given_names: list, yong_shen: str, xi_shen: list, ji_shen: list) -> list[dict]:
    """独立评估候选名（对齐 Go EvaluateNames）。"""
    _ensure_loaded()
    if not given_names:
        raise ValueError("given_names must contain at least one name")
    yong_elem = wuxing_from_chinese(yong_shen)
    if yong_shen and not yong_elem:
        raise ValueError(f"invalid yongshen {yong_shen!r}")
    xi_elems = _parse_wuxing_list("xishen", xi_shen)
    ji_elems = _parse_wuxing_list("jishen", ji_shen)
    _reject_conflicting_wuxing(yong_elem, xi_elems, ji_elems)
    return [_evaluate_name(n, yong_elem, xi_elems, ji_elems) for n in given_names]


def _evaluate_name(given_name: str, yong_elem: str, xi_elems: list, ji_elems: list) -> dict:
    evaluation: dict = {"given_name": given_name, "valid": False}
    if not 1 <= len(given_name) <= 2:
        evaluation["errors"] = [{"code": "invalid_name_length"}]
        return evaluation
    characters: list[Character] = []
    for char in given_name:
        if char not in _char_by_char:
            _append_error(evaluation, "character_not_found", char)
            continue
        if char in _negative_chars:
            _append_error(evaluation, "negative_character_forbidden", char)
            continue
        characters.append(_char_by_char[char])
    if evaluation.get("errors"):
        return evaluation

    evaluation["valid"] = True
    evaluation["characters"] = [_character_json(c) for c in characters]
    evaluation["phonetic"] = {"tones": "-".join(str(c.tone) for c in characters)}
    if yong_elem or xi_elems or ji_elems:
        wuxing: dict = {}
        if yong_elem:
            wuxing["yong"] = _contains_element(characters, yong_elem)
        if xi_elems:
            wuxing["xi"] = _contains_any_element(characters, xi_elems)
        if ji_elems:
            wuxing["ji"] = _contains_any_element(characters, ji_elems)
        evaluation["wuxing"] = wuxing
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


def _contains_element(chars: list[Character], elem: str) -> bool:
    return any(c.element == elem for c in chars)


def _contains_any_element(chars: list[Character], elems: list[str]) -> bool:
    return any(_contains_element(chars, e) for e in elems)


def _parse_wuxing_list(field: str, values: list) -> list[str]:
    if not values:
        return []
    out: list[str] = []
    for value in values:
        elem = wuxing_from_chinese(value)
        if not elem:
            raise ValueError(f"invalid {field} value {value!r}")
        if elem in out:
            raise ValueError(f"duplicate {field} value {value!r}")
        out.append(elem)
    return out


def _reject_conflicting_wuxing(yong: str, xi: list, ji: list) -> None:
    if yong and (yong in xi or yong in ji):
        raise ValueError("yongshen must be disjoint with xishen/jishen")
    if set(xi) & set(ji):
        raise ValueError("xishen and jishen must be disjoint")


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
    matches.sort(key=lambda m: (_match_order(m[0]["match_level"]), -m[1], m[0]["baijiaxing_index"]))

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
        "match_level": level,
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