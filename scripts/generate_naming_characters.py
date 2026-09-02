#!/usr/bin/env python3
"""Generate the runtime qiming character table from the source GSC table."""
from __future__ import annotations

import argparse
import csv
from pathlib import Path

from qiming_projection import load_runtime_naming_rows


REPO = Path(__file__).resolve().parents[1]
DEFAULT_SOURCE = REPO / "engine/internal/engine/qiming/data/gsc_pinyin_with_tone.csv"
DEFAULT_RADICALS = REPO / "engine/internal/engine/qiming/data/radicals.yaml"
DEFAULT_OUTPUT = REPO / "engine/internal/engine/qiming/data/naming_characters.csv"
FIELDS = ["char", "pinyin", "radical", "stroke", "wuxing", "tone"]
SOURCE_FIELDS = {"word", "pinyin", "radical", "stroke_count", "wuxing", "tone"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_SOURCE)
    parser.add_argument("--radicals", type=Path, default=DEFAULT_RADICALS)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    args = parser.parse_args()

    with args.source.open(encoding="utf-8-sig", newline="") as source:
        rows = csv.DictReader(source)
        missing = SOURCE_FIELDS - set(rows.fieldnames or [])
        if missing:
            raise ValueError(f"source table missing columns: {sorted(missing)}")
        with args.output.open("w", encoding="utf-8", newline="") as output:
            writer = csv.DictWriter(output, fieldnames=FIELDS, lineterminator="\n")
            writer.writeheader()
            count = 0
            for row in load_runtime_naming_rows(args.source, args.radicals):
                writer.writerow({
                    "char": row["word"],
                    "pinyin": row["pinyin"],
                    "radical": row["radical"],
                    "stroke": row["stroke_count"],
                    "wuxing": row["wuxing"],
                    "tone": row["tone"],
                })
                count += 1

    if count != 7734:
        raise ValueError(f"generated {count} rows, want 7734")
    print(f"wrote {count} rows to {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
