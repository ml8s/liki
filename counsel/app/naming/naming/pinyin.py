"""Latin/pinyin compatibility helpers."""

import unicodedata

_NON_DECOMPOSING_LATIN = {
    "æ": "ae", "œ": "oe", "ß": "ss", "ø": "o", "đ": "d", "ð": "d",
    "ł": "l", "þ": "th", "ı": "i", "ħ": "h", "ŋ": "n", "ſ": "s",
}


def fold_latin(value: str) -> str:
    lowered = value.strip().lower()
    out: list[str] = []
    for ch in unicodedata.normalize("NFD", lowered):
        if unicodedata.combining(ch):
            continue
        out.append(_NON_DECOMPOSING_LATIN.get(ch, ch))
    return "".join(out)


def is_latin_lower(value: str) -> bool:
    return bool(value) and all("a" <= ch <= "z" for ch in value)


def first_vowel(s: str) -> str:
    for ch in s:
        if ch in "aeiouv":
            return ch
    return ""


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

