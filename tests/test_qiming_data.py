import csv
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "engine/internal/engine/qiming/data"
SCRIPTS = ROOT / "scripts"
sys.path.insert(0, str(SCRIPTS))

from qiming_projection import load_runtime_naming_rows
SCRIPTS = ROOT / "scripts"
sys.path.insert(0, str(SCRIPTS))


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
            "pinyin": source_row["pinyin"],
            "radical": source_row["radical"],
            "stroke": source_row["stroke_count"],
            "wuxing": source_row["wuxing"],
            "tone": source_row["tone"],
        }


def test_runtime_kangxi_character_projection():
    with (DATA / "unihan_char_strokes.csv").open(encoding="utf-8-sig", newline="") as fh:
        source = list(csv.DictReader(fh))
    with (DATA / "kangxi_character_strokes.csv").open(encoding="utf-8-sig", newline="") as fh:
        runtime = list(csv.DictReader(fh))
    naming_rows = load_runtime_naming_rows(
        DATA / "gsc_pinyin_with_tone.csv",
        DATA / "radicals.yaml",
    )
    naming_chars = {row["word"] for row in naming_rows}

    assert len(runtime) == 7734
    source = [row for row in source if row["char"] in naming_chars]
    assert len(source) == len(runtime)
    for source_row, row in zip(source, runtime):
        assert row == {
            "char": source_row["char"],
            "kangxi_form": source_row["unihan_form"],
            "kangxi_stroke": source_row["unihan_stroke"],
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
