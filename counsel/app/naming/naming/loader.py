"""Read-only naming data loading."""

import csv
import threading
from pathlib import Path

from app.naming.naming.models import Character, SurnameRecord
from app.naming.naming.pinyin import fold_latin, is_latin_lower, tone_from_pinyin
from app.natal.tools.errors import LikiToolError

DATA_DIR = Path(__file__).resolve().parents[1] / "data"

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
            raise LikiToolError(
                "naming_characters.csv columns mismatch; "
                f"expected {sorted(columns)}, got {reader.fieldnames}"
            )
        for i, row in enumerate(reader, start=2):
            word = row["char"]
            if len(word) != 1:
                raise LikiToolError(f"naming_characters.csv row {i}: invalid character {word!r}")
            radical = row["radical"] if row["radical"] != "NULL" else ""
            stroke = int(row["stroke"])
            if stroke <= 0:
                raise LikiToolError(f"naming_characters.csv row {i}: invalid stroke for {word!r}")
            tone = int(row["tone"])
            if not 1 <= tone <= 5:
                raise LikiToolError(f"naming_characters.csv row {i}: invalid tone for {word!r}")
            pinyin = row["pinyin"].strip()
            if not pinyin:
                raise LikiToolError(f"naming_characters.csv row {i}: missing pinyin for {word!r}")
            frequency = row["frequency"]
            if frequency not in ("common", "standard", "rare"):
                raise LikiToolError(f"naming_characters.csv row {i}: invalid frequency {frequency!r}")
            element = wuxing_from_chinese(row["wuxing"])
            if not element:
                raise LikiToolError(f"naming_characters.csv row {i}: missing naming element for {word!r}")
            if word in _char_by_char:
                raise LikiToolError(f"naming_characters.csv row {i}: duplicate character {word!r}")
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
            raise LikiToolError(f"negative_chars.txt line {line_no}: unknown character {excluded!r}")
        if excluded in _negative_chars:
            raise LikiToolError(f"negative_chars.txt line {line_no}: duplicate character {excluded!r}")
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
                raise LikiToolError(f"surnames.csv row {line_no}: empty surname or pinyin")
            if not 1 <= len(surname) <= 2:
                raise LikiToolError(f"surnames.csv row {line_no}: invalid surname {surname!r}")
            if " " in plain or not is_latin_lower(plain):
                raise LikiToolError(f"surnames.csv row {line_no}: invalid plain pinyin {plain!r}")
            if tone_from_pinyin(pinyin) == 5:
                raise LikiToolError(f"surnames.csv row {line_no}: pinyin {pinyin!r} must contain a tone mark")
            idx = int(row["baijiaxing_index"])
            if idx != line_no - 1 or idx > 504:
                raise LikiToolError(f"surnames.csv row {line_no}: invalid baijiaxing index")
            if idx in seen_indexes:
                raise LikiToolError(f"surnames.csv row {line_no}: duplicate baijiaxing index {idx}")
            seen_indexes.add(idx)
            aliases: list[str] = []
            for alias in row["aliases"].split(","):
                alias = fold_latin(alias)
                if not alias:
                    continue
                if not is_latin_lower(alias) or " " in alias or alias == plain or alias in aliases:
                    raise LikiToolError(f"surnames.csv row {line_no}: invalid alias {alias!r}")
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

