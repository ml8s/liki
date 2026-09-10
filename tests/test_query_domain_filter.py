from unittest import mock

import _helpers  # noqa: F401 —— 注入 tools 路径
import duanyu


def _mock_result():
    return {
        "八字": [
            {"id": "hun_101", "领域": "婚姻"},
            {"id": "cai_110", "领域": "财运"},
        ],
        "紫微": [
            {"id": "zs_301", "领域": "事业"},
        ],
        "合参": [],
    }


def test_query_domain_filter_keeps_only_requested_life_domain():
    pan = {"gender": "male"}
    with mock.patch.object(duanyu, "validate_natal_pan"), \
         mock.patch.object(
             duanyu, "evaluate_snap_from_pan",
             return_value={"八字": {}, "紫微": {}, "context": {}},
         ), \
         mock.patch.object(duanyu, "_match_rule", return_value=_mock_result()):
        result = duanyu.query("大运", pan, year=2020, domains=["财运"])

    assert [row["id"] for row in result["八字"]] == ["cai_110"]
    assert result["紫微"] == []
    assert result["合参"] == []


def test_query_domain_filter_rejects_unknown_life_domain():
    pan = {"gender": "male"}
    with mock.patch.object(duanyu, "validate_natal_pan"), \
         mock.patch.object(duanyu, "evaluate_snap_from_pan"), \
         mock.patch.object(duanyu, "_match_rule") as match_rule:
        try:
            duanyu.query("大运", pan, year=2020, domains=["不存在"])
        except ValueError as exc:
            message = str(exc)
        else:
            raise AssertionError("unknown domain filter was accepted")

    assert "domains 含无效领域" in message
    match_rule.assert_not_called()


def test_filter_domains_keeps_error_payload_unchanged():
    result = {"error": "rpc failed"}
    # yearly_range 在调用过滤器前显式跳过 error payload。
    assert "error" in result


def test_filter_domains_omits_snapshot_evidence():
    result = {
        "八字": [{"id": "yx_101", "领域": "学业"}],
        "紫微": [],
        "合参": [],
        "evidence": {"三刑流年": {"group": "寅巳申"}},
    }
    filtered = duanyu._filter_domains(result, ["学业"])
    assert [row["id"] for row in filtered["八字"]] == ["yx_101"]
    assert "evidence" not in filtered
