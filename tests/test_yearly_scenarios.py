"""Scenario alias must stay domain-scoped."""

from pathlib import Path

import sys

import json
import csv

ROOT = Path(__file__).resolve().parents[1]
constants = json.loads((ROOT / "skills/liki-bazi/tools/constants.json").read_text(encoding="utf-8"))
sys.path.insert(0, str(ROOT / "skills/liki-bazi/tools"))

from duanyu import SCENE_ALIASES  # noqa: E402
from duanyu import _default_scene_domains  # noqa: E402


def test_yearly_study_excludes_family_and_marriage_domain() -> None:
    source = constants["命理域"]["场景别名"]["yearly_study"]
    resolved = SCENE_ALIASES["yearly_study"]
    assert "年六亲" not in source
    assert "年六亲" not in resolved
    assert {"年十神", "年神煞", "年用神", "年官禄"} <= set(resolved)


def test_yearly_study_scene_has_dedicated_study_assertions():
    source = ROOT / "skills/liki-bazi/tools/assertions/assertions.csv"
    with source.open(encoding="utf-8", newline="") as fh:
        rows = list(csv.DictReader(fh))

    alias = set(constants["命理域"]["场景别名"]["yearly_study"])
    study_ids = {
        row["assertion_id"]
        for row in rows
        if row["rule"] in alias and row["领域"] == "学业"
    }
    # 覆盖八字结构 / 神煞 / 用神与紫微官禄四类学业信号，避免场景聚合退化为空集。
    assert {"yx_101", "yx_102", "yx_205", "yx_101"} <= study_ids
    assert {"yxz_201", "yxz_202", "yxz_203"} <= study_ids


def test_yearly_study_has_default_domain_scope():
    assert _default_scene_domains(["yearly_study"]) == ["学业"]
    assert _default_scene_domains(["yingqi"]) is None
