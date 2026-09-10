import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools"))

from paipan import _shichen_boundary_hint


def test_shichen_boundary_hint_near_hour_boundary():
    hint = _shichen_boundary_hint("1981-08-25T00:35:00+08:00")
    assert hint is not None
    assert hint["near_boundary"] is True
    assert hint["minutes_to_boundary"] == 25
    assert hint["boundary_solar"].endswith("T01:00:00+08:00")
    assert "3-5 件" in hint["message"]


def test_shichen_boundary_hint_absent_far_from_boundary():
    assert _shichen_boundary_hint("1990-06-01T12:00:00+08:00") is None
