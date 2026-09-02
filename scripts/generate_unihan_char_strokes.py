#!/usr/bin/env python3
"""Generate qiming source and runtime Kangxi character tables from Unicode Unihan.

The input name table supplies the set of publishable characters and its preferred
traditional form for ambiguous simplified characters.  Stroke values are derived
only from Unihan kRSUnicode; kTotalStrokes and kKangXi are never used as stroke
values.  For characters with multiple kRSUnicode entries, the Kangxi dictionary
page selects the radical entry that belongs to the Kangxi classification.
"""
from __future__ import annotations

import argparse
import csv
import json
import zipfile
from collections import defaultdict
from pathlib import Path

from qiming_projection import load_runtime_naming_rows

REPO = Path(__file__).resolve().parents[1]
DEFAULT_CHAR_TABLE = REPO / "engine/internal/engine/qiming/data/gsc_pinyin_with_tone.csv"
DEFAULT_RADICALS = REPO / "engine/internal/engine/qiming/data/radicals.yaml"
DEFAULT_OUTPUT = REPO / "engine/internal/engine/qiming/data/unihan_char_strokes.csv"
DEFAULT_RUNTIME_OUTPUT = REPO / "engine/internal/engine/qiming/data/kangxi_character_strokes.csv"
DEFAULT_UNIHAN = Path("/tmp/Unihan/Unihan.zip")

# Kangxi radical number -> radical stroke count.
RADICAL_STROKES = {
    **{n: 1 for n in range(1, 7)},
    **{n: 2 for n in range(7, 30)},
    **{n: 3 for n in range(30, 61)},
    **{n: 4 for n in range(61, 95)},
    **{n: 5 for n in range(95, 118)},
    **{n: 6 for n in range(118, 147)},
    **{n: 7 for n in range(147, 167)},
    **{n: 8 for n in range(167, 176)},
    **{n: 9 for n in range(176, 187)},
    **{n: 10 for n in range(187, 195)},
    **{n: 11 for n in range(195, 201)},
    **{n: 12 for n in range(201, 205)},
    **{n: 13 for n in range(205, 209)},
    **{n: 14 for n in range(209, 211)},
    **{n: 15 for n in range(211, 212)},
    **{n: 16 for n in range(212, 214)},
    214: 17,
}

# Keep a semantic default only where Unihan's traditional variants have no single
# non-identity candidate or where the character table's preferred form is not a
# candidate.  These are defaults for the given-name context.
FORM_OVERRIDES = {
    "宁": "寧",
    "𫇭": "蒍",
}

def parse_unihan_files(files: dict[str, str]) -> tuple[dict, dict, str]:
    data: dict[tuple[str, str], list[str]] = defaultdict(list)
    unicode_version = ""

    for name in ("Unihan_IRGSources.txt", "Unihan_Variants.txt", "Unihan_DictionaryIndices.txt"):
        for line in files[name].splitlines():
            if not line.strip():
                continue
            if line.startswith("# Unicode Version "):
                unicode_version = line.removeprefix("# Unicode Version ").strip()
            if line.startswith("#"):
                continue
            codepoint, field, value = line.split("\t")
            data[(codepoint, field)].append(value)

    if not unicode_version:
        raise ValueError("Unihan data does not declare its Unicode version")
    return data, dict(data), unicode_version


def read_unihan(path: Path) -> dict[str, str]:
    if path.is_dir():
        return {p.name: p.read_text(encoding="utf-8") for p in (
            path / "Unihan_IRGSources.txt",
            path / "Unihan_Variants.txt",
            path / "Unihan_DictionaryIndices.txt",
        )}
    with zipfile.ZipFile(path) as archive:
        return {name: archive.read(name).decode("utf-8") for name in (
            "Unihan_IRGSources.txt",
            "Unihan_Variants.txt",
            "Unihan_DictionaryIndices.txt",
        )}


