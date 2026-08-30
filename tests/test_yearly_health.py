"""流年健康表条件契约：禁止无条件恒命中。"""
import _helpers  # noqa: F401 —— 注入 tools 路径
from duanyu import query_yearly


def _flow(bazi: dict | None = None) -> dict:
    return {
        "_snapshot_type": "liunian",
        "八字": bazi or {},
        "紫微": {},
    }


def _bazi_ids(bazi: dict | None = None) -> set[str]:
    return {item["id"] for item in query_yearly("yearly_health", _flow(bazi))["八字"]}


def test_empty_flow_snapshot_does_not_hit_unconditional_health_rules() -> None:
    assert not {"yj_103", "yj_206", "yj_207", "yj_208"} & _bazi_ids()


def test_health_rules_require_their_named_factors() -> None:
    assert _bazi_ids({"流年财坏印": 1}) == {"yj_103"}
    assert _bazi_ids({"流年日主长生状态": "墓"}) == {"yj_206"}
    assert _bazi_ids({"流年日主长生状态": "绝"}) == {"yj_207"}
    assert _bazi_ids({"流年日主长生状态": "病"}) == {"yj_208"}
