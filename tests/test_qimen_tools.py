"""奇门 Python 编排层：事象路由与解释表只消费 engine 输出。"""
from __future__ import annotations

import json
import csv
import importlib.util
import sys
from pathlib import Path

import pytest
from jsonschema import validate


TOOLS = Path(__file__).resolve().parents[1] / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import divination_rpc  # noqa: E402
import qimen_paipan as paipan  # noqa: E402
from qimen_duanyu import query  # noqa: E402
from qimen_factors import (  # noqa: E402
    load_snapshot_contract,
    project_qimen_snapshot,
    validate_qimen_snapshot,
)
from qimen_errors import TableError  # noqa: E402
from qimen_interpretations import (  # noqa: E402
    load_interpretation_index,
    load_rule_table,
)
from qimen_matters import load_matter_table as load_routing_matter_table  # noqa: E402

sys.path.remove(str(TOOLS))


def _load_agent_cli():
    spec = importlib.util.spec_from_file_location(
        "qimen_agent_cli", TOOLS / "agent_cli.py"
    )
    module = importlib.util.module_from_spec(spec)
    assert spec is not None and spec.loader is not None
    sys.path.insert(0, str(TOOLS))
    try:
        spec.loader.exec_module(module)
    finally:
        sys.path.remove(str(TOOLS))
    return module


def _pan(**changes):
    chart = {
        "method": {
            "scope": "hour", "school": "zhuanpan",
            "spirit_mode": "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu",
        },
        "pan": {
            "yin_dun": False,
        "ri_gan": "甲",
        "ri_zhi": "辰",
        "nian_gan": "丙",
        "nian_zhi": "午",
        "yue_gan": "丙",
        "yue_zhi": "申",
        "shi_gan": "甲",
        "shi_zhi": "午",
        "wu_bu_yu_shi": False,
            "gong_wei": [
                {"gong": {"name": "坎", "luoshu": 1}},
                {"gong": {"name": "坤", "luoshu": 2}},
                {"gong": {"name": "震", "luoshu": 3}},
                {"gong": {"name": "巽", "luoshu": 4}},
                {"gong": {"name": "离", "luoshu": 9}},
                {"gong": {"name": "艮", "luoshu": 8}},
                {"gong": {"name": "兑", "luoshu": 7}},
                {"gong": {"name": "乾", "luoshu": 6}},
            ],
        },
        "ri_gan_gong": "震",
        "shi_gan_gong": "兑",
        "shi_gan_domain": "outer",
        "ri_shi_relation": {
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "same",
            "name": "日干宫与时干宫比和",
        },
        "patterns": [],
        "kong_wang_affected": [],
        "ma_xing_affected": [],
        "palace_wang_shuai": [],
        "shi_gan_gong_wang_shuai": {
            "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
        },
        "ying_qi": {"candidates": []},
        "gan_interaction": [],
        "men_interaction": [],
        "xing_interaction": [],
        "xing_gong_wu_xing": [],
        "men_po": [],
        "men_zhi": [],
        "zhi_fu_xing_gong": "坎",
        "zhi_shi_men_gong": "坎",
    }
    chart.update(changes)
    return {"matter": None, "chart": chart}


def _lost_pan(**changes):
    pan = _pan()
    pan["chart"].update({
        "patterns": [{
            "name": "反吟",
            "gong_wei": ["坎"],
            "auspicious": False,
        }],
        "kong_wang_affected": [{
            "symbol": "时干",
            "role": "pillar",
            "branch": "戌",
            "gong": "乾",
        }],
        "ri_shi_relation": {
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "shi_generates_ri",
            "name": "时干宫生日干宫",
        },
        "palace_wang_shuai": [{
            "gong": "兑",
            "wuxing": "金",
            "wang_shuai": "旺",
            "wang_shuai_name": "旺",
        }],
        "shi_gan_gong_wang_shuai": {
            "gong": "兑",
            "wuxing": "金",
            "wang_shuai": "旺",
            "wang_shuai_name": "旺",
        },
    })
    pan["chart"].update(changes)
    return pan


