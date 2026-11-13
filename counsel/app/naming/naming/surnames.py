"""Surname phonetic matching."""

import unicodedata

from app.naming.naming.loader import _ensure_loaded, _surname_records
from app.naming.naming.models import SurnameRecord
from app.naming.naming.pinyin import first_vowel, fold_latin, is_latin_lower, tone_from_pinyin

MATCH_PINYIN_EXACT = "pinyin_exact"
MATCH_ROMANIZATION_EXACT = "romanization_exact"
MATCH_PHONETIC_CLOSE = "phonetic_close"
MATCH_FALLBACK_BAIJIAXING = "fallback_baijiaxing"

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

    result: dict = {"source_surname": source}
    if matches:
        result["candidates"] = [m[0] for m in matches[:max_candidates]]
        return result

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


def latin_phonetic_close(source: str, plain: str) -> bool:
    if len(source) < 2 or len(plain) < 2:
        return False
    if source.startswith(plain) or plain.startswith(source):
        return True
    if source[0] != plain[0]:
        return False
    return first_vowel(source) == first_vowel(plain)


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




def _has_combining_diaeresis(value: str) -> bool:
    return any(ch == "\u0308" for ch in unicodedata.normalize("NFD", value))


def _split_non_alpha(value: str) -> list[str]:
    import re
    return re.split(r"[^a-zA-Z]+", value)
