import csv
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "engine/internal/engine/qiming/data"
SCRIPTS = ROOT / "scripts"
sys.path.insert(0, str(SCRIPTS))

from qiming_projection import load_runtime_naming_rows


def test_runtime_naming_characters_projection():
    source = list(load_runtime_naming_rows(
        DATA / "gsc_pinyin_with_tone.csv",
        DATA / "radicals.yaml",
    ))
    with (DATA / "naming_characters.csv").open(encoding="utf-8-sig", newline="") as fh:
        runtime = list(csv.DictReader(fh))

    assert len(runtime) == 7734
    assert len(runtime) == len(source)
    for source_row, row in zip(source, runtime):
        assert row == {
            "char": source_row["word"],
            "frequency": (
                "common" if int(source_row["num"]) <= 3500
                else "standard" if int(source_row["num"]) <= 6500
                else "rare"
            ),
            "pinyin": source_row["pinyin"],
            "radical": source_row["radical"],
            "stroke": source_row["stroke_count"],
            "wuxing": source_row["wuxing"],
            "tone": source_row["tone"],
        }


def test_naming_character_generator_output(tmp_path):
    output = tmp_path / "naming_characters.csv"
    completed = subprocess.run(
        [sys.executable, str(ROOT / "scripts" / "generate_naming_characters.py"), "--output", str(output)],
        check=False,
        capture_output=True,
        text=True,
    )
    assert completed.returncode == 0, completed.stderr
    assert output.read_bytes() == (DATA / "naming_characters.csv").read_bytes()


def test_naming_character_generator_rejects_missing_source(tmp_path):
    output = tmp_path / "naming_characters.csv"
    completed = subprocess.run(
        [
            sys.executable,
            str(ROOT / "scripts" / "generate_naming_characters.py"),
            "--source",
            str(tmp_path / "missing.csv"),
            "--output",
            str(output),
        ],
        check=False,
        capture_output=True,
        text=True,
    )
    assert completed.returncode != 0
    assert not output.exists()


def test_baijiaxing_surname_table_is_complete_and_canonical():
    path = DATA / "surnames.csv"
    with path.open(encoding="utf-8", newline="") as fh:
        rows = list(csv.DictReader(fh))

    # The classic source has 504 entries. Two surname variants repeat; the
    # runtime matcher therefore exposes 502 unique surnames.
    assert len(rows) == 504
    assert len({row["surname"] for row in rows}) == 502
    assert sum(len(row["surname"]) > 1 for row in rows) == 60
    assert [int(row["baijiaxing_index"]) for row in rows] == list(range(1, 505))
    for row in rows:
        assert 1 <= len(row["surname"]) <= 2
        assert re.fullmatch(r"[a-z]+", row["plain_pinyin"])
        assert re.search(r"[āáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜ]", row["pinyin"])
        aliases = [alias for alias in row["aliases"].split(",") if alias]
        assert len(aliases) == len(set(aliases))
        assert row["plain_pinyin"] not in aliases
        assert all(re.fullmatch(r"[a-z]+", alias) for alias in aliases)
