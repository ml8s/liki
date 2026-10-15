"""回归测试：#123 透干口径、#129 七杀制化、#131 性别门控。"""

import duanyu
import factors
from _helpers import mock_base_context


def _hidden_shangguan_ke_guan_snapshot():
    base = mock_base_context(
        伤官={
            "count": 1,
            "wuxing": "金",
            "transparent": False,
            "hidden": True,
            "rooted": True,
            "timely": True,
        },
        正官={"count": 1, "wuxing": "木", "transparent": True},
    )
    base["element_states"] = {
        "金": {"controls": "木"},
        "木": {"controls": "土"},
    }
    return factors.evaluate_factors("female", base, shushi="bazi")


def test_hun_203_does_not_claim_transparent_shangguan_when_hidden():
    snap = _hidden_shangguan_ke_guan_snapshot()
    assert snap["伤官克官"] == 1
    assert snap["伤官透干克官"] == 0

    snapshots = {
        "八字": {**snap, "配偶星得地": 1},
        "紫微": {},
        "context": {"性别": "female"},
    }
    ids = [row["id"] for row in duanyu.match_rule("十神", snapshots)["八字"]]
    assert "hun_203" not in ids


def test_hun_203_accepts_genuine_transparent_shangguan_ke_guan():
    snap = _hidden_shangguan_ke_guan_snapshot()
    snap["伤官透干克官"] = 1
    snapshots = {
        "八字": {**snap, "配偶星得地": 1},
        "紫微": {},
        "context": {"性别": "female"},
    }
    ids = [row["id"] for row in duanyu.match_rule("十神", snapshots)["八字"]]
    assert "hun_203" in ids


def test_ge_202_treats_wang_pianyin_as_killing_control():
    snapshots = {
        "八字": {
            "月令格": "七杀格",
            "食神旺": 0,
            "正印旺": 0,
            "偏印旺": 1,
            "印星旺": 1,
        },
        "紫微": {},
        "context": {"性别": "male"},
    }
    ids = [row["id"] for row in duanyu.match_rule("格局", snapshots)["八字"]]
    assert "ge_202" not in ids


def test_shi_108_is_gated_to_female_charts():
    factors = {
        "女命官星得地格": 0,
        "财星透根": 1,
    }
    male = {
        "八字": factors,
        "紫微": {},
        "context": {"性别": "male"},
    }
    female = {
        "八字": factors,
        "紫微": {},
        "context": {"性别": "female"},
    }
    assert "shi_108" not in [row["id"] for row in duanyu.match_rule("十神", male)["八字"]]
    assert "shi_108" in [row["id"] for row in duanyu.match_rule("十神", female)["八字"]]
