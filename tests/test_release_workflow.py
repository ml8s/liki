from __future__ import annotations

import os
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPTS = ROOT / "scripts"


def test_promotion_manifest_script_rejects_missing_input():
    result = subprocess.run(
        [str(SCRIPTS / "validate_candidate_manifest.py"), "--source-commit", "a" * 40],
        env={**os.environ, "DIGESTS_JSON": ""},
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode != 0
    assert "invalid candidate manifest" in result.stderr
