import csv
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
