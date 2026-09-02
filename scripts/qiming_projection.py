"""Shared projection rules for qiming runtime data."""
from __future__ import annotations

import csv
from pathlib import Path
from typing import Iterator

import yaml


NAMING_ELEMENTS = {"木", "火", "土", "金", "水"}


def load_runtime_naming_rows(source: Path, radicals: Path) -> Iterator[dict[str, str]]:
    radical_data = yaml.safe_load(radicals.read_text(encoding="utf-8"))
    radical_elements = {
        radical: element
        for element, values in radical_data.items()
        for radical in values
    }
    with source.open(encoding="utf-8-sig", newline="") as fh:
        for row in csv.DictReader(fh):
            wuxing = row["wuxing"]
            if wuxing not in NAMING_ELEMENTS:
                wuxing = radical_elements.get(row["radical"])
            if wuxing is None:
                continue
            yield {**row, "wuxing": wuxing}