def radical_number(expression: str) -> int | None:
    token = expression.split(".", 1)[0]
    return int(token) if token.isdigit() else None


def load_unihan(data: dict) -> tuple[dict[str, list[str]], dict[str, list[str]], dict[str, str]]:
    rs_by_cp = {cp: values[0] for (cp, field), values in data.items() if field == "kRSUnicode"}
    variants_by_cp = {
        cp: values[0].split()
        for (cp, field), values in data.items()
        if field == "kTraditionalVariant"
    }

    kangxi_by_cp = {}
    for cp in rs_by_cp:
        for field in ("kIRGKangXi", "kKangXi"):
            if data.get((cp, field)):
                kangxi_by_cp[cp] = data[(cp, field)][0]
                break

    # Infer the first page occupied by each radical from characters with one
    # radical classification. Kangxi pages are ordered by radical number, so the
    # nearest preceding start page resolves characters with multiple kRSUnicode
    # entries (for example 萬 has both 114.8 and 140.9).
    pages_by_radical: dict[int, list[int]] = defaultdict(list)
    for cp, expressions in rs_by_cp.items():
        numbers = {radical_number(item) for item in expressions.split()}
        if len(numbers) != 1 or None in numbers:
            continue
        page_source = kangxi_by_cp.get(cp)
        if not page_source:
            continue
        try:
            page = int(page_source.split(".", 1)[0])
        except ValueError:
            continue
        codepoint = int(cp[2:], 16)
        if 0x4E00 <= codepoint <= 0x9FFF:
            pages_by_radical[numbers.pop()].append(page)

    starts = {number: min(pages) for number, pages in pages_by_radical.items()}
    missing = [n for n in RADICAL_STROKES if n not in starts]
    if missing:
        raise ValueError(f"cannot infer Kangxi start pages for radicals: {missing}")
    return rs_by_cp, variants_by_cp, starts


def select_krs_entry(expressions: str, starts: dict[int, int], page: int | None) -> str:
    items = expressions.split()
    if len(items) == 1:
        return items[0]
    if page is None:
        raise ValueError(f"multiple kRSUnicode entries without a Kangxi page: {expressions}")

    candidates = []
    for item in items:
        number = radical_number(item)
        if number is not None and number in starts and starts[number] <= page:
            candidates.append((starts[number], number, item))
    if not candidates:
        raise ValueError(f"no Kangxi radical entry matches page {page}: {expressions}")
    return max(candidates)[2]


def char_codepoint(char: str) -> str:
    return f"U+{ord(char):04X}"


