from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]


def test_release_profiles_match_runtime() -> None:
    runtime = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()

    for profile in (ROOT / "profiles").glob("*.json"):
        document = yaml.safe_load(profile.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == runtime
