"""考时工具契约：流年域白名单与同年事件复用。"""
from unittest import mock

import _helpers  # noqa: F401 —— 注入 tools 路径
import calibrate


def test_calibrate_accepts_yingqi_and_reuses_same_year_snapshot() -> None:
    candidate = {
        "label": "25日",
        "gregorian": "1981-08-25T00:15:00+08:00",
        "gender": "male",
        "longitude": 130.3,
    }
    second_candidate = {
        "label": "26日",
        "gregorian": "1981-08-26T00:15:00+08:00",
        "gender": "male",
        "longitude": 130.3,
    }
    events = [
        {"year": 2010, "label": "感情重大问题", "rule": "yingqi"},
        {"year": 2010, "label": "同一年验证", "rule": "yearly_marriage"},
        {"year": 2010, "label": "同年第三件事", "rule": "yearly_career"},
    ]
    pan = {"base": {}, "gender": "male"}
    context = object()

    def fake_year(pan, year, rules, detail=False, natal_context=None):
        return {rule: {"八字": [], "紫微": []} for rule in rules}

    with mock.patch.object(calibrate, "full_paipan", return_value=pan) as paipan_mock, \
         mock.patch.object(calibrate, "prepare_natal_context", return_value=context) as prepare_mock, \
         mock.patch.object(calibrate, "yearly_snapshot", return_value={"_snapshot_type": "liunian", "八字": {}, "紫微": {}}) as snapshot_mock:
        result = calibrate.calibrate([candidate, second_candidate], events)

    assert len(result["25日"]) == 3
    assert len(result["26日"]) == 3
    assert paipan_mock.call_count == 2
    assert prepare_mock.call_count == 2
    assert snapshot_mock.call_count == 2   # 同一年份快照只生成一次（每候选）


def test_calibrate_enforces_documented_candidate_and_event_counts() -> None:
    candidate = {
        "label": "25日",
        "gregorian": "1981-08-25T00:15:00+08:00",
        "gender": "male",
        "longitude": 130.3,
    }
    event = {"year": 2010, "label": "感情重大问题", "rule": "yingqi"}

    with mock.patch.object(calibrate, "full_paipan") as paipan_mock:
        try:
            calibrate.calibrate([candidate], [event] * 3)
        except ValueError as exc:
            assert "candidates 必须 2-3 个" in str(exc)
        else:
            raise AssertionError("单个候选未拒绝")

        try:
            calibrate.calibrate([candidate, {**candidate, "label": "26日"}], [event] * 2)
        except ValueError as exc:
            assert "events 必须 3-5 件" in str(exc)
        else:
            raise AssertionError("两个事件未拒绝")

    paipan_mock.assert_not_called()
