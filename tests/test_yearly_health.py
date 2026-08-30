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
    flow = _flow(bazi)
    ids: set[str] = set()
    for rule in ("年十神", "年旺衰"):   # yj_103(财坏印)在年十神；yj_206/207/208(长生)在年旺衰
        ids |= {item["id"] for item in query_yearly(rule, flow)["八字"]}
    return ids


def test_empty_flow_snapshot_does_not_hit_unconditional_health_rules() -> None:
    assert not {"yj_103", "yj_206", "yj_207", "yj_208"} & _bazi_ids()


def test_health_rules_require_their_named_factors() -> None:
    # 财坏印/日主长生 断语（yj_*）须在对应因子满足时命中（年十神/年旺衰域内与其它取象并存）
    assert "yj_103" in _bazi_ids({"流年财坏印": 1})
    assert "yj_206" in _bazi_ids({"流年日主长生状态": "墓"})
    assert "yj_207" in _bazi_ids({"流年日主长生状态": "绝"})
    assert "yj_208" in _bazi_ids({"流年日主长生状态": "病"})