def preferred_form(char: str, traditional: str, variants: list[str], unicode_version: str) -> tuple[str, str]:
    override = FORM_OVERRIDES.get(char)
    if override:
        if not variants or f"U+{ord(override):04X}" in variants:
            return override, "repo:form-override"

    if not variants:
        return char, "unihan:identity"

    candidates = [item for item in variants if item != char_codepoint(char)]
    if not candidates:
        return char, "unihan:identity"

    first_declared = traditional.split("、", 1)[0] if traditional else ""
    if len(first_declared) == 1:
        declared_cp = char_codepoint(first_declared)
        if declared_cp in candidates:
            return first_declared, "char-table:preferred-traditional+unihan:kTraditionalVariant"

    selected_cp = candidates[0]
    return chr(int(selected_cp[2:], 16)), f"unihan:kTraditionalVariant[{unicode_version}]"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--unihan", type=Path, default=DEFAULT_UNIHAN)
    parser.add_argument("--char-table", type=Path, default=DEFAULT_CHAR_TABLE)
    parser.add_argument("--radicals", type=Path, default=DEFAULT_RADICALS)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--runtime-output", type=Path, default=DEFAULT_RUNTIME_OUTPUT)
    parser.add_argument("--manifest", type=Path, default=DEFAULT_OUTPUT.with_suffix(".json"))
    args = parser.parse_args()

    files = read_unihan(args.unihan)
    data, _, unicode_version = parse_unihan_files(files)
    rs_by_cp, variants_by_cp, radical_starts = load_unihan(data)

    seen: set[str] = set()
    declared_traditional: dict[str, str] = {}
    with args.char_table.open(encoding="utf-8-sig", newline="") as fh:
        for row in csv.DictReader(fh):
            char = row["word"]
            if not char or len(char) != 1 or char in seen:
                continue
            seen.add(char)
            traditional = row["traditional"]
            declared_traditional[char] = "" if traditional in ("", "NULL") else traditional

    rows = []
    for char in sorted(seen):
        source_cp = char_codepoint(char)
        variants = variants_by_cp.get(source_cp, [])
        form, form_source = preferred_form(char, declared_traditional[char], variants, unicode_version)
        form_cp = char_codepoint(form)
        expressions = rs_by_cp.get(form_cp)
        if not expressions:
            raise ValueError(f"{char}: form {form} has no kRSUnicode")

        page_source = data.get((form_cp, "kIRGKangXi")) or data.get((form_cp, "kKangXi"))
        page = int(page_source[0].split(".", 1)[0]) if page_source else None
        expression = select_krs_entry(expressions, radical_starts, page)
        radical_no, residual = (int(part) for part in expression.split("."))
        stroke = RADICAL_STROKES[radical_no] + residual

        rows.append({
            "char": char,
            "unihan_stroke": stroke,
            "unihan_form": form,
            "form_source": form_source,
            "radical_no": radical_no,
            "residual": residual,
            "unicode_version": unicode_version,
        })

    expected_cases = {
        "万": 15, "叶": 15, "冯": 12, "刘": 15, "张": 11,
        "陈": 16, "郑": 19, "沈": 19, "胡": 19, "姜": 19,
        "温": 14, "蔡": 17, "魏": 18,
    }
    generated = {row["char"]: row["unihan_stroke"] for row in rows}
    for char, expected in expected_cases.items():
        actual = generated.get(char)
        if actual != expected:
            raise ValueError(f"{char}: unihan_stroke={actual}, want {expected}")

    with args.output.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=list(rows[0]), lineterminator="\n")
        writer.writeheader()
        writer.writerows(rows)

    runtime_fields = ["char", "kangxi_form", "kangxi_stroke"]
    runtime_chars = {
        row["word"]
        for row in load_runtime_naming_rows(args.char_table, args.radicals)
    }
    runtime_rows = [row for row in rows if row["char"] in runtime_chars]
    with args.runtime_output.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=runtime_fields, lineterminator="\n")
        writer.writeheader()
        writer.writerows({
            "char": row["char"],
            "kangxi_form": row["unihan_form"],
            "kangxi_stroke": row["unihan_stroke"],
        } for row in runtime_rows)

    manifest = {
        "unicode_version": unicode_version,
        "source": "Unicode Unihan Database (Unihan.zip)",
        "source_url": "https://www.unicode.org/Public/UCD/latest/ucd/Unihan.zip",
        "license": "Unicode License V3",
        "license_url": "https://www.unicode.org/license.txt",
        "copyright": "Copyright © 1991-2026 Unicode, Inc.",
        "fields": ["kRSUnicode", "kTraditionalVariant", "kIRGKangXi", "kKangXi"],
        "stroke_rule": "Unihan kRSUnicode radical strokes + residual",
        "multi_radical_rule": "select the kRSUnicode entry whose radical starts at the nearest preceding Kangxi page",
        "character_count": len(rows),
        "runtime_character_count": len(runtime_rows),
        "runtime_output": args.runtime_output.name,
        "runtime_fields": runtime_fields,
        "generated_from": args.unihan.name,
        "expected_cases": expected_cases,
    }
    args.manifest.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {len(rows)} rows to {args.output}")
    print(f"wrote {len(runtime_rows)} rows to {args.runtime_output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