def _thief_pan(**changes):
    pan = _pan()
    pan["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坤", "luoshu": 2},
            "tian_pan": [{"gan": "己", "xing": "天芮"}],
            "men": "死门",
            "men_present": True,
            "shen": "勾陈",
            "shen_present": True,
        },
        {
            "gong": {"name": "坎", "luoshu": 1},
            "tian_pan": [{"gan": "戊", "xing": "天蓬"}],
            "men": "休门",
            "men_present": True,
            "shen": "朱雀",
            "shen_present": True,
        },
    ]
    pan["chart"].update({
        "palace_wang_shuai": [
            {"gong": "坤", "wuxing": "土", "wang_shuai": "相", "wang_shuai_name": "相"},
            {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
    })
    pan["chart"].update(changes)
    return pan


def _rich_pan():
    rich = _pan(
        ri_shi_relation={
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "shi_generates_ri",
            "name": "时干宫生日干宫",
        },
        patterns=[
            {"name": "反吟", "gong_wei": ["坎"], "auspicious": False},
            {"name": "青龙返首", "gong_wei": ["震"], "auspicious": True},
        ],
        kong_wang_affected=[{
            "symbol": "时干", "role": "pillar", "branch": "戌", "gong": "乾",
        }],
        ma_xing_affected=[{
            "symbol": "日干", "role": "pillar", "branch": "寅", "gong": "艮",
        }],
        ying_qi={"candidates": [
            {
                "type": "ma_xing", "branch": "寅", "gong": "艮",
                "related_to": [{"symbol": "日干", "role": "pillar"}],
            },
            {
                "type": "kong_wang", "branch": "戌", "gong": "乾",
                "related_to": [{"symbol": "时干", "role": "pillar"}],
            },
        ]},
        gan_interaction=[{
            "name": "青龙返首", "gong": "震",
            "tian_pan_gan": "戊", "di_pan_gan": "丙", "auspicious": True,
        }],
        men_interaction=[{"name": "门迫", "door": "休门", "gong": "坎"}],
        xing_interaction=[{
            "name": "星克宫", "xing": "天蓬", "gong": "离", "auspicious": False,
        }],
        xing_gong_wu_xing=[{
            "xing": "天蓬", "gong": "离", "relation": "xing_controls_gong",
            "relation_name": "星克宫", "traditional_label": "囚",
        }],
        palace_wang_shuai=[
            {"gong": "离", "wuxing": "火", "wang_shuai": "旺", "wang_shuai_name": "旺"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
        shi_gan_gong_wang_shuai={
            "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
        },
        men_po=["坎"],
        men_zhi=["离"],
        zhi_fu_xing_gong="乾",
        zhi_shi_men_gong="开",
        yong_shen={
            "nian_gan_gong": "乾",
            "symbols": [
                {
                    "symbol": "生门",
                    "palace": "艮",
                    "tian_gan": ["戊"],
                    "kong_wang_branches": [],
                    "ma_xing_branch": None,
                },
                {
                    "symbol": "戊",
                    "palace": "震",
                    "tian_gan": ["庚"],
                    "kong_wang_branches": [],
                    "ma_xing_branch": None,
                },
            ],
        },
        method={
            "scope": "hour", "school": "zhuanpan",
            "spirit_mode": "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu",
        },
    )
    rich["chart"]["pan"]["gong_wei"] = [{
        "gong": {"name": "兑", "luoshu": 7},
        "tian_pan": [{"gan": "庚", "xing": "天蓬"}],
        "men": "杜门",
        "men_present": True,
        "shen": "勾陈",
        "shen_present": True,
    }]
    return rich


def _missing_pan(**changes):
    pan = _pan()
    pan["chart"]["shi_gan_gong"] = "坎"
    pan["chart"]["yong_shen"] = {
        "nian_gan_gong": None,
        "symbols": [{
            "symbol": "六合", "palace": "坎", "tian_gan": ["戊"],
            "kong_wang_branches": [], "ma_xing_branch": None,
        }],
    }
    pan["chart"]["gong_domains"] = [
        {"gong": "坎", "domain": "inner"},
        {"gong": "兑", "domain": "outer"},
    ]
    pan["chart"].update(changes)
    return pan


def _escape_pan(**changes):
    pan = _pan()
    pan["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坎", "luoshu": 1},
            "tian_pan": [{"gan": "戊", "xing": "天蓬"}],
            "men": "休门",
            "men_present": True,
            "shen": "六合",
            "shen_present": True,
        },
        {
            "gong": {"name": "震", "luoshu": 3},
            "tian_pan": [{"gan": "庚", "xing": "天冲"}],
            "men": "伤门",
            "men_present": True,
            "shen": "值符",
            "shen_present": True,
        },
    ]
    pan["chart"].update({
        "palace_wang_shuai": [
            {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
            {"gong": "震", "wuxing": "木", "wang_shuai": "休", "wang_shuai_name": "休"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
        "gan_interaction": [{
            "name": "伏吟", "gong": "坎", "tian_pan_gan": "戊",
            "di_pan_gan": "戊", "auspicious": False,
        }],
    })
    pan["chart"].update(changes)
    return pan


def test_matter_table_is_the_python_routing_fact(monkeypatch) -> None:
    matters = load_routing_matter_table()
    assert matters["wealth"]["name"] == "求财"
    assert matters["wealth"]["symbols"] == ["生门", "戊"]
    assert matters["missing_person"]["symbols"] == ["六合"]
    assert all(row["basis"] for row in matters.values())


def test_all_matter_requests_satisfy_engine_schema(monkeypatch) -> None:
    schema = json.loads(
        (Path(__file__).resolve().parents[1] / "engine/internal/agent/qimen_schema.json").read_text(
            encoding="utf-8"
        )
    )["params"]
    requests = []

    def fake_call(method: str, params: dict) -> dict:
        assert method == "qimen.chart"
        requests.append(params)
        return {"data": {}}

    monkeypatch.setattr(paipan, "call", fake_call)
    for matter in load_routing_matter_table():
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", matter=matter)
    assert len(requests) == len(load_routing_matter_table())
    for params in requests:
        validate(params, schema)


def test_qimen_chart_maps_matter_before_engine_call(monkeypatch) -> None:
    calls = []

    def fake_call(method: str, params: dict) -> dict:
        calls.append((method, params))
        return {"data": {"ri_gan_gong": "震", "shi_gan_gong": "兑"}}

    monkeypatch.setattr(paipan, "call", fake_call)
    result = paipan.qimen_chart(
        "2026-09-07T12:00:00+08:00",
        matter="wealth",
        birth_date="1984-02-15",
        scope="hour",
        school="zhuanpan",
    )

    assert len(calls) == 1
    method, params = calls[0]
    assert method == "qimen.chart"
    assert "matter" not in params
    assert params["yong_shen"] == ["生门", "戊"]
    assert params["birth_date"] == "1984-02-15"
    assert result["matter"]["matter"] == "wealth"
    assert result["matter"]["name"] == "求财"
    assert result["chart"]["ri_gan_gong"] == "震"


def test_qimen_chart_rejects_duplicate_yong_shen_sources(monkeypatch) -> None:
    monkeypatch.setattr(paipan, "call", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    try:
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", matter="wealth", yong_shen=["生门"])
    except ValueError as exc:
        assert "互斥" in str(exc)
    else:
        raise AssertionError("matter and yong_shen must conflict")


def test_qimen_chart_rejects_empty_explicit_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "call",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 不能为空"):
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", yong_shen=[])


def test_python_tool_schema_rejects_empty_or_duplicate_yong_shen() -> None:
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    qimen_read = next(
        item["function"]["parameters"]
        for item in schema["tools"]
        if item["function"]["name"] == "qimen_snapshot"
    )["properties"]["yong_shen"]
    assert qimen_read["minItems"] == 1
    assert qimen_read["uniqueItems"] is True


def test_qimen_chart_rejects_duplicate_explicit_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "call",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 符号重复"):
        paipan.qimen_chart(
            "2026-09-07T12:00:00+08:00", yong_shen=["生门", "生门"]
        )


def test_qimen_chart_rejects_string_yong_shen(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "call",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="yong_shen 必须是字符串数组"):
        paipan.qimen_chart("2026-09-07T12:00:00+08:00", yong_shen="生门")


def test_matter_table_rejects_duplicate_symbols(tmp_path, monkeypatch) -> None:
    import qimen_matters

    matters = tmp_path / "matters.csv"
    matters.write_text(
        "matter,name,symbols,basis\n"
        "wealth,求财,生门、生门、戊,test\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_matters, "_MATTER_TABLE", None)
    monkeypatch.setattr(qimen_matters, "MATTERS_PATH", matters)
    with pytest.raises(TableError, match="重复符号"):
        qimen_matters.load_matter_table()


def test_rpc_call_reports_malformed_engine_payload(monkeypatch) -> None:
    class Response:
        def __init__(self, payload):
            self.payload = payload

        def read(self):
            return self.payload

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

    monkeypatch.setattr(
        divination_rpc.urllib.request,
        "urlopen",
        lambda *_, **__: Response(b"not-json"),
    )
    with pytest.raises(paipan.RPCError, match="malformed JSON-RPC response"):
        paipan.call("qimen.chart", {"solar_time": "x"})


def test_rpc_call_reports_missing_result(monkeypatch) -> None:
    class Response:
        def __init__(self, payload):
            self.payload = payload

        def read(self):
            return self.payload

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

    monkeypatch.setattr(
        divination_rpc.urllib.request,
        "urlopen",
        lambda *_, **__: Response(b'{"jsonrpc":"2.0","id":1}'),
    )
    with pytest.raises(paipan.RPCError, match="missing result"):
        paipan.call("qimen.chart", {"solar_time": "x"})


def test_qimen_chart_reports_missing_engine_data(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan,
        "call",
        lambda *_: {"_product": "qimen"},
    )
    with pytest.raises(paipan.RPCError, match="missing data object"):
        paipan.qimen_chart("2026-09-07T12:00:00+08:00")


def test_rpc_retries_server_error_but_not_client_error(monkeypatch) -> None:
    class Response:
        def __init__(self, payload):
            self.payload = payload

        def read(self):
            return self.payload

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

    calls = []

    def urlopen(*_, **__):
        if not calls:
            calls.append("server-error")
            raise paipan.HTTPError(
                "https://example.test", 500, "server error", {}, None
            )
        calls.append("ok")
        return Response(b'{"jsonrpc":"2.0","id":1,"result":{"data":{"ok":true}}}')

    monkeypatch.setattr(divination_rpc.urllib.request, "urlopen", urlopen)
    result = paipan.engine_data("qimen.chart", {"solar_time": "x"})
    assert result == {"ok": True}
    assert calls == ["server-error", "ok"]

    client_error = paipan.HTTPError(
        "https://example.test", 403, "forbidden", {}, None
    )
    monkeypatch.setattr(
        divination_rpc.urllib.request,
        "urlopen",
        lambda *_, **__: (_ for _ in ()).throw(client_error),
    )
    with pytest.raises(paipan.RPCError, match="HTTP 403"):
        paipan.engine_data("qimen.chart", {"solar_time": "x"})


def test_rpc_endpoint_reads_environment_on_each_call(monkeypatch) -> None:
    seen = []

    class Response:
        def read(self):
            return b'{"jsonrpc":"2.0","id":1,"result":{"data":{"ok":true}}}'

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

    def urlopen(request, *_, **__):
        seen.append(request.full_url)
        return Response()

    monkeypatch.setenv("LIKI_RPC_URL", "https://dynamic.example/jsonrpc")
    monkeypatch.setattr(divination_rpc.urllib.request, "urlopen", urlopen)
    assert paipan.engine_data("qimen.chart", {"solar_time": "x"}) == {"ok": True}
    assert seen == ["https://dynamic.example/jsonrpc"]


def test_specialized_lost_property_chart_keeps_base_day_hour_facts(monkeypatch) -> None:
    calls = []

    def fake_call(method: str, params: dict) -> dict:
        calls.append((method, params))
        return {"data": {"ri_gan_gong": "震", "shi_gan_gong": "兑"}}

    monkeypatch.setattr(paipan, "call", fake_call)
    result = paipan.qimen_chart("2026-09-07T12:00:00+08:00")
    assert calls == [("qimen.chart", {"solar_time": "2026-09-07T12:00:00+08:00"})]
    assert result["matter"] is None
    assert result["chart"]["ri_gan_gong"] == "震"


def test_qimen_snapshot_is_readonly_contract_projection() -> None:
    pan = _rich_pan()
    before = json.loads(json.dumps(pan, ensure_ascii=False))
    snapshot = project_qimen_snapshot(pan)
    assert pan == before
    assert snapshot == {
        "scope": "hour",
        "school": "zhuanpan",
        "wu_bu_yu_shi": False,
        "pillars": {
            "ri_gan": "甲",
            "ri_zhi": "辰",
            "shi_gan": "甲",
            "shi_zhi": "午",
            "nian_gan": "丙",
            "nian_zhi": "午",
            "yue_gan": "丙",
            "yue_zhi": "申",
        },
        "ri_gan_gong": "震",
        "shi_gan_gong": "兑",
        "shi_gan_domain": "outer",
        "ri_shi_relation": {
            "subject": "ri_gan_gong",
            "object": "shi_gan_gong",
            "relation": "shi_generates_ri",
            "name": "时干宫生日干宫",
        },
        "nian_gan_gong": "乾",
        "palace_wang_shuai": [
            {"gong": "离", "wuxing": "火", "wang_shuai": "旺", "wang_shuai_name": "旺"},
            {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
        ],
        "shi_gan_gong_wang_shuai": {
            "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
        },
        "patterns": [
            {"name": "反吟", "palaces": ["坎"], "auspicious": False},
            {"name": "青龙返首", "palaces": ["震"], "auspicious": True},
        ],
        "gan_interactions": [{
            "name": "青龙返首", "gong": "震", "range": "low",
            "direction": "东",
            "heaven_gan": "戊", "earth_gan": "丙", "auspicious": True,
        }],
        "men_interactions": [{
            "name": "门迫", "door": "休门", "gong": "坎",
        }],
        "xing_interactions": [{
            "name": "星克宫", "star": "天蓬", "gong": "离", "auspicious": False,
        }],
        "xing_gong_wuxing": [{
            "star": "天蓬", "gong": "离",
            "relation": "xing_controls_gong",
            "relation_name": "星克宫",
            "traditional_label": "囚",
        }],
        "men_po": ["坎"],
        "men_zhi": ["离"],
        "stars": [{"star": "天蓬", "gong": "兑", "heaven_gan": "庚"}],
        "doors": [{"door": "杜门", "gong": "兑"}],
        "spirits": [{"name": "勾陈", "archetype": "勾陈", "gong": "兑"}],
        "spirit_star_relations": [{
            "spirit": "勾陈", "spirit_name": "勾陈", "spirit_gong": "兑",
            "star": "天蓬", "star_gong": "兑", "relation": "same",
            "same_gong": True,
        }],
        "spirit_spirit_relations": [],
        "spirit_door_relations": [{
            "spirit": "勾陈", "spirit_gong": "兑", "door": "杜门",
            "door_gong": "兑", "relation": "same", "same_gong": True,
        }],
        "hour_polarity": "yang",
        "kong_wang": [{
            "symbol": "时干", "role": "pillar", "branch": "戌", "gong": "乾",
        }],
        "ma_xing": [{
            "symbol": "日干", "role": "pillar", "branch": "寅", "gong": "艮",
        }],
        "ying_qi": [
            {
                "type": "ma_xing", "branch": "寅", "gong": "艮",
                "related_to": [{"symbol": "日干", "role": "pillar"}],
            },
            {
                "type": "kong_wang", "branch": "戌", "gong": "乾",
                "related_to": [{"symbol": "时干", "role": "pillar"}],
            },
        ],
        "yong_shen": [
            {
                "symbol": "生门", "palace": "艮",
                "tian_gan": ["戊"], "kong_wang_branches": [],
                "ma_xing_branch": None,
            },
            {
                "symbol": "戊", "palace": "震",
                "tian_gan": ["庚"], "kong_wang_branches": [],
                "ma_xing_branch": None,
            },
        ],
        "yong_shen_palaces": [
            {
                "symbol": "生门", "palace": "艮", "domain": "outer",
                "direction": "东北",
                "spirit": None, "door": None, "wang_shuai": None,
                "star": None, "star_traditional_label": None,
            },
            {
                "symbol": "戊", "palace": "震", "domain": "inner",
                "direction": "东",
                "spirit": None, "door": None, "wang_shuai": None,
                "star": None, "star_traditional_label": None,
            },
        ],
        "zhi_fu_xing_gong": "乾",
        "zhi_shi_men_gong": "开",
        "gong_domains": [{"gong": "兑", "domain": "outer"}],
        "palace_directions": [{"gong": "兑", "direction": "西"}],
        "lost_context": [{
            "gong": "兑", "direction": "西", "domain": "outer",
            "door": "杜门", "wang_shuai": "死",
        }],
        "geng_ge": [],
        "geng_ge_levels": [],
        "geng_ge_status": "absent",
        "thief_context": [],
        "tianwang_context": [],
    }


def test_snapshot_contract_covers_all_projection_fields() -> None:
    contract = load_snapshot_contract()
    snapshot = project_qimen_snapshot(_rich_pan())
    assert set(snapshot) == set(contract["fields"])
    assert "chart_required" not in contract


def test_snapshot_reserves_stable_unconsumed_qimen_factors() -> None:
    reserved = {
        "gan_interactions", "men_interactions", "xing_interactions",
        "xing_gong_wuxing", "men_po", "men_zhi",
        "zhi_fu_xing_gong", "zhi_shi_men_gong",
    }
    assert reserved <= set(load_snapshot_contract()["fields"])


def test_snapshot_projects_stable_star_door_spirit_factors() -> None:
    snapshot = project_qimen_snapshot(_rich_pan())
    assert snapshot["stars"] == [
        {"star": "天蓬", "gong": "兑", "heaven_gan": "庚"},
    ]
    assert snapshot["doors"] == [{"door": "杜门", "gong": "兑"}]
    assert snapshot["spirits"] == [
        {"name": "勾陈", "archetype": "勾陈", "gong": "兑"},
    ]


def test_spirit_archetypes_preserve_hidden_domain_pairs() -> None:
    pan = _pan()
    pan["chart"]["pan"]["gong_wei"] = [
        {"gong": {"name": "坎", "luoshu": 1}, "shen": "白虎", "shen_present": True},
        {"gong": {"name": "兑", "luoshu": 7}, "shen": "朱雀", "shen_present": True},
    ]
    pan["chart"]["palace_wang_shuai"] = [
        {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
        {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
    ]
    snapshot = project_qimen_snapshot(pan)
    by_name = {item["name"]: item for item in snapshot["spirits"]}
    assert by_name["白虎"]["archetype"] == "勾陈"
    assert by_name["朱雀"]["archetype"] == "玄武"


def test_palace_domains_follow_yang_and_yin_dun() -> None:
    yang = project_qimen_snapshot(_pan())
    assert {
        item["gong"]: item["domain"]
        for item in yang["gong_domains"]
    } == {
        "坎": "inner", "坤": "inner", "震": "inner", "巽": "inner",
        "离": "outer", "艮": "outer", "兑": "outer", "乾": "outer",
    }


def test_center_palace_has_no_domain_direction_candidate() -> None:
    pan = _missing_pan()
    pan["chart"]["yong_shen"]["symbols"][0]["palace"] = "中"
    pan["chart"]["gong_domains"] = []
    pan["chart"]["gong_wei"] = [{
        "gong": {"name": "中", "luoshu": 5},
        "men_present": False,
        "shen_present": False,
    }]
    snapshot = project_qimen_snapshot(pan)
    assert snapshot["yong_shen_palaces"][0]["domain"] == "center"
    assert snapshot["yong_shen_palaces"][0]["direction"] == "center"
    ids = {
        item["id"]
        for item in query("missing_person", snapshot)["assertions"]
    }
    assert "qimen_missing_person_direction" not in ids
    assert "qimen_missing_person_inner_inner" not in ids


def test_python_wuxing_relations_match_engine_semantics() -> None:
    root = Path(__file__).resolve().parents[1]
    engine_relations = {
        row["relation"]
        for row in json.loads(
            (root / "engine/internal/engine/qimen/data/plate.json").read_text(
                encoding="utf-8"
            )
        )["wuxing_relations"]
    }
    python_rows = list(csv.DictReader(
        (TOOLS / "data/qimen_wuxing_relations.csv").open(encoding="utf-8-sig")
    ))
    assert {row["relation"] for row in python_rows} == {
        "same", "generates", "controls", "generated_by", "controlled_by",
    }
    assert "same" in engine_relations
    assert {
        "gong_generates_xing": "generated_by",
        "xing_generates_gong": "generates",
        "xing_controls_gong": "controls",
        "gong_controls_xing": "controlled_by",
    }.keys() <= engine_relations
    assert len({(row["source"], row["target"]) for row in python_rows}) == 25

    yin = _pan()
    yin["chart"]["pan"]["yin_dun"] = True
    yin_snapshot = project_qimen_snapshot(yin)
    assert {
        item["gong"]: item["domain"]
        for item in yin_snapshot["gong_domains"]
    } == {
        "坎": "outer", "坤": "outer", "震": "outer", "巽": "outer",
        "离": "inner", "艮": "inner", "兑": "inner", "乾": "inner",
    }


def test_snapshot_projects_stable_spirit_spirit_relations() -> None:
    snapshot = project_qimen_snapshot(_thief_pan())
    assert {
        (item["spirit"], item["other"], item["relation"], item["same_gong"])
        for item in snapshot["spirit_spirit_relations"]
    } == {
        ("勾陈", "玄武", "controls", False),
        ("玄武", "勾陈", "controlled_by", False),
    }


def test_snapshot_projects_real_engine_golden_chart() -> None:
    golden = json.loads(
        (
            Path(__file__).resolve().parents[1]
            / "engine/internal/engine/qimen/testdata/chart_golden.json"
        ).read_text(encoding="utf-8")
    )
    snapshot = project_qimen_snapshot({"matter": None, "chart": golden})
    validate_qimen_snapshot(snapshot)
    assert snapshot["scope"] == "hour"
    assert snapshot["school"] == "zhuanpan"
    assert snapshot["ri_gan_gong"]
    assert snapshot["shi_gan_gong"]
    assert snapshot["pillars"]["shi_gan"] == "戊"
    assert snapshot["hour_polarity"] == "yang"
    assert {
        item["gong"]: item["direction"]
        for item in snapshot["palace_directions"]
    } == {
        "坎": "北", "艮": "东北", "震": "东", "巽": "东南",
        "离": "南", "坤": "西南", "兑": "西", "乾": "西北",
    }
    assert {
        item["gong"]: item["domain"]
        for item in snapshot["gong_domains"]
    } == {
        "坎": "outer", "坤": "outer", "震": "outer", "巽": "outer",
        "乾": "inner", "兑": "inner", "艮": "inner", "离": "inner",
    }
    assert {
        (item["gong"], item["heaven_gan"]): item["range"]
        for item in snapshot["gan_interactions"]
    }[("震", "癸")] == "low"
    assert isinstance(snapshot["patterns"], list)
    thief_roles = {item["thief_symbol"]: item for item in snapshot["thief_context"]}
    assert {"天蓬", "玄武"} <= set(thief_roles)
    assert thief_roles["天蓬"]["gong"]
    assert thief_roles["玄武"]["gong"]
    assert isinstance(snapshot["gan_interactions"], list)
    assert isinstance(snapshot["men_interactions"], list)
    assert isinstance(snapshot["xing_interactions"], list)
    assert isinstance(snapshot["kong_wang"], list)
    assert isinstance(snapshot["ma_xing"], list)
    assert snapshot["stars"]
    assert snapshot["doors"]
    assert snapshot["spirits"]
    assert snapshot["spirit_star_relations"]
    assert isinstance(snapshot["spirit_spirit_relations"], list)
    assert snapshot["gong_domains"]
    assert snapshot["yong_shen"] == []


def test_missing_person_direction_evidence_keeps_table_direction() -> None:
    golden = json.loads(
        (
            Path(__file__).resolve().parents[1]
            / "engine/internal/engine/qimen/testdata/chart_golden.json"
        ).read_text(encoding="utf-8")
    )
    golden["yong_shen"] = {"symbols": [{
        "symbol": "六合", "palace": "震", "tian_gan": ["癸"],
        "kong_wang_branches": [], "ma_xing_branch": None,
    }]}
    snapshot = project_qimen_snapshot({"matter": None, "chart": golden})
    result = query("missing_person", snapshot)
    direction = next(
        item for item in result["assertions"]
        if item["id"] == "qimen_missing_person_direction"
    )
    evidence = next(
        item for item in direction["evidence"]
        if item["field"] == "yong_shen_palaces" and item["item_key"] == "symbol"
    )
    assert evidence["matched"]["palace"] == "震"
    assert evidence["matched"]["direction"] == "东"


def test_missing_person_candidates_are_conservative_and_unranked() -> None:
    snapshot = project_qimen_snapshot(_missing_pan())
    result = query("missing_person", snapshot)
    ids = [item["id"] for item in result["assertions"]]
    assert "qimen_missing_person_direction" in ids
    assert "qimen_missing_person_inner_inner" in ids
    assert all(item["basis"] for item in result["assertions"])


def test_missing_person_wang_star_four_door_candidate_is_table_driven() -> None:
    pan = _missing_pan()
    pan["chart"]["pan"]["gong_wei"] = [{
        "gong": {"name": "坎", "luoshu": 1},
        "tian_pan": [{"gan": "戊", "xing": "天蓬"}],
        "men": "伤门", "men_present": True,
        "shen": "六合", "shen_present": True,
    }]
    pan["chart"]["palace_wang_shuai"] = [{
        "gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚",
    }]
    pan["chart"]["xing_gong_wu_xing"] = [{
        "xing": "天蓬", "gong": "坎", "relation": "same",
        "relation_name": "比和", "traditional_label": "旺",
    }]
    snapshot = project_qimen_snapshot(pan)
    ids = {item["id"] for item in query("missing_person", snapshot)["assertions"]}
    assert "qimen_missing_person_wang_star_four_doors" in ids

    pan["chart"]["xing_gong_wu_xing"][0]["traditional_label"] = "囚"
    snapshot = project_qimen_snapshot(pan)
    ids = {item["id"] for item in query("missing_person", snapshot)["assertions"]}
    assert "qimen_missing_person_wang_star_four_doors" not in ids

    pan["chart"]["xing_gong_wu_xing"][0]["traditional_label"] = "旺"
    pan["chart"]["pan"]["gong_wei"][0]["men"] = "开门"
    snapshot = project_qimen_snapshot(pan)
    ids = {item["id"] for item in query("missing_person", snapshot)["assertions"]}
    assert "qimen_missing_person_wang_star_four_doors" not in ids


def test_missing_person_outer_hour_inner_liuhe_is_findable_candidate() -> None:
    snapshot = project_qimen_snapshot(_missing_pan(
        gong_domains=[
            {"gong": "坎", "domain": "inner"},
            {"gong": "兑", "domain": "outer"},
        ],
        shi_gan_gong="兑",
    ))
    ids = {item["id"] for item in query("missing_person", snapshot)["assertions"]}
    assert "qimen_missing_person_outer_hour_inner_liuhe" in ids
    assert "qimen_missing_person_inner_inner" not in ids


def test_missing_person_spirit_signal_must_share_liuhe_palace() -> None:
    pan = _missing_pan()
    pan["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坎", "luoshu": 1},
            "men": "休门", "men_present": True,
            "shen": "九天", "shen_present": True,
        },
        {
            "gong": {"name": "离", "luoshu": 9},
            "men": "景门", "men_present": True,
            "shen": "九天", "shen_present": True,
        },
        {
            "gong": {"name": "兑", "luoshu": 7},
            "men": "惊门", "men_present": True,
            "shen": "九天", "shen_present": True,
        },
    ]
    pan["chart"]["palace_wang_shuai"] = [
        {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
        {"gong": "离", "wuxing": "火", "wang_shuai": "旺", "wang_shuai_name": "旺"},
        {"gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
    ]
    snapshot = project_qimen_snapshot(pan)
    ids = {item["id"] for item in query("missing_person", snapshot)["assertions"]}
    assert "qimen_missing_person_jiutian" in ids

    cross_palace = _missing_pan()
    cross_palace["chart"]["pan"]["gong_wei"] = [
        {
            "gong": {"name": "坎", "luoshu": 1},
            "men": "休门", "men_present": True,
            "shen": "太阴", "shen_present": True,
        },
        {
            "gong": {"name": "离", "luoshu": 9},
            "men": "景门", "men_present": True,
            "shen": "九天", "shen_present": True,
        },
    ]
    cross_palace["chart"]["palace_wang_shuai"] = [
        {"gong": "坎", "wuxing": "水", "wang_shuai": "囚", "wang_shuai_name": "囚"},
        {"gong": "离", "wuxing": "火", "wang_shuai": "旺", "wang_shuai_name": "旺"},
    ]
    cross_ids = {
        item["id"]
        for item in query("missing_person", project_qimen_snapshot(cross_palace))["assertions"]
    }
    assert "qimen_missing_person_jiutian" not in cross_ids


def test_missing_person_rejects_incompatible_school() -> None:
    snapshot = project_qimen_snapshot(_missing_pan(method={
        "scope": "hour", "school": "jinhan_yujing",
        "spirit_mode": "jinhan_day_spirits",
    }))
    with pytest.raises(ValueError, match="missing_person"):
        query("missing_person", snapshot)


def test_snapshot_projects_escape_capture_factors() -> None:
    snapshot = project_qimen_snapshot(_escape_pan())
    assert snapshot["hour_polarity"] == "yang"
    assert snapshot["spirit_door_relations"] == [
        {
            "spirit": "六合", "spirit_gong": "坎", "door": "休门",
            "door_gong": "坎", "relation": "same", "same_gong": True,
        },
        {
            "spirit": "六合", "spirit_gong": "坎", "door": "伤门",
            "door_gong": "震", "relation": "generates", "same_gong": False,
        },
        {
            "spirit": "值符", "spirit_gong": "震", "door": "休门",
            "door_gong": "坎", "relation": "generated_by", "same_gong": False,
        },
        {
            "spirit": "值符", "spirit_gong": "震", "door": "伤门",
            "door_gong": "震", "relation": "same", "same_gong": True,
        },
    ]
    assert snapshot["gan_interactions"] == [{
        "name": "伏吟", "gong": "坎", "range": "low",
        "direction": "北",
        "heaven_gan": "戊", "earth_gan": "戊", "auspicious": False,
    }]


def test_escape_capture_candidates_are_table_driven_and_unranked() -> None:
    snapshot = project_qimen_snapshot(_escape_pan(
        patterns=[{
            "name": "太白入荧", "gong_wei": ["震"], "auspicious": False,
        }],
        gan_interaction=[{
            "name": "天网", "gong": "坎", "tian_pan_gan": "癸",
            "di_pan_gan": "壬", "auspicious": False,
        }],
    ))
    result = query("capture_escape", snapshot)
    ids = [item["id"] for item in result["assertions"]]
    assert ids == [
        "qimen_capture_escape_escape_self_return",
        "qimen_capture_escape_taibai",
        "qimen_capture_escape_tianwang_direction",
        "qimen_capture_escape_tianwang_yang_low",
    ]
    assert all(item["basis"] for item in result["assertions"])


def test_escape_capture_door_spirit_relations_are_table_driven() -> None:
    gong_wuxing = {"坎": "水", "兑": "金", "震": "木", "离": "火"}
    cases = [
        ("坎", "震", "qimen_capture_escape_escape_self_return"),
        ("兑", "震", "qimen_capture_escape_escape_not_caught"),
        ("震", "兑", "qimen_capture_escape_easy_capture"),
        ("离", "震", "qimen_capture_escape_bribe_release"),
        ("震", "震", "qimen_capture_escape_collusion"),
    ]
    for spirit_gong, door_gong, assertion_id in cases:
        pan = _escape_pan()
        pan["chart"]["pan"]["gong_wei"] = [
            {
                "gong": {"name": spirit_gong, "luoshu": 1},
                "men": "休门", "men_present": True,
                "shen": "六合", "shen_present": True,
            },
            {
                "gong": {"name": door_gong, "luoshu": 2},
                "men": "伤门", "men_present": True,
                "shen": "值符", "shen_present": True,
            },
        ]
        pan["chart"]["palace_wang_shuai"] = [
            {"gong": spirit_gong, "wuxing": gong_wuxing[spirit_gong], "wang_shuai": "休", "wang_shuai_name": "休"},
            {"gong": door_gong, "wuxing": gong_wuxing[door_gong], "wang_shuai": "囚", "wang_shuai_name": "囚"},
        ]
        ids = {
            item["id"]
            for item in query("capture_escape", project_qimen_snapshot(pan))["assertions"]
        }
        assert assertion_id in ids


def test_escape_tianwang_requires_low_palace_and_yin_hour() -> None:
    tianwang = [{
        "name": "天网", "gong": "坎", "tian_pan_gan": "癸",
        "di_pan_gan": "壬", "auspicious": False,
    }]
    capture = project_qimen_snapshot(_escape_pan(gan_interaction=tianwang))
    capture_ids = {
        item["id"] for item in query("capture_escape", capture)["assertions"]
    }
    assert "qimen_capture_escape_tianwang_yin_low" not in capture_ids
    assert "qimen_capture_escape_tianwang_yang_low" in capture_ids
    assert "qimen_capture_escape_tianwang_direction" in capture_ids

    yin_hour = _escape_pan(gan_interaction=tianwang)
    yin_hour["chart"]["pan"]["shi_gan"] = "癸"
    yin_snapshot = project_qimen_snapshot(yin_hour)
    assert yin_snapshot["hour_polarity"] == "yin"
    yin_ids = {
        item["id"] for item in query("capture_escape", yin_snapshot)["assertions"]
    }
    assert "qimen_capture_escape_tianwang_yin_low" in yin_ids
    assert "qimen_capture_escape_tianwang_yang_low" not in yin_ids

    high = _escape_pan(gan_interaction=[{
        "name": "天网", "gong": "乾", "tian_pan_gan": "癸",
        "di_pan_gan": "壬", "auspicious": False,
    }])
    high["chart"]["pan"]["shi_gan"] = "癸"
    high["chart"]["palace_wang_shuai"].extend([
        {"gong": "乾", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死"},
    ])
    high_snapshot = project_qimen_snapshot(high)
    high_ids = {
        item["id"] for item in query("capture_escape", high_snapshot)["assertions"]
    }
    assert "qimen_capture_escape_tianwang_yin_low" not in high_ids
    assert "qimen_capture_escape_tianwang_high" in high_ids
    assert "qimen_capture_escape_tianwang_direction" in high_ids


def test_array_rule_evidence_preserves_matched_domain_row() -> None:
    pan = _escape_pan(gan_interaction=[{
        "name": "天网", "gong": "震", "tian_pan_gan": "癸",
        "di_pan_gan": "壬", "auspicious": False,
    }])
    snapshot = project_qimen_snapshot(pan)
    result = query("capture_escape", snapshot)
    direction = next(
        item for item in result["assertions"]
        if item["id"] == "qimen_capture_escape_tianwang_direction"
    )
    evidence = next(
        item for item in direction["evidence"]
        if item["field"] == "gan_interactions" and item["item_key"] == "heaven_gan"
    )
    assert evidence["matched"] == {
        "name": "天网", "gong": "震", "range": "low", "direction": "东",
        "heaven_gan": "癸", "earth_gan": "壬", "auspicious": False,
    }


def test_capture_escape_rejects_incompatible_school() -> None:
    snapshot = project_qimen_snapshot(_escape_pan(method={
        "scope": "hour", "school": "jinhan_yujing",
        "spirit_mode": "jinhan_day_spirits",
    }))
    with pytest.raises(ValueError, match="capture_escape"):
        query("capture_escape", snapshot)


def test_optional_snapshot_branch_must_be_complete_when_present() -> None:
    pan = _pan()
    pan["chart"]["yong_shen"] = {}
    with pytest.raises(ValueError, match="yong_shen.symbols"):
        project_qimen_snapshot(pan)

    pan["chart"]["yong_shen"] = None
    with pytest.raises(ValueError, match="yong_shen.symbols"):
        project_qimen_snapshot(pan)


def test_lost_property_does_not_infer_wang_shuai_from_relation_only() -> None:
    snapshot = project_qimen_snapshot(_pan(ri_shi_relation={
        "subject": "ri_gan_gong",
        "object": "shi_gan_gong",
        "relation": "shi_generates_ri",
        "name": "时干宫生日干宫",
    }, palace_wang_shuai=[], shi_gan_gong_wang_shuai={
        "gong": "兑", "wuxing": "金", "wang_shuai": "死", "wang_shuai_name": "死",
    }))
    result = query("lost_property", snapshot)
    assert result["rule"] == "lost_property"
    assert not {"qimen_lost_property_shi_sheng_ri_wang", "qimen_lost_property_shi_sheng_ri_xiang"} & {
        item["id"] for item in result["assertions"]
    }


def test_lost_property_candidates_can_coexist_without_ranking() -> None:
    snapshot = project_qimen_snapshot(_lost_pan())
    result = query("lost_property", snapshot)
    ids = [item["id"] for item in result["assertions"]]
    assert ids == [
        "qimen_lost_property_direction",
        "qimen_lost_property_fan_yin",
        "qimen_lost_property_outer",
        "qimen_lost_property_shi_gan_void",
        "qimen_lost_property_shi_sheng_ri_wang",
    ]
    assert all(item["basis"] for item in result["assertions"])


def test_thief_capture_candidates_are_table_driven_and_unranked() -> None:
    snapshot = project_qimen_snapshot(_thief_pan())
    result = query("thief_capture", snapshot)
    ids = [item["id"] for item in result["assertions"]]
    assert ids == [
        "qimen_thief_capture_controls_peng",
        "qimen_thief_capture_controls_xuanwu",
        "qimen_thief_capture_geng_absent",
    ]
    assert all(item["basis"] for item in result["assertions"])


def test_geng_ge_levels_project_all_four_pillars() -> None:
    cases = {
        "year": ("nian", "丙", "午"),
        "month": ("yue", "丙", "申"),
        "day": ("ri", "庚", "辰"),
        "hour": ("shi", "庚", "午"),
    }
    for level, (prefix, gan, zhi) in cases.items():
        pan = _thief_pan()
        pan["chart"]["pan"].update({
            f"{prefix}_gan": gan,
            f"{prefix}_zhi": zhi,
        })
        pan["chart"]["gan_interaction"] = [{
            "name": "庚格", "gong": "坎", "tian_pan_gan": "庚",
            "di_pan_gan": gan, "auspicious": False,
        }]
        snapshot = project_qimen_snapshot(pan)
        assert snapshot["geng_ge_status"] == "present"
        assert any(
            row["level"] == level and row["gong"] == "坎" and row["earth_gan"] == gan
            for row in snapshot["geng_ge"]
        )
        assert level in snapshot["geng_ge_levels"]
        result = query("thief_capture", snapshot)
        expected = {
            "year": "qimen_thief_capture_geng_year",
            "month": "qimen_thief_capture_geng_month",
            "day": "qimen_thief_capture_geng_day",
            "hour": "qimen_thief_capture_geng_hour",
        }[level]
        assert expected in {item["id"] for item in result["assertions"]}
        evidence = next(
            item for item in next(
                row for row in result["assertions"] if row["id"] == expected
            )["evidence"]
            if item["field"] == "geng_ge" and item["item_key"] == "level"
        )
        assert evidence["matched"]["gong"] == "坎"
        assert evidence["matched"]["direction"] == "北"


def test_thief_profile_noble_and_common_are_table_driven() -> None:
    base = _thief_pan()
    base["chart"]["xing_gong_wu_xing"] = [{
        "xing": "天蓬", "gong": "坎", "relation": "same",
        "relation_name": "比和", "traditional_label": "旺",
    }]
    base["chart"]["patterns"] = [{
        "name": "青龙返首", "gong_wei": ["坎"], "auspicious": True,
    }]
    noble = query("thief_profile", project_qimen_snapshot(base))
    assert "qimen_thief_profile_noble" in {
        item["id"] for item in noble["assertions"]
    }

    base["chart"]["patterns"] = []
    base["chart"]["xing_gong_wu_xing"][0]["traditional_label"] = "囚"
    common = query("thief_profile", project_qimen_snapshot(base))
    assert "qimen_thief_profile_common" in {
        item["id"] for item in common["assertions"]
    }


def test_thief_capture_rejects_incompatible_school() -> None:
    snapshot = project_qimen_snapshot(
        _thief_pan(method={
            "scope": "hour", "school": "jinhan_yujing",
            "spirit_mode": "jinhan_day_spirits",
        })
    )
    with pytest.raises(ValueError, match="thief_capture"):
        query("thief_capture", snapshot)


def test_interpretation_output_does_not_use_table_row_order_as_priority(monkeypatch) -> None:
    import qimen_duanyu

    def reversed_index() -> dict:
        return {
            "lost_property": [
                {
                    "id": "qimen_lost_property_shi_gan_void",
                    "name": "时干落空亡",
                    "conclusion": "难复得候选",
                    "basis": "test",
                    "conditions": [[{
                        "field": "kong_wang",
                        "item_key": "symbol",
                        "operator": "array_some_equals",
                        "expected": "时干",
                    }]],
                },
                {
                    "id": "qimen_lost_property_fan_yin",
                    "name": "反吟",
                    "conclusion": "复得候选",
                    "basis": "test",
                    "conditions": [[{
                        "field": "patterns",
                        "item_key": "name",
                        "operator": "array_some_equals",
                        "expected": "反吟",
                    }]],
                },
            ]
        }

    monkeypatch.setattr(qimen_duanyu, "load_interpretation_index", reversed_index)
    result = qimen_duanyu.query("lost_property", project_qimen_snapshot(_lost_pan()))
    assert [item["id"] for item in result["assertions"]] == [
        "qimen_lost_property_fan_yin",
        "qimen_lost_property_shi_gan_void",
    ]


def test_query_rejects_partial_chart() -> None:
    pan = _pan()
    del pan["chart"]["kong_wang_affected"]
    with pytest.raises(ValueError, match="kong_wang_affected"):
        project_qimen_snapshot(pan)


def test_interpretation_table_has_complete_rules() -> None:
    index = load_interpretation_index()
    assert index, "奇门解释表不能为空"
    for rows in index.values():
        for row in rows:
            assert row["name"] and row["conclusion"] and row["basis"]
            assert row["conditions"]
    paths = {
        condition["field"]
        for rows in index.values()
        for row in rows
        for conditions in row["conditions"]
        for condition in conditions
    }
    assert paths <= set(load_snapshot_contract()["fields"])


def test_interpretation_table_rejects_unknown_snapshot_path(tmp_path, monkeypatch) -> None:
    import qimen_interpretations

    assertions = tmp_path / "assertions.csv"
    assertions.write_text(
        "assertion_id,rule,name,conclusion,basis\n"
        "qimen_lost_property_fan_yin,lost_property,反吟,候选,test\n",
        encoding="utf-8",
    )
    conditions = tmp_path / "conditions.csv"
    conditions.write_text(
        "assertion_id,condition_group_id,field,item_key,operator,expected\n"
        "qimen_lost_property_fan_yin,1,unknown_path,,array_some_equals,反吟\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_interpretations, "_INTERPRETATION_INDEX", None)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATIONS_PATH", assertions)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATION_CONDITIONS_PATH", conditions)
    with pytest.raises(TableError, match="unknown_path"):
        qimen_interpretations.load_interpretation_index()


def test_interpretation_condition_group_ids_are_valid(tmp_path, monkeypatch) -> None:
    import qimen_interpretations

    assertions = tmp_path / "assertions.csv"
    assertions.write_text(
        "assertion_id,rule,name,conclusion,basis\n"
        "qimen_lost_property_fan_yin,lost_property,反吟,候选,test\n",
        encoding="utf-8",
    )
    conditions = tmp_path / "conditions.csv"
    conditions.write_text(
        "assertion_id,condition_group_id,field,item_key,operator,expected\n"
        "qimen_lost_property_fan_yin,a,patterns,name,array_some_equals,反吟\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_interpretations, "_INTERPRETATION_INDEX", None)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATIONS_PATH", assertions)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATION_CONDITIONS_PATH", conditions)
    with pytest.raises(TableError, match="condition_group_id"):
        qimen_interpretations.load_interpretation_index()


def test_interpretation_condition_groups_sort_numerically(tmp_path, monkeypatch) -> None:
    import qimen_interpretations

    rules = tmp_path / "rules.csv"
    rules.write_text(
        "rule,name,scopes,schools,basis\n"
        "lost_property,失物,hour,zhuanpan,test\n",
        encoding="utf-8",
    )
    assertions = tmp_path / "assertions.csv"
    assertions.write_text(
        "assertion_id,rule,name,conclusion,basis\n"
        "qimen_lost_property_fan_yin,lost_property,反吟,候选,test\n"
        "qimen_lost_property_shi_gan_void,lost_property,空亡,候选,test\n"
        "qimen_lost_property_shi_sheng_ri_wang,lost_property,旺生,候选,test\n"
        "qimen_lost_property_shi_sheng_ri_xiang,lost_property,相生,候选,test\n",
        encoding="utf-8",
    )
    conditions = tmp_path / "conditions.csv"
    conditions.write_text(
        "assertion_id,condition_group_id,field,item_key,operator,expected\n"
        "qimen_lost_property_fan_yin,2,kong_wang,symbol,array_some_equals,时干\n"
        "qimen_lost_property_fan_yin,10,patterns,name,array_some_equals,反吟\n"
        "qimen_lost_property_fan_yin,1,ri_shi_relation,relation,equals,same\n"
        "qimen_lost_property_fan_yin,1,shi_gan_gong_wang_shuai,wang_shuai,equals,旺\n"
        "qimen_lost_property_shi_sheng_ri_wang,1,ri_shi_relation,relation,equals,shi_generates_ri\n"
        "qimen_lost_property_shi_sheng_ri_wang,1,shi_gan_gong_wang_shuai,wang_shuai,equals,旺\n"
        "qimen_lost_property_shi_sheng_ri_xiang,1,ri_shi_relation,relation,equals,shi_generates_ri\n"
        "qimen_lost_property_shi_sheng_ri_xiang,1,shi_gan_gong_wang_shuai,wang_shuai,equals,相\n"
        "qimen_lost_property_shi_gan_void,1,kong_wang,symbol,array_some_equals,时干",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_interpretations, "_INTERPRETATION_INDEX", None)
    monkeypatch.setattr(qimen_interpretations, "_RULE_TABLE", None)
    monkeypatch.setattr(qimen_interpretations, "RULES_PATH", rules)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATIONS_PATH", assertions)
    monkeypatch.setattr(qimen_interpretations, "INTERPRETATION_CONDITIONS_PATH", conditions)
    index = qimen_interpretations.load_interpretation_index()
    groups = index["lost_property"][0]["conditions"]
    assert [group[0]["field"] for group in groups] == [
        "ri_shi_relation", "kong_wang", "patterns",
    ]


def test_interpretation_rule_applicability_is_table_driven() -> None:
    rules = load_rule_table()
    assert rules["lost_property"]["name"] == "失物"
    assert rules["lost_property"]["scopes"] == ["hour"]
    assert rules["lost_property"]["schools"] == [
        "zhuanpan", "luoshu_feipan", "mingfa_feipan",
    ]
    assert rules["lost_property"]["basis"]

    snapshot = project_qimen_snapshot(
        _lost_pan(method={
            "scope": "hour", "school": "jinhan_yujing",
            "spirit_mode": "jinhan_day_spirits",
        })
    )
    with pytest.raises(ValueError, match="lost_property"):
        query("lost_property", snapshot)


def test_rule_axes_come_from_engine_enum_without_duplicates() -> None:
    root = Path(__file__).resolve().parents[1]
    engine = json.loads(
        (root / "engine/internal/agent/qimen_schema.json").read_text(encoding="utf-8")
    )["params"]["properties"]
    scope_enum = set(engine["scope"]["enum"])
    school_enum = set(engine["school"]["enum"])
    for metadata in load_rule_table().values():
        assert len(metadata["scopes"]) == len(set(metadata["scopes"]))
        assert len(metadata["schools"]) == len(set(metadata["schools"]))
        assert set(metadata["scopes"]) <= scope_enum
        assert set(metadata["schools"]) <= school_enum


def test_rule_table_rejects_duplicate_axes(tmp_path, monkeypatch) -> None:
    import qimen_interpretations

    rules = tmp_path / "rules.csv"
    rules.write_text(
        "rule,name,scopes,schools,basis\n"
        "lost_property,失物,hour|hour,zhuanpan,test\n",
        encoding="utf-8",
    )
    monkeypatch.setattr(qimen_interpretations, "_RULE_TABLE", None)
    monkeypatch.setattr(qimen_interpretations, "RULES_PATH", rules)
    with pytest.raises(TableError, match="scopes"):
        qimen_interpretations.load_rule_table()


def test_python_tool_schema_stays_aligned_with_fact_sources() -> None:
    root = Path(__file__).resolve().parents[1]
    tool = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    functions = {item["function"]["name"]: item["function"] for item in tool["tools"]}
    assert set(functions["qimen_snapshot"]["parameters"]["properties"]["matter"]["enum"]) == set(
        load_routing_matter_table()
    )
    assert functions["qimen_snapshot"]["parameters"]["properties"]["rule"]["enum"] == list(
        load_rule_table()
    )

    engine = json.loads(
        (root / "engine/internal/agent/qimen_schema.json").read_text(encoding="utf-8")
    )["params"]["properties"]
    python = functions["qimen_snapshot"]["parameters"]["properties"]
    for field in (
        "scope", "school", "dingju_method", "quarter_rule",
        "base_dingju_method", "dun_source", "hour_boundary",
    ):
        assert python[field]["enum"] == engine[field]["enum"]


def test_snapshot_contract_version_and_engine_paths_are_bound() -> None:
    root = Path(__file__).resolve().parents[1]
    contract = load_snapshot_contract()
    version = (root / "skills/liki-divination/VERSION").read_text(encoding="utf-8").strip()
    assert contract["version"] == version

    schema = json.loads(
        (root / "engine/internal/agent/qimen_schema.json").read_text(encoding="utf-8")
    )["result"]


    def schema_node_for_path(value: dict, segments: list[str]) -> dict | None:
        if not segments:
            return value
        properties = value.get("properties", {})
        if segments[0] in properties:
            return schema_node_for_path(properties[segments[0]], segments[1:])
        for key in ("oneOf", "anyOf"):
            for branch in value.get(key, []):
                if found := schema_node_for_path(branch, segments):
                    return found
        for branch in value.get("allOf", []):
            if found := schema_node_for_path(branch, segments):
                return found
        return None

    for spec in contract["fields"].values():
        node = schema_node_for_path(schema, spec["path"].split("."))
        assert node is not None, spec["path"]
        if spec.get("projection"):
            continue
        if spec["kind"] == "boolean":
            expected_type = "boolean"
        elif spec["kind"] == "scalar":
            expected_type = spec["type"]
        elif spec["kind"] == "object":
            expected_type = "object"
        else:
            expected_type = "array"
        assert node.get("type") == expected_type
        if spec["kind"] == "object_array":
            item_properties = node.get("items", {}).get("properties", {})
            for field_name, field_spec in spec["fields"].items():
                segments = field_spec["path"].split(".")
                assert segments[0] in item_properties, (
                    f"array source lacks item property {segments[0]}"
                )
                if len(segments) > 1:
                    nested = item_properties[segments[0]].get("properties", {})
                    assert segments[1] in nested, field_spec["path"]
                    expected_type = field_spec["type"]
                    assert nested[segments[1]].get("type") == expected_type
                else:
                    expected_field_type = (
                        "array" if field_spec["type"] == "array" else field_spec["type"]
                    )
                    assert item_properties[segments[0]].get("type") == expected_field_type


def test_tool_schema_matches_cli_dispatch() -> None:
    agent_cli = _load_agent_cli()
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    names = {item["function"]["name"] for item in schema["tools"]}
    assert names == {
        "liuyao_snapshot", "liuyao_ask",
        "qimen_snapshot", "qimen_ask", "huangli_days",
    }
    assert names == set(agent_cli._DISPATCH)


def test_python_layers_are_orthogonal() -> None:
    paipan_source = (TOOLS / "qimen_paipan.py").read_text(encoding="utf-8")
    matters_source = (TOOLS / "qimen_matters.py").read_text(encoding="utf-8")
    interpretations_source = (TOOLS / "qimen_interpretations.py").read_text(encoding="utf-8")
    duanyu_source = (TOOLS / "qimen_duanyu.py").read_text(encoding="utf-8")
    factors_source = (TOOLS / "qimen_factors.py").read_text(encoding="utf-8")
    assert "from qimen_duanyu" not in paipan_source and "import qimen_duanyu" not in paipan_source
    assert "from qimen_interpretations" not in paipan_source
    assert "from qimen_duanyu" not in matters_source and "import qimen_duanyu" not in matters_source
    assert "from qimen_paipan" not in interpretations_source
    assert "from qimen_interpretations" not in factors_source
    assert "from qimen_paipan" not in duanyu_source and "import qimen_paipan" not in duanyu_source
    assert "from qimen_paipan" not in factors_source and "import qimen_paipan" not in factors_source
    assert "patterns[]" not in duanyu_source
    assert "kong_wang_affected" not in duanyu_source


def test_windows_launcher_matches_bazi_compatibility() -> None:
    agent_cli = (TOOLS / "agent_cli.py").read_text(encoding="utf-8")
    launcher = (TOOLS / "agent_cli.cmd").read_text(encoding="ascii")
    assert "if os.name != \"nt\":" in agent_cli
    assert "reconfigure(encoding=\"utf-8\")" in agent_cli
    assert "ensure_ascii=True" in agent_cli
    for required in (
        "setlocal",
        "set \"PYTHONUTF8=1\"",
        "set \"PYTHONIOENCODING=utf-8\"",
        "where py >nul 2>nul",
        "set \"PYTHON_CMD=py -3\"",
        "-X utf8",
        "%~dp0agent_cli.py",
    ):
        assert required in launcher


def test_query_compatibility_without_network() -> None:
    from qimen_duanyu import query
    from qimen_factors import project_qimen_snapshot

    pan = _lost_pan(kong_wang_affected=[])
    result = query("lost_property", project_qimen_snapshot(pan))
    assert result["assertions"][0]["id"] == "qimen_lost_property_direction"


def test_query_rejects_incompatible_rule_before_projection() -> None:
    import pytest

    from qimen_interpretations import assert_rule_for_pan
    from qimen_factors import project_qimen_snapshot

    pan = _lost_pan(
        method={
            "scope": "hour", "school": "jinhan_yujing",
            "spirit_mode": "jinhan_day_spirits",
        }
    )
    with pytest.raises(ValueError, match="lost_property requires scope"):
        assert_rule_for_pan("lost_property", pan)
        project_qimen_snapshot(pan)
